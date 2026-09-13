package world

import "fmt"

type VendorItem struct {
	Item, MaxCount, Slot int64
}

func (s *Store) VendorItems(entry int64, template bool) ([]VendorItem, error) {
	table := "npc_vendor"
	if template {
		table = "npc_vendor_template"
	}
	rows, err := s.db.Query(`SELECT item, maxcount, slot FROM `+table+` WHERE entry = ? ORDER BY slot, item`, entry)
	if err != nil {
		return nil, fmt.Errorf("query vendor items: %w", err)
	}
	defer rows.Close()
	items := make([]VendorItem, 0)
	for rows.Next() {
		var item VendorItem
		if err := rows.Scan(&item.Item, &item.MaxCount, &item.Slot); err != nil {
			return nil, fmt.Errorf("scan vendor item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read vendor items: %w", err)
	}
	return items, nil
}
