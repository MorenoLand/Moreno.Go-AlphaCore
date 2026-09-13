package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
)

func TestBuyBankSlot(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO BankBagSlotPrices (ID, Cost) VALUES (1, 100)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, npc_flags) VALUES (100, 20, 'Banker', 32); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z) VALUES (1, 100, 0, 0, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Banker", Health: 1, Money: 200})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Banker", Health: 1, Money: 200, Map: 0}
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, 0xf130000000000001)
	responses, err := server.buyBankSlot(&active, data)
	if err != nil || len(responses) != 2 || active.Bankslots != 1 || active.Money != 100 {
		t.Fatalf("responses=%d slots=%d money=%d err=%v", len(responses), active.Bankslots, active.Money, err)
	}
	stored, found, err := characters.CharacterByGUID(guid)
	if err != nil || !found || stored.Bankslots != 1 || stored.Money != 100 {
		t.Fatalf("stored=%#v found=%v err=%v", stored, found, err)
	}
}
