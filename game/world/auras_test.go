package world

import (
	"context"
	"testing"
	"time"

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

func TestPeriodicAuraChangesHealth(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO SpellDuration (ID, Duration, MaxDuration) VALUES (1, 100, 100)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Map: 0, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Caster", Map: 0, Health: 10}
	server.registerPlayer(active)
	server.setPlayerMaxHealth(guid, 100)
	spell := dbc.Spell{ID: 44, DurationIndex: 1}
	effect := dbc.SpellEffect{Type: int64(packet.SpellEffectApplyAura), Aura: int64(packet.AuraPeriodicHeal), AuraPeriod: 5, BasePoints: 2}
	server.applyAura(&spellCast{caster: active, spell: spell}, active, 0, effect)
	defer server.removeAura(active, 0)
	deadline := time.After(500 * time.Millisecond)
	for {
		stored, found, err := characters.Character(guid, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if found && stored.Health >= 12 {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("periodic health=%d found=%v", stored.Health, found)
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func TestControlAuraUpdatesUnitFlags(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	server := &WorldServer{DBC: dbc.NewStore(databases)}
	active := realm.Character{GUID: 1, Map: 0, Health: 10}
	server.registerPlayer(active)
	spell := dbc.Spell{ID: 45}
	effect := dbc.SpellEffect{Type: int64(packet.SpellEffectApplyAura), Aura: int64(packet.AuraModStealth)}
	server.applyAura(&spellCast{caster: active, spell: spell}, active, 0, effect)
	if server.unitFlags(active.GUID)&0x00008000 == 0 {
		t.Fatalf("flags=%x", server.unitFlags(active.GUID))
	}
	server.removeAura(active, 0)
	if server.unitFlags(active.GUID)&0x00008000 != 0 {
		t.Fatalf("flags after remove=%x", server.unitFlags(active.GUID))
	}
}
