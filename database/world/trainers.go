package world

import (
	"database/sql"
	"fmt"
)

type TrainerSpell struct {
	TemplateEntry, Spell, PlayerSpell, SpellCost, TalentPointCost, SkillPointCost int64
	ReqSkill, ReqSkillValue, ReqLevel, ReqSpell1, ReqSpell2, ReqSpell3            int64
}

type TrainerGreeting struct {
	Entry   int64
	Content string
}

func (s *Store) TrainerSpells(template int64) ([]TrainerSpell, error) {
	rows, err := s.db.Query(`SELECT template_entry, spell, playerspell, spellcost, talentpointcost, skillpointcost, reqskill, reqskillvalue, reqlevel, req_spell_1, req_spell_2, req_spell_3 FROM trainer_template WHERE template_entry = ? ORDER BY spell`, template)
	if err != nil {
		return nil, fmt.Errorf("query trainer spells: %w", err)
	}
	defer rows.Close()
	spells := make([]TrainerSpell, 0)
	for rows.Next() {
		var spell TrainerSpell
		if err := rows.Scan(&spell.TemplateEntry, &spell.Spell, &spell.PlayerSpell, &spell.SpellCost, &spell.TalentPointCost, &spell.SkillPointCost, &spell.ReqSkill, &spell.ReqSkillValue, &spell.ReqLevel, &spell.ReqSpell1, &spell.ReqSpell2, &spell.ReqSpell3); err != nil {
			return nil, fmt.Errorf("scan trainer spell: %w", err)
		}
		spells = append(spells, spell)
	}
	return spells, rows.Err()
}

func (s *Store) TrainerGreeting(entry int64) (TrainerGreeting, bool, error) {
	var greeting TrainerGreeting
	err := s.db.QueryRow(`SELECT entry, content_default FROM npc_trainer_greeting WHERE entry = ?`, entry).Scan(&greeting.Entry, &greeting.Content)
	if err == sql.ErrNoRows {
		return TrainerGreeting{}, false, nil
	}
	if err != nil {
		return TrainerGreeting{}, false, fmt.Errorf("query trainer greeting: %w", err)
	}
	return greeting, true, nil
}
