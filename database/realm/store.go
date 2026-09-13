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

type InventoryItem struct {
	GUID, Owner, Creator, Bag, Slot, ItemTemplate, StackCount, Duration, Flags int64
	SpellCharges                                                               [5]int64
}

type Spell struct {
	ID     int64
	Active bool
}

type Store struct{ db *sql.DB }

const characterColumns = `guid, account_id, realm_id, name, race, "class", gender, level, xp, money, skin, face, hairstyle, haircolour, facialhair, bankslots, talentpoints, skillpoints, position_x, position_y, position_z, map, orientation, COALESCE(taximask, ''), COALESCE(explored_areas, ''), online, totaltime, leveltime, extra_flags, zone, COALESCE(taxi_path, ''), drunk, health, power1, power2, power3, power4, power5`

type rowScanner interface{ Scan(...any) error }

func scanCharacter(row rowScanner) (Character, error) {
	var character Character
	err := row.Scan(&character.GUID, &character.AccountID, &character.RealmID, &character.Name, &character.Race, &character.Class, &character.Gender, &character.Level, &character.XP, &character.Money, &character.Skin, &character.Face, &character.Hairstyle, &character.Haircolour, &character.Facialhair, &character.Bankslots, &character.Talentpoints, &character.Skillpoints, &character.PositionX, &character.PositionY, &character.PositionZ, &character.Map, &character.Orientation, &character.Taximask, &character.ExploredAreas, &character.Online, &character.Totaltime, &character.Leveltime, &character.ExtraFlags, &character.Zone, &character.TaxiPath, &character.Drunk, &character.Health, &character.Power1, &character.Power2, &character.Power3, &character.Power4, &character.Power5)
	return character, err
}

func NewStore(databases *database.Databases) *Store { return &Store{db: databases.DB(database.Realm)} }

func (s *Store) Characters(accountID, realmID int64) ([]Character, error) {
	rows, err := s.db.Query(`SELECT `+characterColumns+` FROM characters WHERE account_id = ? AND realm_id = ? ORDER BY guid LIMIT 10`, accountID, realmID)
	if err != nil {
		return nil, fmt.Errorf("query characters: %w", err)
	}
	defer rows.Close()
	var characters []Character
	for rows.Next() {
		character, err := scanCharacter(rows)
		if err != nil {
			return nil, fmt.Errorf("scan character: %w", err)
		}
		characters = append(characters, character)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read characters: %w", err)
	}
	return characters, nil
}

func (s *Store) Character(guid, accountID, realmID int64) (Character, bool, error) {
	character, err := scanCharacter(s.db.QueryRow(`SELECT `+characterColumns+` FROM characters WHERE guid = ? AND account_id = ? AND realm_id = ? LIMIT 1`, guid, accountID, realmID))
	if err == sql.ErrNoRows {
		return Character{}, false, nil
	}
	if err != nil {
		return Character{}, false, fmt.Errorf("query character: %w", err)
	}
	return character, true, nil
}

func (s *Store) CharacterByGUID(guid int64) (Character, bool, error) {
	character, err := scanCharacter(s.db.QueryRow(`SELECT `+characterColumns+` FROM characters WHERE guid = ? LIMIT 1`, guid))
	if err == sql.ErrNoRows {
		return Character{}, false, nil
	}
	if err != nil {
		return Character{}, false, fmt.Errorf("query character by guid: %w", err)
	}
	return character, true, nil
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

func (s *Store) SetOnline(guid, accountID, realmID int64, online bool) error {
	value := 0
	if online {
		value = 1
	}
	_, err := s.db.Exec(`UPDATE characters SET online = ? WHERE guid = ? AND account_id = ? AND realm_id = ?`, value, guid, accountID, realmID)
	return err
}

func (s *Store) UpdatePosition(guid, accountID, realmID int64, x, y, z, o float32) error {
	_, err := s.db.Exec(`UPDATE characters SET position_x = ?, position_y = ?, position_z = ?, orientation = ? WHERE guid = ? AND account_id = ? AND realm_id = ?`, x, y, z, o, guid, accountID, realmID)
	return err
}

func (s *Store) UpdateZone(guid, accountID, realmID, zone int64) error {
	_, err := s.db.Exec(`UPDATE characters SET zone = ? WHERE guid = ? AND account_id = ? AND realm_id = ?`, zone, guid, accountID, realmID)
	return err
}

func (s *Store) AddInventoryItem(owner, itemTemplate, slot, amount int64) error {
	return s.AddInventoryItemAt(owner, itemTemplate, 23, slot, amount)
}

func (s *Store) AddInventoryItemAt(owner, itemTemplate, bag, slot, amount int64) error {
	_, err := s.db.Exec(`INSERT INTO character_inventory (owner, bag, slot, item_template, stackcount, enchantments) VALUES (?, ?, ?, ?, ?, '')`, owner, bag, slot, itemTemplate, amount)
	return err
}

func (s *Store) AddSpell(owner, spell int64) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO character_spells (guid, spell, active, disabled) VALUES (?, ?, 1, 0)`, owner, spell)
	return err
}

func (s *Store) AddButton(owner, index, action int64) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO character_buttons (owner, "index", action) VALUES (?, ?, ?)`, owner, index, action)
	return err
}

