package login

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/network/packet"
)

func TestSRP6LoginListener(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	accounts := auth.NewStore(databases)
	if err := accounts.CreateAccount("PLAYER", "PASSWORD", "", 0); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	listener, err := (LoginServer{Address: "127.0.0.1:0", Accounts: accounts}).Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	client, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	username := []byte("PLAYER")
	begin := make([]byte, 9+len(username))
	begin[8] = byte(len(username))
	copy(begin[9:], username)
	message, err := packet.Encode(packet.CMSGAuthSRP6Begin, begin)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Write(message); err != nil {
		t.Fatal(err)
	}
	challenge, err := readSRP6(client)
	if err != nil {
		t.Fatal(err)
	}
	if len(challenge) != 101 || challenge[0] != byte(packet.AuthOK) || challenge[1] != byte(packet.SRP6Challenge) {
		t.Fatalf("unexpected challenge length=%d data=%x", len(challenge), challenge)
	}
	salt := challenge[37:69]
	serverPublic := challenge[69:101]
	clientPrivate := bytes.Repeat([]byte{0x23}, 32)
	clientPublic := packet.ClientPublicKey(clientPrivate)
	u := packet.ScramblingParameter(clientPublic, serverPublic)
	x := packet.CalculateX("PLAYER", "PASSWORD", salt)
	shared := packet.ClientSKey(clientPrivate, serverPublic, x, u)
	session, err := packet.Interleaved(shared)
	if err != nil {
		t.Fatal(err)
	}
	proof := packet.ClientProof(packet.XorNg, "PLAYER", session, clientPublic, serverPublic, salt)
	message, err = packet.Encode(packet.CMSGAuthSRP6Proof, append(clientPublic, proof...))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Write(message); err != nil {
		t.Fatal(err)
	}
	response, err := readSRP6(client)
	if err != nil {
		t.Fatal(err)
	}
	if len(response) != 26 || response[0] != byte(packet.AuthOK) || response[1] != byte(packet.SRP6Proof) {
		t.Fatalf("unexpected proof response length=%d data=%x", len(response), response)
	}
	if !bytes.Equal(response[2:22], packet.ServerProof(clientPublic, proof, session)) {
		t.Fatal("server SRP6 proof mismatch")
	}
}

func readSRP6(reader io.Reader) ([]byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(reader, header); err != nil {
		return nil, err
	}
	data := make([]byte, int(binary.BigEndian.Uint16(header)))
	_, err := io.ReadFull(reader, data)
	return data, err
}
