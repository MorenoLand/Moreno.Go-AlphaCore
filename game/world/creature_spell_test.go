package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestSpellDamagesCreatureTarget(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, Effect_1, EffectBasePoints_1) VALUES (42, 2, 7)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, level_min, level_max, unit_class, health_multiplier, mana_multiplier, damage_multiplier) VALUES (100, 123, 'Wolf', 1, 1, 1, 1, 1, 1); INSERT INTO creature_classlevelstats ("class", level, health, mana) VALUES (1, 1, 20, 0); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z, health_percent, mana_percent) VALUES (1, 100, 0, 3, 0, 0, 100, 100)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	casterID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(casterID, 42); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	caster := realm.Character{GUID: casterID, AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, Health: 10}
	server.registerPlayer(caster)
	data := append(encodeUint32(42), byte(packet.SpellTargetUnit), 0)
	data = append(data, make([]byte, 8)...)
	binary.LittleEndian.PutUint64(data[6:], 0xf130000000000001)
	if responses, err := server.castSpellPacket(caster, data); err != nil || len(responses) != 0 {
		t.Fatalf("responses=%d err=%v", len(responses), err)
	}
	health, found := server.creatureHealth(0xf130000000000001)
	if !found || health != 13 {
		t.Fatalf("health=%d found=%v", health, found)
	}
}