func (s *Store) SetButton(owner, index, action int64) error {
	if _, err := s.db.Exec(`DELETE FROM character_buttons WHERE owner = ? AND "index" = ?`, owner, index); err != nil {
		return err
	}
	if action == 0 {
		return nil
	}
	return s.AddButton(owner, index, action)
}

func (s *Store) SetSpellButton(owner, spell, index int64) error {
	_, err := s.db.Exec(`INSERT INTO character_spell_book (owner, "index", spell) VALUES (?, ?, ?) ON CONFLICT(owner, spell) DO UPDATE SET "index" = excluded."index"`, owner, index, spell)
	return err
}

func (s *Store) Spells(owner int64) ([]Spell, error) {
	rows, err := s.db.Query(`SELECT spell, active FROM character_spells WHERE guid = ? ORDER BY spell`, owner)
	if err != nil {
		return nil, fmt.Errorf("query character spells: %w", err)
	}
	defer rows.Close()
	var spells []Spell
	for rows.Next() {
		var spell Spell
		var active int
		if err := rows.Scan(&spell.ID, &active); err != nil {
			return nil, fmt.Errorf("scan character spell: %w", err)
		}
		spell.Active = active != 0
		spells = append(spells, spell)
	}
	return spells, rows.Err()
}

func (s *Store) Buttons(owner int64) (map[int64]int64, error) {
	rows, err := s.db.Query(`SELECT "index", action FROM character_buttons WHERE owner = ?`, owner)
	if err != nil {
		return nil, fmt.Errorf("query character buttons: %w", err)
	}
	defer rows.Close()
	buttons := make(map[int64]int64)
	for rows.Next() {
		var index, action int64
		if err := rows.Scan(&index, &action); err != nil {
			return nil, fmt.Errorf("scan character button: %w", err)
		}
		buttons[index] = action
	}
	return buttons, rows.Err()
}

func (s *Store) SpellButtons(owner int64) (map[int64]int64, error) {
	rows, err := s.db.Query(`SELECT spell, "index" FROM character_spell_book WHERE owner = ?`, owner)
	if err != nil {
		return nil, fmt.Errorf("query character spell buttons: %w", err)
	}
	defer rows.Close()
	buttons := make(map[int64]int64)
	for rows.Next() {
		var spell, index int64
		if err := rows.Scan(&spell, &index); err != nil {
			return nil, fmt.Errorf("scan character spell button: %w", err)
		}
		buttons[spell] = index
	}
	return buttons, rows.Err()
}

func (s *Store) Inventory(owner int64) ([]InventoryItem, error) {
	rows, err := s.db.Query(`SELECT slot, item_template FROM character_inventory WHERE owner = ? AND bag = 23 AND slot BETWEEN 0 AND 19`, owner)
	if err != nil {
		return nil, fmt.Errorf("query character inventory: %w", err)
	}
	defer rows.Close()
	var items []InventoryItem
	for rows.Next() {
		var item InventoryItem
		if err := rows.Scan(&item.Slot, &item.ItemTemplate); err != nil {
			return nil, fmt.Errorf("scan character inventory: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) WorldInventory(owner int64) ([]InventoryItem, error) {
	rows, err := s.db.Query(`SELECT guid, owner, creator, bag, slot, item_template, stackcount, duration, item_flags, SpellCharges1, SpellCharges2, SpellCharges3, SpellCharges4, SpellCharges5 FROM character_inventory WHERE owner = ? AND bag = 23 AND slot BETWEEN 0 AND 39 ORDER BY slot, guid`, owner)
	if err != nil {
		return nil, fmt.Errorf("query world inventory: %w", err)
	}
	defer rows.Close()
	var items []InventoryItem
	for rows.Next() {
		var item InventoryItem
		values := []interface{}{&item.GUID, &item.Owner, &item.Creator, &item.Bag, &item.Slot, &item.ItemTemplate, &item.StackCount, &item.Duration, &item.Flags}
		for index := range item.SpellCharges {
			values = append(values, &item.SpellCharges[index])
		}
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("scan world inventory: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
