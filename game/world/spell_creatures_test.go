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

func TestSummonSpellCreature(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, faction, unit_class, level_min, health_multiplier, mana_multiplier) VALUES (200, 20, 'Guardian', 1, 1, 1, 1, 1); INSERT INTO creature_classlevelstats (class, level, health, mana, base_health, base_mana) VALUES (1, 1, 40, 0, 40, 0)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{WorldData: worlddb.NewStore(databases)}
	caster := realm.Character{GUID: 1, Level: 1, Map: 0, PositionX: 1, PositionY: 2, PositionZ: 3}
	server.registerPlayer(caster)
	cast := &spellCast{caster: caster, spell: dbc.Spell{ID: 99, Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectSummonWild), MiscValue: 200}}}}
	server.applySpellEffects(cast)
	server.creatures.mu.Lock()
	if len(server.creatures.active) != 1 {
		server.creatures.mu.Unlock()
		t.Fatalf("creatures=%d", len(server.creatures.active))
	}
	var state *creatureState
	for _, value := range server.creatures.active {
		state = value
	}
	server.creatures.mu.Unlock()
	if state == nil || state.OwnerGUID != uint64(caster.GUID) || state.CreatedBySpell != 99 || state.Spawn.Entry != 200 {
		t.Fatalf("state=%#v", state)
	}
	server.despawnCreature(state.GUID)
}
