package realm

import (
	"database/sql"
	"fmt"

	"Moreno.AlphaCore/database"
)

type Character struct {
	GUID          int64
	AccountID     int64
	RealmID       int64
	Name          string
	Race          uint8
	Class         uint8
	Gender        uint8
	Level         uint8
	XP            int64
	Money         int64
	Skin          uint8
	Face          uint8
	Hairstyle     uint8
	Haircolour    uint8
	Facialhair    uint8
	Bankslots     uint8
	Talentpoints  int64
	Skillpoints   int64
	PositionX     float32
	PositionY     float32
	PositionZ     float32
	Map           int64
	Orientation   float32
	Taximask      string
	ExploredAreas string
	Online        uint8
	Totaltime     int64
	Leveltime     int64
	ExtraFlags    int64
	Zone          int64
	TaxiPath      string
	Drunk         int64
	Health        int64
	Power1        int64
	Power2        int64
	Power3        int64
	Power4        int64
	Power5        int64
}

type Store struct{ db *sql.DB }

func NewStore(databases *database.Databases) *Store { return &Store{db: databases.DB(database.Realm)} }

func (s *Store) Characters(accountID, realmID int64) ([]Character, error) {
	rows, err := s.db.Query(`SELECT guid, account_id, realm_id, name, race, "class", gender, level, xp, money, skin, face, hairstyle, haircolour, facialhair, bankslots, talentpoints, skillpoints, position_x, position_y, position_z, map, orientation, COALESCE(taximask, ''), COALESCE(explored_areas, ''), online, totaltime, leveltime, extra_flags, zone, COALESCE(taxi_path, ''), drunk, health, power1, power2, power3, power4, power5 FROM characters WHERE account_id = ? AND realm_id = ? ORDER BY guid LIMIT 10`, accountID, realmID)
	if err != nil {
		return nil, fmt.Errorf("query characters: %w", err)
	}
	defer rows.Close()
	var characters []Character
	for rows.Next() {
		var character Character
		if err := rows.Scan(&character.GUID, &character.AccountID, &character.RealmID, &character.Name, &character.Race, &character.Class, &character.Gender, &character.Level, &character.XP, &character.Money, &character.Skin, &character.Face, &character.Hairstyle, &character.Haircolour, &character.Facialhair, &character.Bankslots, &character.Talentpoints, &character.Skillpoints, &character.PositionX, &character.PositionY, &character.PositionZ, &character.Map, &character.Orientation, &character.Taximask, &character.ExploredAreas, &character.Online, &character.Totaltime, &character.Leveltime, &character.ExtraFlags, &character.Zone, &character.TaxiPath, &character.Drunk, &character.Health, &character.Power1, &character.Power2, &character.Power3, &character.Power4, &character.Power5); err != nil {
			return nil, fmt.Errorf("scan character: %w", err)
		}
		characters = append(characters, character)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read characters: %w", err)
	}
	return characters, nil
}

func (s *Store) NameExists(name string, realmID int64) (bool, error) {
	var value string
	err := s.db.QueryRow(`SELECT name FROM characters WHERE name = ? COLLATE NOCASE AND realm_id = ? LIMIT 1`, name, realmID).Scan(&value)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("query character name: %w", err)
	}
	return true, nil
}

func (s *Store) Count(accountID, realmID int64) (int, error) {
	var count int
	if err := s.db.QueryRow(`SELECT count(*) FROM characters WHERE account_id = ? AND realm_id = ?`, accountID, realmID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count characters: %w", err)
	}
	return count, nil
}

func (s *Store) Create(character Character) (int64, error) {
	result, err := s.db.Exec(`INSERT INTO characters (account_id, realm_id, name, race, "class", gender, level, xp, money, skin, face, hairstyle, haircolour, facialhair, bankslots, talentpoints, skillpoints, position_x, position_y, position_z, map, orientation, online, totaltime, leveltime, extra_flags, zone, drunk, health, power1, power2, power3, power4, power5) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, 0, ?, 0, ?, ?, ?, ?, ?, ?)`, character.AccountID, character.RealmID, character.Name, character.Race, character.Class, character.Gender, character.Level, character.XP, character.Money, character.Skin, character.Face, character.Hairstyle, character.Haircolour, character.Facialhair, character.Bankslots, character.Talentpoints, character.Skillpoints, character.PositionX, character.PositionY, character.PositionZ, character.Map, character.Orientation, character.Zone, character.Health, character.Power1, character.Power2, character.Power3, character.Power4, character.Power5)
	if err != nil {
		return 0, fmt.Errorf("create character: %w", err)
	}
	guid, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read character id: %w", err)
	}
	return guid, nil
}

func (s *Store) Delete(guid, accountID, realmID int64) (bool, error) {
	result, err := s.db.Exec(`DELETE FROM characters WHERE guid = ? AND account_id = ? AND realm_id = ?`, guid, accountID, realmID)
	if err != nil {
		return false, fmt.Errorf("delete character: %w", err)
	}
	count, err := result.RowsAffected()
	return count == 1, err
}
