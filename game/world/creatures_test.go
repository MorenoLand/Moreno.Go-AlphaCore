package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestNearbyCreaturePackets(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, level_min, level_max, faction, scale, unit_class, unit_flags, npc_flags, type, beast_family, health_multiplier, mana_multiplier, armor_multiplier, damage_multiplier, damage_variance, base_attack_time, ranged_attack_time) VALUES (100, 123, 'Wolf', 2, 2, 14, 1, 1, 0, 0, 1, 1, 1, 1, 1, 1, 0.14, 2000, 2000)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_classlevelstats ("class", level, melee_damage, ranged_damage, attack_power, ranged_attack_power, health, base_health, mana, base_mana, strength, agility, stamina, intellect, spirit, armor) VALUES (1, 2, 2.3, 2.3, 4, 4, 55, 29, 0, 0, 20, 21, 23, 20, 20, 10)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_model_info (modelid, bounding_radius, combat_reach, gender) VALUES (123, 0.3, 1.5, 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z, orientation) VALUES (1, 100, 0, 10, 20, 30, 1), (2, 100, 0, 200, 200, 200, 1)`); err != nil {
		t.Fatal(err)
	}
	world := worlddb.NewStore(databases)
	spawns, err := world.CreatureSpawns(0, 10, 20, 30, 100)
	if err != nil || len(spawns) != 1 || spawns[0].SpawnID != 1 {
		t.Fatalf("spawns=%#v err=%v", spawns, err)
	}
	server := &WorldServer{WorldData: world}
	packets, err := server.nearbyCreaturePackets(realm.Character{Map: 0, PositionX: 10, PositionY: 20, PositionZ: 30})
	if err != nil || len(packets) != 1 {
		t.Fatalf("creature packets=%d err=%v", len(packets), err)
	}
	parsed, err := packet.Parse(packets[0])
	if err != nil || parsed.Opcode != packet.SMSGCompressedUpdateObject {
		t.Fatalf("creature packet=%#v err=%v", parsed, err)
	}
}
