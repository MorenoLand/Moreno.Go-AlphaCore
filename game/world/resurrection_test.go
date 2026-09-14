package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestResurrectionRequestAndResponse(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	casterID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Map: 1, PositionX: 10, PositionY: 20, PositionZ: 30, Orientation: 1, Health: 100})
	if err != nil {
		t.Fatal(err)
	}
	targetID, err := characters.Create(realm.Character{AccountID: 2, RealmID: 1, Name: "Target", Map: 1, Health: 100})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters}
	caster := realm.Character{GUID: casterID, AccountID: 1, RealmID: 1, Name: "Caster", Map: 1, PositionX: 10, PositionY: 20, PositionZ: 30, Orientation: 1, Health: 100}
	target := realm.Character{GUID: targetID, AccountID: 2, RealmID: 1, Name: "Target", Map: 1, Health: 0}
	server.registerPlayer(caster)
	server.registerPlayer(realm.Character{GUID: targetID, AccountID: 2, RealmID: 1, Name: "Target", Map: 1, Health: 100})
	server.updatePlayer(target)
	server.requestResurrection(&spellCast{caster: caster}, target, 50)
	if responses, err := server.resurrectResponse(&target, append(encodeGUID(casterID+1), 1)); err != nil || len(responses) != 0 {
		t.Fatalf("mismatched responses=%d err=%v", len(responses), err)
	}
	responses, err := server.resurrectResponse(&target, append(encodeGUID(casterID), 1))
	if err != nil || len(responses) != 3 || target.Health != 50 || target.Map != 1 || target.PositionX != 10 || target.PositionY != 20 || target.PositionZ != 30 {
		t.Fatalf("responses=%d target=%#v err=%v", len(responses), target, err)
	}
	if parsed, err := packet.Parse(responses[2]); err != nil || parsed.Opcode != packet.SMSGNewWorld {
		t.Fatalf("teleport=%#v err=%v", parsed, err)
	}
	stored, found, err := characters.Character(targetID, 2, 1)
	if err != nil || !found || stored.Health != 50 {
		t.Fatalf("stored=%#v found=%v err=%v", stored, found, err)
	}
}
