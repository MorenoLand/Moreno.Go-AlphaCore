package world

import (
	"sync"

	"Moreno.AlphaCore/database/realm"
)

type playerRegistry struct {
	mu      sync.RWMutex
	players map[int64]realm.Character
}

func (s *WorldServer) registerPlayer(character realm.Character) {
	s.players.mu.Lock()
	if s.players.players == nil {
		s.players.players = make(map[int64]realm.Character)
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
	s.players.mu.Unlock()
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
