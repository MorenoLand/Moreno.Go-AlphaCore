package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestLeapSpellEffect(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	casterGUID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Health: 100})
	if err != nil {
		t.Fatal(err)
	}
	targetGUID, err := characters.Create(realm.Character{AccountID: 2, RealmID: 1, Name: "Target", Health: 100, PositionX: 10})
	if err != nil {
		t.Fatal(err)
	}
	caster := realm.Character{GUID: casterGUID, AccountID: 1, RealmID: 1, Name: "Caster", Health: 100}
	target := realm.Character{GUID: targetGUID, AccountID: 2, RealmID: 1, Name: "Target", Health: 100, PositionX: 10}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	server.registerPlayer(caster)
	server.registerPlayer(target)
	server.applySpellEffects(&spellCast{caster: caster, target: spellTarget{UnitGUID: uint64(targetGUID)}, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectLeap)}}}})
	updated, found := server.playerByGUID(targetGUID)
	if !found {
		t.Fatal("target player missing")
	}
	leaper, found := server.playerByGUID(casterGUID)
	if !found || leaper.PositionX <= 0 || leaper.PositionX >= updated.PositionX {
		t.Fatalf("leaper=%#v target=%#v", leaper, updated)
	}
	server.applySpellEffects(&spellCast{caster: leaper, target: spellTarget{Dest: &spellVector{X: 20, Y: 30, Z: 40}}, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectLeap)}}}})
	leaper, found = server.playerByGUID(casterGUID)
	if !found || leaper.PositionX != 20 || leaper.PositionY != 30 || leaper.PositionZ != 40 {
		t.Fatalf("terrain leaper=%#v", leaper)
	}
}
