package world

import (
	"testing"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
)

func TestDummySpellDisplayEffect(t *testing.T) {
	server := &WorldServer{}
	target := realm.Character{GUID: 2, Name: "Target", Race: 1, Class: 1, Health: 1}
	server.registerPlayer(target)
	server.dummySpellEffect(&spellCast{spell: dbc.Spell{ID: 6236}}, target)
	display, found := server.playerDisplayID(target.GUID)
	if !found || display != 2279 {
		t.Fatalf("display=%d found=%v", display, found)
	}
}
