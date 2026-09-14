package realm

import (
	"database/sql"
	"fmt"
)

type Skill struct {
	ID, Value, Max int64
}

func (s *Store) Skills(owner int64) ([]Skill, error) {
	rows, err := s.db.Query(`SELECT skill, value, max FROM character_skills WHERE guid = ? ORDER BY skill`, owner)
	if err != nil {
		return nil, fmt.Errorf("query character skills: %w", err)
	}
	defer rows.Close()
	skills := make([]Skill, 0)
	for rows.Next() {
		var skill Skill
		if err := rows.Scan(&skill.ID, &skill.Value, &skill.Max); err != nil {
			return nil, fmt.Errorf("scan character skill: %w", err)
		}
		skills = append(skills, skill)
	}
	return skills, rows.Err()
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

func (s *Store) AddSkill(owner, skill, value, maximum int64) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO character_skills (guid, skill, value, max) VALUES (?, ?, ?, ?)`, owner, skill, value, maximum)
	return err
}

func (s *Store) UpdateSkill(owner, skill, value, maximum int64) error {
	_, err := s.db.Exec(`UPDATE character_skills SET value = ?, max = ? WHERE guid = ? AND skill = ?`, value, maximum, owner, skill)
	return err
}
