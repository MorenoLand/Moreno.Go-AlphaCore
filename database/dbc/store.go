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
	BaseLanguage    int64
	CreatureType    int64
}

type Area struct {
	ID            int64
	AreaNumber    int64
	ContinentID   int64
	ParentAreaNum int64
}

type Store struct{ db *sql.DB }

func NewStore(databases *database.Databases) *Store { return &Store{db: databases.DB(database.DBC)} }

func (s *Store) Race(id uint8) (Race, bool, error) {
	var race Race
	err := s.db.QueryRow(`SELECT ID, FactionID, MaleDisplayId, FemaleDisplayId, BaseLanguage, CreatureType FROM ChrRaces WHERE ID = ?`, id).Scan(&race.ID, &race.FactionID, &race.MaleDisplayID, &race.FemaleDisplayID, &race.BaseLanguage, &race.CreatureType)
	if err == sql.ErrNoRows {
		return Race{}, false, nil
	}
	if err != nil {
		return Race{}, false, fmt.Errorf("query race: %w", err)
	}
	return race, true, nil
}

func (s *Store) AreaByIDAndMap(id, mapID int64) (Area, bool, error) {
	var area Area
	err := s.db.QueryRow(`SELECT ID, AreaNumber, ContinentID, ParentAreaNum FROM AreaTable WHERE ID = ? AND ContinentID = ?`, id, mapID).Scan(&area.ID, &area.AreaNumber, &area.ContinentID, &area.ParentAreaNum)
	if err == sql.ErrNoRows {
		return Area{}, false, nil
	}
	if err != nil {
		return Area{}, false, fmt.Errorf("query area: %w", err)
	}
	return area, true, nil
}

func (s *Store) AreaByAreaNumber(number, mapID int64) (Area, bool, error) {
	var area Area
	err := s.db.QueryRow(`SELECT ID, AreaNumber, ContinentID, ParentAreaNum FROM AreaTable WHERE AreaNumber = ? AND ContinentID = ?`, number, mapID).Scan(&area.ID, &area.AreaNumber, &area.ContinentID, &area.ParentAreaNum)
	if err == sql.ErrNoRows {
		return Area{}, false, nil
	}
	if err != nil {
		return Area{}, false, fmt.Errorf("query area by number: %w", err)
	}
	return area, true, nil
}
