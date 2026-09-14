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

func TestFarsightSpellEffect(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO SpellDuration (ID, Duration, DurationPerLevel, MaxDuration) VALUES (1, 5000, 0, 5000); INSERT INTO SpellRadius (ID, Radius, RadiusPerLevel, RadiusMax) VALUES (2, 12, 0, 12)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	caster := realm.Character{GUID: 1, Map: 0, Health: 100}
	server.registerPlayer(caster)
	server.applySpellEffects(&spellCast{caster: caster, target: spellTarget{Dest: &spellVector{X: 10, Y: 20, Z: 30}}, spell: dbc.Spell{ID: 42, DurationIndex: 1, Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectAddFarsight), RadiusIndex: 2}}}})
	if guid := server.farSight(caster.GUID); guid == 0 {
		t.Fatal("farsight field was not set")
	} else {
		server.dynamicMu.Lock()
		object, found := server.dynamicObjects[guid]
		server.dynamicMu.Unlock()
		if !found || object.caster != uint64(caster.GUID) || object.spell != 42 || object.x != 10 || object.y != 20 || object.z != 30 || object.radius != 12 {
			t.Fatalf("dynamic object=%#v found=%v", object, found)
		}
		server.destroyDynamicObject(guid)
		if server.farSight(caster.GUID) != 0 {
			t.Fatal("farsight field remained set")
		}
	}
}
