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

func TestItemLootLifecycle(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id, flags, stackable) VALUES (100, 'Container', 10, 4, 1), (200, 'Loot', 20, 0, 20); INSERT INTO item_loot_template (entry, item, ChanceOrQuestChance, mincountOrRef, maxcount) VALUES (100, 200, 100, 1, 2)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Looter", Health: 20})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddInventoryItemAt(guid, 100, 23, 23, 1); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Looter", Health: 20, Map: 0}
	responses, err := server.openItem(active, []byte{0xff, 23})
	if err != nil || len(responses) != 2 {
		t.Fatalf("open responses=%d err=%v", len(responses), err)
	}
	lootGUID := uint64(0)
	if item, found, itemErr := characters.ItemAt(guid, 23, 23); itemErr != nil || !found {
		t.Fatalf("container found=%v err=%v", found, itemErr)
	} else {
		lootGUID = uint64(item.GUID) | 0x4000000000000000
	}
	loot, err := packet.Parse(responses[1])
	if err != nil || loot.Opcode != packet.SMSGLootResponse || binary.LittleEndian.Uint64(loot.Data) != lootGUID || loot.Data[16] != 1 || binary.LittleEndian.Uint32(loot.Data[18:]) != 200 {
		t.Fatalf("loot=%#v err=%v", loot, err)
	}
	responses, err = server.lootItem(active, []byte{0})
	if err != nil || len(responses) != 4 {
		t.Fatalf("autostore responses=%d err=%v", len(responses), err)
	}
	if item, found, itemErr := characters.ItemAt(guid, 23, 24); itemErr != nil || !found || item.ItemTemplate != 200 {
		t.Fatalf("looted item=%#v found=%v err=%v", item, found, itemErr)
	}
	if responses, err = server.lootRelease(active, encodeUint64(lootGUID)); err != nil || len(responses) != 1 {
		t.Fatalf("release responses=%d err=%v", len(responses), err)
	}
	if _, found, err := characters.ItemAt(guid, 23, 23); err != nil || found {
		t.Fatalf("container remains found=%v err=%v", found, err)
	}
}

func TestGameObjectLootLifecycle(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id, stackable) VALUES (200, 'Chest Loot', 20, 1); INSERT INTO gameobject_template (entry, type, displayId, name, data1, mingold, maxgold) VALUES (300, 3, 10, 'Chest', 500, 5, 5); INSERT INTO gameobject_loot_template (entry, item, ChanceOrQuestChance, mincountOrRef, maxcount) VALUES (500, 200, 100, 1, 1); INSERT INTO spawns_gameobjects (spawn_id, spawn_entry, spawn_map, spawn_positionX, spawn_positionY, spawn_positionZ) VALUES (1, 300, 0, 0, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: realm.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: 1, Map: 0, Health: 20}
	guid := uint64(0xf110000000000001)
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, guid)
	responses, err := server.gameObjectUse(active, data)
	if err != nil || len(responses) != 2 {
		t.Fatalf("use responses=%d err=%v", len(responses), err)
	}
	loot, err := packet.Parse(responses[1])
	if err != nil || loot.Opcode != packet.SMSGLootResponse || binary.LittleEndian.Uint32(loot.Data[12:]) != 5 || loot.Data[16] != 1 {
		t.Fatalf("loot=%#v err=%v", loot, err)
	}
	responses, err = server.lootMoney(&active)
	if err != nil || len(responses) != 2 || active.Money != 5 {
		t.Fatalf("money responses=%d money=%d err=%v", len(responses), active.Money, err)
	}
	if clear, err := packet.Parse(responses[1]); err != nil || clear.Opcode != packet.SMSGLootClearMoney {
		t.Fatalf("clear=%#v err=%v", clear, err)
	}
	responses, err = server.lootRelease(active, data)
	if err != nil || len(responses) != 3 {
		t.Fatalf("release responses=%d err=%v", len(responses), err)
	}
	state := server.gameObjectStateFor(guid, worlddb.GameObjectSpawn{})
	if state.flags&gameObjectFlagInUse != 0 || state.state != gameObjectStateReady {
		t.Fatalf("state=%#v", state)
	}
}
