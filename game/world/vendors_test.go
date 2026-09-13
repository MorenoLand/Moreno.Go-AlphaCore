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

func TestListInventory(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id, buy_price, buy_count, max_durability) VALUES (200, 'Test Item', 12, 99, 1, 40); INSERT INTO creature_template (entry, display_id1, name, npc_flags, vendor_id) VALUES (100, 20, 'Vendor', 1, 0); INSERT INTO npc_vendor (entry, item, maxcount, slot) VALUES (100, 200, 0, 0); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z) VALUES (1, 100, 0, 1, 1, 1)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{WorldData: worlddb.NewStore(databases)}
	active := realm.Character{Map: 0, PositionX: 0, PositionY: 0, PositionZ: 0}
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, 0xf110000000000001)
	responses, err := server.listInventory(active, data)
	if err != nil || len(responses) != 2 {
		t.Fatalf("responses=%d err=%v", len(responses), err)
	}
	list, err := packet.Parse(responses[1])
	if err != nil || list.Opcode != packet.SMSGListInventory || len(list.Data) != 37 || binary.LittleEndian.Uint64(list.Data) != 0xf110000000000001 || list.Data[8] != 1 || binary.LittleEndian.Uint32(list.Data[9:]) != 1 || binary.LittleEndian.Uint32(list.Data[13:]) != 200 || binary.LittleEndian.Uint32(list.Data[17:]) != 12 || binary.LittleEndian.Uint32(list.Data[21:]) != 0xffffffff || binary.LittleEndian.Uint32(list.Data[25:]) != 99 || binary.LittleEndian.Uint32(list.Data[29:]) != 40 {
		t.Fatalf("list=%#v err=%v", list, err)
	}
}
