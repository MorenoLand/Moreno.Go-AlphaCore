package world

import (
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) applySpellObjectEffects(cast *spellCast) {
	for _, effect := range cast.spell.Effects {
		switch packet.SpellEffect(effect.Type) {
		case packet.SpellEffectSummonObject, packet.SpellEffectSummonObjectWild, packet.SpellEffectCreateHouse:
			s.summonGameObject(cast, effect, packet.SpellEffect(effect.Type) == packet.SpellEffectSummonObjectWild)
		case packet.SpellEffectActivateObject:
			s.activateGameObject(cast)
		}
	}
	if cast.target.GameObjectGUID != 0 {
		for _, effect := range cast.spell.Effects {
			if !spellOpenLockEffect(effect) {
				continue
			}
			responses, err := s.gameObjectUse(cast.caster, encodeGUID(int64(cast.target.GameObjectGUID)))
			if err == nil {
				for _, response := range responses {
					s.sendPlayer(cast.caster.GUID, response)
				}
			}
		}
	}
	if cast.target.ItemGUID != 0 && s.Characters != nil && s.WorldData != nil {
		for _, effect := range cast.spell.Effects {
			if !spellOpenLockEffect(effect) {
				continue
			}
			item, found, err := s.Characters.ItemByGUID(cast.caster.GUID, int64(cast.target.ItemGUID))
			if err != nil || !found {
				continue
			}
			template, found, err := s.WorldData.ItemTemplate(item.ItemTemplate)
			if err != nil || !found {
				continue
			}
			item.Flags |= itemDynUnlocked
			if err := s.Characters.UpdateItemFlags(item.GUID, item.Owner, item.Flags); err != nil {
				continue
			}
			if update, err := packet.EncodeFieldUpdate(uint64(item.GUID)|0x4000000000000000, itemFieldFlag, encodedItemFlags(template, item.Flags)); err == nil {
				s.sendPlayer(cast.caster.GUID, update)
			}
		}
	}
}

func spellOpenLockEffect(effect dbc.SpellEffect) bool {
	return packet.SpellEffect(effect.Type) == packet.SpellEffectOpenLock || packet.SpellEffect(effect.Type) == packet.SpellEffectOpenLockItem
}
