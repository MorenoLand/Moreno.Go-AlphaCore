package dbc

import (
	"database/sql"
	"fmt"
)

func (s *Store) CinematicSequenceExists(id int64) (bool, error) {
	var value int
	err := s.db.QueryRow(`SELECT 1 FROM CinematicSequences WHERE ID = ?`, id).Scan(&value)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("query cinematic sequence: %w", err)
	}
	return true, nil
}
