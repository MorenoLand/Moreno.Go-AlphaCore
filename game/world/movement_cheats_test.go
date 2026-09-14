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

func TestMovementSpeedCheat(t *testing.T) {
	server := &WorldServer{}
	active := realm.Character{GUID: 1, Name: "GM", Health: 1}
	server.registerPlayer(active)
	client, connection := net.Pipe()
	defer client.Close()
	server.attachPlayer(active.GUID, connection)
	data := make([]byte, 52)
	binary.LittleEndian.PutUint32(data[48:], math.Float32bits(12.5))
	done := make(chan error, 1)
	go func() { done <- server.movementSpeedCheat(&active, packet.MSGMoveSetRunSpeedCheat, data, 1) }()
	response, err := sockets.ReadPacket(client)
	if err != nil || response.Opcode != packet.SMSGForceSpeedChange || len(response.Data) != 4 || math.Float32frombits(binary.LittleEndian.Uint32(response.Data)) != 12.5 {
		t.Fatalf("response=%#v err=%v", response, err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if speed := server.playerSpeeds(active.GUID); speed[1] != 12.5 {
		t.Fatalf("speeds=%#v", speed)
	}
	if err := server.movementSpeedCheat(&active, packet.MSGMoveSetWalkSpeedCheat, data[:48], 1); err != nil {
		t.Fatal(err)
	}
}
