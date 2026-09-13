package realm

import (
	"database/sql"
	"fmt"
)

type Gift struct {
	ID, Creator, ItemGUID, Entry, Flags int64
}

func (s *Store) GiftByItemGUID(itemGUID int64) (Gift, bool, error) {
	var gift Gift
	err := s.db.QueryRow(`SELECT guid, creator, item_guid, entry, flags FROM character_gifts WHERE item_guid = ? LIMIT 1`, itemGUID).Scan(&gift.ID, &gift.Creator, &gift.ItemGUID, &gift.Entry, &gift.Flags)
	if err == sql.ErrNoRows {
		return Gift{}, false, nil
	}
	if err != nil {
		return Gift{}, false, fmt.Errorf("query character gift: %w", err)
	}
	return gift, true, nil
}

func (s *Store) WrapItem(item InventoryItem, wrappedEntry, wrappedCreator int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin item wrap: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO character_gifts (creator, item_guid, entry, flags) VALUES (?, ?, ?, ?)`, item.Creator, item.GUID, item.ItemTemplate, item.Flags); err != nil {
		return fmt.Errorf("create character gift: %w", err)
	}
	result, err := tx.Exec(`UPDATE character_inventory SET item_template = ?, creator = ?, item_flags = item_flags | 8 WHERE guid = ? AND owner = ?`, wrappedEntry, wrappedCreator, item.GUID, item.Owner)
	if err != nil {
		return fmt.Errorf("wrap inventory item: %w", err)
	}
	if changed, _ := result.RowsAffected(); changed != 1 {
		return fmt.Errorf("wrap inventory item was changed")
	}
	return tx.Commit()
}

func (s *Store) UnwrapItem(item InventoryItem, gift Gift, entry, creator int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin item unwrap: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE character_inventory SET item_template = ?, creator = ?, item_flags = item_flags & 4294967287 WHERE guid = ? AND owner = ?`, entry, creator, item.GUID, item.Owner)
	if err != nil {
		return fmt.Errorf("unwrap inventory item: %w", err)
	}
	if changed, _ := result.RowsAffected(); changed != 1 {
		return fmt.Errorf("unwrap inventory item was changed")
	}
	if _, err := tx.Exec(`DELETE FROM character_gifts WHERE guid = ? AND item_guid = ?`, gift.ID, item.GUID); err != nil {
		return fmt.Errorf("delete character gift: %w", err)
	}
	return tx.Commit()
}
