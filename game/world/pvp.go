package world

import (
	"fmt"
	"math/rand"

	"Moreno.AlphaCore/database/realm"
)

type pvpLocation struct {
	Map        int64
	X, Y, Z, O float32
}

func (s *WorldServer) pvpPort(active *realm.Character) ([]byte, error) {
	if active == nil || s.DBC == nil || s.WorldData == nil || s.Characters == nil || s.combatTarget(active.GUID) != 0 {
		return nil, nil
	}
	pvp, err := s.DBC.MapIsPVP(active.Map)
	if err != nil {
		return nil, err
	}
	if pvp {
		if source, found := s.takePVPSource(active.GUID); found {
			return s.teleportPlayer(active, source.Map, source.X, source.Y, source.Z, source.O)
		}
		bind, found, err := s.Characters.Deathbind(active.GUID)
		if err != nil || !found {
			return nil, err
		}
		return s.teleportPlayer(active, bind.Map, bind.X, bind.Y, bind.Z, active.Orientation)
	}
	port, found, err := s.WorldData.WorldportByName(fmt.Sprintf("PvPZone0%d", rand.Intn(2)+1))
	if err != nil || !found {
		return nil, err
	}
	s.setPVPSource(active.GUID, pvpLocation{Map: active.Map, X: active.PositionX, Y: active.PositionY, Z: active.PositionZ, O: active.Orientation})
	return s.teleportPlayer(active, port.Map, port.X, port.Y, port.Z, port.O)
}

func (s *WorldServer) setPVPSource(guid int64, source pvpLocation) {
	s.players.mu.Lock()
	if s.players.pvpSource == nil {
		s.players.pvpSource = make(map[int64]pvpLocation)
	}
	s.players.pvpSource[guid] = source
	s.players.mu.Unlock()
}

func (s *WorldServer) takePVPSource(guid int64) (pvpLocation, bool) {
	s.players.mu.Lock()
	source, found := s.players.pvpSource[guid]
	delete(s.players.pvpSource, guid)
	s.players.mu.Unlock()
	return source, found
}
