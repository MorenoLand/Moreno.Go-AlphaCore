package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
)

func TestCreatureTemplatePassiveSpellInitialization(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, unit_class, level_min, health_multiplier, mana_multiplier, spell_id1) VALUES (100, 20, 'Passive Creature', 1, 1, 1, 1, 900); INSERT INTO creature_classlevelstats (class, level, health, mana, base_health, base_mana) VALUES (1, 1, 40, 0, 40, 0); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z, health_percent, mana_percent) VALUES (7, 100, 0, 1, 2, 3, 100, 100)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, Attributes, Effect_1, EffectAura_1) VALUES (900, 64, 6, 16)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases), Characters: realm.NewStore(databases)}
	state, found, err := server.creatureStateAt(realm.Character{GUID: 1, Map: 0, PositionX: 1, PositionY: 2, PositionZ: 3, Health: 100}, 0xf130000000000007, creatureViewDistance)
	if err != nil || !found {
		t.Fatalf("state=%#v found=%v err=%v", state, found, err)
	}
	server.auras.mu.Lock()
	_, auraFound := server.auras.active[int64(state.GUID)]
	server.auras.mu.Unlock()
	if !state.PassiveSpellsInitialized || !auraFound {
		t.Fatalf("state=%#v aura=%v", state, auraFound)
	}
	registered := *state
	registered.GUID = 0xf130000000000008
	registered.PassiveSpellsInitialized = false
	server.setCreatureState(registered)
	server.creatures.mu.Lock()
	registeredInitialized := server.creatures.active[registered.GUID].PassiveSpellsInitialized
	server.creatures.mu.Unlock()
	server.auras.mu.Lock()
	_, registeredAuraFound := server.auras.active[int64(registered.GUID)]
	server.auras.mu.Unlock()
	if !registeredInitialized || !registeredAuraFound {
		t.Fatalf("registered state=%#v aura=%v", registered, registeredAuraFound)
	}
}
