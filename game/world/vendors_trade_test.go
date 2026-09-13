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

func TestVendorBuyAndSell(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Buyer", Money: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id, buy_price, sell_price, buy_count, max_durability, stackable) VALUES (200, 'Test Item', 12, 10, 10, 1, 40, 20); INSERT INTO creature_template (entry, display_id1, name, npc_flags, vendor_id) VALUES (100, 20, 'Vendor', 1, 0); INSERT INTO npc_vendor (entry, item, maxcount, slot) VALUES (100, 200, 0, 0); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z) VALUES (1, 100, 0, 1, 1, 1)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Buyer", Money: 1000}
	vendorGUID := uint64(0xf110000000000001)
	buyData := make([]byte, 14)
	binary.LittleEndian.PutUint64(buyData, vendorGUID)
	binary.LittleEndian.PutUint32(buyData[8:], 200)
	buyData[12] = 1
	responses, err := server.buyItem(&active, buyData, false)
	if err != nil || len(responses) != 3 {
		t.Fatalf("buy responses=%d err=%v", len(responses), err)
	}
	push, err := packet.Parse(responses[1])
	if err != nil || push.Opcode != packet.SMSGItemPushResult || len(push.Data) != 21 || push.Data[16] != 0xff || binary.LittleEndian.Uint32(push.Data[17:]) != 200 {
		t.Fatalf("push=%#v err=%v", push, err)
	}
	item, found, err := characters.ItemAt(guid, 23, 23)
	if err != nil || !found || item.ItemTemplate != 200 || item.StackCount != 1 || active.Money != 990 {
		t.Fatalf("bought=%#v found=%v money=%d err=%v", item, found, active.Money, err)
	}
	sellData := make([]byte, 17)
	binary.LittleEndian.PutUint64(sellData, vendorGUID)
	binary.LittleEndian.PutUint64(sellData[8:], uint64(item.GUID))
	sellData[16] = 1
	responses, err = server.sellItem(&active, sellData)
	if err != nil || len(responses) != 2 {
		t.Fatalf("sell responses=%d err=%v", len(responses), err)
	}
	if _, found, err := characters.ItemByGUID(guid, item.GUID); err != nil || found || active.Money != 1000 {
		t.Fatalf("sold found=%v money=%d err=%v", found, active.Money, err)
	}
}

func TestVendorErrorsDistinguishNonVendor(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id, buy_price, buy_count) VALUES (200, 'Test Item', 12, 10, 1); INSERT INTO creature_template (entry, display_id1, name, npc_flags) VALUES (100, 20, 'Creature', 0); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z) VALUES (1, 100, 0, 1, 1, 1)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: 1, Map: 0}
	vendorGUID := uint64(0xf110000000000001)
	buyData := make([]byte, 14)
	binary.LittleEndian.PutUint64(buyData, vendorGUID)
	binary.LittleEndian.PutUint32(buyData[8:], 200)
	responses, err := server.buyItem(&active, buyData, false)
	if err != nil || len(responses) != 1 {
		t.Fatalf("buy responses=%d err=%v", len(responses), err)
	}
	buyFailure, err := packet.Parse(responses[0])
	if err != nil || buyFailure.Opcode != packet.SMSGBuyFailed || buyFailure.Data[len(buyFailure.Data)-1] != 11 {
		t.Fatalf("buy failure=%#v err=%v", buyFailure, err)
	}
	sellData := make([]byte, 17)
	binary.LittleEndian.PutUint64(sellData, vendorGUID)
	responses, err = server.sellItem(&active, sellData)
	if err != nil || len(responses) != 1 {
		t.Fatalf("sell responses=%d err=%v", len(responses), err)
	}
	sellFailure, err := packet.Parse(responses[0])
	if err != nil || sellFailure.Opcode != packet.SMSGSellItem || sellFailure.Data[len(sellFailure.Data)-1] != 2 {
		t.Fatalf("sell failure=%#v err=%v", sellFailure, err)
	}
}
