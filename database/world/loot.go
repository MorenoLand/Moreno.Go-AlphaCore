package world

import "fmt"

type LootTemplate struct {
	Entry, Item, GroupID, MinCountOrRef, MaxCount int64
	Chance                                        float64
}

func (s *Store) lootTemplates(table string, entry int64) ([]LootTemplate, error) {
	allowed := map[string]bool{"item_loot_template": true, "creature_loot_template": true, "gameobject_loot_template": true, "fishing_loot_template": true, "reference_loot_template": true}
	if !allowed[table] {
		return nil, fmt.Errorf("unknown loot table %q", table)
	}
	rows, err := s.db.Query(`SELECT entry, item, ChanceOrQuestChance, groupid, mincountOrRef, maxcount FROM `+table+` WHERE entry = ? ORDER BY groupid, item`, entry)
	if err != nil {
		return nil, fmt.Errorf("query loot templates: %w", err)
	}
	defer rows.Close()
	items := make([]LootTemplate, 0)
	for rows.Next() {
		var item LootTemplate
		if err := rows.Scan(&item.Entry, &item.Item, &item.Chance, &item.GroupID, &item.MinCountOrRef, &item.MaxCount); err != nil {
			return nil, fmt.Errorf("scan loot template: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read loot templates: %w", err)
	}
	return items, nil
}

func (s *Store) ItemLootTemplates(entry int64) ([]LootTemplate, error) {
	return s.lootTemplates("item_loot_template", entry)
}

func (s *Store) CreatureLootTemplates(entry int64) ([]LootTemplate, error) {
	return s.lootTemplates("creature_loot_template", entry)
}

func (s *Store) GameObjectLootTemplates(entry int64) ([]LootTemplate, error) {
	return s.lootTemplates("gameobject_loot_template", entry)
}

func (s *Store) FishingLootTemplates(entry int64) ([]LootTemplate, error) {
	return s.lootTemplates("fishing_loot_template", entry)
}

func (s *Store) ReferenceLootTemplates(entry int64) ([]LootTemplate, error) {
	return s.lootTemplates("reference_loot_template", entry)
}
