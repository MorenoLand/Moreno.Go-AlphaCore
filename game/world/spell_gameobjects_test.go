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

func TestSummonGameObjectLifecycle(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO gameobject_template (entry, type, displayId, name, faction, size) VALUES (200, 5, 321, 'Summoned Object', 1, 1)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{WorldData: worlddb.NewStore(databases)}
	owner := realm.Character{GUID: 1, Map: 2, PositionX: 10, PositionY: 20, PositionZ: 30, Orientation: 1}
	destination := &spellVector{X: 40, Y: 50, Z: 60}
	server.summonGameObject(&spellCast{caster: owner, target: spellTarget{Dest: destination}, spell: dbc.Spell{ID: 7}}, dbc.SpellEffect{Type: int64(packet.SpellEffectSummonObject), MiscValue: 200}, false)
	server.gameObjectMu.Lock()
	if len(server.dynamicGameObjects) != 1 {
		server.gameObjectMu.Unlock()
		t.Fatalf("dynamic objects=%d", len(server.dynamicGameObjects))
	}
	var guid uint64
	for value := range server.dynamicGameObjects {
		guid = value
	}
	instance := server.dynamicGameObjects[guid]
	server.gameObjectMu.Unlock()
	if guid&0xffff000000000000 != 0xf110000000000000 || instance.spawn.PositionX != 40 || instance.spawn.PositionY != 50 || instance.spawn.PositionZ != 60 {
		t.Fatalf("guid=%x instance=%#v", guid, instance)
	}
	spawn, template, found, err := server.gameObjectAt(owner, guid, 100)
	if err != nil || !found || spawn.Entry != 200 || template.DisplayID != 321 {
		t.Fatalf("spawn=%#v template=%#v found=%v err=%v", spawn, template, found, err)
	}
	server.destroyGameObject(guid)
	_, _, found, err = server.gameObjectAt(owner, guid, 100)
	if err != nil || found {
		t.Fatalf("destroyed found=%v err=%v", found, err)
	}
}
