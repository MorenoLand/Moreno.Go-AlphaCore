package world

import (
	"database/sql"
	"fmt"

	"Moreno.AlphaCore/database"
)

type StartLocation struct {
	Map         int64
	Zone        int64
	PositionX   float32
	PositionY   float32
	PositionZ   float32
	Orientation float32
}

type Store struct{ db *sql.DB }

func NewStore(databases *database.Databases) *Store { return &Store{db: databases.DB(database.World)} }

func (s *Store) StartingLocation(race, class uint8) (StartLocation, bool, error) {
	var location StartLocation
	err := s.db.QueryRow(`SELECT map, zone, position_x, position_y, position_z, orientation FROM playercreateinfo WHERE race = ? AND "class" = ? ORDER BY id LIMIT 1`, race, class).Scan(&location.Map, &location.Zone, &location.PositionX, &location.PositionY, &location.PositionZ, &location.Orientation)
	if err == sql.ErrNoRows {
		return StartLocation{}, false, nil
	}
	if err != nil {
		return StartLocation{}, false, fmt.Errorf("query starting location: %w", err)
	}
	return location, true, nil
}
