package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestGameObjectUse(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO gameobject_template (entry, type, displayId, name, flags, size) VALUES (200, 0, 10, 'Door', 0, 1); INSERT INTO spawns_gameobjects (spawn_id, spawn_entry, spawn_map, spawn_positionX, spawn_positionY, spawn_positionZ) VALUES (1, 200, 0, 0, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: 1, Map: 0}
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, 0xf110000000000001)
	responses, err := server.gameObjectUse(active, data)
	if err != nil || len(responses) != 2 {
		t.Fatalf("responses=%d err=%v", len(responses), err)
	}
	for _, response := range responses {
		parsed, parseErr := packet.Parse(response)
		if parseErr != nil || parsed.Opcode != packet.SMSGUpdateObject {
			t.Fatalf("response=%#v err=%v", parsed, parseErr)
		}
	}
	state := server.gameObjectStateFor(0xf110000000000001, worlddb.GameObjectSpawn{State: 0})
	if state.state != gameObjectStateReady || state.flags&gameObjectFlagInUse == 0 {
		t.Fatalf("state=%#v", state)
	}
	if _, err := databases.DB(database.World).Exec(`UPDATE gameobject_template SET type = 5 WHERE entry = 200`); err != nil {
		t.Fatal(err)
	}
	if responses, err := server.gameObjectUse(active, data); err != nil || responses != nil {
		t.Fatalf("generic responses=%v err=%v", responses, err)
	}
}
