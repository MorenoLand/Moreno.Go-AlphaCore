package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
)

func TestGroupAstralRecallScript(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Traveler", Map: 2, PositionX: 10, PositionY: 20, PositionZ: 30, Health: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.SaveDeathbind(realm.Deathbind{PlayerGUID: guid, Map: 1, X: 2, Y: 3, Z: 4}); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters}
	target := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Traveler", Map: 2, PositionX: 10, PositionY: 20, PositionZ: 30, Health: 1}
	server.registerPlayer(target)
	server.scriptSpellEffect(&spellCast{spell: dbc.Spell{ID: 966}}, target)
	updated, found := server.playerByGUID(guid)
	if !found || updated.Map != 1 || updated.PositionX != 2 || updated.PositionY != 3 || updated.PositionZ != 4 {
		t.Fatalf("updated=%#v found=%v", updated, found)
	}
}
