package dbc

import (
	"database/sql"
	"fmt"
)

type Lock struct {
	ID                             int64
	Skill1, Skill2, Skill3, Skill4 int64
}

func (s *Store) LockByID(id int64) (Lock, bool, error) {
	var lock Lock
	err := s.db.QueryRow(`SELECT ID, Skill_1, Skill_2, Skill_3, Skill_4 FROM Lock WHERE ID = ?`, id).Scan(&lock.ID, &lock.Skill1, &lock.Skill2, &lock.Skill3, &lock.Skill4)
	if err == sql.ErrNoRows {
		return Lock{}, false, nil
	}
	if err != nil {
		return Lock{}, false, fmt.Errorf("query lock: %w", err)
	}
	return lock, true, nil
}
