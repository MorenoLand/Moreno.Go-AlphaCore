package world

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"strings"

	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

type WorldServer struct {
	Address           string
	Accounts          *auth.Store
	Characters        *realm.Store
	SupportedClient   uint32
	AutoCreateAccount bool
	ServerSeed        []byte
}

func (s *WorldServer) Start(ctx context.Context) (net.Listener, error) {
	if len(s.ServerSeed) != 4 {
		s.ServerSeed = make([]byte, 4)
		if _, err := rand.Read(s.ServerSeed); err != nil {
			return nil, fmt.Errorf("generate world server seed: %w", err)
		}
	}
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return nil, fmt.Errorf("listen world server: %w", err)
	}
	go func() {
		<-ctx.Done()
		listener.Close()
	}()
	go s.accept(ctx, listener)
	return listener, nil
}

func (s *WorldServer) accept(ctx context.Context, listener net.Listener) {
	for {
		connection, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return
			}
			continue
		}
		go s.handle(connection)
	}
}

func (s *WorldServer) handle(connection net.Conn) {
	defer connection.Close()
	challenge, err := packet.Encode(packet.SMSGAuthChallenge, s.ServerSeed)
	if err != nil || sockets.WriteAll(connection, challenge) != nil {
		return
	}
	message, err := sockets.ReadPacket(connection)
	if err != nil || message.Opcode != packet.CMSGAuthSession {
		return
	}
	account, code := s.authenticate(message.Data)
	response, err := packet.Encode(packet.SMSGAuthResponse, []byte{byte(code)})
	if err != nil || sockets.WriteAll(connection, response) != nil || code != packet.AuthOK {
		return
	}
	for {
		message, err = sockets.ReadPacket(connection)
		if err != nil {
			return
		}
		switch message.Opcode {
		case packet.CMSGCharEnum:
			response, err = s.characterList(account.ID)
		case packet.CMSGPing:
			if len(message.Data) < 4 {
				return
			}
			response, err = packet.Encode(packet.SMSGPong, message.Data)
		default:
			return
		}
		if err != nil || sockets.WriteAll(connection, response) != nil {
			return
		}
	}
}

func (s *WorldServer) authenticate(data []byte) (*auth.Account, packet.AuthCode) {
	if len(data) < 8 {
		return nil, packet.AuthFailed
	}
	if binary.LittleEndian.Uint32(data[:4]) != s.SupportedClient {
		return nil, packet.AuthVersionMismatch
	}
	username, err := packet.ReadString(data, 8, 0)
	if err != nil || strings.TrimSpace(username) == "" {
		return nil, packet.AuthUnknownAccount
	}
	username = strings.TrimSpace(username)
	offset := 8 + len(username) + 1
	remaining := data[offset:]
	password, legacy := legacyPassword(remaining)
	account, err := s.Accounts.Account(username)
	if err != nil {
		return nil, packet.AuthFailed
	}
	if account == nil && legacy && s.AutoCreateAccount {
		if err := s.Accounts.CreateAccount(username, password, "", 0); err != nil {
			return nil, packet.AuthFailed
		}
		account, err = s.Accounts.Account(username)
		if err != nil {
			return nil, packet.AuthFailed
		}
	}
	if account == nil {
		return nil, packet.AuthUnknownAccount
	}
	if legacy {
		return account, verifyPassword(account.Password, password)
	}
	if len(remaining) < 24 {
		return nil, packet.AuthSessionExpired
	}
	key, err := hexDecode(account.SessionKey)
	if err != nil || len(key) == 0 {
		return nil, packet.AuthSessionExpired
	}
	expected := packet.WorldServerProof(account.Name, remaining[:4], s.ServerSeed, key)
	if !equal(expected, remaining[4:24]) {
		return nil, packet.AuthIncorrectPassword
	}
	return account, packet.AuthOK
}

func legacyPassword(data []byte) (string, bool) {
	if len(data) == 0 {
		return "", false
	}
	end := strings.IndexByte(string(data), 0)
	if end <= 0 || end > 64 {
		return "", false
	}
	for _, value := range data[:end] {
		if value < 0x20 || value > 0x7e {
			return "", false
		}
	}
	return string(data[:end]), true
}

func (s *WorldServer) characterList(accountID int64) ([]byte, error) {
	characters, err := s.Characters.Characters(accountID, 1)
	if err != nil {
		return nil, err
	}
	if len(characters) > 255 {
		return nil, fmt.Errorf("character list exceeds protocol limit")
	}
	data := []byte{byte(len(characters))}
	for _, character := range characters {
		name, err := packet.StringBytes(character.Name)
		if err != nil {
			return nil, err
		}
		value := make([]byte, 0, 8+len(name)+9+8+12+16+100)
		guid := make([]byte, 8)
		binary.LittleEndian.PutUint64(guid, uint64(character.GUID))
		value = append(value, guid...)
		value = append(value, name...)
		value = append(value, character.Race, character.Class, character.Gender, character.Skin, character.Face, character.Hairstyle, character.Haircolour, character.Facialhair, character.Level)
		for _, number := range []int64{character.Zone, character.Map} {
			encoded := make([]byte, 4)
			binary.LittleEndian.PutUint32(encoded, uint32(number))
			value = append(value, encoded...)
		}
		for _, number := range []float32{character.PositionX, character.PositionY, character.PositionZ} {
			encoded := make([]byte, 4)
			binary.LittleEndian.PutUint32(encoded, math.Float32bits(number))
			value = append(value, encoded...)
		}
		value = append(value, make([]byte, 16)...)
		value = append(value, make([]byte, 100)...)
		data = append(data, value...)
	}
	return packet.Encode(packet.SMSGCharEnum, data)
}
