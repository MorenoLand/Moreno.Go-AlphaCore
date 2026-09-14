package world

import (
	"sync"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

type resurrectionRegistry struct {
	mu       sync.Mutex
	requests map[int64]resurrectionRequest
}

type resurrectionRequest struct {
	CasterGUID int64
	Recovery   float64
	Map        int64
	X, Y, Z, O float32
}

func (s *WorldServer) requestResurrection(cast *spellCast, target realm.Character, points int64) {
	if target.GUID == 0 {
		return
	}
	s.resurrections.mu.Lock()
	if s.resurrections.requests == nil {
		s.resurrections.requests = make(map[int64]resurrectionRequest)
	}
	s.resurrections.requests[target.GUID] = resurrectionRequest{CasterGUID: cast.caster.GUID, Recovery: float64(points), Map: cast.caster.Map, X: cast.caster.PositionX, Y: cast.caster.PositionY, Z: cast.caster.PositionZ, O: cast.caster.Orientation}
	s.resurrections.mu.Unlock()
	if request, err := packet.Encode(packet.SMSGResurrectRequest, encodeGUID(cast.caster.GUID)); err == nil {
		s.sendPlayer(target.GUID, request)
	}
}
