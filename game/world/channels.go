package world

import (
	"encoding/binary"
	"fmt"
	"strings"
	"sync"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const (
	channelPlayerJoined       byte = 0x00
	channelPlayerLeft         byte = 0x01
	channelYouJoined          byte = 0x02
	channelYouLeft            byte = 0x03
	channelWrongPassword      byte = 0x04
	channelNotMember          byte = 0x05
	channelNotModerator       byte = 0x06
	channelPasswordChanged    byte = 0x07
	channelOwnerChanged       byte = 0x08
	channelPlayerNotFound     byte = 0x09
	channelNotOwner           byte = 0x0a
	channelOwner              byte = 0x0b
	channelMemberFlagChange   byte = 0x0c
	channelAnnouncementsOn    byte = 0x0d
	channelAnnouncementsOff   byte = 0x0e
	channelModerationOn       byte = 0x0f
	channelModerationOff      byte = 0x10
	channelSelfMuted          byte = 0x11
	channelKicked             byte = 0x12
	channelPlayerBanned       byte = 0x13
	channelBanned             byte = 0x14
	channelUnbanned           byte = 0x15
	channelPlayerNotBanned    byte = 0x16
	channelAlreadyMember      byte = 0x17
	channelInvite             byte = 0x18
	channelInviteWrongFaction byte = 0x19
	channelWrongFaction       byte = 0x1a
)

type channelState struct {
	Name, Password string
	Team           int64
	Default        bool
	Announce       bool
	Moderated      bool
	Owner          int64
	Members        map[int64]bool
	Moderators     map[int64]bool
	Muted          map[int64]bool
	Banned         map[int64]bool
}

type channelRegistry struct {
	mu       sync.RWMutex
	channels map[string]*channelState
}

func channelDisplayName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func channelKey(team int64, name string) string {
	return fmt.Sprintf("%d:%s", team, strings.ToLower(name))
}

func (s *WorldServer) channel(team int64, name string) *channelState {
	s.channels.mu.RLock()
	channel := s.channels.channels[channelKey(team, name)]
	s.channels.mu.RUnlock()
	return channel
}

func (s *WorldServer) createChannel(team int64, name, password string) *channelState {
	channel := &channelState{Name: channelDisplayName(name), Password: password, Team: team, Announce: true, Owner: 0, Members: make(map[int64]bool), Moderators: make(map[int64]bool), Muted: make(map[int64]bool), Banned: make(map[int64]bool)}
	s.channels.mu.Lock()
	if s.channels.channels == nil {
		s.channels.channels = make(map[string]*channelState)
	}
	key := channelKey(team, name)
	if existing := s.channels.channels[key]; existing != nil {
		channel = existing
	} else {
		s.channels.channels[key] = channel
	}
	s.channels.mu.Unlock()
	return channel
}

func (s *WorldServer) channelTeam(player realm.Character) int64 {
	if s.DBC == nil {
		return 0
	}
	race, found, err := s.DBC.Race(player.Race)
	if err == nil && found {
		return race.BaseLanguage
	}
	return 0
}

func (s *WorldServer) joinChannel(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 1 {
		return nil, nil
	}
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	name = channelDisplayName(name)
	if name == "" {
		return nil, nil
	}
	password := ""
	offset := len(name) + 1
	if offset < len(data) {
		password, _ = packet.ReadString(data, offset, 0)
		password = strings.TrimSpace(password)
	}
	team := s.channelTeam(active)
	channel := s.channel(team, name)
	if channel == nil {
		channel = s.createChannel(team, name, password)
		channel.Owner = active.GUID
		channel.Moderators[active.GUID] = true
	}
	if channel.Members[active.GUID] {
		return s.channelResponses(channelNotify(channel.Name, channelAlreadyMember, active.GUID, 0, "", nil, true))
	}
	if channel.Password != password {
		return s.channelResponses(channelNotify(channel.Name, channelWrongPassword, 0, 0, "", nil, false))
	}
	if channel.Banned[active.GUID] {
		return s.channelResponses(channelNotify(channel.Name, channelPlayerBanned, 0, 0, "", nil, false))
	}
	for member := range channel.Members {
		if channel.Announce {
			notification, err := channelNotify(channel.Name, channelPlayerJoined, active.GUID, 0, "", nil, true)
			if err != nil {
				return nil, err
			}
			s.sendPlayer(member, notification)
		}
	}
	channel.Members[active.GUID] = true
	return s.channelResponses(channelNotify(channel.Name, channelYouJoined, 0, 0, "", nil, false))
}

func (s *WorldServer) leaveChannel(active realm.Character, data []byte) ([][]byte, error) {
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	channel := s.channel(s.channelTeam(active), channelDisplayName(name))
	if channel == nil || !channel.Members[active.GUID] {
		return s.channelResponses(channelNotify(channelDisplayName(name), channelNotMember, 0, 0, "", nil, false))
	}
	delete(channel.Members, active.GUID)
	for member := range channel.Members {
		if channel.Announce {
			notification, err := channelNotify(channel.Name, channelPlayerLeft, active.GUID, 0, "", nil, true)
			if err != nil {
				return nil, err
			}
			s.sendPlayer(member, notification)
		}
	}
	return s.channelResponses(channelNotify(channel.Name, channelYouLeft, 0, 0, "", nil, false))
}

func (s *WorldServer) listChannel(active realm.Character, data []byte) ([][]byte, error) {
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	channel := s.channel(s.channelTeam(active), channelDisplayName(name))
	if channel == nil || !channel.Members[active.GUID] {
		return s.channelResponses(channelNotify(channelDisplayName(name), channelNotMember, 0, 0, "", nil, false))
	}
	nameBytes, err := packet.StringBytes(channel.Name)
	if err != nil {
		return nil, err
	}
	flags := byte(1)
	if channel.Default {
		flags = 0
	}
	body := append(nameBytes, flags)
	body = append(body, encodeUint32(int64(len(channel.Members)))...)
	for member := range channel.Members {
		body = append(body, encodeGUID(member)...)
		mode := byte(0)
		if channel.Moderators[member] || channel.Owner == member {
			mode |= 2
		}
		if channel.Muted[member] {
			mode |= 4
		}
		body = append(body, mode)
	}
	response, err := packet.Encode(packet.SMSGChannelList, body)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) channelChat(active realm.Character, data []byte) ([][]byte, error) {
	name, err := packet.ReadString(data, 8, 0)
	if err != nil {
		return nil, nil
	}
	offset := 8 + len(name) + 1
	message, err := packet.ReadString(data, offset, 0)
	if err != nil || message == "" {
		return nil, err
	}
	channel := s.channel(s.channelTeam(active), channelDisplayName(name))
	if channel == nil || !channel.Members[active.GUID] {
		return s.channelResponses(channelNotify(channelDisplayName(name), channelNotMember, 0, 0, "", nil, false))
	}
	if channel.Muted[active.GUID] {
		return s.channelResponses(channelNotify(channel.Name, channelSelfMuted, 0, 0, "", nil, false))
	}
	response, err := channelMessagePacket(data[0], binary.LittleEndian.Uint32(data[4:8]), channel.Name, active.GUID, message)
	if err != nil {
		return nil, err
	}
	for member := range channel.Members {
		if member != active.GUID {
			s.sendPlayer(member, response)
		}
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) channelResponses(data []byte, err error) ([][]byte, error) {
	if err != nil {
		return nil, err
	}
	return [][]byte{data}, nil
}

func channelNotify(name string, notification byte, target1, target2 int64, playerName string, flags []byte, includeTarget bool) ([]byte, error) {
	nameBytes, err := packet.StringBytes(name)
	if err != nil {
		return nil, err
	}
	body := append([]byte{notification}, nameBytes...)
	if includeTarget {
		body = append(body, encodeGUID(target1)...)
	}
	if target2 != 0 {
		body = append(body, encodeGUID(target2)...)
	}
	if playerName != "" {
		playerBytes, err := packet.StringBytes(playerName)
		if err != nil {
			return nil, err
		}
		body = append(body, playerBytes...)
	}
	if len(flags) == 2 {
		body = append(body, flags...)
	}
	return packet.Encode(packet.SMSGChannelNotify, body)
}

func channelMessagePacket(chatType byte, language uint32, name string, guid int64, message string) ([]byte, error) {
	channel, err := packet.StringBytes(name)
	if err != nil {
		return nil, err
	}
	text, err := packet.StringBytes(message)
	if err != nil {
		return nil, err
	}
	body := append([]byte{chatType}, encodeUint32(int64(language))...)
	body = append(body, channel...)
	body = append(body, encodeGUID(guid)...)
	body = append(body, text...)
	body = append(body, 0)
	return packet.Encode(packet.SMSGMessageChat, body)
}

func (s *WorldServer) channelPassword(active realm.Character, data []byte) ([]byte, error) {
	name, offset, err := readCString(data, 0)
	if err != nil {
		return nil, nil
	}
	password, _ := packet.ReadString(data, offset, 0)
	channel := s.channel(s.channelTeam(active), channelDisplayName(name))
	if response, err := s.channelCheck(channel, active, true, true); response != nil || err != nil {
		return response, err
	}
	channel.Password = strings.TrimSpace(password)
	response, err := channelNotify(channel.Name, channelPasswordChanged, 0, 0, active.Name, nil, false)
	if err != nil {
		return nil, err
	}
	if channel.Announce {
		s.broadcastChannel(channel, response, 0)
		return nil, nil
	}
	return response, nil
}

func (s *WorldServer) channelOwner(active realm.Character, data []byte) ([]byte, error) {
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	channel := s.channel(s.channelTeam(active), channelDisplayName(name))
	if response, err := s.channelCheck(channel, active, false, false); response != nil || err != nil {
		return response, err
	}
	owner, found := s.playerByGUID(channel.Owner)
	if !found {
		owner, found, err = s.Characters.CharacterByGUID(channel.Owner)
		if err != nil || !found {
			return nil, err
		}
	}
	return channelNotify(channel.Name, channelOwner, 0, 0, owner.Name, nil, false)
}

func (s *WorldServer) channelSetOwner(active realm.Character, data []byte) ([]byte, error) {
	channelName, targetName, err := channelTarget(data)
	if err != nil {
		return nil, nil
	}
	channel := s.channel(s.channelTeam(active), channelDisplayName(channelName))
	if response, err := s.channelCheck(channel, active, true, true); response != nil || err != nil {
		return response, err
	}
	target, found := s.playerByName(strings.TrimSpace(targetName))
	if !found || !channel.Members[target.GUID] || target.GUID == active.GUID {
		return nil, nil
	}
	channel.Owner = target.GUID
	response, err := channelNotify(channel.Name, channelOwnerChanged, target.GUID, 0, "", nil, true)
	if err != nil {
		return nil, err
	}
	if channel.Announce {
		s.broadcastChannel(channel, response, 0)
	} else {
		s.sendPlayer(target.GUID, response)
	}
	return response, nil
}

func (s *WorldServer) channelModerator(active realm.Character, data []byte, remove bool) ([]byte, error) {
	channelName, targetName, err := channelTarget(data)
	if err != nil {
		return nil, nil
	}
	channel := s.channel(s.channelTeam(active), channelDisplayName(channelName))
	if response, err := s.channelCheck(channel, active, true, !remove); response != nil || err != nil {
		return response, err
	}
	target, found := s.playerByName(strings.TrimSpace(targetName))
	if !found || !channel.Members[target.GUID] || target.GUID == active.GUID {
		return nil, nil
	}
	if remove {
		if !channel.Moderators[target.GUID] {
			return nil, nil
		}
		delete(channel.Moderators, target.GUID)
	} else {
		if channel.Moderators[target.GUID] {
			return nil, nil
		}
		channel.Moderators[target.GUID] = true
	}
	flags := []byte{1, 2}
	if remove {
		flags = []byte{2, 1}
	}
	response, err := channelNotify(channel.Name, channelMemberFlagChange, target.GUID, 0, "", flags, true)
	if err != nil {
		return nil, err
	}
	if channel.Announce {
		s.broadcastChannel(channel, response, 0)
	} else {
		s.sendPlayer(target.GUID, response)
	}
	return response, nil
}

func (s *WorldServer) channelMute(active realm.Character, data []byte, remove bool) ([]byte, error) {
	channelName, targetName, err := channelTarget(data)
	if err != nil {
		return nil, nil
	}
	channel := s.channel(s.channelTeam(active), channelDisplayName(channelName))
	if response, err := s.channelCheck(channel, active, true, true); response != nil || err != nil {
		return response, err
	}
	target, found := s.playerByName(strings.TrimSpace(targetName))
	if !found || !channel.Members[target.GUID] || target.GUID == active.GUID {
		return nil, nil
	}
	if remove {
		if !channel.Muted[target.GUID] {
			return nil, nil
		}
		delete(channel.Muted, target.GUID)
	} else {
		if channel.Muted[target.GUID] {
			return nil, nil
		}
		channel.Muted[target.GUID] = true
	}
	flags := []byte{4, 1}
	if remove {
		flags = []byte{1, 4}
	}
	response, err := channelNotify(channel.Name, channelMemberFlagChange, target.GUID, 0, "", flags, true)
	if err != nil {
		return nil, err
	}
	if channel.Announce {
		s.broadcastChannel(channel, response, 0)
	} else {
		s.sendPlayer(target.GUID, response)
	}
	return response, nil
}

func (s *WorldServer) channelInvite(active realm.Character, data []byte) ([]byte, error) {
	channelName, targetName, err := channelTarget(data)
	if err != nil {
		return nil, nil
	}
	channel := s.channel(s.channelTeam(active), channelDisplayName(channelName))
	if response, err := s.channelCheck(channel, active, true, true); response != nil || err != nil {
		return response, err
	}
	target, found := s.playerByName(strings.TrimSpace(targetName))
	if !found || channel.Members[target.GUID] || channel.Banned[target.GUID] || s.channelTeam(target) != channel.Team {
		return nil, nil
	}
	response, err := channelNotify(channel.Name, channelInvite, active.GUID, 0, "", nil, true)
	if err != nil {
		return nil, err
	}
	s.sendPlayer(target.GUID, response)
	return nil, nil
}

func (s *WorldServer) channelKick(active realm.Character, data []byte, ban bool) ([]byte, error) {
	channelName, targetName, err := channelTarget(data)
	if err != nil {
		return nil, nil
	}
	channel := s.channel(s.channelTeam(active), channelDisplayName(channelName))
	if response, err := s.channelCheck(channel, active, true, true); response != nil || err != nil {
		return response, err
	}
	target, found := s.playerByName(strings.TrimSpace(targetName))
	if !found || !channel.Members[target.GUID] || target.GUID == active.GUID {
		return nil, nil
	}
	if ban {
		channel.Banned[target.GUID] = true
	}
	notification := channelKicked
	if ban {
		notification = channelBanned
	}
	response, err := channelNotify(channel.Name, notification, target.GUID, active.GUID, "", nil, true)
	if err != nil {
		return nil, err
	}
	if channel.Announce {
		s.broadcastChannel(channel, response, target.GUID)
	}
	s.sendPlayer(target.GUID, response)
	delete(channel.Members, target.GUID)
	delete(channel.Moderators, target.GUID)
	delete(channel.Muted, target.GUID)
	return nil, nil
}

func (s *WorldServer) channelUnban(active realm.Character, data []byte) ([]byte, error) {
	channelName, targetName, err := channelTarget(data)
	if err != nil {
		return nil, nil
	}
	channel := s.channel(s.channelTeam(active), channelDisplayName(channelName))
	if response, err := s.channelCheck(channel, active, true, true); response != nil || err != nil {
		return response, err
	}
	target, found := s.playerByName(strings.TrimSpace(targetName))
	if !found {
		return nil, nil
	}
	if !channel.Banned[target.GUID] {
		return channelNotify(channel.Name, channelPlayerNotFound, 0, 0, target.Name, nil, false)
	}
	delete(channel.Banned, target.GUID)
	return channelNotify(channel.Name, channelUnbanned, target.GUID, active.GUID, "", nil, true)
}

func (s *WorldServer) channelToggle(active realm.Character, data []byte, moderation bool) ([]byte, error) {
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	channel := s.channel(s.channelTeam(active), channelDisplayName(name))
	if response, err := s.channelCheck(channel, active, true, true); response != nil || err != nil {
		return response, err
	}
	if moderation {
		channel.Moderated = !channel.Moderated
		code := channelModerationOff
		if !channel.Moderated {
			code = channelModerationOn
		}
		response, err := channelNotify(channel.Name, code, 0, 0, "", nil, false)
		if err != nil {
			return nil, err
		}
		if channel.Announce {
			s.broadcastChannel(channel, response, 0)
			return nil, nil
		}
		return response, nil
	}
	channel.Announce = !channel.Announce
	code := channelAnnouncementsOff
	if !channel.Announce {
		code = channelAnnouncementsOn
	}
	return channelNotify(channel.Name, code, 0, 0, "", nil, false)
}

func (s *WorldServer) channelCheck(channel *channelState, active realm.Character, owner, moderator bool) ([]byte, error) {
	if channel == nil || !channel.Members[active.GUID] {
		return channelNotify(channelDisplayName(active.Name), channelNotMember, 0, 0, "", nil, false)
	}
	if owner && channel.Owner != active.GUID {
		return channelNotify(channel.Name, channelNotOwner, 0, 0, "", nil, false)
	}
	if moderator && !channel.Moderators[active.GUID] {
		return channelNotify(channel.Name, channelNotModerator, 0, 0, "", nil, false)
	}
	return nil, nil
}

func (s *WorldServer) broadcastChannel(channel *channelState, data []byte, exclude int64) {
	for member := range channel.Members {
		if member != exclude {
			s.sendPlayer(member, data)
		}
	}
}

func channelTarget(data []byte) (string, string, error) {
	channel, offset, err := readCString(data, 0)
	if err != nil {
		return "", "", err
	}
	target, _, err := readCString(data, offset)
	return channel, target, err
}
