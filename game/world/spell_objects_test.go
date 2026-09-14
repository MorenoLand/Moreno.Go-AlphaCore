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

func TestSpellObjectTargets(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO gameobject_template (entry, type, displayId, name, flags, size) VALUES (300, 0, 10, 'Door', 0, 1); INSERT INTO spawns_gameobjects (spawn_id, spawn_entry, spawn_map, spawn_positionX, spawn_positionY, spawn_positionZ) VALUES (1, 300, 0, 0, 0, 0); INSERT INTO item_template (entry, name) VALUES (400, 'Locked Box')`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, Effect_1) VALUES (42, 33), (43, 59)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Player", Class: 1, Level: 1, Health: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(guid, 42); err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(guid, 43); err != nil {
		t.Fatal(err)
	}
	item, err := characters.CreateInventoryItem(guid, 0, 23, 23, 400, 1)
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Player", Class: 1, Level: 1, Health: 1}
	server.registerPlayer(active)
	gameObjectGUID := uint64(0xf110000000000001)
	gameObjectMask := uint16(packet.SpellTargetGameObject)
	data := append(encodeUint32(42), byte(gameObjectMask), byte(gameObjectMask>>8))
	data = append(data, encodeUint64(gameObjectGUID)...)
	if responses, err := server.castSpellPacket(active, data); err != nil || len(responses) != 0 {
		t.Fatalf("gameobject responses=%d err=%v", len(responses), err)
	}
	state := server.gameObjectStateFor(gameObjectGUID, worlddb.GameObjectSpawn{State: 0})
	if state.state != gameObjectStateReady || state.flags&gameObjectFlagInUse == 0 {
		t.Fatalf("gameobject state=%#v", state)
	}
	data = append(encodeUint32(43), byte(packet.SpellTargetItem), 0)
	data = append(data, encodeUint64(uint64(item.GUID))...)
	if responses, err := server.castSpellPacket(active, data); err != nil || len(responses) != 0 {
		t.Fatalf("item responses=%d err=%v", len(responses), err)
	}
	updated, found, err := characters.ItemByGUID(guid, item.GUID)
	if err != nil || !found || updated.Flags&itemDynUnlocked == 0 {
		t.Fatalf("item=%#v found=%v err=%v", updated, found, err)
	}
}
