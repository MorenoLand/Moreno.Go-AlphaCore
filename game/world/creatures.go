package world

import (
	"math"
	"math/rand"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const creatureViewDistance float32 = 100

func (s *WorldServer) nearbyCreaturePackets(player realm.Character) ([][]byte, error) {
	spawns, err := s.WorldData.CreatureSpawns(player.Map, player.PositionX, player.PositionY, player.PositionZ, creatureViewDistance)
	if err != nil {
		return nil, err
	}
	packets := make([][]byte, 0, len(spawns))
	for _, spawn := range spawns {
		template, found, err := s.WorldData.CreatureTemplate(spawn.Entry)
		if err != nil {
			return nil, err
		}
		if !found || template.DisplayID1 == 0 || template.UnitClass == 0 {
			continue
		}
		level := template.LevelMin
		if template.LevelMax > level {
			level += int64(rand.Intn(int(template.LevelMax - level + 1)))
		}
		stats, found, err := s.WorldData.CreatureClassLevelStats(template.UnitClass, level)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		scale, boundingRadius, combatReach := template.Scale, float32(0), float32(0)
		if scale <= 0 {
			scale = 1
		}
		if info, infoFound, infoErr := s.WorldData.CreatureModelInfo(template.DisplayID1); infoErr != nil {
			return nil, infoErr
		} else if infoFound {
			boundingRadius, combatReach = info.BoundingRadius, info.CombatReach
		}
		if s.DBC != nil {
			if display, displayFound, displayErr := s.DBC.CreatureDisplayInfo(template.DisplayID1); displayErr != nil {
				return nil, displayErr
			} else if displayFound && template.Scale <= 0 && display.ModelScale > 0 {
				scale = display.ModelScale
			}
		}
		health := int64(float64(stats.Health) * float64(template.HealthMultiplier) * float64(spawn.HealthPercent) / 100)
		if health < 1 {
			health = 1
		}
		mana := int64(float64(stats.Mana) * float64(template.ManaMultiplier) * float64(spawn.ManaPercent) / 100)
		averageDamage := stats.MeleeDamage * template.DamageMultiplier
		variance := averageDamage * template.DamageVariance
		minimumDamage := int64(averageDamage - variance)
		maximumDamage := int64(math.Round(float64(averageDamage + variance)))
		guid := uint64(spawn.SpawnID) | 0xf130000000000000
		fields := make([]uint32, packet.UnitFieldCount)
		packet.SetUint64(fields, 0, guid)
		fields[2] = 9
		fields[3] = uint32(template.Entry)
		fields[4] = math.Float32bits(scale)
		fields[22] = uint32(health)
		fields[23] = uint32(mana)
		fields[27] = uint32(health)
		fields[28] = uint32(mana)
		fields[32] = uint32(level)
		fields[33] = uint32(template.Faction)
		fields[34] = byteValue(0, 0, uint8(template.UnitClass), 0)
		fields[54] = uint32(template.UnitFlags)
		fields[140] = uint32(template.BaseAttackTime)
		fields[141] = uint32(template.BaseAttackTime)
		fields[142] = uint32(int64(float64(stats.Armor) * float64(template.ArmorMultiplier)))
		fields[148] = math.Float32bits(boundingRadius)
		fields[149] = math.Float32bits(combatReach)
		fields[151] = uint32(template.DisplayID1)
		fields[153] = uint32(maximumDamage)<<16 | uint32(minimumDamage)
		fields[172] = byteValue(1, 0, uint8(template.NPCFlags), 0)
		createPacket, err := packet.EncodeUnitCreate(guid, fields, packet.Movement{X: spawn.PositionX, Y: spawn.PositionY, Z: spawn.PositionZ, O: spawn.Orientation, WalkSpeed: 2.5, RunSpeed: 7, SwimSpeed: 4.722222, TurnRate: 3.141594})
		if err != nil {
			return nil, err
		}
		packets = append(packets, createPacket)
	}
	return packets, nil
}
