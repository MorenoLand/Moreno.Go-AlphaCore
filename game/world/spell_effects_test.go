package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestSpellEffectStateTransitions(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID) VALUES (9010)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	casterID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Health: 100, Map: 1, PositionX: 1, PositionY: 2, PositionZ: 3, Orientation: 4})
	if err != nil {
		t.Fatal(err)
	}
	targetID, err := characters.Create(realm.Character{AccountID: 2, RealmID: 1, Name: "Target", Health: 100, Map: 1, PositionX: 10, PositionY: 20, PositionZ: 30, Orientation: 40})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases)}
	caster := realm.Character{GUID: casterID, AccountID: 1, RealmID: 1, Name: "Caster", Health: 100, Map: 1, PositionX: 1, PositionY: 2, PositionZ: 3, Orientation: 4}
	target := realm.Character{GUID: targetID, AccountID: 2, RealmID: 1, Name: "Target", Health: 100, Map: 1, PositionX: 10, PositionY: 20, PositionZ: 30, Orientation: 40}
	server.registerPlayer(caster)
	server.registerPlayer(target)
	server.auras.active = map[int64]map[int]*auraState{targetID: {0: {target: target, harmful: true}, 1: {target: target}}}
	server.dispelSpellAuras(&spellCast{caster: caster}, target, 1)
	server.auras.mu.Lock()
	_, harmful := server.auras.active[targetID][0]
	_, beneficial := server.auras.active[targetID][1]
	server.auras.mu.Unlock()
	if harmful || !beneficial {
		t.Fatalf("auras after dispel harmful=%v beneficial=%v", harmful, beneficial)
	}
	server.spells.casts = map[int64]*spellCast{targetID: {caster: target, spell: dbc.Spell{ID: 77}}}
	server.applySpellEffects(&spellCast{caster: caster, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectInterruptCast)}}}, effectTargets: map[int][]realm.Character{0: {target}}})
	server.spells.mu.Lock()
	_, casting := server.spells.casts[targetID]
	server.spells.mu.Unlock()
	if casting {
		t.Fatal("interrupt effect left target cast active")
	}
	server.applySpellEffects(&spellCast{caster: caster, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectLearnSpell), TriggerSpell: 9010}}}, effectTargets: map[int][]realm.Character{0: {target}}})
	learned, err := characters.Spells(targetID)
	if err != nil || len(learned) != 1 || learned[0].ID != 9010 {
		t.Fatalf("learned=%#v err=%v", learned, err)
	}
	server.setCombatTarget(casterID, uint64(targetID))
	server.applySpellEffects(&spellCast{caster: caster, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectSanctuary)}}}, effectTargets: map[int][]realm.Character{0: {caster}}})
	if !server.isSanctuary(casterID) || server.combatTarget(casterID) != 0 {
		t.Fatalf("sanctuary=%v combat=%d", server.isSanctuary(casterID), server.combatTarget(casterID))
	}
	destination := &spellVector{X: 100, Y: 200, Z: 300}
	server.applySpellEffects(&spellCast{caster: caster, target: spellTarget{UnitGUID: uint64(targetID), Dest: destination}, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectTeleportUnits)}}}, effectTargets: map[int][]realm.Character{0: {target}}})
	teleported, found := server.playerByGUID(targetID)
	if !found || teleported.Map != caster.Map || teleported.PositionX != 100 || teleported.PositionY != 200 || teleported.PositionZ != 300 {
		t.Fatalf("teleported=%#v found=%v", teleported, found)
	}
	server.applySpellEffects(&spellCast{caster: caster, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectSummonPlayer)}}}, effectTargets: map[int][]realm.Character{0: {target}}})
	summoned, found := server.playerByGUID(targetID)
	if !found || summoned.PositionX != caster.PositionX || summoned.PositionY != caster.PositionY || summoned.PositionZ != caster.PositionZ {
		t.Fatalf("summoned=%#v found=%v", summoned, found)
	}
}

func TestSpellDamageAndHealingSemantics(t *testing.T) {
	server := &WorldServer{}
	caster := realm.Character{GUID: 1, Name: "Caster", Health: 50}
	target := realm.Character{GUID: 2, Name: "Target", Health: 20}
	server.registerPlayer(caster)
	server.registerPlayer(target)
	server.setPlayerMaxHealth(caster.GUID, 100)
	server.setPlayerMaxHealth(target.GUID, 100)
	server.applySpellEffects(&spellCast{caster: caster, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectHealthLeech), BasePoints: 5}}}, effectTargets: map[int][]realm.Character{0: {target}}})
	updatedCaster, casterFound := server.playerByGUID(caster.GUID)
	updatedTarget, targetFound := server.playerByGUID(target.GUID)
	if !casterFound || !targetFound || updatedCaster.Health != 55 || updatedTarget.Health != 15 {
		t.Fatalf("leech caster=%#v target=%#v", updatedCaster, updatedTarget)
	}
	updatedTarget.Health = 20
	server.updatePlayer(updatedTarget)
	server.applySpellEffects(&spellCast{caster: caster, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectHealMaxHealth)}}}, effectTargets: map[int][]realm.Character{0: {updatedTarget}}})
	updatedTarget, targetFound = server.playerByGUID(target.GUID)
	if !targetFound || updatedTarget.Health != 100 {
		t.Fatalf("max heal target=%#v", updatedTarget)
	}
}
