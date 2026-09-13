package world

import (
	"context"
	"net"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

func TestChannelJoinChatListLeave(t *testing.T) {
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
	responses, err := server.joinChannel(owner, []byte("Trade\x00"))
	if err != nil || len(responses) != 1 {
		t.Fatalf("owner join responses=%d err=%v", len(responses), err)
	}
	assertChannelNotification(t, responses[0], channelYouJoined)
	client, connection := net.Pipe()
	defer client.Close()
	defer connection.Close()
	server.attachPlayer(targetGUID, connection)
	forwarded := make(chan packet.Packet, 1)
	go func() {
		value, err := sockets.ReadPacket(client)
		if err == nil {
			forwarded <- value
		}
	}()
	responses, err = server.joinChannel(target, []byte("Trade\x00"))
	if err != nil || len(responses) != 1 {
		t.Fatalf("target join responses=%d err=%v", len(responses), err)
	}
	assertChannelNotification(t, responses[0], channelYouJoined)
	chatData := append([]byte{13, 0, 0, 0, 0, 0, 0, 0}, []byte("Trade\x00Hello\x00")...)
	responses, err = server.chat(owner, chatData)
	if err != nil || len(responses) != 1 {
		t.Fatalf("channel chat responses=%d err=%v", len(responses), err)
	}
	chat, err := packet.Parse(responses[0])
	if err != nil || chat.Opcode != packet.SMSGMessageChat || chat.Data[0] != 13 {
		t.Fatalf("owner chat=%#v err=%v", chat, err)
	}
	if received := <-forwarded; received.Opcode != packet.SMSGMessageChat || received.Data[0] != 13 {
		t.Fatalf("target chat=%#v", received)
	}
	responses, err = server.listChannel(owner, []byte("Trade\x00"))
	if err != nil || len(responses) != 1 {
		t.Fatalf("channel list responses=%d err=%v", len(responses), err)
	}
	listed, err := packet.Parse(responses[0])
	if err != nil || listed.Opcode != packet.SMSGChannelList {
		t.Fatalf("channel list=%#v err=%v", listed, err)
	}
	responses, err = server.leaveChannel(target, []byte("Trade\x00"))
	if err != nil || len(responses) != 1 {
		t.Fatalf("target leave responses=%d err=%v", len(responses), err)
	}
	assertChannelNotification(t, responses[0], channelYouLeft)
}

func assertChannelNotification(t *testing.T, data []byte, expected byte) {
	t.Helper()
	parsed, err := packet.Parse(data)
	if err != nil || parsed.Opcode != packet.SMSGChannelNotify || len(parsed.Data) < 2 || parsed.Data[0] != expected {
		t.Fatalf("channel notification=%#v err=%v", parsed, err)
	}
}
