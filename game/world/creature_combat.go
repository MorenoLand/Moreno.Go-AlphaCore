package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) makeMonsterAttackMe(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) != 8 || s.WorldData == nil {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	state, found, err := s.creatureStateAt(active, guid, creatureViewDistance)
	if err != nil || !found || state.Health <= 0 || !s.creatureHostile(active, state) {
		return nil, err
	}
	s.creatures.mu.Lock()
	if current := s.creatures.active[state.GUID]; current != nil {
		current.CombatTarget = uint64(active.GUID)
	}
	s.creatures.mu.Unlock()
	body := append(encodeGUID(int64(state.GUID)), encodeGUID(active.GUID)...)
	start, err := packet.Encode(packet.SMSGAttackStart, body)
	if err != nil {
		return nil, err
	}
	s.broadcastPlayer(active, start)
	return [][]byte{start}, nil
}

func (s *WorldServer) creatureHostile(active realm.Character, state *creatureState) bool {
	if state.Template.Faction == 0 || s.DBC == nil {
		return true
	}
	race, found, err := s.DBC.Race(active.Race)
	return err == nil && (!found || race.FactionID != state.Template.Faction)
}
