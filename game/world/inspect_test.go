package world

import (
	"encoding/binary"
	"net"
	"testing"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

func TestInspectNotifiesTarget(t *testing.T) {
	server := &WorldServer{}
	active, target := realm.Character{GUID: 1, Name: "Active", Map: 0}, realm.Character{GUID: 2, Name: "Target", Map: 0, Health: 1}
	server.registerPlayer(active)
	server.registerPlayer(target)
	client, connection := net.Pipe()
	defer client.Close()
	server.attachPlayer(target.GUID, connection)
	read := make(chan packet.Packet, 1)
	go func() { value, _ := sockets.ReadPacket(client); read <- value }()
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(target.GUID))
	if err := server.inspect(active, data); err != nil {
		t.Fatal(err)
	}
	message := <-read
	if message.Opcode != packet.SMSGInspect || len(message.Data) != 8 || binary.LittleEndian.Uint64(message.Data) != uint64(active.GUID) {
		t.Fatalf("inspect=%#v", message)
	}
	server.players.mu.RLock()
	selection := server.players.selection[active.GUID]
	server.players.mu.RUnlock()
	if selection != uint64(target.GUID) {
		t.Fatalf("selection=%d", selection)
	}
}
