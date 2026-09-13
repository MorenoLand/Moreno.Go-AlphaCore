package world

import (
	"database/sql"
	"fmt"
)

type ClassStats struct {
	BaseHealth int64
	BaseMana   int64
}

type StartingItem struct {
	ItemID int64
	Amount int64
}

type ItemTemplate struct {
	Entry         int64
	DisplayID     int64
	InventoryType int64
}

func (s *Store) ClassStats(class, level uint8) (ClassStats, bool, error) {
	var stats ClassStats
	err := s.db.QueryRow(`SELECT basehp, basemana FROM player_classlevelstats WHERE "class" = ? AND level = ?`, class, level).Scan(&stats.BaseHealth, &stats.BaseMana)
	if err == sql.ErrNoRows {
		return ClassStats{}, false, nil
	}
	if err != nil {
		return ClassStats{}, false, fmt.Errorf("query class stats: %w", err)
	}
	return stats, true, nil
}

func (s *Store) StartingItems(race, class uint8) ([]StartingItem, error) {
	rows, err := s.db.Query(`SELECT itemid, amount FROM playercreateinfo_item WHERE race = ? AND "class" = ? ORDER BY id`, race, class)
	if err != nil {
		return nil, fmt.Errorf("query starting items: %w", err)
	}
	defer rows.Close()
	var items []StartingItem
	for rows.Next() {
		var item StartingItem
		if err := rows.Scan(&item.ItemID, &item.Amount); err != nil {
			return nil, fmt.Errorf("scan starting item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) StartingSpells(race, class uint8) ([]int64, error) {
	rows, err := s.db.Query(`SELECT Spell FROM playercreateinfo_spell WHERE race = ? AND "class" = ? ORDER BY Spell`, race, class)
	if err != nil {
		return nil, fmt.Errorf("query starting spells: %w", err)
	}
	defer rows.Close()
	var spells []int64
	for rows.Next() {
		var spell int64
		if err := rows.Scan(&spell); err != nil {
			return nil, fmt.Errorf("scan starting spell: %w", err)
		}
		spells = append(spells, spell)
	}
	return spells, rows.Err()
}

func (s *Store) ItemTemplate(entry int64) (ItemTemplate, bool, error) {
	var item ItemTemplate
	err := s.db.QueryRow(`SELECT entry, display_id, inventory_type FROM item_template WHERE entry = ?`, entry).Scan(&item.Entry, &item.DisplayID, &item.InventoryType)
	if err == sql.ErrNoRows {
		return ItemTemplate{}, false, nil
	}
	if err != nil {
		return ItemTemplate{}, false, fmt.Errorf("query item template: %w", err)
	}
	return item, true, nil
}
