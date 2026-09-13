package world

import (
	"encoding/binary"
	"net"
	"testing"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

func TestPlayerMacroAndMountBroadcast(t *testing.T) {
	server := &WorldServer{}
	active, target := realm.Character{GUID: 1, Name: "Active"}, realm.Character{GUID: 2, Name: "Target"}
	server.registerPlayer(active)
	server.registerPlayer(target)
	client, connection := net.Pipe()
	defer client.Close()
	server.attachPlayer(target.GUID, connection)
	read := make(chan packet.Packet, 2)
	go func() {
		for range 2 {
			value, _ := sockets.ReadPacket(client)
			read <- value
		}
	}()
	macro, err := server.playerMacro(active, []byte{3, 0, 0, 0})
	if err != nil {
		t.Fatal(err)
	}
	message, err := packet.Parse(macro)
	if err != nil || message.Opcode != packet.SMSGPlayerMacro {
		t.Fatalf("macro=%#v err=%v", message, err)
	}
	if message = <-read; message.Opcode != packet.SMSGPlayerMacro {
		t.Fatalf("macro broadcast=%#v", message)
	}
	if err := server.mountSpecialAnim(active); err != nil {
		t.Fatal(err)
	}
	message = <-read
	if message.Opcode != packet.SMSGMountSpecialAnim || len(message.Data) != 8 || binary.LittleEndian.Uint64(message.Data) != 1 {
		t.Fatalf("mount=%#v", message)
	}
}
