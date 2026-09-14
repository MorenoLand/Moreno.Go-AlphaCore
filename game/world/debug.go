package world

import (
	"encoding/binary"
	"fmt"
	"math"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) debugAIState(active realm.Character, data []byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	high := guid & 0xffff000000000000
	var messages []string
	switch high {
	case 0xf110000000000000:
		spawn, template, found, err := s.gameObjectAt(active, guid, gameObjectViewDistance)
		if err != nil || !found {
			return nil, err
		}
		distance := distance3(active.PositionX, active.PositionY, active.PositionZ, spawn.PositionX, spawn.PositionY, spawn.PositionZ)
		messages = []string{fmt.Sprintf("Spawn ID %d, Guid: %d, Entry: %d, Display ID: %d", spawn.SpawnID, uint32(guid), template.Entry, template.DisplayID), fmt.Sprintf("X: %.3f, Y: %.3f, Z: %.3f, O: %.3f", spawn.PositionX, spawn.PositionY, spawn.PositionZ, spawn.Orientation), fmt.Sprintf("Distance: %.3f yd", distance)}
	case 0xf130000000000000:
		state, found, err := s.creatureStateAt(active, guid, creatureViewDistance)
		if err != nil || !found {
			return nil, err
		}
		distance := distance3(active.PositionX, active.PositionY, active.PositionZ, state.Spawn.PositionX, state.Spawn.PositionY, state.Spawn.PositionZ)
		messages = []string{fmt.Sprintf("Spawn ID %d, Guid: %d, Entry: %d, Display ID: %d", state.Spawn.SpawnID, uint32(guid), state.Template.Entry, state.Template.DisplayID1), fmt.Sprintf("X: %.3f, Y: %.3f, Z: %.3f, O: %.3f", state.Spawn.PositionX, state.Spawn.PositionY, state.Spawn.PositionZ, state.Spawn.Orientation), fmt.Sprintf("Distance: %.3f yd", distance)}
	default:
		player, found := s.playerByGUID(int64(guid))
		if !found || player.Map != active.Map {
			return nil, nil
		}
		distance := distance3(active.PositionX, active.PositionY, active.PositionZ, player.PositionX, player.PositionY, player.PositionZ)
		if distance > creatureViewDistance {
			return nil, nil
		}
		displayID := int64(0)
		if s.DBC != nil {
			if race, found, err := s.DBC.Race(player.Race); err == nil && found {
				displayID = playerDisplayID(player, race)
			}
		}
		messages = []string{fmt.Sprintf("Guid: %d, Entry: 0, Display ID: %d", uint32(guid), displayID), fmt.Sprintf("X: %.3f, Y: %.3f, Z: %.3f, O: %.3f", player.PositionX, player.PositionY, player.PositionZ, player.Orientation), fmt.Sprintf("Distance: %.3f yd", distance)}
	}
	return debugAIStatePacket(guid, messages)
}

func debugAIStatePacket(guid uint64, messages []string) ([]byte, error) {
	body := append(encodeUint64(guid), encodeUint32(int64(len(messages)))...)
	for _, message := range messages {
		if len(message) > 127 {
			message = message[:127]
		}
		value, err := packet.StringBytes(message)
		if err != nil {
			return nil, err
		}
		body = append(body, value...)
	}
	return packet.Encode(packet.SMSGDebugAIState, body)
}

func distance3(x1, y1, z1, x2, y2, z2 float32) float32 {
	dx, dy, dz := x1-x2, y1-y2, z1-z2
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}
