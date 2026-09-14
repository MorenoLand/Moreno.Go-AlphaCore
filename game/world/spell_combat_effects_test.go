package world

import (
	"testing"
	"time"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestCombatSpellEffects(t *testing.T) {
	server := &WorldServer{}
	caster := realm.Character{GUID: 1, Health: 100}
	player := realm.Character{GUID: 2, Health: 100}
	server.registerPlayer(player)
	server.applySpellEffects(&spellCast{caster: caster, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectExtraAttacks), BasePoints: 2}}}, effectTargets: map[int][]realm.Character{0: {player}}})
	if amount := server.extraAttacks(player.GUID); amount != 2 {
		t.Fatalf("player extra attacks=%d", amount)
	}
	extraTarget := &creatureState{GUID: 3, Health: 100, MaxHealth: 100}
	server.applySpellEffects(&spellCast{caster: caster, targetCreature: extraTarget, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectExtraAttacks), BasePoints: 3}}}})
	if extraTarget.ExtraAttacks != 3 {
		t.Fatalf("creature extra attacks=%d", extraTarget.ExtraAttacks)
	}
	threatTarget := &creatureState{GUID: 4, Health: 100, MaxHealth: 100}
	server.applySpellEffects(&spellCast{caster: caster, targetCreature: threatTarget, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectThreat), BasePoints: 4}}}})
	if threatTarget.Threat[uint64(caster.GUID)] != 4 || threatTarget.CombatTarget != uint64(caster.GUID) || server.combatTarget(caster.GUID) != threatTarget.GUID {
		t.Fatalf("threat=%#v combat=%d player combat=%d", threatTarget.Threat, threatTarget.CombatTarget, server.combatTarget(caster.GUID))
	}
	pullTarget := &creatureState{GUID: 5, Health: 100, MaxHealth: 100}
	server.applySpellEffects(&spellCast{caster: caster, targetCreature: pullTarget, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectPull)}}}})
	if pullTarget.CombatTarget != uint64(caster.GUID) || !pullTarget.PullUntil.After(time.Now()) || pullTarget.Threat[uint64(caster.GUID)] != creatureThreatNotToLeaveCombat {
		t.Fatalf("pull target=%#v", pullTarget)
	}
}
