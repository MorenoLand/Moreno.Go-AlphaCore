package realm

import (
	"database/sql"
	"fmt"
)

type Skill struct {
	ID, Value, Max int64
}

func (s *Store) SkillValue(owner, skill int64) (int64, bool, error) {
	var value int64
	err := s.db.QueryRow(`SELECT value FROM character_skills WHERE guid = ? AND skill = ?`, owner, skill).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("query character skill: %w", err)
	}
	return value, true, nil
}
