package realm

import (
	"database/sql"
	"encoding/binary"
	"fmt"
)

type Pet struct {
	ID, OwnerGUID, CreatureID, CreatedBySpell, Level, XP, ReactState, CommandState int64
	Name                                                                           string
	RenameTime, Health, Mana                                                       int64
	ActionBar                                                                      [10]int64
	Active                                                                         bool
}

func scanPet(row rowScanner) (Pet, error) {
	var pet Pet
	var actionBar []byte
	var active int
	err := row.Scan(&pet.ID, &pet.OwnerGUID, &pet.CreatureID, &pet.CreatedBySpell, &pet.Level, &pet.XP, &pet.ReactState, &pet.CommandState, &pet.Name, &pet.RenameTime, &pet.Health, &pet.Mana, &actionBar, &active)
	if err != nil {
		return Pet{}, err
	}
	for index := range pet.ActionBar {
		offset := index * 4
		if len(actionBar) >= offset+4 {
			pet.ActionBar[index] = int64(binary.LittleEndian.Uint32(actionBar[offset:]))
		}
	}
	pet.Active = active != 0
	return pet, nil
}

const petColumns = `pet_id, owner_guid, creature_id, created_by_spell, level, xp, react_state, command_state, name, rename_time, health, mana, action_bar, is_active`

func petActionBar(values [10]int64) []byte {
	data := make([]byte, len(values)*4)
	for index, value := range values {
		binary.LittleEndian.PutUint32(data[index*4:], uint32(value))
	}
	return data
}

func (s *Store) Pets(ownerGUID int64) ([]Pet, error) {
	rows, err := s.db.Query(`SELECT `+petColumns+` FROM character_pets WHERE owner_guid = ? ORDER BY pet_id`, ownerGUID)
	if err != nil {
		return nil, fmt.Errorf("query pets: %w", err)
	}
	defer rows.Close()
	pets := make([]Pet, 0)
	for rows.Next() {
		pet, err := scanPet(rows)
		if err != nil {
			return nil, fmt.Errorf("scan pet: %w", err)
		}
		pets = append(pets, pet)
	}
	return pets, rows.Err()
}

func (s *Store) Pet(id int64) (Pet, bool, error) {
	pet, err := scanPet(s.db.QueryRow(`SELECT `+petColumns+` FROM character_pets WHERE pet_id = ?`, id))
	if err == sql.ErrNoRows {
		return Pet{}, false, nil
	}
	if err != nil {
		return Pet{}, false, fmt.Errorf("query pet: %w", err)
	}
	return pet, true, nil
}

func (s *Store) CreatePet(pet Pet) (int64, error) {
	active := 0
	if pet.Active {
		active = 1
	}
	result, err := s.db.Exec(`INSERT INTO character_pets (owner_guid, creature_id, created_by_spell, level, xp, react_state, command_state, name, rename_time, health, mana, action_bar, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, pet.OwnerGUID, pet.CreatureID, pet.CreatedBySpell, pet.Level, pet.XP, pet.ReactState, pet.CommandState, pet.Name, pet.RenameTime, pet.Health, pet.Mana, petActionBar(pet.ActionBar), active)
	if err != nil {
		return 0, fmt.Errorf("create pet: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read pet id: %w", err)
	}
	return id, nil
}

func (s *Store) UpdatePet(pet Pet) error {
	active := 0
	if pet.Active {
		active = 1
	}
	_, err := s.db.Exec(`UPDATE character_pets SET owner_guid = ?, creature_id = ?, created_by_spell = ?, level = ?, xp = ?, react_state = ?, command_state = ?, name = ?, rename_time = ?, health = ?, mana = ?, action_bar = ?, is_active = ? WHERE pet_id = ?`, pet.OwnerGUID, pet.CreatureID, pet.CreatedBySpell, pet.Level, pet.XP, pet.ReactState, pet.CommandState, pet.Name, pet.RenameTime, pet.Health, pet.Mana, petActionBar(pet.ActionBar), active, pet.ID)
	return err
}

func (s *Store) DeletePet(id int64) error {
	_, err := s.db.Exec(`DELETE FROM character_pets WHERE pet_id = ?`, id)
	return err
}

func (s *Store) PetSpells(ownerGUID, petID int64) ([]int64, error) {
	rows, err := s.db.Query(`SELECT spell_id FROM character_pet_spells WHERE guid = ? AND pet_id = ? ORDER BY spell_id`, ownerGUID, petID)
	if err != nil {
		return nil, fmt.Errorf("query pet spells: %w", err)
	}
	defer rows.Close()
	spells := make([]int64, 0)
	for rows.Next() {
		var spell int64
		if err := rows.Scan(&spell); err != nil {
			return nil, fmt.Errorf("scan pet spell: %w", err)
		}
		spells = append(spells, spell)
	}
	return spells, rows.Err()
}

func (s *Store) AddPetSpell(ownerGUID, petID, spellID int64) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO character_pet_spells (guid, pet_id, spell_id) VALUES (?, ?, ?)`, ownerGUID, petID, spellID)
	return err
}

func (s *Store) DeletePetSpell(ownerGUID, petID, spellID int64) error {
	_, err := s.db.Exec(`DELETE FROM character_pet_spells WHERE guid = ? AND pet_id = ? AND spell_id = ?`, ownerGUID, petID, spellID)
	return err
}
