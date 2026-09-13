package world

import (
	"context"
	"encoding/binary"
	"net"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

func TestGroupInviteAcceptAndDisband(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO ChrRaces (ID, FactionID, MaleDisplayId, FemaleDisplayId, BaseLanguage, CreatureType) VALUES (1, 1, 49, 50, 1, 7)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	ownerGUID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Owner", Race: 1, Class: 1})
	if err != nil {
		t.Fatal(err)
	}
	targetGUID, err := characters.Create(realm.Character{AccountID: 2, RealmID: 1, Name: "Target", Race: 1, Class: 1})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	owner := realm.Character{GUID: ownerGUID, Name: "Owner", Race: 1, Class: 1}
	target := realm.Character{GUID: targetGUID, Name: "Target", Race: 1, Class: 1}
	server.registerPlayer(owner)
	server.registerPlayer(target)
	client, connection := net.Pipe()
	defer client.Close()
	defer connection.Close()
	server.attachPlayer(targetGUID, connection)
	packets := make(chan packet.Packet, 4)
	go func() {
		for {
			value, err := sockets.ReadPacket(client)
			if err != nil {
				return
			}
			packets <- value
		}
	}()
	responses, err := server.groupInvite(owner, []byte("Target\x00"))
	if err != nil || len(responses) != 1 {
		t.Fatalf("invite responses=%d err=%v", len(responses), err)
	}
	invite, err := packet.Parse(responses[0])
	if err != nil || invite.Opcode != packet.SMSGPartyCommandResult || binary.LittleEndian.Uint32(invite.Data) != partyInvite || binary.LittleEndian.Uint32(invite.Data[len(invite.Data)-4:]) != partyOK {
		t.Fatalf("invite response=%#v err=%v", invite, err)
	}
	invitePacket := <-packets
	if invitePacket.Opcode != packet.SMSGGroupInvite {
		t.Fatalf("invite packet=%#v", invitePacket)
	}
	name, err := packet.ReadString(invitePacket.Data, 0, 0)
	if err != nil || name != owner.Name {
		t.Fatalf("invite name=%q err=%v", name, err)
	}
	if err := server.groupAccept(target); err != nil {
		t.Fatal(err)
	}
	groupPacket := <-packets
	if groupPacket.Opcode != packet.SMSGGroupList {
		t.Fatalf("group list=%#v", groupPacket)
	}
	group, found, err := characters.GroupByPlayer(targetGUID)
	if err != nil || !found || group.LeaderGUID != ownerGUID {
		t.Fatalf("group=%#v found=%v err=%v", group, found, err)
	}
	members, err := characters.GroupMembers(group.ID)
	if err != nil || len(members) != 2 {
		t.Fatalf("members=%v err=%v", members, err)
	}
	if _, err := server.groupSetLeader(owner, []byte("Target\x00")); err != nil {
		t.Fatal(err)
	}
	_ = <-packets
	_ = <-packets
	if _, err := server.groupDisband(target); err != nil {
		t.Fatal(err)
	}
	if _, found, err := characters.GroupByPlayer(ownerGUID); err != nil || found {
		t.Fatalf("group after disband found=%v err=%v", found, err)
	}
}
