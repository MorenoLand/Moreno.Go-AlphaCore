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

type SkillLine struct {
	ID, RaceMask, ClassMask, ExcludeRace, ExcludeClass int64
}

func (s *Store) SpellSkillLine(spell int64, race, class uint8) (int64, bool, error) {
	rows, err := s.db.Query(`SELECT SkillLine, RaceMask, ClassMask, ExcludeRace, ExcludeClass FROM SkillLineAbility WHERE Spell = ? ORDER BY ID`, spell)
	if err != nil {
		return 0, false, fmt.Errorf("query spell skill line: %w", err)
	}
	defer rows.Close()
	var races, classes int64
	if race > 0 && race <= 63 {
		races = int64(1) << (race - 1)
	}
	if class > 0 && class <= 63 {
		classes = int64(1) << (class - 1)
	}
	for rows.Next() {
		var skillLine, raceMask, classMask, excludeRace, excludeClass int64
		if err := rows.Scan(&skillLine, &raceMask, &classMask, &excludeRace, &excludeClass); err != nil {
			return 0, false, fmt.Errorf("scan spell skill line: %w", err)
		}
		if excludeRace != 0 {
			raceMask = ^raceMask
		}
		if excludeClass != 0 {
			classMask = ^classMask
		}
		if raceMask == 0 || raceMask&races != 0 {
			if classMask == 0 || classMask&classes != 0 {
				return skillLine, true, nil
			}
		}
	}
	if err := rows.Err(); err != nil {
		return 0, false, fmt.Errorf("read spell skill line: %w", err)
	}
	return 0, false, nil
}

func (s *Store) SpellAllowedForRaceClass(spell int64, race, class uint8) (bool, error) {
	rows, err := s.db.Query(`SELECT SkillLine, RaceMask, ClassMask, ExcludeRace, ExcludeClass FROM SkillLineAbility WHERE Spell = ? ORDER BY ID`, spell)
	if err != nil {
		return false, fmt.Errorf("query spell skill lines: %w", err)
	}
	defer rows.Close()
	var races, classes int64
	if race > 0 && race <= 63 {
		races = int64(1) << (race - 1)
	}
	if class > 0 && class <= 63 {
		classes = int64(1) << (class - 1)
	}
	found := false
	for rows.Next() {
		var ability SkillLineAbility
		if err := rows.Scan(&ability.SkillLine, &ability.RaceMask, &ability.ClassMask, &ability.ExcludeRace, &ability.ExcludeClass); err != nil {
			return false, fmt.Errorf("scan spell skill line: %w", err)
		}
		found = true
		raceMask, classMask := ability.RaceMask, ability.ClassMask
		if ability.ExcludeRace != 0 {
			raceMask = ^raceMask
		}
		if ability.ExcludeClass != 0 {
			classMask = ^classMask
		}
		if raceMask != 0 && raceMask&int64(races) == 0 || classMask != 0 && classMask&int64(classes) == 0 {
			continue
		}
		var skill SkillLine
		err := s.db.QueryRow(`SELECT ID, RaceMask, ClassMask, ExcludeRace, ExcludeClass FROM SkillLine WHERE ID = ?`, ability.SkillLine).Scan(&skill.ID, &skill.RaceMask, &skill.ClassMask, &skill.ExcludeRace, &skill.ExcludeClass)
		if err == sql.ErrNoRows {
			return true, nil
		}
		if err != nil {
			return false, fmt.Errorf("query skill line: %w", err)
		}
		raceMask, classMask = skill.RaceMask, skill.ClassMask
		if skill.ExcludeRace != 0 {
			raceMask = ^raceMask
		}
		if skill.ExcludeClass != 0 {
			classMask = ^classMask
		}
		if raceMask == 0 || raceMask&int64(races) != 0 {
			if classMask == 0 || classMask&int64(classes) != 0 {
				return true, nil
			}
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("read spell skill lines: %w", err)
	}
	return !found, nil
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
