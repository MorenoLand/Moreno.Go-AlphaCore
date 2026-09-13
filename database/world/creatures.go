package world

import (
	"database/sql"
	"fmt"
)

type CreatureClassLevelStats struct {
	MeleeDamage, RangedDamage             float32
	AttackPower, RangedAttackPower        int64
	Health, BaseHealth, Mana, BaseMana    int64
	Strength, Agility, Stamina, Intellect int64
	Spirit, Armor                         int64
}

type CreatureSpawn struct {
	SpawnID, Entry, Map, MovementType, SpawnFlags int64
	PositionX, PositionY, PositionZ, Orientation  float32
	HealthPercent, ManaPercent                    float32
}

type CreatureModelInfo struct {
	ModelID, Gender int64
	BoundingRadius  float32
	CombatReach     float32
}

func (s *Store) CreatureClassLevelStats(class, level int64) (CreatureClassLevelStats, bool, error) {
	var stats CreatureClassLevelStats
	err := s.db.QueryRow(`SELECT melee_damage, ranged_damage, attack_power, ranged_attack_power, health, base_health, mana, base_mana, strength, agility, stamina, intellect, spirit, armor FROM creature_classlevelstats WHERE "class" = ? AND level = ? LIMIT 1`, class, level).Scan(&stats.MeleeDamage, &stats.RangedDamage, &stats.AttackPower, &stats.RangedAttackPower, &stats.Health, &stats.BaseHealth, &stats.Mana, &stats.BaseMana, &stats.Strength, &stats.Agility, &stats.Stamina, &stats.Intellect, &stats.Spirit, &stats.Armor)
	if err == sql.ErrNoRows {
		return CreatureClassLevelStats{}, false, nil
	}
	if err != nil {
		return CreatureClassLevelStats{}, false, fmt.Errorf("query creature class stats: %w", err)
	}
	return stats, true, nil
}

func (s *Store) CreatureSpawns(mapID int64, x, y, z, distance float32) ([]CreatureSpawn, error) {
	rows, err := s.db.Query(`SELECT spawn_id, spawn_entry1, map, position_x, position_y, position_z, orientation, health_percent, mana_percent, movement_type, spawn_flags FROM spawns_creatures WHERE map = ? AND ignored = 0 AND spawn_entry1 > 0 AND position_x BETWEEN ? AND ? AND position_y BETWEEN ? AND ? ORDER BY spawn_id`, mapID, x-distance, x+distance, y-distance, y+distance)
	if err != nil {
		return nil, fmt.Errorf("query creature spawns: %w", err)
	}
	defer rows.Close()
	spawns := make([]CreatureSpawn, 0)
	for rows.Next() {
		var spawn CreatureSpawn
		if err := rows.Scan(&spawn.SpawnID, &spawn.Entry, &spawn.Map, &spawn.PositionX, &spawn.PositionY, &spawn.PositionZ, &spawn.Orientation, &spawn.HealthPercent, &spawn.ManaPercent, &spawn.MovementType, &spawn.SpawnFlags); err != nil {
			return nil, fmt.Errorf("scan creature spawn: %w", err)
		}
		dx, dy, dz := spawn.PositionX-x, spawn.PositionY-y, spawn.PositionZ-z
		if dx*dx+dy*dy+dz*dz <= distance*distance {
			spawns = append(spawns, spawn)
		}
	}
	return spawns, rows.Err()
}

func (s *Store) CreatureSpawnByID(id int64) (CreatureSpawn, bool, error) {
	var spawn CreatureSpawn
	err := s.db.QueryRow(`SELECT spawn_id, spawn_entry1, map, position_x, position_y, position_z, orientation, health_percent, mana_percent, movement_type, spawn_flags FROM spawns_creatures WHERE spawn_id = ?`, id).Scan(&spawn.SpawnID, &spawn.Entry, &spawn.Map, &spawn.PositionX, &spawn.PositionY, &spawn.PositionZ, &spawn.Orientation, &spawn.HealthPercent, &spawn.ManaPercent, &spawn.MovementType, &spawn.SpawnFlags)
	if err == sql.ErrNoRows {
		return CreatureSpawn{}, false, nil
	}
	if err != nil {
		return CreatureSpawn{}, false, fmt.Errorf("query creature spawn by id: %w", err)
	}
	return spawn, true, nil
}

func (s *Store) CreatureModelInfo(modelID int64) (CreatureModelInfo, bool, error) {
	var info CreatureModelInfo
	err := s.db.QueryRow(`SELECT modelid, bounding_radius, combat_reach, gender FROM creature_model_info WHERE modelid = ?`, modelID).Scan(&info.ModelID, &info.BoundingRadius, &info.CombatReach, &info.Gender)
	if err == sql.ErrNoRows {
		return CreatureModelInfo{}, false, nil
	}
	if err != nil {
		return CreatureModelInfo{}, false, fmt.Errorf("query creature model info: %w", err)
	}
	return info, true, nil
}
