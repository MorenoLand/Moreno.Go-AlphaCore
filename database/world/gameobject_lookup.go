package world

import "database/sql"

func (s *Store) GameObjectSpawnByID(id int64) (GameObjectSpawn, bool, error) {
	var spawn GameObjectSpawn
	err := s.db.QueryRow(`SELECT spawn_id, spawn_entry, spawn_map, spawn_positionX, spawn_positionY, spawn_positionZ, spawn_orientation, spawn_rotation0, spawn_rotation1, spawn_rotation2, spawn_rotation3, spawn_animprogress, spawn_state, spawn_flags FROM spawns_gameobjects WHERE spawn_id = ?`, id).Scan(&spawn.SpawnID, &spawn.Entry, &spawn.Map, &spawn.PositionX, &spawn.PositionY, &spawn.PositionZ, &spawn.Orientation, &spawn.Rotation0, &spawn.Rotation1, &spawn.Rotation2, &spawn.Rotation3, &spawn.AnimProgress, &spawn.State, &spawn.Flags)
	if err == sql.ErrNoRows {
		return GameObjectSpawn{}, false, nil
	}
	return spawn, err == nil, err
}
