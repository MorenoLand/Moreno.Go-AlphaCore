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

func TestQuestProgression(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Player", Level: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id, stackable) VALUES (300, 'Reward', 33, 1); INSERT INTO quest_template (entry, QuestLevel, Title, Details, Objectives, OfferRewardText, RequestItemsText, RewItemId1, RewItemCount1, RewXP, RewOrReqMoney) VALUES (500, 1, 'Quest', 'Details', 'Objectives', 'Reward', 'Bring it', 300, 1, 10, 5); INSERT INTO creature_template (entry, display_id1, name, npc_flags) VALUES (100, 20, 'Quest Giver', 2); INSERT INTO creature_quest_starter (entry, quest) VALUES (100, 500); INSERT INTO creature_quest_finisher (entry, quest) VALUES (100, 500); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z) VALUES (1, 100, 0, 0, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Player", Level: 1, Map: 0}
	giver := make([]byte, 12)
	binary.LittleEndian.PutUint64(giver, 0xf130000000000001)
	binary.LittleEndian.PutUint32(giver[8:], 500)
	if responses, err := server.questAccept(active, giver); err != nil || len(responses) == 0 {
		t.Fatalf("accept responses=%d err=%v", len(responses), err)
	}
	state, found, err := characters.QuestState(guid, 500)
	if err != nil || !found || state.State != questAccepted {
		t.Fatalf("accepted=%#v found=%v err=%v", state, found, err)
	}
	responses, err := server.questComplete(active, giver)
	if err != nil || len(responses) != 2 {
		t.Fatalf("complete responses=%d err=%v", len(responses), err)
	}
	offer, err := packet.Parse(responses[1])
	if err != nil || offer.Opcode != packet.SMSGQuestGiverOfferReward {
		t.Fatalf("offer=%#v err=%v", offer, err)
	}
	state, found, err = characters.QuestState(guid, 500)
	if err != nil || !found || state.State != questReward {
		t.Fatalf("reward state=%#v found=%v err=%v", state, found, err)
	}
	choose := append(append([]byte(nil), giver...), make([]byte, 4)...)
	responses, err = server.questChooseReward(&active, choose)
	if err != nil || len(responses) != 2 {
		t.Fatalf("choose responses=%d err=%v", len(responses), err)
	}
	complete, err := packet.Parse(responses[0])
	if err != nil || complete.Opcode != packet.SMSGQuestGiverQuestComplete || binary.LittleEndian.Uint32(complete.Data) != 500 {
		t.Fatalf("complete=%#v err=%v", complete, err)
	}
	item, found, err := characters.ItemAt(guid, 23, 23)
	if err != nil || !found || item.ItemTemplate != 300 || active.Money != 5 {
		t.Fatalf("reward item=%#v found=%v money=%d err=%v", item, found, active.Money, err)
	}
	state, found, err = characters.QuestState(guid, 500)
	if err != nil || !found || !state.Rewarded {
		t.Fatalf("rewarded=%#v found=%v err=%v", state, found, err)
	}
}
