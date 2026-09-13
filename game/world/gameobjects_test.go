package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestNearbyGameObjectPackets(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO gameobject_template (entry, type, displayId, name, faction, flags, size) VALUES (200, 3, 10, 'Chest', 1, 4, 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO spawns_gameobjects (spawn_id, spawn_entry, spawn_map, spawn_positionX, spawn_positionY, spawn_positionZ, spawn_orientation, spawn_rotation2, spawn_rotation3, spawn_animprogress, spawn_state, spawn_flags) VALUES (1, 200, 0, 10, 20, 30, 1, 0.5, 0.5, 100, 1, 2), (2, 200, 0, 200, 200, 200, 1, 0, 0, 0, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	world := worlddb.NewStore(databases)
	spawns, err := world.GameObjectSpawns(0, 10, 20, 30, 100)
	if err != nil || len(spawns) != 1 || spawns[0].SpawnID != 1 {
		t.Fatalf("spawns=%#v err=%v", spawns, err)
	}
	server := &WorldServer{WorldData: world}
	packets, err := server.nearbyGameObjectPackets(realm.Character{Map: 0, PositionX: 10, PositionY: 20, PositionZ: 30})
	if err != nil || len(packets) != 1 {
		t.Fatalf("gameobject packets=%d err=%v", len(packets), err)
	}
	parsed, err := packet.Parse(packets[0])
	if err != nil || parsed.Opcode != packet.SMSGCompressedUpdateObject {
		t.Fatalf("gameobject packet=%#v err=%v", parsed, err)
	}
}
