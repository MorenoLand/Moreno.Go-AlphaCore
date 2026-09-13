package dbc

import (
	"database/sql"
	"fmt"

	"Moreno.AlphaCore/database"
)

type Race struct {
	ID              int64
	FactionID       int64
	MaleDisplayID   int64
	FemaleDisplayID int64
	CreatureType    int64
}

type Store struct{ db *sql.DB }

func NewStore(databases *database.Databases) *Store { return &Store{db: databases.DB(database.DBC)} }

func (s *Store) Race(id uint8) (Race, bool, error) {
	var race Race
	err := s.db.QueryRow(`SELECT ID, FactionID, MaleDisplayId, FemaleDisplayId, CreatureType FROM ChrRaces WHERE ID = ?`, id).Scan(&race.ID, &race.FactionID, &race.MaleDisplayID, &race.FemaleDisplayID, &race.CreatureType)
	if err == sql.ErrNoRows {
		return Race{}, false, nil
	}
	if err != nil {
		return Race{}, false, fmt.Errorf("query race: %w", err)
	}
	return race, true, nil
}
