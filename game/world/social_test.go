package world

import (
	"net"
	"testing"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

func TestChatBroadcast(t *testing.T) {
	server := &WorldServer{}
	owner := realm.Character{GUID: 1, Name: "Owner", Map: 0}
	target := realm.Character{GUID: 2, Name: "Target", Map: 0}
	server.registerPlayer(owner)
	server.registerPlayer(target)
	client, connection := net.Pipe()
	defer client.Close()
	defer connection.Close()
	server.attachPlayer(target.GUID, connection)
	forwarded := make(chan packet.Packet, 1)
	go func() {
		value, err := sockets.ReadPacket(client)
		if err == nil {
			forwarded <- value
		}
	}()
	data := append([]byte{0, 0, 0, 0, 0, 0, 0, 0}, []byte("Hello\x00")...)
	responses, err := server.chat(owner, data)
	if err != nil || len(responses) != 1 {
		t.Fatalf("chat responses=%d err=%v", len(responses), err)
	}
	sender, err := packet.Parse(responses[0])
	if err != nil || sender.Opcode != packet.SMSGMessageChat || len(sender.Data) != 20 {
		t.Fatalf("sender=%#v err=%v", sender, err)
	}
	message, err := packet.ReadString(sender.Data, 13, 0)
	if err != nil || message != "Hello" || sender.Data[19] != 0 {
		t.Fatalf("sender data=%#v message=%q err=%v", sender.Data, message, err)
	}
	received := <-forwarded
	if received.Opcode != packet.SMSGMessageChat || string(received.Data) != string(sender.Data) {
		t.Fatalf("received=%#v sender=%#v", received, sender)
	}
}
