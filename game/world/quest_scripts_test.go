package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
)

func TestQuestStartScriptUsesGiverSource(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, unit_class, level_min, health_multiplier, mana_multiplier) VALUES (100, 20, 'Quest Giver', 1, 1, 1, 1); INSERT INTO creature_classlevelstats (class, level, health, mana, base_health, base_mana) VALUES (1, 1, 40, 0, 40, 0); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z, health_percent, mana_percent) VALUES (7, 100, 0, 10, 20, 30, 100, 100); INSERT INTO quest_start_scripts (id, command, datalong, datalong2, x, y, z, o) VALUES (701, 10, 100, 1000, 10, 20, 30, 1)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: 1, Map: 0, Health: 100, PositionX: 10, PositionY: 20, PositionZ: 30}
	giverGUID := uint64(0xf130000000000007)
	server.startQuestScript(active, int64(giverGUID), 701, true)
	server.creatures.mu.Lock()
	defer server.creatures.mu.Unlock()
	if len(server.creatures.active) != 1 {
		t.Fatalf("creatures=%d", len(server.creatures.active))
	}
	for _, state := range server.creatures.active {
		if state.Spawn.PositionX != 10 || state.Spawn.PositionY != 20 || state.OwnerGUID != giverGUID {
			t.Fatalf("state=%#v", state)
		}
	}
}
