package world

import (
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const (
	passiveBlock uint32 = 1 << iota
	passiveParry
	passiveDodge
	passiveDefense
	passiveDualWield
)

func (s *WorldServer) applyPassiveSkill(target realm.Character, spell dbc.Spell, effect dbc.SpellEffect) error {
	if s.Characters == nil || s.DBC == nil {
		return nil
	}
	skill, found, err := s.DBC.SpellSkillLine(spell.ID, target.Race, target.Class)
	if err != nil || !found {
		return err
	}
	value, found, err := s.Characters.SkillValue(target.GUID, skill)
	if err != nil {
		return err
	}
	maximum := int64(1)
	if !found {
		value = 1
		if err := s.Characters.AddSkill(target.GUID, skill, value, maximum); err != nil {
			return err
		}
	} else if data, skillFound, skillErr := s.skillData(target.GUID, skill); skillErr != nil {
		return skillErr
	} else if skillFound {
		maximum = data.Max
	}
	s.setPassiveSkill(target.GUID, effect.Type)
	s.sendSkillFields(target, skill, value, maximum)
	return nil
}

func (s *WorldServer) skillStep(target realm.Character, effect dbc.SpellEffect, points int64) {
	if points < 0 || s.Characters == nil {
		return
	}
	skill := effect.MiscValue
	if skill <= 0 {
		return
	}
	value, found, err := s.Characters.SkillValue(target.GUID, skill)
	if err != nil {
		return
	}
	maximum := points * 5
	if !found {
		value = 1
		if maximum <= 0 {
			maximum = 1
		}
		if err := s.Characters.AddSkill(target.GUID, skill, value, maximum); err != nil {
			return
		}
	} else {
		if value < 1 {
			value = 1
		}
		if maximum <= 0 || maximum >= 1000 {
			skillData, skillFound, skillErr := s.skillData(target.GUID, skill)
			if skillErr != nil || !skillFound {
				return
			}
			maximum = skillData.Max
		}
		if err := s.Characters.UpdateSkill(target.GUID, skill, value, maximum); err != nil {
			return
		}
	}
	s.sendSkillFields(target, skill, value, maximum)
}

func (s *WorldServer) skillData(guid, skill int64) (realm.Skill, bool, error) {
	skills, err := s.Characters.Skills(guid)
	if err != nil {
		return realm.Skill{}, false, err
	}
	for _, value := range skills {
		if value.ID == skill {
			return value, true, nil
		}
	}
	return realm.Skill{}, false, nil
}

func (s *WorldServer) setPassiveSkill(guid int64, effect int64) {
	var flag uint32
	switch packet.SpellEffect(effect) {
	case packet.SpellEffectBlock:
		flag = passiveBlock
	case packet.SpellEffectParry:
		flag = passiveParry
	case packet.SpellEffectDodge:
		flag = passiveDodge
	case packet.SpellEffectDefense:
		flag = passiveDefense
	case packet.SpellEffectDualWield:
		flag = passiveDualWield
	}
	if flag == 0 {
		return
	}
	s.players.mu.Lock()
	if s.players.passiveSkills == nil {
		s.players.passiveSkills = make(map[int64]uint32)
	}
	s.players.passiveSkills[guid] |= flag
	s.players.mu.Unlock()
}

func (s *WorldServer) sendSkillFields(target realm.Character, skill, value, maximum int64) {
	if s.Characters == nil {
		return
	}
	skills, err := s.Characters.Skills(target.GUID)
	if err != nil {
		return
	}
	for index, current := range skills {
		if current.ID != skill || index >= 64 {
			continue
		}
		base := 334 + index*3
		for _, update := range []struct {
			field int
			value uint32
		}{{base, uint32(value&0xffff)<<16 | uint32(skill&0xffff)}, {base + 1, uint32(maximum&0xffff) << 16}, {base + 2, 0}} {
			if fieldUpdate, err := packet.EncodeFieldUpdate(uint64(target.GUID), update.field, update.value); err == nil {
				s.sendSpell(target, fieldUpdate)
			}
		}
		return
	}
}
