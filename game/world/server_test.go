package world

import (
	"context"
	"encoding/binary"
	"net"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

func TestLegacyAuthAndCharacterEnum(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	server := &WorldServer{Accounts: auth.NewStore(databases), Characters: realm.NewStore(databases), SupportedClient: 3368, AutoCreateAccount: true, ServerSeed: []byte{1, 2, 3, 4}}
	client, connection := net.Pipe()
	defer client.Close()
	done := make(chan struct{})
	go func() {
		server.handle(connection)
		close(done)
	}()
	challenge, err := sockets.ReadPacket(client)
	if err != nil || challenge.Opcode != packet.SMSGAuthChallenge || string(challenge.Data) != string(server.ServerSeed) {
		t.Fatalf("challenge=%#v err=%v", challenge, err)
	}
	authData := make([]byte, 8)
	binary.LittleEndian.PutUint32(authData, 3368)
	authData = append(authData, []byte("PLAYER\x00PASSWORD\x00")...)
	authPacket, err := packet.Encode(packet.CMSGAuthSession, authData)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Write(authPacket); err != nil {
		t.Fatal(err)
	}
	response, err := sockets.ReadPacket(client)
	if err != nil || response.Opcode != packet.SMSGAuthResponse || len(response.Data) != 1 || response.Data[0] != byte(packet.AuthOK) {
		t.Fatalf("response=%#v err=%v", response, err)
	}
	enumPacket, err := packet.Encode(packet.CMSGCharEnum, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Write(enumPacket); err != nil {
		t.Fatal(err)
	}
	characters, err := sockets.ReadPacket(client)
	if err != nil || characters.Opcode != packet.SMSGCharEnum || len(characters.Data) != 1 || characters.Data[0] != 0 {
		t.Fatalf("characters=%#v err=%v", characters, err)
	}
	client.Close()
	<-done
}
