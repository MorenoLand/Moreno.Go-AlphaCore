package realm

import (
	"database/sql"
	"fmt"
)

type Social struct {
	GUID      int64
	OtherGUID int64
	Ignore    bool
}

func (s *Store) CharacterByName(name string) (Character, bool, error) {
	character, err := scanCharacter(s.db.QueryRow(`SELECT `+characterColumns+` FROM characters WHERE name = ? COLLATE NOCASE LIMIT 1`, name))
	if err == sql.ErrNoRows {
		return Character{}, false, nil
	}
	if err != nil {
		return Character{}, false, fmt.Errorf("query character by name: %w", err)
	}
	return character, true, nil
}

func (s *Store) Social(guid int64) ([]Social, error) {
	rows, err := s.db.Query(`SELECT guid, other_guid, ignore FROM character_social WHERE guid = ? ORDER BY other_guid, ignore`, guid)
	if err != nil {
		return nil, fmt.Errorf("query character social: %w", err)
	}
	defer rows.Close()
	var social []Social
	for rows.Next() {
		var entry Social
		var ignored int
		if err := rows.Scan(&entry.GUID, &entry.OtherGUID, &ignored); err != nil {
			return nil, fmt.Errorf("scan character social: %w", err)
		}
		entry.Ignore = ignored != 0
		social = append(social, entry)
	}
	return social, rows.Err()
}

func (s *Store) AddSocial(guid, otherGUID int64, ignored bool) error {
	value := 0
	if ignored {
		value = 1
	}
	_, err := s.db.Exec(`INSERT INTO character_social (guid, other_guid, ignore) VALUES (?, ?, ?)`, guid, otherGUID, value)
	if err != nil {
		return fmt.Errorf("add character social: %w", err)
	}
	return nil
}

func (s *Store) DeleteSocial(guid, otherGUID int64, ignored bool) error {
	value := 0
	if ignored {
		value = 1
	}
	_, err := s.db.Exec(`DELETE FROM character_social WHERE guid = ? AND other_guid = ? AND ignore = ?`, guid, otherGUID, value)
	if err != nil {
		return fmt.Errorf("delete character social: %w", err)
	}
	return nil
}
