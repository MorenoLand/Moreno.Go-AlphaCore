package world

import (
	"math"
	"time"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) applySpellCreatureEffects(cast *spellCast) {
	for _, effect := range cast.spell.Effects {
		kind := packet.SpellEffect(effect.Type)
		switch kind {
		case packet.SpellEffectSummonWild, packet.SpellEffectSummonGuardian, packet.SpellEffectSummonPossessed, packet.SpellEffectSummonTotem:
			s.summonSpellCreatures(cast, effect, kind)
		}
	}
}

func (s *WorldServer) summonSpellCreatures(cast *spellCast, effect dbc.SpellEffect, kind packet.SpellEffect) {
	if cast == nil || s.WorldData == nil || effect.MiscValue <= 0 {
		return
	}
	template, found, err := s.WorldData.CreatureTemplate(effect.MiscValue)
	if err != nil || !found {
		return
	}
	level := int64(cast.caster.Level)
	if level < 1 {
		level = template.LevelMin
	}
	stats, found, err := s.WorldData.CreatureClassLevelStats(template.UnitClass, level)
	if err != nil || !found {
		return
	}
	if s.DBC != nil {
		if race, raceFound, raceErr := s.DBC.Race(cast.caster.Race); raceErr == nil && raceFound {
			template.Faction = race.FactionID
		}
	}
	amount := spellEffectPoints(effect, cast.effectLevel)
	if amount < 1 {
		amount = 1
	}
	if kind != packet.SpellEffectSummonGuardian && amount > 1 {
		amount = 1
	}
	duration := s.spellDurationMillis(cast)
	if duration == 0 {
		switch kind {
		case packet.SpellEffectSummonWild:
			duration = 120000
		case packet.SpellEffectSummonTotem:
			duration = 300000
		}
	}
	for index := int64(0); index < amount; index++ {
		spawn := summonCreatureSpawn(cast, effect, index)
		state := s.newCreatureState(spawn, template, stats, level)
		state.GUID = s.nextCreatureGUID()
		state.OwnerGUID = uint64(cast.caster.GUID)
		state.CreatedBySpell = cast.spell.ID
		s.setCreatureState(state)
		if duration > 0 {
			state.Timer = time.AfterFunc(time.Duration(duration)*time.Millisecond, func() { s.despawnCreature(state.GUID) })
			s.creatures.mu.Lock()
			if current := s.creatures.active[state.GUID]; current != nil {
				current.Timer = state.Timer
			}
			s.creatures.mu.Unlock()
		}
		if create, err := s.creatureCreatePacket(state, template.Scale, 0, 0); err == nil {
			s.sendSpell(cast.caster, create)
		}
	}
}

func summonCreatureSpawn(cast *spellCast, effect dbc.SpellEffect, index int64) worlddb.CreatureSpawn {
	spawn := worlddb.CreatureSpawn{Entry: effect.MiscValue, Map: cast.caster.Map, PositionX: cast.caster.PositionX, PositionY: cast.caster.PositionY, PositionZ: cast.caster.PositionZ, Orientation: cast.caster.Orientation, HealthPercent: 100, ManaPercent: 100}
	if cast.target.Dest != nil {
		spawn.PositionX, spawn.PositionY, spawn.PositionZ = cast.target.Dest.X, cast.target.Dest.Y, cast.target.Dest.Z
	} else if cast.target.Source != nil {
		spawn.PositionX, spawn.PositionY, spawn.PositionZ = cast.target.Source.X, cast.target.Source.Y, cast.target.Source.Z
	}
	if index > 0 {
		angle := float64(cast.caster.Orientation) + float64(index)*math.Pi/2
		spawn.PositionX += float32(math.Cos(angle) * 2)
		spawn.PositionY += float32(math.Sin(angle) * 2)
	}
	return spawn
}

func (s *WorldServer) despawnCreature(guid uint64) {
	state, found := s.removeCreature(guid)
	if !found {
		return
	}
	viewer := realm.Character{GUID: int64(guid), Map: state.Spawn.Map, PositionX: state.Spawn.PositionX, PositionY: state.Spawn.PositionY, PositionZ: state.Spawn.PositionZ}
	if destroy, err := packet.Encode(packet.SMSGDestroyObject, encodeGUID(int64(guid))); err == nil {
		s.broadcastPlayer(viewer, destroy)
	}
}
