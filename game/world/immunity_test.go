package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestSpellDamageImmunity(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	server := &WorldServer{DBC: dbc.NewStore(databases)}
	caster := realm.Character{GUID: 1, Name: "Caster", Health: 100}
	target := realm.Character{GUID: 2, Name: "Target", Health: 100}
	server.registerPlayer(caster)
	server.registerPlayer(target)
	server.setPlayerMaxHealth(target.GUID, 100)
	server.applyAura(&spellCast{caster: caster, spell: dbc.Spell{ID: 1}}, target, 0, dbc.SpellEffect{Type: int64(packet.SpellEffectApplyAura), Aura: int64(packet.AuraModDamageImmunity), MiscValue: -2})
	damage := &spellCast{caster: caster, spell: dbc.Spell{ID: 2, School: 2, Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectSchoolDamage), BasePoints: 25}}}, effectTargets: map[int][]realm.Character{0: {target}}}
	server.applySpellEffects(damage)
	current, found := server.playerByGUID(target.GUID)
	if !found || current.Health != 100 || !server.spellDamageImmune(target.GUID, 2) {
		t.Fatalf("immune target=%#v found=%v immune=%v", current, found, server.spellDamageImmune(target.GUID, 2))
	}
	server.removeAura(target, 0)
	server.applySpellEffects(damage)
	current, found = server.playerByGUID(target.GUID)
	if !found || current.Health != 75 || server.spellDamageImmune(target.GUID, 2) {
		t.Fatalf("after removal target=%#v found=%v immune=%v", current, found, server.spellDamageImmune(target.GUID, 2))
	}
}
