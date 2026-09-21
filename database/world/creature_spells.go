package world

import (
	"database/sql"
	"fmt"
)

type CreatureSpellEntry struct {
	SpellID, Probability, CastTarget, TargetParam1, TargetParam2, CastFlags    int64
	DelayInitialMin, DelayInitialMax, DelayRepeatMin, DelayRepeatMax, ScriptID int64
}

type CreatureSpell struct {
	Entry  int64
	Name   string
	Spells [8]CreatureSpellEntry
}

func (s *Store) CreatureSpell(entry int64) (CreatureSpell, bool, error) {
	columns := "entry, name"
	for index := 1; index <= 8; index++ {
		columns += fmt.Sprintf(", spellId_%d, probability_%d, castTarget_%d, targetParam1_%d, targetParam2_%d, castFlags_%d, delayInitialMin_%d, delayInitialMax_%d, delayRepeatMin_%d, delayRepeatMax_%d, scriptId_%d", index, index, index, index, index, index, index, index, index, index, index)
	}
	var spell CreatureSpell
	values := []any{&spell.Entry, &spell.Name}
	for index := range spell.Spells {
		values = append(values, &spell.Spells[index].SpellID, &spell.Spells[index].Probability, &spell.Spells[index].CastTarget, &spell.Spells[index].TargetParam1, &spell.Spells[index].TargetParam2, &spell.Spells[index].CastFlags, &spell.Spells[index].DelayInitialMin, &spell.Spells[index].DelayInitialMax, &spell.Spells[index].DelayRepeatMin, &spell.Spells[index].DelayRepeatMax, &spell.Spells[index].ScriptID)
	}
	err := s.db.QueryRow(`SELECT `+columns+` FROM creature_spells WHERE entry = ?`, entry).Scan(values...)
	if err == sql.ErrNoRows {
		return CreatureSpell{}, false, nil
	}
	if err != nil {
		return CreatureSpell{}, false, fmt.Errorf("query creature spells: %w", err)
	}
	return spell, true, nil
}
