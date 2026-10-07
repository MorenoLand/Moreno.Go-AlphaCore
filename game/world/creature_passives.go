package world

import (
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) initializeCreaturePassiveSpells(state *creatureState) {
	if state == nil || s.DBC == nil || len(state.Template.SpellIDs) == 0 {
		return
	}
	s.creatures.mu.Lock()
	current := s.creatures.active[state.GUID]
	if current == nil || current.PassiveSpellsInitialized {
		s.creatures.mu.Unlock()
		return
	}
	current.PassiveSpellsInitialized = true
	snapshot := *current
	s.creatures.mu.Unlock()
	target := realm.Character{GUID: int64(snapshot.GUID), Map: snapshot.Spawn.Map, PositionX: snapshot.Spawn.PositionX, PositionY: snapshot.Spawn.PositionY, PositionZ: snapshot.Spawn.PositionZ, Orientation: snapshot.Spawn.Orientation, Level: uint8(snapshot.Level), Health: snapshot.Health}
	for _, spellID := range snapshot.Template.SpellIDs {
		if spellID <= 0 {
			continue
		}
		spell, found, err := s.DBC.Spell(spellID)
		if err != nil || !found || packet.SpellAttributes(spell.Attributes)&packet.SpellAttributePassive == 0 {
			continue
		}
		effectLevel := snapshot.Level - spell.BaseLevel
		if effectLevel < 0 {
			effectLevel = 0
		}
		cast := &spellCast{caster: target, target: spellTarget{UnitGUID: snapshot.GUID}, spell: spell, targetMask: packet.SpellTargetSelf, effectLevel: effectLevel}
		for index, effect := range spell.Effects {
			switch packet.SpellEffect(effect.Type) {
			case packet.SpellEffectApplyAura, packet.SpellEffectApplyAreaAura, packet.SpellEffectPersistentAreaAura:
				s.applyAura(cast, target, index, effect)
			}
		}
	}
}
