package world

import (
	"context"
	"encoding/binary"
	"math"
	"net"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

func TestMovementPersistsAndBroadcasts(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	senderID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Sender", Race: 1, Class: 1, PositionX: 1, PositionY: 2, PositionZ: 3})
	if err != nil {
		t.Fatal(err)
	}
	targetID, err := characters.Create(realm.Character{AccountID: 2, RealmID: 1, Name: "Target", Race: 1, Class: 1, PositionX: 1, PositionY: 2, PositionZ: 3})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters}
	sender := realm.Character{GUID: senderID, AccountID: 1, RealmID: 1, Name: "Sender", PositionX: 1, PositionY: 2, PositionZ: 3}
	target := realm.Character{GUID: targetID, AccountID: 2, RealmID: 1, Name: "Target"}
	server.registerPlayer(sender)
	server.registerPlayer(target)
	client, connection := net.Pipe()
	defer client.Close()
	server.attachPlayer(targetID, connection)
	data := make([]byte, 48)
	binary.LittleEndian.PutUint32(data[24:], math.Float32bits(4))
	binary.LittleEndian.PutUint32(data[28:], math.Float32bits(5))
	binary.LittleEndian.PutUint32(data[32:], math.Float32bits(6))
	binary.LittleEndian.PutUint32(data[36:], math.Float32bits(0.5))
	done := make(chan error, 1)
	go func() { done <- server.updateMovement(&sender, packet.MSGMoveHeartbeat, data) }()
	message, err := sockets.ReadPacket(client)
	if err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if message.Opcode != packet.MSGMoveHeartbeat || len(message.Data) != 56 || int64(binary.LittleEndian.Uint64(message.Data)) != senderID {
		t.Fatalf("movement packet=%#v", message)
	}
	stored, found, err := characters.Character(senderID, 1, 1)
	if err != nil || !found || stored.PositionX != 4 || stored.PositionY != 5 || stored.PositionZ != 6 || stored.Orientation != 0.5 {
		t.Fatalf("stored=%#v found=%v err=%v", stored, found, err)
	}
}
