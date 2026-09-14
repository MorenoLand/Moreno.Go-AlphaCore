package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestSRPWorldAuth(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	accounts := auth.NewStore(databases)
	if err := accounts.CreateAccount("PLAYER", "PASSWORD", "", 0); err != nil {
		t.Fatal(err)
	}
	key := []byte("0123456789abcdefghijklmnopqrstuv")
	if err := accounts.UpdateSessionKey("PLAYER", key); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Accounts: accounts, Characters: realm.NewStore(databases), WorldData: worlddb.NewStore(databases), SupportedClient: 3368, ServerSeed: []byte{1, 2, 3, 4}}
	clientSeed := []byte{5, 6, 7, 8}
	expected := packet.WorldServerProof("PLAYER", clientSeed, server.ServerSeed, key)
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data, 3368)
	data = append(data, []byte("PLAYER\x00")...)
	data = append(data, clientSeed...)
	data = append(data, expected...)
	account, code := server.authenticate(data)
	if code != packet.AuthOK || account == nil || account.Name != "PLAYER" {
		t.Fatalf("account=%#v code=%x", account, code)
	}
}

func TestLegacyWorldAuth(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	accounts := auth.NewStore(databases)
	if err := accounts.CreateAccount("PLAYER", "PASSWORD", "", 0); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Accounts: accounts, Characters: realm.NewStore(databases), WorldData: worlddb.NewStore(databases), SupportedClient: 3368, ServerSeed: []byte{1, 2, 3, 4}}
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data, 3368)
	data = append(data, []byte("PLAYER PASSWORD\x00")...)
	account, code := server.authenticate(data)
	if code != packet.AuthOK || account == nil || account.Name != "PLAYER" {
		t.Fatalf("account=%#v code=%x", account, code)
	}
}
