package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestSendEventSpellSummonsCreature(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, faction, unit_class, level_min, health_multiplier, mana_multiplier) VALUES (200, 20, 'Event Creature', 1, 1, 1, 1, 1); INSERT INTO creature_classlevelstats (class, level, health, mana, base_health, base_mana) VALUES (1, 1, 40, 0, 40, 0); INSERT INTO event_scripts (id, command, datalong, datalong2, x, y, z, o) VALUES (99, 10, 200, 1000, 4, 5, 6, 1)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{WorldData: worlddb.NewStore(databases)}
	caster := realm.Character{GUID: 1, Level: 1, Map: 0, Health: 100}
	server.applySpellEffects(&spellCast{caster: caster, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectSendEvent), MiscValue: 99}}}})
	server.creatures.mu.Lock()
	if len(server.creatures.active) != 1 {
		server.creatures.mu.Unlock()
		t.Fatalf("creatures=%d", len(server.creatures.active))
	}
	for _, state := range server.creatures.active {
		if state.Template.Entry != 200 || state.OwnerGUID != uint64(caster.GUID) || state.Spawn.PositionX != 4 || state.Spawn.PositionY != 5 || state.Spawn.PositionZ != 6 {
			server.creatures.mu.Unlock()
			t.Fatalf("state=%#v", state)
		}
	}
	server.creatures.mu.Unlock()
}
