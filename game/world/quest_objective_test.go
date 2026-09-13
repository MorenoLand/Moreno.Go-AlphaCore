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

func TestQuestObjectiveProgress(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id, stackable) VALUES (200, 'Quest Item', 20, 20); INSERT INTO quest_template (entry, QuestLevel, ReqCreatureOrGOId1, ReqCreatureOrGOCount1) VALUES (500, 1, -300, 1); INSERT INTO quest_template (entry, QuestLevel, ReqItemId1, ReqItemCount1) VALUES (501, 1, 200, 2)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Adventurer", Level: 1, Health: 1})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Adventurer", Level: 1, Health: 1}
	if err := characters.SaveQuestState(realm.QuestState{GUID: guid, Quest: 500, State: questAccepted}); err != nil {
		t.Fatal(err)
	}
	responses, err := server.questProgress(active, -300, 0xf110000000000001)
	if err != nil || len(responses) != 1 {
		t.Fatalf("GO responses=%d err=%v", len(responses), err)
	}
	kill, err := packet.Parse(responses[0])
	if err != nil || kill.Opcode != packet.SMSGQuestUpdateAddKill || binary.LittleEndian.Uint32(kill.Data[4:]) != 0x8000012c {
		t.Fatalf("kill=%#v err=%v", kill, err)
	}
	state, found, err := characters.QuestState(guid, 500)
	if err != nil || !found || state.MobCounts[0] != 1 || state.State != questReward {
		t.Fatalf("GO state=%#v found=%v err=%v", state, found, err)
	}
	if err := characters.SaveQuestState(realm.QuestState{GUID: guid, Quest: 501, State: questAccepted}); err != nil {
		t.Fatal(err)
	}
	if err := characters.AddInventoryItemAt(guid, 200, 23, 23, 2); err != nil {
		t.Fatal(err)
	}
	responses, err = server.questItemProgress(active, 200, 2)
	if err != nil || len(responses) != 1 {
		t.Fatalf("item responses=%d err=%v", len(responses), err)
	}
	item, err := packet.Parse(responses[0])
	if err != nil || item.Opcode != packet.SMSGQuestUpdateAddItem || binary.LittleEndian.Uint32(item.Data) != 200 || binary.LittleEndian.Uint32(item.Data[4:]) != 2 {
		t.Fatalf("item=%#v err=%v", item, err)
	}
	state, found, err = characters.QuestState(guid, 501)
	if err != nil || !found || state.ItemCounts[0] != 2 || state.State != questReward {
		t.Fatalf("item state=%#v found=%v err=%v", state, found, err)
	}
}
