package world

import (
	"math/rand"
	"time"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) startCreatureSpells(state *creatureState) {
	if state == nil || s.WorldData == nil || s.DBC == nil || state.Template.SpellListID <= 0 {
		return
	}
	list, found, err := s.WorldData.CreatureSpell(state.Template.SpellListID)
	if err != nil || !found {
		return
	}
	for index, entry := range list.Spells {
		if entry.SpellID <= 0 || entry.Probability <= 0 {
			continue
		}
		spell, spellFound, spellErr := s.DBC.Spell(entry.SpellID)
		if spellErr != nil || !spellFound {
			continue
		}
		delay := creatureSpellDelay(entry.DelayInitialMin, entry.DelayInitialMax)
		spellEntry, spellValue := entry, spell
		timer := time.AfterFunc(time.Duration(delay)*time.Second, func() { s.castCreatureSpell(state.GUID, index, spellEntry, spellValue) })
		s.creatures.mu.Lock()
		if current := s.creatures.active[state.GUID]; current != nil {
			current.SpellTimers[index] = timer
		}
		s.creatures.mu.Unlock()
	}
}

func creatureSpellDelay(minimum, maximum int64) int64 {
	if maximum <= minimum {
		return minimum
	}
	return minimum + int64(rand.Intn(int(maximum-minimum+1)))
}

func (s *WorldServer) castCreatureSpell(guid uint64, index int, entry worlddb.CreatureSpellEntry, spell dbc.Spell) {
	s.creatures.mu.Lock()
	state := s.creatures.active[guid]
	if state == nil || state.Health <= 0 || state.CombatTarget == 0 {
		s.creatures.mu.Unlock()
		return
	}
	targetGUID := state.CombatTarget
	level, mapID := state.Level, state.Spawn.Map
	x, y, z, o := state.Spawn.PositionX, state.Spawn.PositionY, state.Spawn.PositionZ, state.Spawn.Orientation
	s.creatures.mu.Unlock()
	if entry.Probability < 100 && rand.Intn(100) >= int(entry.Probability) {
		s.scheduleCreatureSpell(guid, index, entry, spell)
		return
	}
	if entry.CastTarget == 6 {
		targetGUID = guid
	}
	caster := realm.Character{GUID: int64(guid), Level: uint8(level), Map: mapID, PositionX: x, PositionY: y, PositionZ: z, Orientation: o, Health: 1}
	target := spellTarget{UnitGUID: targetGUID}
	var targetCreature *creatureState
	if targetGUID == guid {
		s.creatures.mu.Lock()
		targetCreature = s.creatures.active[guid]
		s.creatures.mu.Unlock()
	}
	effectLevel := level - spell.BaseLevel
	if effectLevel < 0 {
		effectLevel = 0
	}
	cast := &spellCast{caster: caster, target: target, spell: spell, targetMask: packet.SpellTargetUnit, targets: s.spellTargets(caster, spell, target), effectTargets: s.spellEffectTargetsAll(caster, spell, target), targetCreature: targetCreature, started: time.Now(), spellLevel: level, effectLevel: effectLevel}
	s.performSpellCast(cast)
	s.scheduleCreatureSpell(guid, index, entry, spell)
}

func (s *WorldServer) scheduleCreatureSpell(guid uint64, index int, entry worlddb.CreatureSpellEntry, spell dbc.Spell) {
	delay := creatureSpellDelay(entry.DelayRepeatMin, entry.DelayRepeatMax)
	if entry.DelayRepeatMin == 0 && entry.DelayRepeatMax == 0 {
		return
	}
	timer := time.AfterFunc(time.Duration(delay)*time.Second, func() { s.castCreatureSpell(guid, index, entry, spell) })
	s.creatures.mu.Lock()
	if current := s.creatures.active[guid]; current != nil {
		if current.SpellTimers[index] != nil {
			current.SpellTimers[index].Stop()
		}
		current.SpellTimers[index] = timer
	}
	s.creatures.mu.Unlock()
}
