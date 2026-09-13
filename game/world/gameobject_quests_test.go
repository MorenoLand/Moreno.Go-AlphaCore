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

func TestGameObjectQuestGiver(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO gameobject_template (entry, type, displayId, name) VALUES (300, 2, 10, 'Quest Object'); INSERT INTO spawns_gameobjects (spawn_id, spawn_entry, spawn_map, spawn_positionX, spawn_positionY, spawn_positionZ) VALUES (1, 300, 0, 0, 0, 0); INSERT INTO quest_template (entry, QuestLevel, Title, Details, Objectives) VALUES (500, 1, 'Object Quest', 'Details', 'Objectives'); INSERT INTO gameobject_quest_starter (entry, quest) VALUES (300, 500)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Adventurer", Level: 1, Health: 1})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Adventurer", Level: 1, Health: 1, Map: 0}
	objectGUID := uint64(0xf110000000000001)
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, objectGUID)
	status, err := server.questGiverStatus(active, data)
	if err != nil {
		t.Fatal(err)
	}
	statusPacket, err := packet.Parse(status)
	if err != nil || statusPacket.Opcode != packet.SMSGQuestGiverStatus || binary.LittleEndian.Uint32(statusPacket.Data[8:]) != questGiverQuest {
		t.Fatalf("status=%#v err=%v", statusPacket, err)
	}
	responses, err := server.gameObjectUse(active, data)
	if err != nil || len(responses) != 1 {
		t.Fatalf("use responses=%d err=%v", len(responses), err)
	}
	if greeting, err := packet.Parse(responses[0]); err != nil || greeting.Opcode != packet.SMSGQuestGiverQuestDetails {
		t.Fatalf("greeting=%#v err=%v", greeting, err)
	}
	accept := make([]byte, 12)
	copy(accept, data)
	binary.LittleEndian.PutUint32(accept[8:], 500)
	responses, err = server.questAccept(active, accept)
	if err != nil || len(responses) != 1 {
		t.Fatalf("accept responses=%d err=%v", len(responses), err)
	}
	if query, err := packet.Parse(responses[0]); err != nil || query.Opcode != packet.SMSGQuestQueryResponse {
		t.Fatalf("query=%#v err=%v", query, err)
	}
}
