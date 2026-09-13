package world

import (
	"database/sql"
	"fmt"
)

type GameObjectSpawn struct {
	SpawnID, Entry, Map, State, AnimProgress, Flags int64
	PositionX, PositionY, PositionZ, Orientation    float32
	Rotation0, Rotation1, Rotation2, Rotation3      float32
}

func (s *Store) GameObjectSpawns(mapID int64, x, y, z, distance float32) ([]GameObjectSpawn, error) {
	rows, err := s.db.Query(`SELECT spawn_id, spawn_entry, spawn_map, spawn_positionX, spawn_positionY, spawn_positionZ, spawn_orientation, spawn_rotation0, spawn_rotation1, spawn_rotation2, spawn_rotation3, spawn_animprogress, spawn_state, spawn_flags FROM spawns_gameobjects WHERE spawn_map = ? AND ignored = 0 AND spawn_entry > 0 AND spawn_positionX BETWEEN ? AND ? AND spawn_positionY BETWEEN ? AND ? ORDER BY spawn_id`, mapID, x-distance, x+distance, y-distance, y+distance)
	if err != nil {
		return nil, fmt.Errorf("query gameobject spawns: %w", err)
	}
	defer rows.Close()
	spawns := make([]GameObjectSpawn, 0)
	for rows.Next() {
		var spawn GameObjectSpawn
		if err := rows.Scan(&spawn.SpawnID, &spawn.Entry, &spawn.Map, &spawn.PositionX, &spawn.PositionY, &spawn.PositionZ, &spawn.Orientation, &spawn.Rotation0, &spawn.Rotation1, &spawn.Rotation2, &spawn.Rotation3, &spawn.AnimProgress, &spawn.State, &spawn.Flags); err != nil {
			return nil, fmt.Errorf("scan gameobject spawn: %w", err)
		}
		dx, dy, dz := spawn.PositionX-x, spawn.PositionY-y, spawn.PositionZ-z
		if dx*dx+dy*dy+dz*dz <= distance*distance {
			spawns = append(spawns, spawn)
		}
	}
	return spawns, rows.Err()
}

func (s *Store) HasGameObjectTemplate(entry int64) (bool, error) {
	var value int
	err := s.db.QueryRow(`SELECT 1 FROM gameobject_template WHERE entry = ?`, entry).Scan(&value)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
