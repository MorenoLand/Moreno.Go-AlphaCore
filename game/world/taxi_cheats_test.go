package world

import (
	"context"
	"strings"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
)

func TestTaxiEnableAndClear(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "GM", Race: 1, Map: 0})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO ChrRaces (ID, BaseLanguage) VALUES (1, 7); INSERT INTO TaxiNodes (ID, ContinentID, custom_Team) VALUES (1, 0, 469), (2, 0, 67)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases)}
	active := &realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Race: 1, Map: 0}
	if err := server.taxiEnableAll(active, true, 1); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(active.Taximask, "10") {
		t.Fatalf("enabled mask=%q", active.Taximask)
	}
	if err := server.taxiEnableAll(active, false, 1); err != nil {
		t.Fatal(err)
	}
	if strings.Trim(active.Taximask, "0") != "" {
		t.Fatalf("cleared mask=%q", active.Taximask)
	}
}
