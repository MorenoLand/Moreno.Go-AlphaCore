package world

import (
	"encoding/binary"
	"strings"
	"sync"
	"time"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const (
	guildCreateCommand  uint32 = 0
	guildInviteCommand  uint32 = 1
	guildLeaveCommand   uint32 = 2
	guildFounderCommand uint32 = 12
	guildInvited        uint32 = 0
	guildInternal       uint32 = 1
	guildAlready        uint32 = 2
	guildNameInvalid    uint32 = 6
	guildNameExists     uint32 = 7
	guildLeaderLeave    uint32 = 8
	guildNotInGuild     uint32 = 9
	guildNotInGuildName uint32 = 10
	guildPlayerMissing  uint32 = 11
	guildNotAllied      uint32 = 12
)

type guildState struct {
	Guild   realm.Guild
	Members map[int64]int64
	Invites map[int64]int64
}

type guildRegistry struct {
	mu      sync.RWMutex
	guilds  map[int64]*guildState
	players map[int64]*guildState
	invites map[int64]*guildState
}

func (r *guildRegistry) ensure() {
	if r.guilds == nil {
		r.guilds = make(map[int64]*guildState)
		r.players = make(map[int64]*guildState)
		r.invites = make(map[int64]*guildState)
	}
}

func (r *guildRegistry) loaded(guild realm.Guild, members []realm.GuildMember) *guildState {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ensure()
	if state := r.guilds[guild.ID]; state != nil {
		return state
	}
	state := &guildState{Guild: guild, Members: make(map[int64]int64), Invites: make(map[int64]int64)}
	for _, member := range members {
		state.Members[member.GUID] = member.Rank
		r.players[member.GUID] = state
	}
	r.guilds[guild.ID] = state
	return state
}

func (r *guildRegistry) created(guild realm.Guild, leader int64) *guildState {
	return r.loaded(guild, []realm.GuildMember{{GuildID: guild.ID, GUID: leader, Rank: 0}})
}

func (r *guildRegistry) forPlayer(guid int64) *guildState {
	r.mu.RLock()
	state := r.players[guid]
	r.mu.RUnlock()
	return state
}

func (r *guildRegistry) forID(id int64) *guildState {
	r.mu.RLock()
	state := r.guilds[id]
	r.mu.RUnlock()
	return state
}

func (r *guildRegistry) pending(guid int64) *guildState {
	r.mu.RLock()
	state := r.invites[guid]
	r.mu.RUnlock()
	return state
}

func (r *guildRegistry) addPlayer(state *guildState, guid, rank int64) {
	r.mu.Lock()
	r.ensure()
	state.Members[guid] = rank
	r.players[guid] = state
	delete(r.invites, guid)
	delete(state.Invites, guid)
	r.mu.Unlock()
}

func (r *guildRegistry) invite(state *guildState, target, inviter int64) {
	r.mu.Lock()
	r.ensure()
	state.Invites[target] = inviter
	r.invites[target] = state
	r.mu.Unlock()
}

func (r *guildRegistry) removeInvite(state *guildState, target int64) int64 {
	r.mu.Lock()
	inviter := state.Invites[target]
	delete(state.Invites, target)
	delete(r.invites, target)
	r.mu.Unlock()
	return inviter
}

func (r *guildRegistry) removePlayer(state *guildState, guid int64) {
	r.mu.Lock()
	delete(state.Members, guid)
	delete(r.players, guid)
	r.mu.Unlock()
}

func (r *guildRegistry) remove(state *guildState) {
	r.mu.Lock()
	delete(r.guilds, state.Guild.ID)
	for guid := range state.Members {
		delete(r.players, guid)
	}
	for guid := range state.Invites {
		delete(r.invites, guid)
	}
	r.mu.Unlock()
}

func (s *WorldServer) loadGuild(player realm.Character) error {
	guild, found, err := s.Characters.GuildByPlayer(player.GUID)
	if err != nil || !found {
		return err
	}
	members, err := s.Characters.GuildMembers(guild.ID)
	if err != nil {
		return err
	}
	s.guilds.loaded(guild, members)
	return nil
}

func (s *WorldServer) guildCreate(active realm.Character, data []byte, gmLevel int) ([][]byte, error) {
	if gmLevel <= 0 || len(data) < 2 {
		return nil, nil
	}
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	name = strings.TrimSpace(name)
	if !validGuildName(name) {
		return s.guildCommand(active, guildCreateCommand, "", guildNameInvalid)
	}
	if s.guilds.forPlayer(active.GUID) != nil {
		return s.guildCommand(active, guildCreateCommand, "", guildAlready)
	}
	if existing, found, err := s.Characters.GuildByName(name); err != nil {
		return nil, err
	} else if found || s.guilds.forID(existing.ID) != nil {
		return s.guildCommand(active, guildCreateCommand, name, guildNameExists)
	}
	guild, err := s.Characters.CreateGuild(name, "", active.GUID)
	if err != nil {
		return nil, err
	}
	if err := s.Characters.AddGuildMember(guild.ID, active.GUID, 0); err != nil {
		return nil, err
	}
	state := s.guilds.created(guild, active.GUID)
	event, err := guildEvent(3, active.Name)
	if err != nil {
		return nil, err
	}
	_ = state
	return [][]byte{event}, nil
}

func (s *WorldServer) guildInvite(active realm.Character, data []byte) ([][]byte, error) {
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	state := s.guilds.forPlayer(active.GUID)
	if state == nil {
		return s.guildCommand(active, guildInviteCommand, "", guildNotInGuild)
	}
	if state.Members[active.GUID] > 1 {
		return s.guildCommand(active, guildInviteCommand, "", guildInternal)
	}
	target, found := s.playerByName(strings.TrimSpace(name))
	if !found {
		return s.guildCommand(active, guildInviteCommand, name, guildPlayerMissing)
	}
	if s.guilds.forPlayer(target.GUID) != nil || s.guilds.pending(target.GUID) != nil {
		return s.guildCommand(active, guildInviteCommand, name, guildAlready)
	}
	if first, second, err := s.teams(active, target); err != nil {
		return nil, err
	} else if first != second {
		return s.guildCommand(active, guildInviteCommand, "", guildNotAllied)
	}
	s.guilds.invite(state, target.GUID, active.GUID)
	inviter, err := packet.StringBytes(active.Name)
	if err != nil {
		return nil, err
	}
	guildName, err := packet.StringBytes(state.Guild.Name)
	if err != nil {
		return nil, err
	}
	invite, err := packet.Encode(packet.SMSGGuildInvite, append(inviter, guildName...))
	if err != nil {
		return nil, err
	}
	s.sendPlayer(target.GUID, invite)
	return s.guildCommand(active, guildInviteCommand, name, guildInvited)
}

func (s *WorldServer) guildAccept(active realm.Character) error {
	state := s.guilds.pending(active.GUID)
	if state == nil {
		return nil
	}
	s.guilds.removeInvite(state, active.GUID)
	state.Members[active.GUID] = 4
	s.guilds.addPlayer(state, active.GUID, 4)
	if err := s.Characters.AddGuildMember(state.Guild.ID, active.GUID, 4); err != nil {
		return err
	}
	event, err := guildEvent(3, active.Name)
	if err != nil {
		return err
	}
	s.broadcastGuild(state, event, 0)
	return nil
}

func (s *WorldServer) guildDecline(active realm.Character) {
	state := s.guilds.pending(active.GUID)
	if state == nil {
		return
	}
	inviter := s.guilds.removeInvite(state, active.GUID)
	if inviter == 0 {
		return
	}
	name, err := packet.StringBytes(active.Name)
	if err != nil {
		return
	}
	response, err := packet.Encode(packet.SMSGGuildDecline, name)
	if err == nil {
		s.sendPlayer(inviter, response)
	}
}

func (s *WorldServer) guildQuery(data []byte) ([]byte, error) {
	if len(data) != 4 {
		return nil, nil
	}
	id := int64(binary.LittleEndian.Uint32(data) & 0x00ffffff)
	state := s.guilds.forID(id)
	if state == nil {
		guild, found, err := s.Characters.GuildByID(id)
		if err != nil || !found {
			return nil, err
		}
		members, err := s.Characters.GuildMembers(id)
		if err != nil {
			return nil, err
		}
		state = s.guilds.loaded(guild, members)
	}
	return guildQueryState(state)
}

func guildQueryState(state *guildState) ([]byte, error) {
	name, err := packet.StringBytes(state.Guild.Name)
	if err != nil {
		return nil, err
	}
	body := append(encodeUint32((state.Guild.RealmID<<24)|state.Guild.ID), name...)
	for _, value := range []int64{state.Guild.EmblemStyle, state.Guild.EmblemColor, state.Guild.BorderStyle, state.Guild.BorderColor, state.Guild.BackgroundColor} {
		body = append(body, encodeUint32(value)...)
	}
	return packet.Encode(packet.SMSGGuildQueryResponse, body)
}

func (s *WorldServer) guildInfo(active realm.Character) ([]byte, error) {
	state := s.guilds.forPlayer(active.GUID)
	if state == nil {
		return s.guildCommandPacket(guildCreateCommand, "", guildNotInGuild)
	}
	name, err := packet.StringBytes(state.Guild.Name)
	if err != nil {
		return nil, err
	}
	date, _ := time.Parse("2006-01-02 15:04:05", state.Guild.CreationDate)
	accounts := make(map[int64]bool)
	for guid := range state.Members {
		if player, found, err := s.Characters.CharacterByGUID(guid); err == nil && found {
			accounts[player.AccountID] = true
		}
	}
	body := append(name, encodeUint32(int64(date.Day()))...)
	body = append(body, encodeUint32(int64(date.Month()))...)
	body = append(body, encodeUint32(int64(date.Year()))...)
	body = append(body, encodeUint32(int64(len(state.Members)))...)
	body = append(body, encodeUint32(int64(len(accounts)))...)
	return packet.Encode(packet.SMSGGuildInfo, body)
}

func (s *WorldServer) guildRoster(active realm.Character) ([]byte, error) {
	state := s.guilds.forPlayer(active.GUID)
	if state == nil {
		return s.guildCommandPacket(guildCreateCommand, "", guildNotInGuild)
	}
	name, err := packet.StringBytes(state.Guild.Name)
	if err != nil {
		return nil, err
	}
	accounts := make(map[int64]bool)
	entries := make([][]byte, 0, len(state.Members))
	for guid, rank := range state.Members {
		player, found, err := s.Characters.CharacterByGUID(guid)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		accounts[player.AccountID] = true
		playerName, err := packet.StringBytes(player.Name)
		if err != nil {
			return nil, err
		}
		entries = append(entries, append(playerName, encodeUint32(rank)...))
	}
	body := append(name, encodeUint32(int64(len(entries)))...)
	body = append(body, encodeUint32(int64(len(accounts)))...)
	for _, entry := range entries {
		body = append(body, entry...)
	}
	return packet.Encode(packet.SMSGGuildRoster, body)
}

func (s *WorldServer) guildLeave(active realm.Character) ([][]byte, error) {
	state := s.guilds.forPlayer(active.GUID)
	if state == nil {
		return s.guildCommand(active, guildInviteCommand, "", guildNotInGuild)
	}
	if state.Members[active.GUID] == 0 && len(state.Members) > 1 {
		return s.guildCommand(active, guildLeaveCommand, "", guildLeaderLeave)
	}
	if state.Members[active.GUID] == 0 {
		return nil, s.guildDisband(state)
	}
	delete(state.Members, active.GUID)
	s.guilds.removePlayer(state, active.GUID)
	if err := s.Characters.DeleteGuildMember(state.Guild.ID, active.GUID); err != nil {
		return nil, err
	}
	event, err := guildEvent(4, active.Name)
	if err != nil {
		return nil, err
	}
	s.broadcastGuild(state, event, 0)
	return nil, nil
}

func (s *WorldServer) guildDisband(state *guildState) error {
	event, err := guildEvent(8, "")
	if err != nil {
		return err
	}
	s.broadcastGuild(state, event, 0)
	if err := s.Characters.DeleteGuild(state.Guild.ID); err != nil {
		return err
	}
	s.guilds.remove(state)
	return nil
}

func (s *WorldServer) guildMOTD(active realm.Character, data []byte) ([]byte, error) {
	state := s.guilds.forPlayer(active.GUID)
	if state == nil {
		return s.guildCommandPacket(guildInviteCommand, "", guildNotInGuild)
	}
	if len(data) == 0 {
		return guildEvent(2, state.Guild.MOTD)
	}
	if state.Members[active.GUID] != 0 {
		return s.guildCommandPacket(guildInviteCommand, "", guildInternal)
	}
	motd, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	state.Guild.MOTD = strings.TrimSpace(motd)
	if err := s.Characters.UpdateGuildMOTD(state.Guild.ID, state.Guild.MOTD); err != nil {
		return nil, err
	}
	return guildEvent(2, state.Guild.MOTD)
}

func (s *WorldServer) guildPromote(active realm.Character, data []byte, demote bool) ([][]byte, error) {
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	state := s.guilds.forPlayer(active.GUID)
	if state == nil {
		return s.guildCommand(active, guildInviteCommand, "", guildNotInGuild)
	}
	if state.Members[active.GUID] > 1 {
		return s.guildCommand(active, guildInviteCommand, "", guildLeaderLeave)
	}
	target, found := s.playerByName(strings.TrimSpace(name))
	if !found {
		return s.guildCommand(active, guildInviteCommand, name, guildPlayerMissing)
	}
	targetRank, member := state.Members[target.GUID]
	if !member {
		return s.guildCommand(active, guildInviteCommand, name, guildNotInGuild)
	}
	if demote {
		if targetRank == 0 || targetRank >= 4 {
			return s.guildCommand(active, guildInviteCommand, "", guildInternal)
		}
		targetRank++
	} else {
		if targetRank <= 1 {
			return s.guildCommand(active, guildInviteCommand, "", guildInternal)
		}
		targetRank--
	}
	state.Members[target.GUID] = targetRank
	if err := s.Characters.UpdateGuildMemberRank(state.Guild.ID, target.GUID, targetRank); err != nil {
		return nil, err
	}
	event, err := guildRankEvent(target.Name, targetRank, demote)
	if err != nil {
		return nil, err
	}
	s.broadcastGuild(state, event, 0)
	return nil, nil
}

func (s *WorldServer) guildRemove(active realm.Character, data []byte) ([][]byte, error) {
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	state := s.guilds.forPlayer(active.GUID)
	if state == nil {
		return s.guildCommand(active, guildLeaveCommand, "", guildNotInGuild)
	}
	target, found := s.playerByName(strings.TrimSpace(name))
	if !found {
		return s.guildCommand(active, guildLeaveCommand, name, guildPlayerMissing)
	}
	targetRank, member := state.Members[target.GUID]
	if !member {
		return s.guildCommand(active, guildLeaveCommand, name, guildNotInGuild)
	}
	if state.Members[active.GUID] > 1 || targetRank <= 1 && state.Members[active.GUID] > 0 {
		return s.guildCommand(active, guildInviteCommand, "", guildLeaderLeave)
	}
	if targetRank == 0 {
		return s.guildCommand(active, guildLeaveCommand, "", guildLeaderLeave)
	}
	delete(state.Members, target.GUID)
	s.guilds.removePlayer(state, target.GUID)
	if err := s.Characters.DeleteGuildMember(state.Guild.ID, target.GUID); err != nil {
		return nil, err
	}
	event, err := guildRemoveEvent(target.Name, active.Name)
	if err != nil {
		return nil, err
	}
	s.broadcastGuild(state, event, 0)
	return nil, nil
}

func (s *WorldServer) guildSetLeader(active realm.Character, data []byte) ([][]byte, error) {
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	state := s.guilds.forPlayer(active.GUID)
	if state == nil {
		return s.guildCommand(active, guildInviteCommand, "", guildNotInGuild)
	}
	target, found := s.playerByName(strings.TrimSpace(name))
	if !found {
		return s.guildCommand(active, guildInviteCommand, name, guildPlayerMissing)
	}
	if _, member := state.Members[target.GUID]; !member {
		return s.guildCommand(active, guildInviteCommand, name, guildNotInGuild)
	}
	if state.Members[active.GUID] != 0 || target.GUID == active.GUID {
		return s.guildCommand(active, guildFounderCommand, "", guildLeaderLeave)
	}
	state.Members[active.GUID] = 1
	state.Members[target.GUID] = 0
	if err := s.Characters.UpdateGuildMemberRank(state.Guild.ID, active.GUID, 1); err != nil {
		return nil, err
	}
	if err := s.Characters.UpdateGuildMemberRank(state.Guild.ID, target.GUID, 0); err != nil {
		return nil, err
	}
	event, err := guildLeaderEvent(active.Name, target.Name)
	if err != nil {
		return nil, err
	}
	s.broadcastGuild(state, event, 0)
	return nil, nil
}

func (s *WorldServer) guildDisbandPlayer(active realm.Character) ([][]byte, error) {
	state := s.guilds.forPlayer(active.GUID)
	if state == nil {
		return s.guildCommand(active, guildCreateCommand, "", guildNotInGuild)
	}
	if state.Members[active.GUID] != 0 {
		return s.guildCommand(active, guildFounderCommand, "", guildLeaderLeave)
	}
	return nil, s.guildDisband(state)
}

func (s *WorldServer) guildChat(active realm.Character, data []byte, chatType byte) ([][]byte, error) {
	if len(data) < 9 {
		return nil, nil
	}
	message, err := packet.ReadString(data, 8, 0)
	if err != nil || message == "" {
		return nil, err
	}
	state := s.guilds.forPlayer(active.GUID)
	if state == nil || chatType == 3 && state.Members[active.GUID] > 1 {
		return nil, nil
	}
	response, err := messageChatPacket(chatType, binary.LittleEndian.Uint32(data[4:8]), active.GUID, message)
	if err != nil {
		return nil, err
	}
	for guid, rank := range state.Members {
		if guid != active.GUID && (chatType != 3 || rank <= 1) {
			s.sendPlayer(guid, response)
		}
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) guildCommand(active realm.Character, command uint32, name string, result uint32) ([][]byte, error) {
	response, err := s.guildCommandPacket(command, name, result)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) guildCommandPacket(command uint32, name string, result uint32) ([]byte, error) {
	nameBytes, err := packet.StringBytes(name)
	if err != nil {
		return nil, err
	}
	body := append(encodeUint32(int64(command)), nameBytes...)
	body = append(body, encodeUint32(int64(result))...)
	return packet.Encode(packet.SMSGGuildCommandResult, body)
}

func guildEvent(event byte, text string) ([]byte, error) {
	name, err := packet.StringBytes(text)
	if err != nil {
		return nil, err
	}
	return packet.Encode(packet.SMSGGuildEvent, append([]byte{event, 1}, name...))
}

func guildRankEvent(name string, targetRank int64, demote bool) ([]byte, error) {
	event := byte(0)
	target, err := packet.StringBytes(name)
	if err != nil {
		return nil, err
	}
	rankName := []string{"Guild Master", "Officer", "Veteran", "Member", "Initiate"}[targetRank]
	rankBytes, err := packet.StringBytes(rankName)
	if err != nil {
		return nil, err
	}
	if demote {
		event = 1
	} else {
		event = 0
	}
	return packet.Encode(packet.SMSGGuildEvent, append([]byte{event, 2}, append(target, rankBytes...)...))
}

func guildRemoveEvent(name, remover string) ([]byte, error) {
	first, err := packet.StringBytes(name)
	if err != nil {
		return nil, err
	}
	second, err := packet.StringBytes(remover)
	if err != nil {
		return nil, err
	}
	return packet.Encode(packet.SMSGGuildEvent, append([]byte{5, 2}, append(first, second...)...))
}

func guildLeaderEvent(previous, current string) ([]byte, error) {
	first, err := packet.StringBytes(previous)
	if err != nil {
		return nil, err
	}
	second, err := packet.StringBytes(current)
	if err != nil {
		return nil, err
	}
	return packet.Encode(packet.SMSGGuildEvent, append([]byte{7, 2}, append(first, second...)...))
}

func (s *WorldServer) broadcastGuild(state *guildState, data []byte, exclude int64) {
	for guid := range state.Members {
		if guid != exclude {
			s.sendPlayer(guid, data)
		}
	}
}

func validGuildName(name string) bool {
	if len(name) < 2 || len(name) > 24 || name[0] == ' ' || name[len(name)-1] == ' ' {
		return false
	}
	for _, value := range strings.ReplaceAll(name, " ", "") {
		if (value < 'A' || value > 'Z') && (value < 'a' || value > 'z') {
			return false
		}
	}
	return true
}
