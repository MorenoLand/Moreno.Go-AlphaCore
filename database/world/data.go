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

type StartingAction struct {
	Button int64
	Action int64
}

type AreaTriggerTeleport struct {
	ID, RequiredLevel, RequiredItem, RequiredItem2, RequiredQuestDone, TargetMap int64
	Name                                                                         string
	TargetPositionX, TargetPositionY, TargetPositionZ, TargetOrientation         float32
}

type Worldport struct {
	Entry, Map int64
	X, Y, Z, O float32
	Name       string
}

func (s *Store) AreaTriggerTeleport(id int64) (AreaTriggerTeleport, bool, error) {
	var teleport AreaTriggerTeleport
	err := s.db.QueryRow(`SELECT id, COALESCE(name, ''), required_level, required_item, required_item2, required_quest_done, target_map, target_position_x, target_position_y, target_position_z, target_orientation FROM areatrigger_teleport WHERE id = ?`, id).Scan(&teleport.ID, &teleport.Name, &teleport.RequiredLevel, &teleport.RequiredItem, &teleport.RequiredItem2, &teleport.RequiredQuestDone, &teleport.TargetMap, &teleport.TargetPositionX, &teleport.TargetPositionY, &teleport.TargetPositionZ, &teleport.TargetOrientation)
	if err == sql.ErrNoRows {
		return AreaTriggerTeleport{}, false, nil
	}
	if err != nil {
		return AreaTriggerTeleport{}, false, fmt.Errorf("query area trigger teleport: %w", err)
	}
	return teleport, true, nil
}

func (s *Store) WorldportByName(name string) (Worldport, bool, error) {
	var port Worldport
	err := s.db.QueryRow(`SELECT entry, x, y, z, o, map, name FROM worldports WHERE name LIKE ? ORDER BY entry LIMIT 1`, "%"+name+"%").Scan(&port.Entry, &port.X, &port.Y, &port.Z, &port.O, &port.Map, &port.Name)
	if err == sql.ErrNoRows {
		return Worldport{}, false, nil
	}
	if err != nil {
		return Worldport{}, false, fmt.Errorf("query worldport: %w", err)
	}
	return port, true, nil
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

func (s *Store) StartingActions(race, class uint8) ([]StartingAction, error) {
	rows, err := s.db.Query(`SELECT button, action FROM playercreateinfo_action WHERE race = ? AND "class" = ? ORDER BY button`, race, class)
	if err != nil {
		return nil, fmt.Errorf("query starting actions: %w", err)
	}
	defer rows.Close()
	var actions []StartingAction
	for rows.Next() {
		var action StartingAction
		if err := rows.Scan(&action.Button, &action.Action); err != nil {
			return nil, fmt.Errorf("scan starting action: %w", err)
		}
		actions = append(actions, action)
	}
	return actions, rows.Err()
}
