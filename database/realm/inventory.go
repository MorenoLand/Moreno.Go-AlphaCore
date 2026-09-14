package realm

import (
	"database/sql"
	"fmt"
)

const inventoryColumns = `guid, owner, creator, bag, slot, item_template, stackcount, duration, item_flags, SpellCharges1, SpellCharges2, SpellCharges3, SpellCharges4, SpellCharges5`

func scanInventory(row rowScanner) (InventoryItem, error) {
	var item InventoryItem
	values := []interface{}{&item.GUID, &item.Owner, &item.Creator, &item.Bag, &item.Slot, &item.ItemTemplate, &item.StackCount, &item.Duration, &item.Flags}
	for index := range item.SpellCharges {
		values = append(values, &item.SpellCharges[index])
	}
	return item, row.Scan(values...)
}

func (s *Store) ItemAt(owner, bag, slot int64) (InventoryItem, bool, error) {
	item, err := scanInventory(s.db.QueryRow(`SELECT `+inventoryColumns+` FROM character_inventory WHERE owner = ? AND bag = ? AND slot = ? LIMIT 1`, owner, bag, slot))
	if err == sql.ErrNoRows {
		return InventoryItem{}, false, nil
	}
	if err != nil {
		return InventoryItem{}, false, fmt.Errorf("query inventory item: %w", err)
	}
	return item, true, nil
}

func (s *Store) ItemByGUID(owner, guid int64) (InventoryItem, bool, error) {
	item, err := scanInventory(s.db.QueryRow(`SELECT `+inventoryColumns+` FROM character_inventory WHERE owner = ? AND guid = ? LIMIT 1`, owner, guid))
	if err == sql.ErrNoRows {
		return InventoryItem{}, false, nil
	}
	if err != nil {
		return InventoryItem{}, false, fmt.Errorf("query inventory item by guid: %w", err)
	}
	return item, true, nil
}

func (s *Store) UpdateItemLocation(guid, owner, bag, slot int64) error {
	_, err := s.db.Exec(`UPDATE character_inventory SET bag = ?, slot = ? WHERE guid = ? AND owner = ?`, bag, slot, guid, owner)
	return err
}

func (s *Store) TransferItem(guid, owner, newOwner, bag, slot int64) error {
	result, err := s.db.Exec(`UPDATE character_inventory SET owner = ?, bag = ?, slot = ? WHERE guid = ? AND owner = ?`, newOwner, bag, slot, guid, owner)
	if err != nil {
		return fmt.Errorf("transfer inventory item: %w", err)
	}
	if changed, _ := result.RowsAffected(); changed != 1 {
		return fmt.Errorf("transfer inventory item was changed")
	}
	return nil
}

func (s *Store) UpdateItemStack(guid, owner, stack int64) error {
	_, err := s.db.Exec(`UPDATE character_inventory SET stackcount = ? WHERE guid = ? AND owner = ?`, stack, guid, owner)
	return err
}

func (s *Store) UpdateItemSpellCharges(guid, owner, slot, charges int64) error {
	if slot < 0 || slot >= 5 {
		return fmt.Errorf("invalid item spell slot: %d", slot)
	}
	column := fmt.Sprintf("SpellCharges%d", slot+1)
	_, err := s.db.Exec(`UPDATE character_inventory SET `+column+` = ? WHERE guid = ? AND owner = ?`, charges, guid, owner)
	return err
}

func (s *Store) InventoryItems(owner, bag, start, end int64) ([]InventoryItem, error) {
	rows, err := s.db.Query(`SELECT `+inventoryColumns+` FROM character_inventory WHERE owner = ? AND bag = ? AND slot >= ? AND slot < ? ORDER BY slot, guid`, owner, bag, start, end)
	if err != nil {
		return nil, fmt.Errorf("query inventory items: %w", err)
	}
	defer rows.Close()
	items := make([]InventoryItem, 0)
	for rows.Next() {
		item, err := scanInventory(rows)
		if err != nil {
			return nil, fmt.Errorf("scan inventory item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read inventory items: %w", err)
	}
	return items, nil
}

func (s *Store) ItemCount(owner, template int64) (int64, error) {
	var count int64
	err := s.db.QueryRow(`SELECT COALESCE(SUM(stackcount), 0) FROM character_inventory WHERE owner = ? AND item_template = ?`, owner, template).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count inventory item: %w", err)
	}
	return count, nil
}

