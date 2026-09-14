package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestSpellScalingUsesSkillRank(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, BaseLevel, CastingTimeIndex, ManaCostPerLevel) VALUES (42, 1, 1, 3); INSERT INTO SpellCastTimes (ID, Base, PerLevel, Minimum) VALUES (1, 100, 10, 0); INSERT INTO SkillLineAbility (ID, SkillLine, Spell) VALUES (1, 55, 42)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Race: 1, Class: 1, Level: 20, Health: 100, Power1: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(guid, 42); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.Realm).Exec(`INSERT INTO character_skills (guid, skill, value, max) VALUES (?, 55, 25, 25)`, guid); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Caster", Race: 1, Class: 1, Level: 20, Health: 100, Power1: 1000}
	server.registerPlayer(active)
	spell := dbc.Spell{ID: 42}
	level, err := server.spellCasterLevel(active, spell)
	if err != nil || level != 5 {
		t.Fatalf("spell level=%d err=%v", level, err)
	}
	if _, err := server.startSpellCast(active, 42, spellTarget{UnitGUID: uint64(guid)}, packet.SpellTargetUnit); err != nil {
		t.Fatal(err)
	}
	server.spells.mu.Lock()
	cast := server.spells.casts[guid]
	if cast != nil && cast.timer != nil {
		cast.timer.Stop()
	}
	delete(server.spells.casts, guid)
	server.spells.mu.Unlock()
	if cast == nil || cast.spellLevel != 5 || cast.castTime != 150 {
		t.Fatalf("cast=%#v", cast)
	}
}
