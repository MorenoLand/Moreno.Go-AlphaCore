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

func TestCreatureLootUsesTemplateIDsGoldAndSkinning(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id, stackable) VALUES (200, 'Creature Drop', 20, 1), (201, 'Skinning Drop', 21, 1); INSERT INTO creature_template (entry, display_id1, name, unit_class, level_min, health_multiplier, mana_multiplier, loot_id, skinning_loot_id, gold_min, gold_max) VALUES (100, 20, 'Loot Creature', 1, 1, 1, 1, 500, 600, 7, 7); INSERT INTO creature_loot_template (entry, item, ChanceOrQuestChance, mincountOrRef, maxcount) VALUES (500, 200, 100, 1, 1); INSERT INTO skinning_loot_template (entry, item, ChanceOrQuestChance, mincountOrRef, maxcount) VALUES (600, 201, 100, 1, 1)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{WorldData: worlddb.NewStore(databases), Characters: realm.NewStore(databases)}
	active := realm.Character{GUID: 1, Map: 0, Health: 20}
	responses, err := server.sendLoot(active, 0xf130000000000001, lootCreatureSource, 100, worlddb.GameObjectTemplate{}, 0)
	if err != nil || len(responses) != 2 {
		t.Fatalf("responses=%d err=%v", len(responses), err)
	}
	loot, err := packet.Parse(responses[1])
	if err != nil || loot.Opcode != packet.SMSGLootResponse || binary.LittleEndian.Uint32(loot.Data[12:]) != 7 || loot.Data[16] != 2 {
		t.Fatalf("loot=%#v err=%v", loot, err)
	}
	entries := []uint32{binary.LittleEndian.Uint32(loot.Data[18:]), binary.LittleEndian.Uint32(loot.Data[31:])}
	if !((entries[0] == 200 && entries[1] == 201) || (entries[0] == 201 && entries[1] == 200)) {
		t.Fatalf("items=%#v", loot.Data)
	}
}
