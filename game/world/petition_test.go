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

func TestPetitionLifecycle(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id, stackable) VALUES (5863, 'Guild Charter', 9199, 1); INSERT INTO creature_template (entry, display_id1, name, npc_flags) VALUES (100, 20, 'Guild Master', 128); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z) VALUES (1, 100, 0, 0, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	ownerGUID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Founder", Health: 1, Money: 2000})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	owner := realm.Character{GUID: ownerGUID, AccountID: 1, RealmID: 1, Name: "Founder", Health: 1, Money: 2000, Map: 0}
	npc := make([]byte, 8)
	binary.LittleEndian.PutUint64(npc, 0xf130000000000001)
	responses, err := server.petitionShowList(owner, npc)
	if err != nil || len(responses) != 1 {
		t.Fatalf("show list responses=%d err=%v", len(responses), err)
	}
	show, err := packet.Parse(responses[0])
	if err != nil || show.Opcode != packet.SMSGPetitionShowlist || show.Data[8] != 1 || binary.LittleEndian.Uint32(show.Data[13:]) != uint32(petitionCharterEntry) {
		t.Fatalf("show=%#v err=%v", show, err)
	}
	buyData := make([]byte, 20)
	copy(buyData, npc)
	buyData = append(buyData, []byte("Moreno\x00")...)
	responses, err = server.petitionBuy(&owner, buyData)
	if err != nil || len(responses) != 4 || owner.Money != 1000 {
		t.Fatalf("buy responses=%d money=%d err=%v", len(responses), owner.Money, err)
	}
	item, found, err := characters.ItemAt(ownerGUID, 23, 23)
	if err != nil || !found || item.ItemTemplate != petitionCharterEntry {
		t.Fatalf("charter=%#v found=%v err=%v", item, found, err)
	}
	petition, found, err := characters.PetitionByItemGUID(item.GUID)
	if err != nil || !found || petition.Name != "Moreno" {
		t.Fatalf("petition=%#v found=%v err=%v", petition, found, err)
	}
	queryData := make([]byte, 12)
	binary.LittleEndian.PutUint32(queryData, uint32(petition.ID))
	binary.LittleEndian.PutUint64(queryData[4:], uint64(item.GUID)|0x4000000000000000)
	responses, err = server.petitionQuery(owner, queryData)
	if err != nil || len(responses) != 1 {
		t.Fatalf("query responses=%d err=%v", len(responses), err)
	}
	if query, err := packet.Parse(responses[0]); err != nil || query.Opcode != packet.SMSGPetitionQueryResponse {
		t.Fatalf("query=%#v err=%v", query, err)
	}
	signData := queryData[4:]
	responses, err = server.petitionSignatures(owner, signData)
	if err != nil || len(responses) != 1 {
		t.Fatalf("signatures responses=%d err=%v", len(responses), err)
	}
	signatures, err := packet.Parse(responses[0])
	if err != nil || signatures.Opcode != packet.SMSGPetitionShowSignatures || signatures.Data[20] != 0 {
		t.Fatalf("signatures=%#v err=%v", signatures, err)
	}
	for index := 0; index < petitionMaxSignatures; index++ {
		guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Signer" + string(rune('A'+index)), Health: 1})
		if err != nil {
			t.Fatal(err)
		}
		if index == 0 {
			signer := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "SignerA", Health: 1}
			server.registerPlayer(signer)
			signData = make([]byte, 8)
			binary.LittleEndian.PutUint64(signData, uint64(item.GUID)|0x4000000000000000)
			responses, err = server.petitionSign(signer, signData)
			if err != nil || len(responses) != 1 {
				t.Fatalf("sign responses=%d err=%v", len(responses), err)
			}
		} else if err := characters.AddPetitionSigner(petition.ID, guid); err != nil {
			t.Fatal(err)
		}
	}
	turnInData := make([]byte, 8)
	binary.LittleEndian.PutUint64(turnInData, uint64(item.GUID)|0x4000000000000000)
	responses, err = server.petitionTurnIn(owner, turnInData)
	if err != nil || len(responses) != 2 || server.guilds.forPlayer(ownerGUID) == nil {
		t.Fatalf("turn in responses=%d guild=%v err=%v", len(responses), server.guilds.forPlayer(ownerGUID), err)
	}
	if _, found, err := characters.PetitionByItemGUID(item.GUID); err != nil || found {
		t.Fatalf("petition remains found=%v err=%v", found, err)
	}
	if _, found, err := characters.ItemByGUID(ownerGUID, item.GUID); err != nil || found {
		t.Fatalf("charter remains found=%v err=%v", found, err)
	}
}
