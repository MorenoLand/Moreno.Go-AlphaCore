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
	"time"

	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

type WorldServer struct {
	Address           string
	Accounts          *auth.Store
	Characters        *realm.Store
	DBC               *dbc.Store
	WorldData         *worlddb.Store
	SupportedClient   uint32
	AutoCreateAccount bool
	ServerSeed        []byte
	players           playerRegistry
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
	var active *realm.Character
	logoutPending := false
	defer func() {
		if active != nil {
			s.unregisterPlayer(active.GUID)
			s.Characters.SetOnline(active.GUID, active.AccountID, active.RealmID, false)
		}
	}()
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
		response = nil
		responses := make([][]byte, 0, 1)
		switch message.Opcode {
		case packet.CMSGCharEnum:
			response, err = s.characterList(account.ID)
		case packet.CMSGCharCreate:
			response, err = s.characterCreate(account.ID, message.Data)
		case packet.CMSGCharDelete:
			response, err = s.characterDelete(account.ID, message.Data)
		case packet.CMSGPlayerLogin:
			response, err = s.playerLogin(account.ID, message.Data)
			if err == nil && len(message.Data) >= 8 {
				character, found, queryErr := s.Characters.Character(int64(binary.LittleEndian.Uint64(message.Data)), account.ID, 1)
				if queryErr != nil {
					err = queryErr
				} else if found {
					active = &character
					err = s.Characters.SetOnline(character.GUID, account.ID, 1, true)
					if err == nil {
						s.registerPlayer(character)
					}
				}
			}
		case packet.MSGMoveWorldportAck:
			if active == nil {
				return
			}
			response, err = s.initialPlayerPackets(*active)
		case packet.CMSGNameQuery:
			if active == nil {
				return
			}
			response, err = s.nameQuery(message.Data)
		case packet.CMSGItemQuerySingle:
			if active == nil {
				return
			}
			response, err = s.itemQuerySingle(message.Data)
		case packet.CMSGItemQueryMultiple:
			if active == nil {
				return
			}
			responses, err = s.itemQueryMultiple(message.Data)
		case packet.CMSGPageTextQuery:
			if active == nil {
				return
			}
			responses, err = s.pageTextQuery(*active, message.Data)
		case packet.CMSGQuestQuery:
			if active == nil {
				return
			}
			responses, err = s.questQuery(message.Data)
		case packet.CMSGGameObjectQuery:
			if active == nil {
				return
			}
			response, err = s.gameObjectQuery(message.Data)
		case packet.CMSGCreatureQuery:
			if active == nil {
				return
			}
			response, err = s.creatureQuery(message.Data)
		case packet.CMSGWho:
			if active == nil {
				return
			}
			response, err = s.who(*active, message.Data)
		case packet.CMSGFriendList:
			if active == nil {
				return
			}
			responses, err = s.friendList(*active)
		case packet.CMSGAddFriend:
			if active == nil {
				return
			}
			response, err = s.friendAdd(*active, message.Data, false)
		case packet.CMSGAddIgnore:
			if active == nil {
				return
			}
			response, err = s.friendAdd(*active, message.Data, true)
		case packet.CMSGDelFriend:
			if active == nil {
				return
			}
			response, err = s.friendDelete(*active, message.Data, false)
		case packet.CMSGDelIgnore:
			if active == nil {
				return
			}
			response, err = s.friendDelete(*active, message.Data, true)
		case packet.CMSGPlayedTime:
			if active == nil {
				return
			}
			response, err = s.playedTime(*active)
		case packet.CMSGSetActionButton:
			if active == nil {
				return
			}
			err = s.setActionButton(*active, message.Data)
		case packet.CMSGNewSpellSlot:
			if active == nil {
				return
			}
			err = s.newSpellSlot(*active, message.Data)
		case packet.MSGLookingForGroup:
			if active == nil {
				return
			}
			response, err = s.lookingForGroup(*active)
		case packet.CMSGSetLookingForGroup:
			if active == nil {
				return
			}
			s.setLookingForGroup(*active, message.Data)
		case packet.MSGRandomRoll:
			if active == nil {
				return
			}
			response, err = s.randomRoll(*active, message.Data)
		case packet.CMSGSetSelection:
			if active == nil {
				return
			}
			s.setSelection(*active, message.Data)
		case packet.CMSGSetTarget:
			if active == nil {
				return
			}
			s.setTarget(*active, message.Data)
		case packet.CMSGStandStateChange:
			if active == nil || len(message.Data) < 4 {
				return
			}
			s.setStandState(active.GUID, binary.LittleEndian.Uint32(message.Data))
		case packet.CMSGTextEmote:
			if active == nil {
				return
			}
			responses, err = s.textEmote(*active, message.Data)
		case packet.CMSGMessageChat:
			if active == nil {
				return
			}
			response, err = s.chat(active.GUID, message.Data)
		case packet.CMSGZoneUpdate:
			if active == nil {
				return
			}
			err = s.zoneUpdate(active.GUID, active.AccountID, message.Data)
			if err == nil && len(message.Data) >= 4 {
				active.Zone = int64(binary.LittleEndian.Uint32(message.Data))
				s.updatePlayer(*active)
			}
		case packet.CMSGPing:
			if active == nil || len(message.Data) < 4 {
				return
			}
			response, err = packet.Encode(packet.SMSGPong, message.Data)
		case packet.CMSGQueryTime:
			data := make([]byte, 4)
			binary.LittleEndian.PutUint32(data, uint32(time.Now().Unix()))
			response, err = packet.Encode(packet.SMSGQueryTimeResponse, data)
		case packet.CMSGLogoutRequest:
			if active == nil {
				return
			}
			logoutPending = true
			response, err = packet.Encode(packet.SMSGLogoutResponse, []byte{1})
		case packet.CMSGLogoutCancel:
			if active == nil || !logoutPending {
				return
			}
			logoutPending = false
			response, err = packet.Encode(packet.SMSGLogoutCancelAck, nil)
		case packet.CMSGPlayerLogout:
			if active == nil {
				return
			}
			response, err = packet.Encode(packet.SMSGLogoutComplete, nil)
			if err == nil {
				s.unregisterPlayer(active.GUID)
				err = s.Characters.SetOnline(active.GUID, active.AccountID, active.RealmID, false)
				active = nil
			}
		default:
			if active == nil || !packet.IsMovement(message.Opcode) || len(message.Data) < 48 {
				continue
			}
			active.PositionX = math.Float32frombits(binary.LittleEndian.Uint32(message.Data[24:28]))
			active.PositionY = math.Float32frombits(binary.LittleEndian.Uint32(message.Data[28:32]))
			active.PositionZ = math.Float32frombits(binary.LittleEndian.Uint32(message.Data[32:36]))
			active.Orientation = math.Float32frombits(binary.LittleEndian.Uint32(message.Data[36:40]))
			err = s.Characters.UpdatePosition(active.GUID, active.AccountID, active.RealmID, active.PositionX, active.PositionY, active.PositionZ, active.Orientation)
			s.updatePlayer(*active)
		}
		if err != nil {
			return
		}
		if response != nil {
			responses = append(responses, response)
		}
		for _, response := range responses {
			if sockets.WriteAll(connection, response) != nil {
				return
			}
		}
	}
}

