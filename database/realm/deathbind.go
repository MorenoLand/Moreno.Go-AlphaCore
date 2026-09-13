package realm

import (
	"database/sql"
	"fmt"
)

type Deathbind struct {
	PlayerGUID, CreatureBinderGUID, Map, Zone int64
	X, Y, Z                                   float32
}

func (s *Store) Deathbind(guid int64) (Deathbind, bool, error) {
	var bind Deathbind
	err := s.db.QueryRow(`SELECT player_guid, creature_binder_guid, deathbind_map, deathbind_zone, deathbind_position_x, deathbind_position_y, deathbind_position_z FROM character_deathbind WHERE player_guid = ?`, guid).Scan(&bind.PlayerGUID, &bind.CreatureBinderGUID, &bind.Map, &bind.Zone, &bind.X, &bind.Y, &bind.Z)
	if err == sql.ErrNoRows {
		return Deathbind{}, false, nil
	}
	if err != nil {
		return Deathbind{}, false, fmt.Errorf("query deathbind: %w", err)
	}
	return bind, true, nil
}

func (s *Store) SaveDeathbind(bind Deathbind) error {
	_, err := s.db.Exec(`INSERT INTO character_deathbind (player_guid, creature_binder_guid, deathbind_map, deathbind_zone, deathbind_position_x, deathbind_position_y, deathbind_position_z) VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(player_guid) DO UPDATE SET creature_binder_guid = excluded.creature_binder_guid, deathbind_map = excluded.deathbind_map, deathbind_zone = excluded.deathbind_zone, deathbind_position_x = excluded.deathbind_position_x, deathbind_position_y = excluded.deathbind_position_y, deathbind_position_z = excluded.deathbind_position_z`, bind.PlayerGUID, bind.CreatureBinderGUID, bind.Map, bind.Zone, bind.X, bind.Y, bind.Z)
	return err
}
