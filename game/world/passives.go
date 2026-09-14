package world

import (
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) initializePassiveSpells(character realm.Character) error {
	if s.Characters == nil || s.DBC == nil {
		return nil
	}
	spells, err := s.Characters.Spells(character.GUID)
	if err != nil {
		return err
	}
	for _, learned := range spells {
		if !learned.Active {
			continue
		}
		spell, found, err := s.DBC.Spell(learned.ID)
		if err != nil {
			return err
		}
		if !found || packet.SpellAttributes(spell.Attributes)&packet.SpellAttributePassive == 0 {
			continue
		}
		spellLevel, err := s.spellCasterLevel(character, spell)
		if err != nil {
			return err
		}
		effectLevel := spellLevel - spell.BaseLevel
		if effectLevel < 0 {
			effectLevel = 0
		}
		cast := &spellCast{caster: character, spell: spell, spellLevel: spellLevel, effectLevel: effectLevel, ranked: true}
		for index, effect := range spell.Effects {
			switch packet.SpellEffect(effect.Type) {
			case packet.SpellEffectApplyAura:
				s.applyAura(cast, character, index, effect)
			case packet.SpellEffectBlock, packet.SpellEffectParry, packet.SpellEffectDodge, packet.SpellEffectWeapon, packet.SpellEffectDefense, packet.SpellEffectDualWield, packet.SpellEffectProficiency, packet.SpellEffectLanguage:
				if err := s.applyPassiveSkill(character, spell, effect); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *WorldServer) playerAuraFields(values []uint32, guid int64) {
	s.auras.mu.Lock()
	for slot, aura := range s.auras.active[guid] {
		if aura.passive || slot < 0 || slot >= visibleAuraEnd {
			continue
		}
		values[unitAuraField+slot] = uint32(aura.spellID)
	}
	s.auras.mu.Unlock()
	for index, value := range s.auraFlags(guid) {
		values[unitAuraFlagsField+index] = value
	}
}
