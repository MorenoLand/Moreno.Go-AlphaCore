package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
)

func TestWrapAndOpenItem(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Wrapper", Health: 20})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id, inventory_type, stackable) VALUES (100, 'Item', 10, 0, 1), (5014, 'Paper', 11, 0, 20), (5015, 'Gift', 12, 0, 1)`); err != nil {
		t.Fatal(err)
	}
	if err := characters.AddInventoryItemAt(guid, 5014, 23, 0, 2); err != nil {
		t.Fatal(err)
	}
	if err := characters.AddInventoryItemAt(guid, 100, 23, 24, 1); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, Name: "Wrapper", Health: 20}
	responses, err := server.wrapItem(active, []byte{0xff, 0, 0xff, 24})
	if err != nil || len(responses) != 6 {
		t.Fatalf("wrap responses=%d err=%v", len(responses), err)
	}
	target, found, err := characters.ItemAt(guid, 23, 24)
	if err != nil || !found || target.ItemTemplate != 5015 || target.Creator != guid || target.Flags != itemDynWrapped {
		t.Fatalf("wrapped item=%#v found=%v err=%v", target, found, err)
	}
	wrapper, found, err := characters.ItemAt(guid, 23, 0)
	if err != nil || !found || wrapper.StackCount != 1 {
		t.Fatalf("wrapper=%#v found=%v err=%v", wrapper, found, err)
	}
	if gift, found, err := characters.GiftByItemGUID(target.GUID); err != nil || !found || gift.Entry != 100 {
		t.Fatalf("gift=%#v found=%v err=%v", gift, found, err)
	}
	responses, err = server.openItem(active, []byte{0xff, 24})
	if err != nil || len(responses) != 5 {
		t.Fatalf("open responses=%d err=%v", len(responses), err)
	}
	target, found, err = characters.ItemAt(guid, 23, 24)
	if err != nil || !found || target.ItemTemplate != 100 || target.Flags != 0 || target.Creator != 0 {
		t.Fatalf("unwrapped item=%#v found=%v err=%v", target, found, err)
	}
	if _, found, err := characters.GiftByItemGUID(target.GUID); err != nil || found {
		t.Fatalf("gift remains found=%v err=%v", found, err)
	}
}
