package world

import (
	"encoding/binary"
	"strings"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) nameQuery(data []byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, nil
	}
	character, found, err := s.Characters.CharacterByGUID(int64(binary.LittleEndian.Uint64(data)))
	if err != nil || !found {
		return nil, err
	}
	name, err := packet.StringBytes(character.Name)
	if err != nil {
		return nil, err
	}
	response := make([]byte, 0, 8+len(name)+12)
	guid := make([]byte, 8)
	binary.LittleEndian.PutUint64(guid, uint64(character.GUID))
	response = append(response, guid...)
	response = append(response, name...)
	for _, value := range []uint32{uint32(character.Race), uint32(character.Gender), uint32(character.Class)} {
		encoded := make([]byte, 4)
		binary.LittleEndian.PutUint32(encoded, value)
		response = append(response, encoded...)
	}
	return packet.Encode(packet.SMSGNameQueryResponse, response)
}

func (s *WorldServer) chat(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 {
		return nil, nil
	}
	chatType := data[0]
	if chatType != 0 && chatType != 2 && chatType != 3 && chatType != 4 && chatType != 5 && chatType != 7 && chatType != 13 {
		return nil, nil
	}
	if chatType == 2 || chatType == 3 {
		return s.guildChat(active, data, chatType)
	}
	if chatType == 13 {
		return s.channelChat(active, data)
	}
	messageOffset := 8
	var target realm.Character
	if chatType == 5 {
		targetName, err := packet.ReadString(data, messageOffset, 0)
		if err != nil {
			return nil, nil
		}
		messageOffset += len(targetName) + 1
		var found bool
		target, found = s.playerByName(strings.TrimSpace(targetName))
		if !found {
			return nil, nil
		}
	}
	message, err := packet.ReadString(data, messageOffset, 0)
	if err != nil || message == "" {
		return nil, err
	}
	language := binary.LittleEndian.Uint32(data[4:8])
	if chatType == 5 {
		inform, err := messageChatPacket(6, language, target.GUID, message)
		if err != nil {
			return nil, err
		}
		received, err := messageChatPacket(5, language, active.GUID, message)
		if err != nil {
			return nil, err
		}
		s.sendPlayer(target.GUID, received)
		return [][]byte{inform}, nil
	}
	response, err := messageChatPacket(chatType, language, active.GUID, message)
	if err != nil {
		return nil, err
	}
	s.broadcastPlayer(active, response)
	return [][]byte{response}, nil
}

func messageChatPacket(chatType byte, language uint32, guid int64, message string) ([]byte, error) {
	text, err := packet.StringBytes(message)
	if err != nil {
		return nil, err
	}
	response := make([]byte, 0, 1+4+8+len(text)+1)
	response = append(response, chatType)
	languageBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(languageBytes, language)
	response = append(response, languageBytes...)
	response = append(response, encodeGUID(guid)...)
	response = append(response, text...)
	response = append(response, 0)
	return packet.Encode(packet.SMSGMessageChat, response)
}

func (s *WorldServer) zoneUpdate(characterID, accountID int64, data []byte) error {
	if len(data) < 4 {
		return nil
	}
	return s.Characters.UpdateZone(characterID, accountID, 1, int64(binary.LittleEndian.Uint32(data)))
}
