package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) attack(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 {
		return nil, nil
	}
	targetGUID := binary.LittleEndian.Uint64(data)
	target, found := s.playerByGUID(int64(targetGUID))
	if !found || target.Health <= 0 || target.Map != active.Map {
		if s.WorldData == nil {
			return s.attackStop(active)
		}
		if _, _, creatureFound, err := s.creatureAt(active, targetGUID, creatureViewDistance); err != nil {
			return nil, err
		} else if !creatureFound {
			return s.attackStop(active)
		}
	}
	if s.isSanctuary(int64(targetGUID)) {
		return s.attackStop(active)
	}
	s.setCombatTarget(active.GUID, targetGUID)
	body := append(encodeGUID(active.GUID), encodeGUID(int64(targetGUID))...)
	start, err := packet.Encode(packet.SMSGAttackStart, body)
	if err != nil {
		return nil, err
	}
	s.broadcastPlayer(active, start)
	return [][]byte{start}, nil
}

func (s *WorldServer) attackStop(active realm.Character) ([][]byte, error) {
	target := s.combatTarget(active.GUID)
	if target == 0 && active.Health > 0 {
		target = uint64(active.GUID)
	}
	body := append(encodeGUID(active.GUID), encodeGUID(int64(target))...)
	if active.Health <= 0 {
		body = append(body, 1, 0, 0, 0)
	} else {
		body = append(body, 0, 0, 0, 0)
	}
	s.setCombatTarget(active.GUID, 0)
	stop, err := packet.Encode(packet.SMSGAttackStop, body)
	if err != nil {
		return nil, err
	}
	s.broadcastPlayer(active, stop)
	cancel, err := packet.Encode(packet.SMSGCancelCombat, nil)
	if err != nil {
		return nil, err
	}
	return [][]byte{stop, cancel}, nil
}
