package dbc

import (
	"database/sql"
	"fmt"
)

type SkillLineAbility struct {
	ID, SkillLine, Spell, RaceMask, ClassMask, ExcludeRace, ExcludeClass int64
	MinSkillLineRank, SupercededBySpell, TrivialSkillLineRankHigh        int64
	TrivialSkillLineRankLow, Abandonable                                 int64
}

func (s *Store) SpellBaseLevel(id int64) (int64, bool, error) {
	var level int64
	err := s.db.QueryRow(`SELECT BaseLevel FROM Spell WHERE ID = ?`, id).Scan(&level)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("query spell base level: %w", err)
	}
	return level, true, nil
}

func (s *Store) PrecededSpell(id int64) (int64, bool, error) {
	var spell int64
	err := s.db.QueryRow(`SELECT Spell FROM SkillLineAbility WHERE SupercededBySpell = ? ORDER BY ID LIMIT 1`, id).Scan(&spell)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("query preceded spell: %w", err)
	}
	return spell, true, nil
}
