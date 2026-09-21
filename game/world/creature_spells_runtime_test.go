package world

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
)

func TestCreatureSpellListCastsOnCombat(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, faction, unit_class, level_min, health_multiplier, mana_multiplier, spell_list_id) VALUES (100, 20, 'Caster Creature', 1, 1, 1, 1, 1, 400); INSERT INTO creature_classlevelstats (class, level, health, mana, base_health, base_mana) VALUES (1, 1, 40, 0, 40, 0); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z, health_percent, mana_percent) VALUES (7, 100, 0, 0, 0, 0, 100, 100); INSERT INTO creature_spells (entry, spellId_1, probability_1, castTarget_1) VALUES (400, 900, 100, 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, Effect_1, EffectBasePoints_1, ImplicitTargetA_1) VALUES (900, 10, 3, 25)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Target", Map: 0, Health: 5})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Target", Map: 0, Health: 5}
	server.registerPlayer(active)
	server.setPlayerMaxHealth(guid, 100)
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, 0xf130000000000007)
	if _, err := server.makeMonsterAttackMe(active, data); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if player, found := server.playerByGUID(guid); found && player.Health >= 8 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	player, _ := server.playerByGUID(guid)
	t.Fatalf("creature spell did not heal target: health=%d", player.Health)
}
