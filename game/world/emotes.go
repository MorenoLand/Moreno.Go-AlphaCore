package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) textEmote(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 12 {
		return nil, nil
	}
	emote, found, err := s.DBC.EmoteText(int64(binary.LittleEndian.Uint32(data)))
	if err != nil || !found {
		return nil, err
	}
	body := append(encodeGUID(active.GUID), encodeUint32(emote.ID)...)
	if target, found := s.playerByGUID(int64(binary.LittleEndian.Uint64(data[4:]))); found {
		name, err := packet.StringBytes(target.Name)
		if err != nil {
			return nil, err
		}
		body = append(body, name...)
	} else {
		body = append(body, 0)
	}
	textPacket, err := packet.Encode(packet.SMSGTextEmote, body)
	if err != nil {
		return nil, err
	}
	responses := [][]byte{textPacket}
	switch emote.EmoteID {
	case 13:
		s.setStandState(active.GUID, 1)
	case 26:
		s.setStandState(active.GUID, 0)
	case 12:
		s.setStandState(active.GUID, 3)
	case 68:
		s.setStandState(active.GUID, 8)
	default:
		visual := append(encodeUint32(emote.EmoteID), encodeGUID(active.GUID)...)
		visualPacket, err := packet.Encode(packet.SMSGEmote, visual)
		if err != nil {
			return nil, err
		}
		responses = append(responses, visualPacket)
	}
	return responses, nil
}
