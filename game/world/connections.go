package world

import (
	"net"
	"sync"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/sockets"
)

type playerConnection struct {
	mu   sync.Mutex
	conn net.Conn
}

func (c *playerConnection) write(data []byte) error {
	c.mu.Lock()
	err := sockets.WriteAll(c.conn, data)
	c.mu.Unlock()
	return err
}

func (s *WorldServer) attachPlayer(guid int64, connection net.Conn) *playerConnection {
	state := &playerConnection{conn: connection}
	s.players.mu.Lock()
	if s.players.connections == nil {
		s.players.connections = make(map[int64]*playerConnection)
	}
	s.players.connections[guid] = state
	s.players.mu.Unlock()
	return state
}

func (s *WorldServer) broadcastPlayer(sender realm.Character, data []byte) {
	s.players.mu.RLock()
	targets := make([]*playerConnection, 0, len(s.players.connections))
	for guid, player := range s.players.players {
		if guid != sender.GUID && player.Map == sender.Map && s.players.connections[guid] != nil {
			targets = append(targets, s.players.connections[guid])
		}
	}
	s.players.mu.RUnlock()
	for _, target := range targets {
		_ = target.write(data)
	}
}

func (s *WorldServer) sendPlayer(guid int64, data []byte) {
	s.players.mu.RLock()
	target := s.players.connections[guid]
	s.players.mu.RUnlock()
	if target != nil {
		_ = target.write(data)
	}
}
