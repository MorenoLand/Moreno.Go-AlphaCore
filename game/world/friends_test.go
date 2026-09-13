package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestFriendAndIgnorePackets(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO ChrRaces (ID, FactionID, MaleDisplayId, FemaleDisplayId, BaseLanguage, CreatureType) VALUES (1, 1, 49, 50, 1, 7)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	ownerGUID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Owner", Race: 1, Class: 1, Level: 1})
	if err != nil {
		t.Fatal(err)
	}
	targetGUID, err := characters.Create(realm.Character{AccountID: 2, RealmID: 1, Name: "Target", Race: 1, Class: 1, Level: 2})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Accounts: auth.NewStore(databases), Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	owner := realm.Character{GUID: ownerGUID, AccountID: 1, RealmID: 1, Name: "Owner", Race: 1, Class: 1, Level: 1}
	target := realm.Character{GUID: targetGUID, AccountID: 2, RealmID: 1, Name: "Target", Race: 1, Class: 1, Level: 2}
	server.registerPlayer(target)
	response, err := server.friendAdd(owner, []byte("Target\x00"), false)
	assertFriendStatus(t, response, friendAddedOnline, targetGUID, err)
	response, err = server.friendAdd(owner, []byte("Target\x00"), false)
	assertFriendStatus(t, response, friendAlready, targetGUID, err)
	responses, err := server.friendList(owner)
	if err != nil || len(responses) != 3 {
		t.Fatalf("friend list responses=%d err=%v", len(responses), err)
	}
	for index, expected := range []packet.Opcode{packet.SMSGNameQueryResponse, packet.SMSGFriendList, packet.SMSGIgnoreList} {
		parsed, parseErr := packet.Parse(responses[index])
		if parseErr != nil || parsed.Opcode != expected {
			t.Fatalf("friend list[%d]=%#v err=%v", index, parsed, parseErr)
		}
	}
	response, err = server.friendDelete(owner, encodeGUID(targetGUID), false)
	assertFriendStatus(t, response, friendRemoved, targetGUID, err)
	response, err = server.friendAdd(owner, []byte("Target\x00"), true)
	assertFriendStatus(t, response, ignoreAdded, targetGUID, err)
	response, err = server.friendDelete(owner, encodeGUID(targetGUID), true)
	assertFriendStatus(t, response, ignoreRemoved, targetGUID, err)
	server.unregisterPlayer(targetGUID)
	response, err = server.friendAdd(owner, []byte("Target\x00"), false)
	assertFriendStatus(t, response, friendAddedOffline, targetGUID, err)
	response, err = server.friendDelete(owner, encodeGUID(targetGUID), false)
	assertFriendStatus(t, response, friendRemoved, targetGUID, err)
}

func assertFriendStatus(t *testing.T, data []byte, expected byte, guid int64, err error) {
	t.Helper()
	parsed, parseErr := packet.Parse(data)
	if err != nil || parseErr != nil || parsed.Opcode != packet.SMSGFriendStatus || len(parsed.Data) != 9 || parsed.Data[0] != expected || int64(binary.LittleEndian.Uint64(parsed.Data[1:])) != guid {
		t.Fatalf("friend status=%#v err=%v parse=%v", parsed, err, parseErr)
	}
}
