package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestPersistentAreaAuraEffect(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	server := &WorldServer{DBC: dbc.NewStore(databases)}
	target := realm.Character{GUID: 2, Name: "Target", Health: 100}
	server.registerPlayer(target)
	server.applySpellEffects(&spellCast{caster: target, spell: dbc.Spell{ID: 7, Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectPersistentAreaAura), Aura: int64(packet.AuraModIncreaseHealth), BasePoints: 5}}}, effectTargets: map[int][]realm.Character{0: {target}}})
	server.auras.mu.Lock()
	_, found := server.auras.active[target.GUID][0]
	server.auras.mu.Unlock()
	if !found {
		t.Fatal("persistent area aura was not applied")
	}
}
