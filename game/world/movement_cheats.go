package world

import (
	"encoding/binary"
	"math"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func defaultPlayerSpeeds() [4]float32 { return [4]float32{2.5, 7, 4.722222, 3.141594} }

func (s *WorldServer) movementSpeedCheat(active *realm.Character, opcode packet.Opcode, data []byte, gmLevel int) error {
	if active == nil || gmLevel <= 0 || len(data) < 52 {
		return nil
	}
	if opcode == packet.MSGMoveSetAllSpeedCheat {
		return nil
	}
	speed := math.Float32frombits(binary.LittleEndian.Uint32(data[48:]))
	speeds := s.playerSpeeds(active.GUID)
	speedType, responseOpcode := 1, packet.SMSGForceSpeedChange
	switch opcode {
	case packet.MSGMoveSetSwimSpeedCheat:
		speedType, responseOpcode = 2, packet.SMSGForceSwimSpeedChange
	case packet.MSGMoveSetWalkSpeedCheat:
		speedType, responseOpcode = 0, packet.MSGMoveSetWalkSpeed
	case packet.MSGMoveSetTurnRateCheat:
		speedType, responseOpcode = 3, packet.MSGMoveSetTurnRate
	}
	if speed <= 0 {
		speed = defaultPlayerSpeeds()[speedType]
	}
	speeds[speedType] = speed
	s.players.mu.Lock()
	if s.players.speeds == nil {
		s.players.speeds = make(map[int64][4]float32)
	}
	s.players.speeds[active.GUID] = speeds
	s.players.mu.Unlock()
	response, err := packet.Encode(responseOpcode, encodeFloat(speed))
	if err != nil {
		return err
	}
	s.sendPlayer(active.GUID, response)
	return nil
}
