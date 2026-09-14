package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestSpellEffectSkillStep(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Player", Level: 1, Health: 1})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters}
	target := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Player", Level: 1, Health: 1}
	server.registerPlayer(target)
	server.applySpellEffects(&spellCast{caster: target, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectSkillStep), MiscValue: 55, BasePoints: 20}}}, effectTargets: map[int][]realm.Character{0: {target}}})
	value, found, err := characters.SkillValue(guid, 55)
	if err != nil || !found || value != 1 {
		t.Fatalf("skill value=%d found=%v err=%v", value, found, err)
	}
	skill, found, err := server.skillData(guid, 55)
	if err != nil || !found || skill.Max != 100 {
		t.Fatalf("skill=%#v found=%v err=%v", skill, found, err)
	}
}

func TestPassiveSpellAddsSkill(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, Attributes, Effect_1) VALUES (42, 64, 25); INSERT INTO SkillLineAbility (ID, SkillLine, Spell) VALUES (1, 55, 42)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Player", Race: 1, Class: 1, Level: 1, Health: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(guid, 42); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases)}
	server.registerPlayer(realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Player", Race: 1, Class: 1, Level: 1, Health: 1})
	if err := server.initializePassiveSpells(realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Player", Race: 1, Class: 1, Level: 1, Health: 1}); err != nil {
		t.Fatal(err)
	}
	value, found, err := characters.SkillValue(guid, 55)
	if err != nil || !found || value != 1 {
		t.Fatalf("skill value=%d found=%v err=%v", value, found, err)
	}
}
