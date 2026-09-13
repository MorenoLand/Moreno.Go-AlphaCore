package world

import (
	"math"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) nearbyGameObjectPackets(player realm.Character) ([][]byte, error) {
	spawns, err := s.WorldData.GameObjectSpawns(player.Map, player.PositionX, player.PositionY, player.PositionZ, creatureViewDistance)
	if err != nil {
		return nil, err
	}
	packets := make([][]byte, 0, len(spawns))
	for _, spawn := range spawns {
		template, found, err := s.WorldData.GameObjectTemplate(spawn.Entry)
		if err != nil {
			return nil, err
		}
		if !found || template.DisplayID == 0 {
			continue
		}
		guid := uint64(spawn.SpawnID) | 0xf110000000000000
		fields := make([]uint32, 20)
		packet.SetUint64(fields, 0, guid)
		fields[2] = 33
		fields[3] = uint32(template.Entry)
		fields[4] = math.Float32bits(template.Scale)
		fields[6] = uint32(template.DisplayID)
		fields[7] = uint32(template.Flags | spawn.Flags)
		fields[8] = math.Float32bits(spawn.Rotation0)
		fields[9] = math.Float32bits(spawn.Rotation1)
		rotation2, rotation3 := spawn.Rotation2, spawn.Rotation3
		if rotation2 == 0 && rotation3 == 0 {
			rotation2 = float32(math.Sin(float64(spawn.Orientation) / 2))
			rotation3 = float32(math.Cos(float64(spawn.Orientation) / 2))
		}
		fields[10] = math.Float32bits(rotation2)
		fields[11] = math.Float32bits(rotation3)
		fields[12] = uint32(spawn.State)
		fields[13] = uint32(spawn.AnimProgress)
		fields[14] = math.Float32bits(spawn.PositionX)
		fields[15] = math.Float32bits(spawn.PositionY)
		fields[16] = math.Float32bits(spawn.PositionZ)
		fields[17] = math.Float32bits(spawn.Orientation)
		fields[19] = uint32(template.Faction)
		createPacket, err := packet.EncodeGameObjectCreate(guid, fields, packet.Movement{X: spawn.PositionX, Y: spawn.PositionY, Z: spawn.PositionZ, O: spawn.Orientation})
		if err != nil {
			return nil, err
		}
		packets = append(packets, createPacket)
	}
	return packets, nil
}
