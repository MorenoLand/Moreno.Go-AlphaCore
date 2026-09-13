package world

import (
	"strings"
	"sync"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const (
	partyInvite       uint32 = 0
	partyLeave        uint32 = 2
	partyOK           uint32 = 0
	partyBadName      uint32 = 1
	partyNotMember    uint32 = 2
	partyFull         uint32 = 3
	partyAlreadyGroup uint32 = 4
	partyNotInGroup   uint32 = 5
	partyNotLeader    uint32 = 6
	partyWrongFaction uint32 = 7
	partyIgnoring     uint32 = 8
	partyRestricted   uint32 = 9
)

type groupState struct {
	ID, LeaderGUID, LootMaster int64
	LootMethod                 byte
	Members                    []int64
	Invites                    map[int64]bool
}

type groupRegistry struct {
	mu      sync.RWMutex
	nextID  int64
	groups  map[int64]*groupState
	players map[int64]*groupState
	invites map[int64]*groupState
}

func (r *groupRegistry) pending(leader int64) *groupState {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.groups == nil {
		r.groups = make(map[int64]*groupState)
		r.players = make(map[int64]*groupState)
		r.invites = make(map[int64]*groupState)
	}
	r.nextID--
	if r.nextID >= 0 {
		r.nextID = -1
	}
	group := &groupState{ID: r.nextID, LeaderGUID: leader, Invites: make(map[int64]bool)}
	r.groups[group.ID] = group
	r.players[leader] = group
	return group
}

func (r *groupRegistry) group(guid int64) *groupState {
	r.mu.RLock()
	group := r.players[guid]
	r.mu.RUnlock()
	return group
}

func (r *groupRegistry) invite(guid int64) *groupState {
	r.mu.RLock()
	group := r.invites[guid]
	r.mu.RUnlock()
	return group
}

func (r *groupRegistry) addInvite(group *groupState, guid int64) {
	r.mu.Lock()
	if r.invites == nil {
		r.invites = make(map[int64]*groupState)
	}
	r.invites[guid] = group
	group.Invites[guid] = true
	r.mu.Unlock()
}

func (r *groupRegistry) removeInvite(group *groupState, guid int64) {
	r.mu.Lock()
	delete(r.invites, guid)
	delete(group.Invites, guid)
	r.mu.Unlock()
}

func (r *groupRegistry) addMember(group *groupState, guid int64) {
	r.mu.Lock()
	group.Members = append(group.Members, guid)
	if r.players == nil {
		r.players = make(map[int64]*groupState)
	}
	r.players[guid] = group
	r.mu.Unlock()
}

func (r *groupRegistry) removeMember(group *groupState, guid int64) {
	r.mu.Lock()
	delete(r.players, guid)
	for index, member := range group.Members {
		if member == guid {
			group.Members = append(group.Members[:index], group.Members[index+1:]...)
			break
		}
	}
	r.mu.Unlock()
}

func (r *groupRegistry) remove(group *groupState) {
	r.mu.Lock()
	delete(r.groups, group.ID)
	for _, guid := range group.Members {
		delete(r.players, guid)
	}
	for guid := range group.Invites {
		delete(r.invites, guid)
	}
	r.mu.Unlock()
}

func (r *groupRegistry) setID(group *groupState, id int64) {
	r.mu.Lock()
	delete(r.groups, group.ID)
	group.ID = id
	r.groups[id] = group
	r.mu.Unlock()
}

func (r *groupRegistry) loaded(group realm.Group, members []int64) *groupState {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.groups == nil {
		r.groups = make(map[int64]*groupState)
		r.players = make(map[int64]*groupState)
		r.invites = make(map[int64]*groupState)
	}
	if state := r.groups[group.ID]; state != nil {
		return state
	}
	state := &groupState{ID: group.ID, LeaderGUID: group.LeaderGUID, LootMaster: group.LootMaster, LootMethod: byte(group.LootMethod), Members: append([]int64(nil), members...), Invites: make(map[int64]bool)}
	r.groups[state.ID] = state
	for _, member := range state.Members {
		r.players[member] = state
	}
	return state
}

func (s *WorldServer) groupInvite(active realm.Character, data []byte) ([][]byte, error) {
	name, err := packet.ReadString(data, 0, 0)
	if err != nil || strings.TrimSpace(name) == "" {
		return nil, nil
	}
	name = strings.TrimSpace(name)
	target, found := s.playerByName(name)
	if !found {
		return s.partyResult(active, name, partyBadName)
	}
	if target.GUID == active.GUID {
		return s.partyResult(active, target.Name, partyRestricted)
	}
	if targetSocial, socialErr := s.Characters.Social(target.GUID); socialErr != nil {
		return nil, socialErr
	} else if hasSocial(targetSocial, active.GUID, true) {
		return s.partyResult(active, target.Name, partyIgnoring)
	}
	if activeTeam, targetTeam, teamErr := s.teams(active, target); teamErr != nil {
		return nil, teamErr
	} else if activeTeam != targetTeam {
		return s.partyResult(active, target.Name, partyWrongFaction)
	}
	if s.groupFor(active.GUID) != nil {
		group := s.groupFor(active.GUID)
		if group.LeaderGUID != active.GUID {
			return s.partyResult(active, target.Name, partyNotLeader)
		}
		if len(group.Members) >= 5 {
			return s.partyResult(active, target.Name, partyFull)
		}
	}
	if s.groupFor(target.GUID) != nil || s.groups.invite(target.GUID) != nil {
		return s.partyResult(active, target.Name, partyAlreadyGroup)
	}
	group := s.groupFor(active.GUID)
	if group == nil {
		group = s.groups.pending(active.GUID)
	}
	s.groups.addInvite(group, target.GUID)
	inviteName, err := packet.StringBytes(active.Name)
	if err != nil {
		return nil, err
	}
	invite, err := packet.Encode(packet.SMSGGroupInvite, inviteName)
	if err != nil {
		return nil, err
	}
	s.sendPlayer(target.GUID, invite)
	return s.partyResult(active, target.Name, partyOK)
}

func (s *WorldServer) groupAccept(active realm.Character) error {
	group := s.groups.invite(active.GUID)
	if group == nil {
		return nil
	}
	s.groups.removeInvite(group, active.GUID)
	if group.ID < 0 {
		created, err := s.Characters.CreateGroup(group.LeaderGUID)
		if err != nil {
			return err
		}
		s.groups.setID(group, created.ID)
		if err := s.Characters.AddGroupMember(group.ID, group.LeaderGUID); err != nil {
			return err
		}
	} else if len(group.Members) == 0 {
		if err := s.Characters.AddGroupMember(group.ID, group.LeaderGUID); err != nil {
			return err
		}
	}
	if !groupMember(group, group.LeaderGUID) {
		s.groups.addMember(group, group.LeaderGUID)
	}
	s.groups.addMember(group, active.GUID)
	if err := s.Characters.AddGroupMember(group.ID, active.GUID); err != nil {
		return err
	}
	for _, member := range group.Members {
		s.setGroupStatus(member, 1)
	}
	return s.sendGroupLists(group)
}

func (s *WorldServer) groupDecline(active realm.Character) {
	group := s.groups.invite(active.GUID)
	if group == nil {
		return
	}
	s.groups.removeInvite(group, active.GUID)
	if leader, found := s.playerByGUID(group.LeaderGUID); found {
		if name, err := packet.StringBytes(active.Name); err == nil {
			if response, err := packet.Encode(packet.SMSGGroupDecline, name); err == nil {
				s.sendPlayer(leader.GUID, response)
			}
		}
	}
	if len(group.Members) == 0 {
		s.groups.remove(group)
	}
}

func (s *WorldServer) groupDisband(active realm.Character) ([][]byte, error) {
	group := s.groupFor(active.GUID)
	if group == nil {
		return s.partyLeaveResult(active, "", partyNotInGroup)
	}
	for _, member := range group.Members {
		if member != active.GUID {
			s.setGroupStatus(member, 0)
			if response, err := packet.Encode(packet.SMSGGroupDestroyed, nil); err == nil {
				s.sendPlayer(member, response)
			}
		}
	}
	if group.ID > 0 {
		if err := s.Characters.DeleteGroup(group.ID); err != nil {
			return nil, err
		}
	}
	s.groups.remove(group)
	s.setGroupStatus(active.GUID, 0)
	return nil, nil
}

func (s *WorldServer) groupUninvite(active realm.Character, targetGUID int64) ([][]byte, error) {
	group := s.groupFor(active.GUID)
	if group == nil {
		return s.partyResult(active, "", partyNotInGroup)
	}
	if group.LeaderGUID != active.GUID {
		return s.partyLeaveResult(active, "", partyNotLeader)
	}
	if !groupMember(group, targetGUID) {
		return s.partyLeaveResult(active, "", partyNotMember)
	}
	s.groups.removeMember(group, targetGUID)
	s.setGroupStatus(targetGUID, 0)
	if group.ID > 0 {
		if err := s.Characters.DeleteGroupMember(group.ID, targetGUID); err != nil {
			return nil, err
		}
	}
	if response, err := packet.Encode(packet.SMSGGroupUninvite, nil); err == nil {
		s.sendPlayer(targetGUID, response)
	} else {
		return nil, err
	}
	if len(group.Members) < 2 {
		return s.groupDisband(active)
	}
	return nil, s.sendGroupLists(group)
}

func (s *WorldServer) groupUninviteName(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 2 {
		return nil, nil
	}
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	target, found := s.playerByName(strings.TrimSpace(name))
	if !found {
		return s.partyResult(active, name, partyBadName)
	}
	return s.groupUninvite(active, target.GUID)
}

func (s *WorldServer) groupSetLeader(active realm.Character, data []byte) ([][]byte, error) {
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	target, found := s.playerByName(strings.TrimSpace(name))
	group := s.groupFor(active.GUID)
	if group == nil {
		return s.partyLeaveResult(active, "", partyNotInGroup)
	}
	if !found || !groupMember(group, target.GUID) {
		return s.partyLeaveResult(active, name, partyNotMember)
	}
	if group.LeaderGUID != active.GUID {
		return s.partyLeaveResult(active, "", partyNotLeader)
	}
	group.LeaderGUID = target.GUID
	if group.ID > 0 {
		if err := s.Characters.UpdateGroupLeader(group.ID, target.GUID); err != nil {
			return nil, err
		}
	}
	leaderName, err := packet.StringBytes(target.Name)
	if err != nil {
		return nil, err
	}
	response, err := packet.Encode(packet.SMSGGroupSetLeader, leaderName)
	if err != nil {
		return nil, err
	}
	for _, member := range group.Members {
		s.sendPlayer(member, response)
	}
	return nil, s.sendGroupLists(group)
}

func (s *WorldServer) sendGroupLists(group *groupState) error {
	for _, member := range group.Members {
		player, found := s.playerByGUID(member)
		if !found {
			continue
		}
		data, err := s.groupListData(group, player)
		if err != nil {
			return err
		}
		response, err := packet.Encode(packet.SMSGGroupList, data)
		if err != nil {
			return err
		}
		s.sendPlayer(member, response)
	}
	return nil
}

func (s *WorldServer) groupListData(group *groupState, viewer realm.Character) ([]byte, error) {
	leader, found := s.playerByGUID(group.LeaderGUID)
	if !found {
		leader, found, _ = s.Characters.CharacterByGUID(group.LeaderGUID)
	}
	if !found {
		return nil, nil
	}
	leaderName, err := packet.StringBytes(leader.Name)
	if err != nil {
		return nil, err
	}
	count := len(group.Members)
	if viewer.GUID != group.LeaderGUID {
		count--
	}
	data := append(encodeUint32(int64(count)), leaderName...)
	data = append(data, encodeGUID(group.LeaderGUID)...)
	if _, online := s.playerByGUID(group.LeaderGUID); online {
		data = append(data, 1)
	} else {
		data = append(data, 0)
	}
	for _, member := range group.Members {
		if member == group.LeaderGUID || member == viewer.GUID {
			continue
		}
		character, found := s.playerByGUID(member)
		if !found {
			character, found, _ = s.Characters.CharacterByGUID(member)
		}
		if !found {
			continue
		}
		name, err := packet.StringBytes(character.Name)
		if err != nil {
			return nil, err
		}
		data = append(data, name...)
		data = append(data, encodeGUID(character.GUID)...)
		if _, online := s.playerByGUID(character.GUID); online {
			data = append(data, 1)
		} else {
			data = append(data, 0)
		}
	}
	data = append(data, group.LootMethod)
	data = append(data, encodeGUID(group.LootMaster)...)
	return data, nil
}

func (s *WorldServer) partyResult(active realm.Character, name string, result uint32) ([][]byte, error) {
	return s.partyResultOperation(active, partyInvite, name, result)
}

func (s *WorldServer) partyLeaveResult(active realm.Character, name string, result uint32) ([][]byte, error) {
	return s.partyResultOperation(active, partyLeave, name, result)
}

func (s *WorldServer) partyResultOperation(active realm.Character, operation uint32, name string, result uint32) ([][]byte, error) {
	nameBytes, err := packet.StringBytes(name)
	if err != nil {
		return nil, err
	}
	body := append(encodeUint32(int64(operation)), nameBytes...)
	body = append(body, encodeUint32(int64(result))...)
	response, err := packet.Encode(packet.SMSGPartyCommandResult, body)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) groupFor(guid int64) *groupState { return s.groups.group(guid) }

func (s *WorldServer) loadGroup(player realm.Character) error {
	group, found, err := s.Characters.GroupByPlayer(player.GUID)
	if err != nil || !found {
		return err
	}
	members, err := s.Characters.GroupMembers(group.ID)
	if err != nil {
		return err
	}
	s.groups.loaded(group, members)
	s.setGroupStatus(player.GUID, 1)
	return nil
}

func groupMember(group *groupState, guid int64) bool {
	for _, member := range group.Members {
		if member == guid {
			return true
		}
	}
	return false
}
