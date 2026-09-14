package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
)

func TestDuelRequestAcceptCancel(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO gameobject_template (entry, type, displayId, name, faction, size) VALUES (21680, 16, 327, 'Duel Flag', 5, 1)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{WorldData: worlddb.NewStore(databases)}
	caster := realm.Character{GUID: 1, Name: "Caster", Map: 0, PositionX: 0, PositionY: 0, PositionZ: 0}
	target := realm.Character{GUID: 2, Name: "Target", Map: 0, PositionX: 1, PositionY: 0, PositionZ: 0}
	server.registerPlayer(caster)
	server.registerPlayer(target)
	server.requestDuel(&spellCast{caster: caster}, target, dbc.SpellEffect{MiscValue: 21680})
	state := server.duelForPlayer(caster.GUID)
	if state == nil || state.guid&0xffff000000000000 != duelArbiterGUIDHigh || server.duelTarget(caster.GUID) != target.GUID {
		t.Fatalf("duel state=%#v target=%d", state, server.duelTarget(caster.GUID))
	}
	server.duelAccept(caster.GUID)
	server.duelAccept(target.GUID)
	state = server.duelForPlayer(caster.GUID)
	if state == nil || state.phase != duelStateStarted || server.duelForPlayer(target.GUID) == nil {
		t.Fatalf("started state=%#v", state)
	}
	server.duelCancel(caster.GUID)
	if server.duelForPlayer(caster.GUID) != nil || server.duelForPlayer(target.GUID) != nil {
		t.Fatal("duel state remains after cancel")
	}
}
