package world

import (
	"math"
	"math/rand"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const creatureViewDistance float32 = 100

func (s *WorldServer) creatureAt(active realm.Character, guid uint64, distance float32) (worlddb.CreatureSpawn, worlddb.CreatureTemplate, bool, error) {
	s.creatures.mu.Lock()
	if state := s.creatures.active[guid]; state != nil {
		spawn, template := state.Spawn, state.Template
		s.creatures.mu.Unlock()
		if spawn.Map != active.Map {
			return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, nil
		}
		dx, dy, dz := spawn.PositionX-active.PositionX, spawn.PositionY-active.PositionY, spawn.PositionZ-active.PositionZ
		return spawn, template, dx*dx+dy*dy+dz*dz <= distance*distance, nil
	}
	s.creatures.mu.Unlock()
	if s.WorldData == nil {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, nil
	}
	spawn, found, err := s.WorldData.CreatureSpawnByID(int64(uint32(guid)))
	if err != nil || !found || spawn.Map != active.Map {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, err
	}
	dx, dy, dz := spawn.PositionX-active.PositionX, spawn.PositionY-active.PositionY, spawn.PositionZ-active.PositionZ
	if dx*dx+dy*dy+dz*dz > distance*distance {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, nil
	}
	creature, found, err := s.WorldData.CreatureTemplate(spawn.Entry)
	if err != nil || !found {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, err
	}
	return spawn, creature, true, nil
}

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
		state := s.newCreatureState(spawn, template, stats, level)
		guid := uint64(spawn.SpawnID) | 0xf130000000000000
		state.GUID = guid
		s.setCreatureState(state)
		createPacket, err := s.creatureCreatePacket(state, scale, boundingRadius, combatReach)
		if err != nil {
			return nil, err
		}
		packets = append(packets, createPacket)
	}
	return packets, nil
}

func (s *WorldServer) newCreatureState(spawn worlddb.CreatureSpawn, template worlddb.CreatureTemplate, stats worlddb.CreatureClassLevelStats, level int64) creatureState {
	health := int64(float64(stats.Health) * float64(template.HealthMultiplier) * float64(spawn.HealthPercent) / 100)
	if health < 1 {
		health = 1
	}
	mana := int64(float64(stats.Mana) * float64(template.ManaMultiplier) * float64(spawn.ManaPercent) / 100)
	return creatureState{Spawn: spawn, Template: template, Stats: stats, Level: level, Health: health, MaxHealth: health, Mana: mana}
}

func (s *WorldServer) creatureCreatePacket(state creatureState, scale, boundingRadius, combatReach float32) ([]byte, error) {
	if scale <= 0 {
		scale = 1
	}
	averageDamage := state.Stats.MeleeDamage * state.Template.DamageMultiplier
	variance := averageDamage * state.Template.DamageVariance
	minimumDamage := int64(averageDamage - variance)
	maximumDamage := int64(math.Round(float64(averageDamage + variance)))
	fields := make([]uint32, packet.UnitFieldCount)
	packet.SetUint64(fields, 0, state.GUID)
	fields[2] = 9
	fields[3] = uint32(state.Template.Entry)
	fields[4] = math.Float32bits(scale)
	fields[22] = uint32(state.Health)
	fields[23] = uint32(state.Mana)
	fields[27] = uint32(state.MaxHealth)
	fields[28] = uint32(state.Mana)
	fields[32] = uint32(state.Level)
	fields[33] = uint32(state.Template.Faction)
	fields[34] = byteValue(0, 0, uint8(state.Template.UnitClass), 0)
	fields[54] = uint32(state.Template.UnitFlags)
	fields[140] = uint32(state.Template.BaseAttackTime)
	fields[141] = uint32(state.Template.BaseAttackTime)
	fields[142] = uint32(int64(float64(state.Stats.Armor) * float64(state.Template.ArmorMultiplier)))
	fields[148] = math.Float32bits(boundingRadius)
	fields[149] = math.Float32bits(combatReach)
	fields[151] = uint32(state.Template.DisplayID1)
	fields[153] = uint32(maximumDamage)<<16 | uint32(minimumDamage)
	fields[172] = byteValue(1, 0, uint8(state.Template.NPCFlags), 0)
	if state.Pet {
		packet.SetUint64(fields, 12, state.OwnerGUID)
		packet.SetUint64(fields, 14, state.OwnerGUID)
		fields[173] = uint32(state.PetID)
		fields[174] = uint32(state.PetNameTimestamp)
		fields[175] = uint32(state.PetExperience)
		fields[176] = uint32(state.PetNextExperience)
		fields[181] = uint32(state.CreatedBySpell)
	}
	return packet.EncodeUnitCreate(state.GUID, fields, packet.Movement{X: state.Spawn.PositionX, Y: state.Spawn.PositionY, Z: state.Spawn.PositionZ, O: state.Spawn.Orientation, WalkSpeed: 2.5, RunSpeed: 7, SwimSpeed: 4.722222, TurnRate: 3.141594})
}

func (s *WorldServer) createCreature(active realm.Character, entry int64) error {
	if s.WorldData == nil || entry <= 0 {
		return nil
	}
	template, found, err := s.WorldData.CreatureTemplate(entry)
	if err != nil || !found || template.DisplayID1 == 0 || template.UnitClass == 0 {
		return err
	}
	level := template.LevelMin
	if level < 1 {
		level = 1
	}
	spawn := worlddb.CreatureSpawn{Entry: entry, Map: active.Map, PositionX: active.PositionX, PositionY: active.PositionY, PositionZ: active.PositionZ, Orientation: active.Orientation, HealthPercent: 100, ManaPercent: 100}
	stats, found, err := s.WorldData.CreatureClassLevelStats(template.UnitClass, level)
	if err != nil || !found {
		return err
	}
	state := s.newCreatureState(spawn, template, stats, level)
	state.GUID = s.nextCreatureGUID()
	s.setCreatureState(state)
	scale, boundingRadius, combatReach := template.Scale, float32(0), float32(0)
	if info, infoFound, infoErr := s.WorldData.CreatureModelInfo(template.DisplayID1); infoErr != nil {
		s.removeCreature(state.GUID)
		return infoErr
	} else if infoFound {
		boundingRadius, combatReach = info.BoundingRadius, info.CombatReach
	}
	if s.DBC != nil {
		if display, displayFound, displayErr := s.DBC.CreatureDisplayInfo(template.DisplayID1); displayErr != nil {
			s.removeCreature(state.GUID)
			return displayErr
		} else if displayFound && template.Scale <= 0 && display.ModelScale > 0 {
			scale = display.ModelScale
		}
	}
	create, err := s.creatureCreatePacket(state, scale, boundingRadius, combatReach)
	if err != nil {
		s.removeCreature(state.GUID)
		return err
	}
	s.broadcastPlayer(active, create)
	s.sendPlayer(active.GUID, create)
	return nil
}

func (s *WorldServer) destroyCreature(active realm.Character, guid uint64) error {
	state, found := s.removeCreature(guid)
	if !found || state.Spawn.Map != active.Map {
		return nil
	}
	destroy, err := packet.Encode(packet.SMSGDestroyObject, encodeGUID(int64(state.GUID)))
	if err != nil {
		return err
	}
	s.broadcastPlayer(active, destroy)
	s.sendPlayer(active.GUID, destroy)
	return nil
}