func (s *WorldServer) characterCreate(accountID int64, data []byte) ([]byte, error) {
	result := byte(0x28)
	name, err := packet.ReadString(data, 0, 0)
	if err != nil || !validName(name) {
		result = 0x29
	} else {
		start := len(name) + 1
		if len(data) < start+9 {
			result = 0x29
		} else {
			raw := data[start : start+9]
			exists, queryErr := s.Characters.NameExists(name, 1)
			count, countErr := s.Characters.Count(accountID, 1)
			location, found, locationErr := s.WorldData.StartingLocation(raw[0], raw[1])
			if queryErr != nil || countErr != nil || locationErr != nil {
				return nil, firstError(queryErr, countErr, locationErr)
			}
			if exists {
				result = 0x2b
			} else if count >= 10 {
				result = 0x2a
			} else if !found {
				result = 0x29
			} else {
				stats, statsFound, statsErr := s.WorldData.ClassStats(raw[1], 1)
				if statsErr != nil {
					return nil, statsErr
				}
				if !statsFound {
					result = 0x29
					return packet.Encode(packet.SMSGCharCreate, []byte{result})
				}
				character := realm.Character{AccountID: accountID, RealmID: 1, Name: name, Race: raw[0], Class: raw[1], Gender: raw[2], Skin: raw[3], Face: raw[4], Hairstyle: raw[5], Haircolour: raw[6], Facialhair: raw[7], Level: 1, PositionX: location.PositionX, PositionY: location.PositionY, PositionZ: location.PositionZ, Map: location.Map, Orientation: location.Orientation, Zone: location.Zone, Health: stats.BaseHealth, Power1: stats.BaseMana}
				guid, createErr := s.Characters.Create(character)
				if createErr != nil {
					return nil, createErr
				}
				items, itemsErr := s.WorldData.StartingItems(raw[0], raw[1])
				if itemsErr != nil {
					return nil, itemsErr
				}
				for slot, item := range items {
					if err = s.Characters.AddInventoryItem(guid, item.ItemID, int64(23+slot), item.Amount); err != nil {
						return nil, err
					}
				}
				spells, spellsErr := s.WorldData.StartingSpells(raw[0], raw[1])
				if spellsErr != nil {
					return nil, spellsErr
				}
				for _, spell := range spells {
					if err = s.Characters.AddSpell(guid, spell); err != nil {
						return nil, err
					}
				}
				actions, actionsErr := s.WorldData.StartingActions(raw[0], raw[1])
				if actionsErr != nil {
					return nil, actionsErr
				}
				for _, action := range actions {
					if err = s.Characters.AddButton(guid, action.Button, action.Action); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return packet.Encode(packet.SMSGCharCreate, []byte{result})
}

func (s *WorldServer) characterDelete(accountID int64, data []byte) ([]byte, error) {
	result := byte(0x2e)
	if len(data) != 8 {
		result = 0x2f
	} else {
		guid := int64(binary.LittleEndian.Uint64(data))
		deleted, err := s.Characters.Delete(guid, accountID, 1)
		if err != nil {
			return nil, err
		}
		if !deleted {
			result = 0x2f
		}
	}
	return packet.Encode(packet.SMSGCharDelete, []byte{result})
}

func validName(value string) bool {
	if len(value) < 3 || len(value) > 12 || strings.Count(value, "`") > 1 || strings.Contains(value, " ") {
		return false
	}
	if strings.Count(value, "`") == 1 {
		value = strings.Replace(value, "`", "", 1)
	}
	for _, character := range value {
		if (character < 'A' || character > 'Z') && (character < 'a' || character > 'z') {
			return false
		}
	}
	return true
}

func firstError(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
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
		inventory, err := s.Characters.Inventory(character.GUID)
		if err != nil {
			return nil, err
		}
		items := make(map[int64]worlddb.ItemTemplate, len(inventory))
		if s.WorldData != nil {
			for _, item := range inventory {
				template, found, queryErr := s.WorldData.ItemTemplate(item.ItemTemplate)
				if queryErr != nil {
					return nil, queryErr
				}
				if found {
					items[item.Slot] = template
				}
			}
		}
		for slot := int64(0); slot < 20; slot++ {
			var displayID, inventoryType int64
			if item, found := items[slot]; found {
				displayID, inventoryType = item.DisplayID, item.InventoryType
			}
			encoded := make([]byte, 5)
			binary.LittleEndian.PutUint32(encoded, uint32(displayID))
			encoded[4] = byte(inventoryType)
			value = append(value, encoded...)
		}
		data = append(data, value...)
	}
	return packet.Encode(packet.SMSGCharEnum, data)
}
