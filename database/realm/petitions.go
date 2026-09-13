package realm

import (
	"database/sql"
	"fmt"
)

type Petition struct {
	ID, RealmID, OwnerGUID, ItemGUID int64
	Name                             string
}

func scanPetition(row rowScanner) (Petition, error) {
	var petition Petition
	err := row.Scan(&petition.ID, &petition.RealmID, &petition.OwnerGUID, &petition.ItemGUID, &petition.Name)
	return petition, err
}

func (s *Store) CreatePetition(ownerGUID, itemGUID int64, name string) (Petition, error) {
	result, err := s.db.Exec(`INSERT INTO petition (realm_id, owner_guid, item_guid, name) VALUES (1, ?, ?, ?)`, ownerGUID, itemGUID, name)
	if err != nil {
		return Petition{}, fmt.Errorf("create petition: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Petition{}, fmt.Errorf("read petition id: %w", err)
	}
	return Petition{ID: id, RealmID: 1, OwnerGUID: ownerGUID, ItemGUID: itemGUID, Name: name}, nil
}

func (s *Store) PetitionByItemGUID(itemGUID int64) (Petition, bool, error) {
	petition, err := scanPetition(s.db.QueryRow(`SELECT petition_id, realm_id, owner_guid, item_guid, name FROM petition WHERE item_guid = ? LIMIT 1`, itemGUID))
	if err == sql.ErrNoRows {
		return Petition{}, false, nil
	}
	if err != nil {
		return Petition{}, false, fmt.Errorf("query petition by item: %w", err)
	}
	return petition, true, nil
}

func (s *Store) PetitionByOwner(ownerGUID int64) (Petition, bool, error) {
	petition, err := scanPetition(s.db.QueryRow(`SELECT petition_id, realm_id, owner_guid, item_guid, name FROM petition WHERE owner_guid = ? LIMIT 1`, ownerGUID))
	if err == sql.ErrNoRows {
		return Petition{}, false, nil
	}
	if err != nil {
		return Petition{}, false, fmt.Errorf("query petition by owner: %w", err)
	}
	return petition, true, nil
}

func (s *Store) PetitionByName(name string) (Petition, bool, error) {
	petition, err := scanPetition(s.db.QueryRow(`SELECT petition_id, realm_id, owner_guid, item_guid, name FROM petition WHERE name = ? COLLATE NOCASE AND realm_id = 1 LIMIT 1`, name))
	if err == sql.ErrNoRows {
		return Petition{}, false, nil
	}
	if err != nil {
		return Petition{}, false, fmt.Errorf("query petition by name: %w", err)
	}
	return petition, true, nil
}

func (s *Store) PetitionSigners(petitionID int64) ([]int64, error) {
	rows, err := s.db.Query(`SELECT player_guid FROM petition_sign WHERE petition_id = ? ORDER BY player_guid`, petitionID)
	if err != nil {
		return nil, fmt.Errorf("query petition signatures: %w", err)
	}
	defer rows.Close()
	signers := make([]int64, 0)
	for rows.Next() {
		var guid int64
		if err := rows.Scan(&guid); err != nil {
			return nil, fmt.Errorf("scan petition signature: %w", err)
		}
		signers = append(signers, guid)
	}
	return signers, rows.Err()
}

func (s *Store) PetitionSignerExists(petitionID, playerGUID int64) (bool, error) {
	var value int
	err := s.db.QueryRow(`SELECT 1 FROM petition_sign WHERE petition_id = ? AND player_guid = ?`, petitionID, playerGUID).Scan(&value)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (s *Store) AddPetitionSigner(petitionID, playerGUID int64) error {
	_, err := s.db.Exec(`INSERT INTO petition_sign (petition_id, player_guid) VALUES (?, ?)`, petitionID, playerGUID)
	return err
}

func (s *Store) DeletePetition(petitionID int64) error {
	_, err := s.db.Exec(`DELETE FROM petition WHERE petition_id = ?`, petitionID)
	return err
}
