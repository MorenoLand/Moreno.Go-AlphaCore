package world

import (
	"encoding/binary"
	"math"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) updateMovement(active *realm.Character, opcode packet.Opcode, data []byte) error {
	if active == nil || len(data) < 48 {
		return nil
	}
	x := math.Float32frombits(binary.LittleEndian.Uint32(data[24:28]))
	y := math.Float32frombits(binary.LittleEndian.Uint32(data[28:32]))
	z := math.Float32frombits(binary.LittleEndian.Uint32(data[32:36]))
	o := math.Float32frombits(binary.LittleEndian.Uint32(data[36:40]))
	dx, dy, dz := float64(active.PositionX-x), float64(active.PositionY-y), float64(active.PositionZ-z)
	if active.TaxiPath == "" && dx*dx+dy*dy+dz*dz > 4096 {
		return nil
	}
	active.PositionX, active.PositionY, active.PositionZ, active.Orientation = x, y, z, o
	if s.Characters != nil {
		if err := s.Characters.UpdatePosition(active.GUID, active.AccountID, active.RealmID, x, y, z, o); err != nil {
			return err
		}
	}
	if active.TaxiPath != "" && s.taxiAtDestination(*active, x, y, z) {
		active.TaxiPath = ""
		if s.Characters != nil {
			if err := s.Characters.UpdateTaxiPath(active.GUID, active.AccountID, active.RealmID, ""); err != nil {
				return err
			}
		}
	}
	s.updatePlayer(*active)
	if movementForceAck(opcode) {
		return nil
	}
	payload := append(encodeGUID(active.GUID), data...)
	if opcode == packet.MSGMoveCollideRedirect || opcode == packet.MSGMoveCollideStuck {
		flags := binary.LittleEndian.Uint32(payload[52:56])
		binary.LittleEndian.PutUint32(payload[52:56], flags|0x1000)
	}
	update, err := packet.Encode(opcode, payload)
	if err != nil {
		return err
	}
	s.broadcastPlayer(*active, update)
	return nil
}

func movementForceAck(opcode packet.Opcode) bool {
	switch opcode {
	case packet.CMSGForceMoveRootAck, packet.CMSGForceMoveUnrootAck, packet.CMSGForceSpeedChangeAck, packet.CMSGForceSwimSpeedChangeAck:
		return true
	default:
		return false
	}
}