func (s *Store) RemoveItems(owner, template, count int64) error {
	if count <= 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin item removal: %w", err)
	}
	defer tx.Rollback()
	rows, err := tx.Query(`SELECT guid, stackcount FROM character_inventory WHERE owner = ? AND item_template = ? ORDER BY bag, slot, guid`, owner, template)
	if err != nil {
		return fmt.Errorf("query item removal: %w", err)
	}
	type stack struct{ guid, count int64 }
	items := make([]stack, 0)
	for rows.Next() {
		var item stack
		if err := rows.Scan(&item.guid, &item.count); err != nil {
			rows.Close()
			return fmt.Errorf("scan item removal: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("read item removal: %w", err)
	}
	rows.Close()
	remaining := count
	for _, item := range items {
		if remaining <= 0 {
			break
		}
		removed := item.count
		if removed > remaining {
			removed = remaining
		}
		if removed == item.count {
			if _, err := tx.Exec(`DELETE FROM character_inventory WHERE owner = ? AND guid = ?`, owner, item.guid); err != nil {
				return fmt.Errorf("delete quest item: %w", err)
			}
		} else if _, err := tx.Exec(`UPDATE character_inventory SET stackcount = stackcount - ? WHERE owner = ? AND guid = ?`, removed, owner, item.guid); err != nil {
			return fmt.Errorf("update quest item: %w", err)
		}
		remaining -= removed
	}
	if remaining > 0 {
		return fmt.Errorf("not enough inventory items: need %d", count)
	}
	return tx.Commit()
}

func (s *Store) DeleteItem(guid, owner int64) error {
	_, err := s.db.Exec(`DELETE FROM character_inventory WHERE guid = ? AND owner = ?`, guid, owner)
	return err
}

func (s *Store) CreateInventoryItem(owner, creator, bag, slot, itemTemplate, stack int64) (InventoryItem, error) {
	result, err := s.db.Exec(`INSERT INTO character_inventory (owner, creator, bag, slot, item_template, stackcount, enchantments) VALUES (?, ?, ?, ?, ?, ?, '')`, owner, creator, bag, slot, itemTemplate, stack)
	if err != nil {
		return InventoryItem{}, fmt.Errorf("create inventory item: %w", err)
	}
	guid, err := result.LastInsertId()
	if err != nil {
		return InventoryItem{}, fmt.Errorf("read inventory item id: %w", err)
	}
	return InventoryItem{GUID: guid, Owner: owner, Creator: creator, Bag: bag, Slot: slot, ItemTemplate: itemTemplate, StackCount: stack}, nil
}

func (s *Store) SplitItem(source InventoryItem, bag, slot, count int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin inventory split: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE character_inventory SET stackcount = stackcount - ? WHERE guid = ? AND owner = ? AND stackcount >= ?`, count, source.GUID, source.Owner, count)
	if err != nil {
		return fmt.Errorf("update split source: %w", err)
	}
	if changed, _ := result.RowsAffected(); changed != 1 {
		return fmt.Errorf("split source item was changed")
	}
	_, err = tx.Exec(`INSERT INTO character_inventory (owner, creator, bag, slot, item_template, stackcount, SpellCharges1, SpellCharges2, SpellCharges3, SpellCharges4, SpellCharges5, item_flags, duration, enchantments) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '')`, source.Owner, source.Creator, bag, slot, source.ItemTemplate, count, source.SpellCharges[0], source.SpellCharges[1], source.SpellCharges[2], source.SpellCharges[3], source.SpellCharges[4], source.Flags, source.Duration)
	if err != nil {
		return fmt.Errorf("insert split item: %w", err)
	}
	return tx.Commit()
}

func (s *Store) SwapItems(owner, sourceBag, sourceSlot, destBag, destSlot int64) error {
	if sourceBag == destBag && sourceSlot == destSlot {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin inventory swap: %w", err)
	}
	defer tx.Rollback()
	source, err := scanInventory(tx.QueryRow(`SELECT `+inventoryColumns+` FROM character_inventory WHERE owner = ? AND bag = ? AND slot = ? LIMIT 1`, owner, sourceBag, sourceSlot))
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("query swap source: %w", err)
	}
	dest, destErr := scanInventory(tx.QueryRow(`SELECT `+inventoryColumns+` FROM character_inventory WHERE owner = ? AND bag = ? AND slot = ? LIMIT 1`, owner, destBag, destSlot))
	if destErr != nil && destErr != sql.ErrNoRows {
		return fmt.Errorf("query swap destination: %w", destErr)
	}
	if destErr == sql.ErrNoRows {
		if _, err := tx.Exec(`UPDATE character_inventory SET bag = ?, slot = ? WHERE guid = ? AND owner = ?`, destBag, destSlot, source.GUID, owner); err != nil {
			return fmt.Errorf("move inventory item: %w", err)
		}
	} else {
		if _, err := tx.Exec(`UPDATE character_inventory SET bag = -1, slot = -1 WHERE guid IN (?, ?) AND owner = ?`, source.GUID, dest.GUID, owner); err != nil {
			return fmt.Errorf("reserve inventory swap: %w", err)
		}
		if _, err := tx.Exec(`UPDATE character_inventory SET bag = ?, slot = ? WHERE guid = ? AND owner = ?`, destBag, destSlot, source.GUID, owner); err != nil {
			return fmt.Errorf("place swapped source: %w", err)
		}
		if _, err := tx.Exec(`UPDATE character_inventory SET bag = ?, slot = ? WHERE guid = ? AND owner = ?`, sourceBag, sourceSlot, dest.GUID, owner); err != nil {
			return fmt.Errorf("place swapped destination: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Store) FirstEmptySlot(owner, bag, start, end int64) (int64, error) {
	var slot int64
	for slot = start; slot < end; slot++ {
		var exists int
		if err := s.db.QueryRow(`SELECT 1 FROM character_inventory WHERE owner = ? AND bag = ? AND slot = ? LIMIT 1`, owner, bag, slot).Scan(&exists); err == sql.ErrNoRows {
			return slot, nil
		} else if err != nil {
			return -1, fmt.Errorf("query empty inventory slot: %w", err)
		}
	}
	return -1, nil
}
