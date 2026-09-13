package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestInventoryMutations(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, stackable, inventory_type) VALUES (1, 'Potion', 20, 0), (2, 'Sword', 1, 13)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO page_text (entry, text, next_page) VALUES (1, 'Readable', 0); UPDATE item_template SET page_text = 1, page_language = 7 WHERE entry = 1`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Inventory", Race: 1, Class: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddInventoryItem(guid, 1, 23, 5); err != nil {
		t.Fatal(err)
	}
	if err := characters.AddInventoryItem(guid, 2, 25, 1); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, Name: "Inventory", Race: 1, Class: 1}
	readResponse, err := server.readItem(active, []byte{23, 23})
	parsed, parseErr := packet.Parse(readResponse)
	if err != nil || parseErr != nil || parsed.Opcode != packet.SMSGReadItemOK || len(parsed.Data) != 8 {
		t.Fatalf("read item=%#v err=%v parse=%v", parsed, err, parseErr)
	}
	if _, err := databases.DB(database.World).Exec(`UPDATE item_template SET page_language = 1 WHERE entry = 1`); err != nil {
		t.Fatal(err)
	}
	readResponse, err = server.readItem(active, []byte{23, 23})
	parsed, parseErr = packet.Parse(readResponse)
	if err != nil || parseErr != nil || parsed.Opcode != packet.SMSGReadItemFailed || len(parsed.Data) != 9 || parsed.Data[8] != 0 {
		t.Fatalf("foreign read item=%#v err=%v parse=%v", parsed, err, parseErr)
	}
	splitData := []byte{23, 23, 23, 24, 2}
	if response, err := server.splitItem(active, splitData); response != nil || err != nil {
		t.Fatalf("split response=%v err=%v", response, err)
	}
	source, found, err := characters.ItemAt(guid, 23, 23)
	if err != nil || !found || source.StackCount != 3 {
		t.Fatalf("split source=%#v found=%v err=%v", source, found, err)
	}
	destination, found, err := characters.ItemAt(guid, 23, 24)
	if err != nil || !found || destination.StackCount != 2 {
		t.Fatalf("split destination=%#v found=%v err=%v", destination, found, err)
	}
	if err := server.swapInventory(active, []byte{23, 24}); err != nil {
		t.Fatal(err)
	}
	if item, _, err := characters.ItemAt(guid, 23, 23); err != nil || item.StackCount != 2 {
		t.Fatalf("swap item=%#v err=%v", item, err)
	}
	if response, err := server.autoequipItem(active, []byte{23, 25}); response != nil || err != nil {
		t.Fatalf("autoequip response=%v err=%v", response, err)
	}
	if item, _, err := characters.ItemAt(guid, 23, 15); err != nil || item.ItemTemplate != 2 {
		t.Fatalf("equipped item=%#v err=%v", item, err)
	}
	if err := characters.AddInventoryItem(guid, 1, 26, 1); err != nil {
		t.Fatal(err)
	}
	if response, err := server.autostoreItem(active, []byte{23, 26, 23}); response != nil || err != nil {
		t.Fatalf("autostore response=%v err=%v", response, err)
	}
	if _, found, err := characters.ItemAt(guid, 23, 25); err != nil || !found {
		t.Fatalf("autostore found=%v err=%v", found, err)
	}
	response, err := server.destroyItem(active, []byte{23, 25, 0, 0, 0, 0})
	if err != nil || len(response) != 1 {
		t.Fatalf("destroy response=%d err=%v", len(response), err)
	}
	parsed, parseErr = packet.Parse(response[0])
	if err != nil || parsed.Opcode != packet.SMSGDestroyObject {
		t.Fatalf("destroy packet=%#v err=%v", parsed, err)
	}
	if _, found, err := characters.ItemAt(guid, 23, 25); err != nil || found {
		t.Fatalf("destroy found=%v err=%v", found, err)
	}
	failure, err := server.splitItem(active, []byte{23, 23, 23, 28, 9})
	parsed, parseErr = packet.Parse(failure)
	if err != nil || parseErr != nil || parsed.Opcode != packet.SMSGInventoryChangeFailure || parsed.Data[0] != bagItemTooFew {
		t.Fatalf("split failure=%#v err=%v parse=%v", parsed, err, parseErr)
	}
}
