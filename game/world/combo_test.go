package world

import (
	"testing"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestComboPointSpellEffect(t *testing.T) {
	server := &WorldServer{}
	caster := realm.Character{GUID: 1, Name: "Caster", Level: 1}
	target := realm.Character{GUID: 2, Name: "Target", Health: 100}
	server.registerPlayer(caster)
	server.registerPlayer(target)
	effect := dbc.SpellEffect{Type: int64(packet.SpellEffectAddComboPoints), BasePoints: 2}
	cast := &spellCast{caster: caster, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{effect}}, effectTargets: map[int][]realm.Character{0: {target}}}
	server.applySpellEffects(cast)
	points, comboTarget := server.comboState(caster.GUID)
	if points != 2 || comboTarget != uint64(target.GUID) || !server.validComboTarget(caster, spellTarget{UnitGUID: uint64(target.GUID)}) {
		t.Fatalf("first combo state points=%d target=%d", points, comboTarget)
	}
	server.applySpellEffects(cast)
	points, comboTarget = server.comboState(caster.GUID)
	if points != 4 || comboTarget != uint64(target.GUID) {
		t.Fatalf("second combo state points=%d target=%d", points, comboTarget)
	}
	server.removeComboPoints(caster.GUID)
	points, comboTarget = server.comboState(caster.GUID)
	if points != 0 || comboTarget != 0 || server.validComboTarget(caster, spellTarget{UnitGUID: uint64(target.GUID)}) {
		t.Fatalf("cleared combo state points=%d target=%d", points, comboTarget)
	}
}
