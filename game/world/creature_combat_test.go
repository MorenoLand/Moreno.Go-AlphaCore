package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestMakeMonsterAttackMe(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, faction, unit_class, level_min, health_multiplier, mana_multiplier) VALUES (100, 20, 'Wolf', 1, 1, 1, 1, 1); INSERT INTO creature_classlevelstats (class, level, health, mana, base_health, base_mana) VALUES (1, 1, 40, 0, 40, 0); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z, health_percent, mana_percent) VALUES (7, 100, 0, 1, 2, 3, 100, 100)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: 42, Map: 0, PositionX: 1, PositionY: 2, PositionZ: 3, Health: 10}
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, 0xf130000000000007)
	responses, err := server.makeMonsterAttackMe(active, data)
	if err != nil || len(responses) != 1 {
		t.Fatalf("responses=%d err=%v", len(responses), err)
	}
	message, err := packet.Parse(responses[0])
	if err != nil || message.Opcode != packet.SMSGAttackStart || len(message.Data) != 16 || binary.LittleEndian.Uint64(message.Data) != 0xf130000000000007 || binary.LittleEndian.Uint64(message.Data[8:]) != 42 {
		t.Fatalf("message=%#v err=%v", message, err)
	}
	state, found, err := server.creatureStateAt(active, 0xf130000000000007, creatureViewDistance)
	if err != nil || !found || state.CombatTarget != 42 {
		t.Fatalf("state=%#v found=%v err=%v", state, found, err)
	}
}
