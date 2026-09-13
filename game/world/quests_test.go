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

func TestQuestGiverStatusHelloAndQuery(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id) VALUES (300, 'Quest Reward', 33); INSERT INTO quest_template (entry, QuestLevel, Title, Details, Objectives, RewItemId1, RewItemCount1) VALUES (500, 1, 'Test Quest', 'Details', 'Objectives', 300, 2); INSERT INTO creature_template (entry, display_id1, name, npc_flags) VALUES (100, 20, 'Quest Giver', 2); INSERT INTO creature_quest_starter (entry, quest) VALUES (100, 500); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z) VALUES (1, 100, 0, 0, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: realm.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: 10, Name: "Player", Map: 0}
	giverGUID := uint64(0xf130000000000001)
	guidData := make([]byte, 8)
	binary.LittleEndian.PutUint64(guidData, giverGUID)
	status, err := server.questGiverStatus(active, guidData)
	if err != nil {
		t.Fatal(err)
	}
	statusPacket, err := packet.Parse(status)
	if err != nil || statusPacket.Opcode != packet.SMSGQuestGiverStatus || binary.LittleEndian.Uint64(statusPacket.Data) != giverGUID || binary.LittleEndian.Uint32(statusPacket.Data[8:]) != questGiverQuest {
		t.Fatalf("status=%#v err=%v", statusPacket, err)
	}
	responses, err := server.questGiverHello(active, guidData)
	if err != nil || len(responses) != 2 {
		t.Fatalf("hello responses=%d err=%v", len(responses), err)
	}
	details, err := packet.Parse(responses[1])
	if err != nil || details.Opcode != packet.SMSGQuestGiverQuestDetails || binary.LittleEndian.Uint64(details.Data) != giverGUID || binary.LittleEndian.Uint32(details.Data[8:]) != 500 {
		t.Fatalf("details=%#v err=%v", details, err)
	}
	queryData := append(guidData, []byte{0xf4, 0x01, 0, 0}...)
	responses, err = server.questGiverQuery(active, queryData)
	if err != nil || len(responses) != 2 {
		t.Fatalf("query responses=%d err=%v", len(responses), err)
	}
	if message, err := packet.Parse(responses[1]); err != nil || message.Opcode != packet.SMSGQuestGiverQuestDetails {
		t.Fatalf("query details=%#v err=%v", message, err)
	}
}
