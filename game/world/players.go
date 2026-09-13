package world

import (
	"sync"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

type playerRegistry struct {
	mu          sync.RWMutex
	players     map[int64]realm.Character
	groupStatus map[int64]uint32
	selection   map[int64]uint64
	target      map[int64]uint64
	standState  map[int64]uint32
	weaponMode  map[int64]uint32
	connections map[int64]*playerConnection
}

func (s *WorldServer) registerPlayer(character realm.Character) {
	s.players.mu.Lock()
	if s.players.players == nil {
		s.players.players = make(map[int64]realm.Character)
		s.players.groupStatus = make(map[int64]uint32)
		s.players.selection = make(map[int64]uint64)
		s.players.target = make(map[int64]uint64)
		s.players.standState = make(map[int64]uint32)
		s.players.weaponMode = make(map[int64]uint32)
		s.players.connections = make(map[int64]*playerConnection)
	}
	s.players.players[character.GUID] = character
	s.players.mu.Unlock()
}

func (s *WorldServer) updatePlayer(character realm.Character) {
	s.players.mu.Lock()
	if s.players.players != nil {
		s.players.players[character.GUID] = character
	}
	s.players.mu.Unlock()
}

func (s *WorldServer) unregisterPlayer(guid int64) {
	s.players.mu.Lock()
	delete(s.players.players, guid)
	delete(s.players.groupStatus, guid)
	delete(s.players.selection, guid)
	delete(s.players.target, guid)
	delete(s.players.standState, guid)
	delete(s.players.weaponMode, guid)
	delete(s.players.connections, guid)
	s.players.mu.Unlock()
}

func (s *WorldServer) getGroupStatus(guid int64) uint32 {
	s.players.mu.RLock()
	status := s.players.groupStatus[guid]
	s.players.mu.RUnlock()
	return status
}

func (s *WorldServer) setGroupStatus(guid int64, status uint32) {
	s.players.mu.Lock()
	if s.players.groupStatus == nil {
		s.players.groupStatus = make(map[int64]uint32)
	}
	s.players.groupStatus[guid] = status
	s.players.mu.Unlock()
}

func (s *WorldServer) setPlayerSelection(guid int64, value uint64) {
	s.players.mu.Lock()
	if s.players.selection == nil {
		s.players.selection = make(map[int64]uint64)
	}
	s.players.selection[guid] = value
	s.players.mu.Unlock()
}

func (s *WorldServer) setPlayerTarget(guid int64, value uint64) {
	s.players.mu.Lock()
	if s.players.target == nil {
		s.players.target = make(map[int64]uint64)
	}
	s.players.target[guid] = value
	s.players.mu.Unlock()
}

func (s *WorldServer) setStandState(guid int64, value uint32) {
	s.players.mu.Lock()
	if s.players.standState == nil {
		s.players.standState = make(map[int64]uint32)
	}
	s.players.standState[guid] = value
	s.players.mu.Unlock()
}

func (s *WorldServer) setWeaponMode(guid int64, value uint32) {
	s.players.mu.Lock()
	if s.players.weaponMode == nil {
		s.players.weaponMode = make(map[int64]uint32)
	}
	s.players.weaponMode[guid] = value
	s.players.mu.Unlock()
}

func (s *WorldServer) bytes1(guid int64) uint32 {
	s.players.mu.RLock()
	mode, stand := s.players.weaponMode[guid], s.players.standState[guid]
	s.players.mu.RUnlock()
	if mode == 0 {
		mode = 1
	}
	return mode<<24 | stand
}

func (s *WorldServer) bytes1Update(active realm.Character) ([]byte, error) {
	update, err := packet.EncodeFieldUpdate(uint64(active.GUID), 172, s.bytes1(active.GUID))
	if err != nil {
		return nil, err
	}
	s.broadcastPlayer(active, update)
	return update, nil
}

func (s *WorldServer) onlinePlayers() []realm.Character {
	s.players.mu.RLock()
	players := make([]realm.Character, 0, len(s.players.players))
	for _, player := range s.players.players {
		players = append(players, player)
	}
	s.players.mu.RUnlock()
	return players
}
