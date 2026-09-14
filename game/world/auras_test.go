package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestAuraSlotsFlagsAndCancel(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO SpellDuration (ID, Duration, MaxDuration) VALUES (1, 1000, 1000)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: realm.NewStore(databases), DBC: dbc.NewStore(databases)}
	active := realm.Character{GUID: 1, AccountID: 1, RealmID: 1, Map: 0, Health: 10}
	server.registerPlayer(active)
	spell := dbc.Spell{ID: 42, DurationIndex: 1}
	server.applyAura(&spellCast{caster: active, spell: spell}, active, 0, dbc.SpellEffect{Type: int64(packet.SpellEffectApplyAura)})
	aura := server.auras.active[active.GUID][0]
	if aura == nil || aura.spellID != 42 || aura.duration != 1000 || !aura.cancelable || server.auraFlags(active.GUID)[0] != uint32(packet.AuraFlagCancelable|packet.AuraFlagEffect0) {
		t.Fatalf("aura=%#v flags=%x", aura, server.auraFlags(active.GUID))
	}
	if responses, err := server.cancelAura(active, encodeUint32(42)); err != nil || len(responses) != 0 {
		t.Fatalf("responses=%d err=%v", len(responses), err)
	}
	if _, found := server.auras.active[active.GUID]; found {
		t.Fatal("aura remained after cancel")
	}
}

func TestHarmfulAuraUsesHarmfulSlots(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	server := &WorldServer{Characters: realm.NewStore(databases), DBC: dbc.NewStore(databases)}
	active := realm.Character{GUID: 1, AccountID: 1, RealmID: 1, Map: 0, Health: 10}
	server.registerPlayer(active)
	spell := dbc.Spell{ID: 43, Attributes: int64(packet.SpellAttributeAuraDebuff)}
	server.applyAura(&spellCast{caster: active, spell: spell}, active, 0, dbc.SpellEffect{Type: int64(packet.SpellEffectApplyAura)})
	aura := server.auras.active[active.GUID][32]
	if aura == nil || aura.cancelable || server.auraFlags(active.GUID)[4] != uint32(packet.AuraFlagEffect0)<<0 {
		t.Fatalf("aura=%#v flags=%x", aura, server.auraFlags(active.GUID))
	}
}
