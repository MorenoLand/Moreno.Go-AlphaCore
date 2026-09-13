package world

import (
	"encoding/binary"

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

func (s *WorldServer) chat(characterID int64, data []byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, nil
	}
	chatType := data[0]
	if chatType != 0 && chatType != 4 && chatType != 7 {
		return nil, nil
	}
	message, err := packet.ReadString(data, 8, 0)
	if err != nil || message == "" {
		return nil, err
	}
	text, err := packet.StringBytes(message)
	if err != nil {
		return nil, err
	}
	response := make([]byte, 0, 1+4+8+len(text)+1)
	response = append(response, chatType)
	response = append(response, data[4:8]...)
	guid := make([]byte, 8)
	binary.LittleEndian.PutUint64(guid, uint64(characterID))
	response = append(response, guid...)
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
