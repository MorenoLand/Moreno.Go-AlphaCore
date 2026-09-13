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

type AreaTrigger struct {
	ID, ContinentID int64
	X, Y, Z, Radius float32
}

type TaxiNode struct {
	ID, ContinentID, Team int64
	X, Y, Z               float32
}

type TaxiPath struct {
	ID, From, To, Cost int64
}

type EmoteText struct {
	ID      int64
	EmoteID int64
}

type CreatureDisplayInfo struct {
	ID, ModelID int64
	ModelScale  float32
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

func (s *Store) AreaTriggerByID(id int64) (AreaTrigger, bool, error) {
	var trigger AreaTrigger
	err := s.db.QueryRow(`SELECT ID, ContinentID, X, Y, Z, Radius FROM AreaTrigger WHERE ID = ?`, id).Scan(&trigger.ID, &trigger.ContinentID, &trigger.X, &trigger.Y, &trigger.Z, &trigger.Radius)
	if err == sql.ErrNoRows {
		return AreaTrigger{}, false, nil
	}
	if err != nil {
		return AreaTrigger{}, false, fmt.Errorf("query area trigger: %w", err)
	}
	return trigger, true, nil
}

func (s *Store) MapExists(id int64) (bool, error) {
	var value int
	err := s.db.QueryRow(`SELECT 1 FROM Map WHERE ID = ?`, id).Scan(&value)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (s *Store) TaxiNodesByMap(mapID int64) ([]TaxiNode, error) {
	rows, err := s.db.Query(`SELECT ID, ContinentID, X, Y, Z, custom_Team FROM TaxiNodes WHERE ContinentID = ? ORDER BY ID`, mapID)
	if err != nil {
		return nil, fmt.Errorf("query taxi nodes: %w", err)
	}
	defer rows.Close()
	nodes := make([]TaxiNode, 0)
	for rows.Next() {
		var node TaxiNode
		if err := rows.Scan(&node.ID, &node.ContinentID, &node.X, &node.Y, &node.Z, &node.Team); err != nil {
			return nil, fmt.Errorf("scan taxi node: %w", err)
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func (s *Store) TaxiNodesAll() ([]TaxiNode, error) {
	rows, err := s.db.Query(`SELECT ID, ContinentID, X, Y, Z, custom_Team FROM TaxiNodes ORDER BY ID`)
	if err != nil {
		return nil, fmt.Errorf("query all taxi nodes: %w", err)
	}
	defer rows.Close()
	nodes := make([]TaxiNode, 0)
	for rows.Next() {
		var node TaxiNode
		if err := rows.Scan(&node.ID, &node.ContinentID, &node.X, &node.Y, &node.Z, &node.Team); err != nil {
			return nil, fmt.Errorf("scan all taxi node: %w", err)
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func (s *Store) TaxiPath(from, to int64) (TaxiPath, bool, error) {
	var path TaxiPath
	err := s.db.QueryRow(`SELECT ID, FromTaxiNode, ToTaxiNode, Cost FROM TaxiPath WHERE FromTaxiNode = ? AND ToTaxiNode = ? LIMIT 1`, from, to).Scan(&path.ID, &path.From, &path.To, &path.Cost)
	if err == sql.ErrNoRows {
		return TaxiPath{}, false, nil
	}
	if err != nil {
		return TaxiPath{}, false, fmt.Errorf("query taxi path: %w", err)
	}
	return path, true, nil
}

func (s *Store) EmoteText(id int64) (EmoteText, bool, error) {
	var emote EmoteText
	err := s.db.QueryRow(`SELECT ID, EmoteID FROM EmotesText WHERE ID = ?`, id).Scan(&emote.ID, &emote.EmoteID)
	if err == sql.ErrNoRows {
		return EmoteText{}, false, nil
	}
	if err != nil {
		return EmoteText{}, false, fmt.Errorf("query emote text: %w", err)
	}
	return emote, true, nil
}

func (s *Store) CreatureDisplayInfo(id int64) (CreatureDisplayInfo, bool, error) {
	var display CreatureDisplayInfo
	err := s.db.QueryRow(`SELECT ID, ModelID, CreatureModelScale FROM CreatureDisplayInfo WHERE ID = ?`, id).Scan(&display.ID, &display.ModelID, &display.ModelScale)
	if err == sql.ErrNoRows {
		return CreatureDisplayInfo{}, false, nil
	}
	if err != nil {
		return CreatureDisplayInfo{}, false, fmt.Errorf("query creature display info: %w", err)
	}
	return display, true, nil
}
