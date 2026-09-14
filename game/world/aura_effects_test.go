package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestHealthAuraLifecycle(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	server := &WorldServer{DBC: dbc.NewStore(databases)}
	active := realm.Character{GUID: 1, Name: "Player", Health: 100}
	server.registerPlayer(active)
	effect := dbc.SpellEffect{Type: int64(packet.SpellEffectApplyAura), Aura: int64(packet.AuraModIncreaseHealth), BasePoints: 49}
	server.applyAura(&spellCast{caster: active, spell: dbc.Spell{ID: 1}}, active, 0, effect)
	if max := server.playerMaxHealth(active.GUID); max != 149 {
		t.Fatalf("max health after apply=%d", max)
	}
	server.applyAura(&spellCast{caster: active, spell: dbc.Spell{ID: 1}}, active, 0, effect)
	if max := server.playerMaxHealth(active.GUID); max != 149 {
		t.Fatalf("max health after refresh=%d", max)
	}
	if err := server.changePlayerHealth(&active, 100); err != nil || active.Health != 149 {
		t.Fatalf("health after clamp=%d err=%v", active.Health, err)
	}
	server.removeAura(active, 0)
	current, found := server.playerByGUID(active.GUID)
	if max := server.playerMaxHealth(active.GUID); max != 100 || !found || current.Health != 100 {
		t.Fatalf("after remove max=%d health=%d found=%v", max, current.Health, found)
	}
	manaEffect := dbc.SpellEffect{Type: int64(packet.SpellEffectApplyAura), Aura: int64(packet.AuraModIncreaseMana), BasePoints: 49}
	server.applyAura(&spellCast{caster: active, spell: dbc.Spell{ID: 2}}, active, 0, manaEffect)
	if max := server.playerMaxPower(active.GUID, 0); max != 1049 {
		t.Fatalf("max mana after apply=%d", max)
	}
	if err := server.changePlayerPower(&active, 0, 2000); err != nil || active.Power1 != 1049 {
		t.Fatalf("mana after clamp=%d err=%v", active.Power1, err)
	}
	server.removeAura(active, 0)
	current, found = server.playerByGUID(active.GUID)
	if max := server.playerMaxPower(active.GUID, 0); max != 1000 || !found || current.Power1 != 1000 {
		t.Fatalf("mana after remove max=%d power=%d found=%v", max, current.Power1, found)
	}
	server.setUnitFlags(active, unitFlagPlayer|unitFlagDebugCombatLog)
	server.applyAura(&spellCast{caster: active, spell: dbc.Spell{ID: 3}}, active, 0, dbc.SpellEffect{Type: int64(packet.SpellEffectApplyAura), Aura: int64(packet.AuraModStealth)})
	if server.unitFlags(active.GUID)&unitFlagDebugCombatLog == 0 || server.unitFlags(active.GUID)&0x00008000 == 0 {
		t.Fatalf("flags after aura=%x", server.unitFlags(active.GUID))
	}
	server.removeAura(active, 0)
	if server.unitFlags(active.GUID) != unitFlagPlayer|unitFlagDebugCombatLog {
		t.Fatalf("flags after aura removal=%x", server.unitFlags(active.GUID))
	}
}
