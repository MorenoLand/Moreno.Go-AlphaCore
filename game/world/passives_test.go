package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestInitializePassiveSpell(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, Attributes, Effect_1, EffectBasePoints_1, EffectAura_1) VALUES (42, 64, 6, 19, 34)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Player", Class: 1, Level: 1, Health: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(guid, 42); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Player", Class: 1, Level: 1, Health: 100}
	server.registerPlayer(active)
	if err := server.initializePassiveSpells(active); err != nil {
		t.Fatal(err)
	}
	server.auras.mu.Lock()
	aura, found := server.auras.active[guid][56]
	server.auras.mu.Unlock()
	if !found || aura.spellID != 42 || server.playerMaxHealth(guid) != 119 {
		t.Fatalf("aura=%#v found=%v max=%d", aura, found, server.playerMaxHealth(guid))
	}
	values := buildPlayerFields(active, dbc.Race{}, nil)
	server.playerAuraFields(values, guid)
	if values[56] != 0 || values[112] != 0 {
		t.Fatalf("passive fields aura=%d flags=%d", values[56], values[112])
	}
	if packet.SpellEffect(aura.effect.Type) != packet.SpellEffectApplyAura {
		t.Fatalf("effect=%d", aura.effect.Type)
	}
}
