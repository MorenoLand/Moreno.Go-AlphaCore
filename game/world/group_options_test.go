package world

import (
	"encoding/binary"
	"math"
	"net"
	"testing"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

func TestMinimapPingBroadcast(t *testing.T) {
	server := &WorldServer{}
	active, target := realm.Character{GUID: 1, Name: "Active"}, realm.Character{GUID: 2, Name: "Target"}
	server.registerPlayer(active)
	server.registerPlayer(target)
	group := server.groups.pending(active.GUID)
	server.groups.addMember(group, active.GUID)
	server.groups.addMember(group, target.GUID)
	activeClient, activeConnection := net.Pipe()
	targetClient, targetConnection := net.Pipe()
	defer activeClient.Close()
	defer targetClient.Close()
	server.attachPlayer(active.GUID, activeConnection)
	server.attachPlayer(target.GUID, targetConnection)
	read := func(client net.Conn, result chan<- packet.Packet) {
		value, _ := sockets.ReadPacket(client)
		result <- value
	}
	activeResult, targetResult := make(chan packet.Packet, 1), make(chan packet.Packet, 1)
	go read(activeClient, activeResult)
	go read(targetClient, targetResult)
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data, math.Float32bits(12))
	binary.LittleEndian.PutUint32(data[4:], math.Float32bits(34))
	if err := server.minimapPing(active, data); err != nil {
		t.Fatal(err)
	}
	for _, result := range []packet.Packet{<-activeResult, <-targetResult} {
		if result.Opcode != packet.MSGMinimapPing || len(result.Data) != 16 || binary.LittleEndian.Uint64(result.Data) != 1 || math.Float32frombits(binary.LittleEndian.Uint32(result.Data[8:])) != 12 || math.Float32frombits(binary.LittleEndian.Uint32(result.Data[12:])) != 34 {
			t.Fatalf("ping=%#v", result)
		}
	}
}
