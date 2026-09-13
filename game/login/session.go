package login

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/network/packet"
)

const maxUsernameLength = 64

var (
	ErrInvalidChallenge  = errors.New("invalid SRP6 challenge")
	ErrInvalidProof      = errors.New("invalid SRP6 proof")
	ErrUnknownAccount    = errors.New("unknown account")
	ErrIncorrectPassword = errors.New("incorrect password")
)

type Session struct {
	accounts      *auth.Store
	account       *auth.Account
	serverPrivate []byte
	serverPublic  []byte
	clientPublic  []byte
	clientProof   []byte
	sessionKey    []byte
}

func NewSession(accounts *auth.Store) *Session { return &Session{accounts: accounts} }

func (s *Session) Begin(data []byte) ([]byte, error) {
	if len(data) < 9 {
		return authFailure(packet.AuthFailed, packet.SRP6Challenge), ErrInvalidChallenge
	}
	length := int(data[8])
	if length <= 0 || length > maxUsernameLength || len(data) < 9+length {
		return authFailure(packet.AuthFailed, packet.SRP6Challenge), ErrInvalidChallenge
	}
	username := strings.TrimSpace(string(data[9 : 9+length]))
	if index := strings.IndexByte(username, 0); index >= 0 {
		username = username[:index]
	}
	account, err := s.accounts.Account(username)
	if err != nil {
		return authFailure(packet.AuthFailed, packet.SRP6Challenge), err
	}
	if account == nil {
		return authFailure(packet.AuthUnknownAccount, packet.SRP6Challenge), ErrUnknownAccount
	}
	salt, err := hexBytes(account.Salt)
	if err != nil || len(salt) != 32 {
		return authFailure(packet.AuthFailed, packet.SRP6Challenge), fmt.Errorf("invalid account salt: %w", err)
	}
	verifier, err := hexBytes(account.Verifier)
	if err != nil || len(verifier) != 32 {
		return authFailure(packet.AuthFailed, packet.SRP6Challenge), fmt.Errorf("invalid account verifier: %w", err)
	}
	s.serverPrivate = make([]byte, 32)
	if _, err := rand.Read(s.serverPrivate); err != nil {
		return authFailure(packet.AuthFailed, packet.SRP6Challenge), fmt.Errorf("generate SRP6 private key: %w", err)
	}
	s.serverPublic = packet.ServerPublicKey(verifier, s.serverPrivate)
	s.account = account
	response := []byte{byte(packet.AuthOK), byte(packet.SRP6Challenge), 1, packet.GeneratorBytes()[0], 32}
	response = append(response, packet.ModulusBytes()...)
	response = append(response, salt...)
	response = append(response, s.serverPublic...)
	return packet.EncodeSRP6(response)
}

func (s *Session) Proof(data []byte) ([]byte, error) {
	if s.account == nil || len(data) != 52 {
		return authFailure(packet.AuthFailed, packet.SRP6Proof), ErrInvalidProof
	}
	s.clientPublic = append([]byte(nil), data[:32]...)
	if !packet.ValidPublicKey(s.clientPublic) {
		return authFailure(packet.AuthFailed, packet.SRP6Proof), ErrInvalidProof
	}
	s.clientProof = append([]byte(nil), data[32:]...)
	salt, err := hexBytes(s.account.Salt)
	if err != nil {
		return authFailure(packet.AuthFailed, packet.SRP6Proof), err
	}
	verifier, err := hexBytes(s.account.Verifier)
	if err != nil {
		return authFailure(packet.AuthFailed, packet.SRP6Proof), err
	}
	u := packet.ScramblingParameter(s.clientPublic, s.serverPublic)
	shared := packet.ServerSKey(s.clientPublic, verifier, u, s.serverPrivate)
	s.sessionKey, err = packet.Interleaved(shared)
	if err != nil {
		return authFailure(packet.AuthFailed, packet.SRP6Proof), err
	}
	expected := packet.ClientProof(packet.XorNg, s.account.Name, s.sessionKey, s.clientPublic, s.serverPublic, salt)
	if subtle.ConstantTimeCompare(expected, s.clientProof) != 1 {
		return authFailure(packet.AuthIncorrectPassword, packet.SRP6Proof), ErrIncorrectPassword
	}
	if err := s.accounts.UpdateSessionKey(s.account.Name, s.sessionKey); err != nil {
		return authFailure(packet.AuthIncorrectPassword, packet.SRP6Proof), err
	}
	response := []byte{byte(packet.AuthOK), byte(packet.SRP6Proof)}
	response = append(response, packet.ServerProof(s.clientPublic, s.clientProof, s.sessionKey)...)
	zero := make([]byte, 4)
	response = append(response, zero...)
	return packet.EncodeSRP6(response)
}

func (s *Session) Account() *auth.Account { return s.account }

func authFailure(code packet.AuthCode, response packet.SRP6Response) []byte {
	result, _ := packet.EncodeSRP6([]byte{byte(code), byte(response)})
	return result
}

func hexBytes(value string) ([]byte, error) {
	return hex.DecodeString(value)
}
