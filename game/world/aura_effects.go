package world

import (
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) auraEffectChange(target realm.Character, aura *auraState, remove bool) {
	switch packet.AuraType(aura.effect.Aura) {
	case packet.AuraModMounted:
		if remove {
			s.unmountPlayer(target, true, false)
		} else {
			if s.WorldData != nil {
				if template, found, err := s.WorldData.CreatureTemplate(aura.effect.MiscValue); err == nil && found {
					s.mountPlayer(target, uint32(template.DisplayID1), false)
				}
			}
		}
		return
	case packet.AuraModIncreaseMountedSpeed:
		return
	}
	if aura.points == 0 {
		return
	}
	current, found := s.playerByGUID(target.GUID)
	if found {
		target = current
	}
	if packet.AuraType(aura.effect.Aura) == packet.AuraModIncreaseMana {
		if playerPowerType(target.Class) != 0 {
			return
		}
		maxMana := s.playerMaxPower(target.GUID, 0)
		if remove {
			maxMana -= aura.points
		} else {
			maxMana += aura.points
		}
		maxMana = maxInt64(maxMana, 1)
		s.setPlayerMaxPower(target.GUID, 0, maxMana)
		if remove && target.Power1 > maxMana {
			target.Power1 = maxMana
			if s.Characters != nil {
				_ = s.Characters.UpdatePower(target.GUID, target.AccountID, target.RealmID, 0, target.Power1)
			}
			s.updatePlayer(target)
			if update, err := packet.EncodeFieldUpdate(uint64(target.GUID), 23, uint32(target.Power1)); err == nil {
				s.sendSpell(target, update)
			}
		}
		if update, err := packet.EncodeFieldUpdate(uint64(target.GUID), 28, uint32(maxMana)); err == nil {
			s.sendSpell(target, update)
		}
		return
	}
	if packet.AuraType(aura.effect.Aura) != packet.AuraModIncreaseHealth {
		return
	}
	maxHealth := s.playerMaxHealth(target.GUID)
	if remove {
		maxHealth -= aura.points
	} else {
		maxHealth += aura.points
	}
	maxHealth = maxInt64(maxHealth, 1)
	s.setPlayerMaxHealth(target.GUID, maxHealth)
	if remove && target.Health > maxHealth {
		target.Health = maxHealth
		if s.Characters != nil {
			_ = s.Characters.UpdateHealth(target.GUID, target.AccountID, target.RealmID, target.Health)
		}
		s.updatePlayer(target)
		if update, err := packet.EncodeFieldUpdate(uint64(target.GUID), 22, uint32(target.Health)); err == nil {
			s.sendSpell(target, update)
		}
	}
	if update, err := packet.EncodeFieldUpdate(uint64(target.GUID), 27, uint32(maxHealth)); err == nil {
		s.sendSpell(target, update)
	}
}
