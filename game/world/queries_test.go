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

func TestQueryPackets(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	world := worlddb.NewStore(databases)
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, description, display_id, quality, inventory_type, dmg_min1, dmg_max1, dmg_type1, spellid_1, bonding) VALUES (25, 'Worn Shortsword', 'A blade', 1542, 1, 21, 6.5, 10.5, 0, 0, 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO page_text (entry, text, next_page) VALUES (1, '$N$B$C', 2), (2, 'Last page', 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, name, subname, static_flags, type, beast_family) VALUES (100, 'Wolf', 'Forest Hunter', 3, 1, 2)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO gameobject_template (entry, type, displayId, name, data0) VALUES (200, 3, 10, 'Chest', 43)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO quest_template (entry, Method, ZoneOrSort, QuestLevel, Type, NextQuestInChain, SrcItemId, RewOrReqMoney, RewItemId1, RewItemCount1, RewChoiceItemId1, RewChoiceItemCount1, Title, Details, Objectives, EndText, ReqCreatureOrGOId1, ReqCreatureOrGOCount1, ReqItemId1, ReqItemCount1) VALUES (500, 2, 12, 1, 0, 0, 0, 100, 25, 1, 25, 2, 'First Quest', 'Details', 'Objectives', 'End', 100, 3, 25, 2)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{WorldData: world}
	single, err := server.itemQuerySingle(append(make([]byte, 0, 12), 25, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := packet.Parse(single)
	if err != nil || parsed.Opcode != packet.SMSGItemQuerySingleResponse || len(parsed.Data) < 12 {
		t.Fatalf("item single=%#v err=%v", parsed, err)
	}
	multipleData := make([]byte, 12)
	binary.LittleEndian.PutUint32(multipleData, 2)
	binary.LittleEndian.PutUint32(multipleData[4:], 25)
	binary.LittleEndian.PutUint32(multipleData[8:], 25)
	multiple, err := server.itemQueryMultiple(multipleData)
	if err != nil || len(multiple) != 1 {
		t.Fatalf("item multiple packets=%d err=%v", len(multiple), err)
	}
	parsed, err = packet.Parse(multiple[0])
	if err != nil || parsed.Opcode != packet.SMSGItemQueryMultipleResponse || parsed.Data[0] != 2 {
		t.Fatalf("item multiple=%#v err=%v", parsed, err)
	}
	pageData := make([]byte, 12)
	binary.LittleEndian.PutUint32(pageData, 1)
	pages, err := server.pageTextQuery(realm.Character{Name: "Reader", Class: 1}, pageData)
	if err != nil || len(pages) != 2 {
		t.Fatalf("pages=%d err=%v", len(pages), err)
	}
	creatureData := make([]byte, 12)
	binary.LittleEndian.PutUint32(creatureData, 100)
	binary.LittleEndian.PutUint64(creatureData[4:], 1)
	creature, err := server.creatureQuery(creatureData)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err = packet.Parse(creature)
	if err != nil || parsed.Opcode != packet.SMSGCreatureQueryResponse {
		t.Fatalf("creature=%#v err=%v", parsed, err)
	}
	gameObjectData := make([]byte, 12)
	binary.LittleEndian.PutUint32(gameObjectData, 200)
	binary.LittleEndian.PutUint64(gameObjectData[4:], 1)
	gameObject, err := server.gameObjectQuery(gameObjectData)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err = packet.Parse(gameObject)
	if err != nil || parsed.Opcode != packet.SMSGGameObjectQueryResponse {
		t.Fatalf("gameobject=%#v err=%v", parsed, err)
	}
	questData := make([]byte, 4)
	binary.LittleEndian.PutUint32(questData, 500)
	questPackets, err := server.questQuery(questData)
	if err != nil || len(questPackets) != 2 {
		t.Fatalf("quest packets=%d err=%v", len(questPackets), err)
	}
	parsed, err = packet.Parse(questPackets[0])
	if err != nil || parsed.Opcode != packet.SMSGCreatureQueryResponse {
		t.Fatalf("quest creature=%#v err=%v", parsed, err)
	}
	parsed, err = packet.Parse(questPackets[1])
	if err != nil || parsed.Opcode != packet.SMSGQuestQueryResponse || len(parsed.Data) < 32 {
		t.Fatalf("quest=%#v err=%v", parsed, err)
	}
}
