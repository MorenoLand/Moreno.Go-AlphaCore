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

func TestTradeLifecycle(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id, stackable) VALUES (100, 'First Item', 10, 1), (200, 'Second Item', 20, 1)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	activeGUID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "First", Health: 1, Money: 1000})
	if err != nil {
		t.Fatal(err)
	}
	otherGUID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Second", Health: 1, Money: 500})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddInventoryItemAt(activeGUID, 100, 23, 23, 1); err != nil {
		t.Fatal(err)
	}
	if err := characters.AddInventoryItemAt(otherGUID, 200, 23, 23, 1); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: activeGUID, AccountID: 1, RealmID: 1, Name: "First", Health: 1, Money: 1000, Map: 0}
	other := realm.Character{GUID: otherGUID, AccountID: 1, RealmID: 1, Name: "Second", Health: 1, Money: 500, Map: 0}
	server.registerPlayer(active)
	server.registerPlayer(other)
	request := make([]byte, 8)
	binary.LittleEndian.PutUint64(request, uint64(otherGUID))
	if responses, err := server.initiateTrade(active, request); err != nil || len(responses) != 1 {
		t.Fatalf("initiate responses=%d err=%v", len(responses), err)
	}
	if responses, err := server.beginTrade(active); err != nil || len(responses) != 1 {
		t.Fatalf("begin responses=%d err=%v", len(responses), err)
	}
	if responses, err := server.setTradeItem(active, []byte{0, 0xff, 23}); err != nil || len(responses) != 2 {
		t.Fatalf("first item responses=%d err=%v", len(responses), err)
	}
	gold := make([]byte, 4)
	binary.LittleEndian.PutUint32(gold, 100)
	if responses, err := server.setTradeGold(active, gold); err != nil || len(responses) != 2 {
		t.Fatalf("first gold responses=%d err=%v", len(responses), err)
	}
	if responses, err := server.setTradeItem(other, []byte{0, 0xff, 23}); err != nil || len(responses) != 2 {
		t.Fatalf("second item responses=%d err=%v", len(responses), err)
	}
	binary.LittleEndian.PutUint32(gold, 50)
	if responses, err := server.setTradeGold(other, gold); err != nil || len(responses) != 2 {
		t.Fatalf("second gold responses=%d err=%v", len(responses), err)
	}
	if responses, err := server.acceptTrade(active); err != nil || responses != nil {
		t.Fatalf("first accept responses=%v err=%v", responses, err)
	}
	responses, err := server.acceptTrade(other)
	if err != nil || len(responses) == 0 {
		t.Fatalf("second accept responses=%d err=%v", len(responses), err)
	}
	firstItem, found, err := characters.ItemAt(activeGUID, 23, 24)
	if err != nil || !found || firstItem.ItemTemplate != 200 {
		t.Fatalf("first inventory=%#v found=%v err=%v", firstItem, found, err)
	}
	secondItem, found, err := characters.ItemAt(otherGUID, 23, 24)
	if err != nil || !found || secondItem.ItemTemplate != 100 {
		t.Fatalf("second inventory=%#v found=%v err=%v", secondItem, found, err)
	}
	activeStored, _, err := characters.CharacterByGUID(activeGUID)
	if err != nil || activeStored.Money != 950 {
		t.Fatalf("active money=%d err=%v", activeStored.Money, err)
	}
	otherStored, _, err := characters.CharacterByGUID(otherGUID)
	if err != nil || otherStored.Money != 550 {
		t.Fatalf("other money=%d err=%v", otherStored.Money, err)
	}
	if _, _, valid := server.tradePairLocked(activeGUID); valid {
		t.Fatal("trade remains active")
	}
	if packetValue, err := packet.Parse(responses[len(responses)-1]); err != nil || packetValue.Opcode != packet.SMSGTradeStatus || binary.LittleEndian.Uint32(packetValue.Data) != tradeComplete {
		t.Fatalf("complete=%#v err=%v", packetValue, err)
	}
}
