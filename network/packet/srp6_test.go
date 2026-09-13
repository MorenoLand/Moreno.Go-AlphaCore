package packet

import (
	"bytes"
	"testing"
)

func TestSRP6SharedSecret(t *testing.T) {
	salt := bytes.Repeat([]byte{0x31}, 32)
	username, password := "PLAYER", "PASSWORD"
	verifier := PasswordVerifier(username, password, salt)
	clientPrivateKey := bytes.Repeat([]byte{0x41}, 32)
	serverPrivateKey := bytes.Repeat([]byte{0x52}, 32)
	clientPublicKey := ClientPublicKey(clientPrivateKey)
	serverPublicKey := ServerPublicKey(verifier, serverPrivateKey)
	u := ScramblingParameter(clientPublicKey, serverPublicKey)
	x := CalculateX(username, password, salt)
	clientSecret := ClientSKey(clientPrivateKey, serverPublicKey, x, u)
	serverSecret := ServerSKey(clientPublicKey, verifier, u, serverPrivateKey)
	if !bytes.Equal(clientSecret, serverSecret) {
		t.Fatal("client and server SRP secrets differ")
	}
	clientSession, err := Interleaved(clientSecret)
	if err != nil {
		t.Fatal(err)
	}
	serverSession, err := Interleaved(serverSecret)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(clientSession, serverSession) {
		t.Fatal("client and server SRP sessions differ")
	}
}
