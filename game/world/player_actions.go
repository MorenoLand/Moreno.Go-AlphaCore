package world

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"math/big"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) playedTime(active realm.Character) ([]byte, error) {
	body := append(encodeUint32(active.Totaltime), encodeUint32(active.Leveltime)...)
	return packet.Encode(packet.SMSGPlayedTime, body)
}

func (s *WorldServer) setActionButton(active realm.Character, data []byte) error {
	if len(data) < 5 {
		return nil
	}
	return s.Characters.SetButton(active.GUID, int64(data[0]), int64(int32(binary.LittleEndian.Uint32(data[1:]))))
}

func (s *WorldServer) newSpellSlot(active realm.Character, data []byte) error {
	if len(data) < 8 {
		return nil
	}
	spell := int64(int32(binary.LittleEndian.Uint32(data)))
	index := int64(int32(binary.LittleEndian.Uint32(data[4:])))
	return s.Characters.SetSpellButton(active.GUID, spell, index)
}

func (s *WorldServer) lookingForGroup(active realm.Character) ([]byte, error) {
	return packet.Encode(packet.MSGLookingForGroup, encodeUint32(int64(s.getGroupStatus(active.GUID))))
}

func (s *WorldServer) setLookingForGroup(active realm.Character, data []byte) {
	if len(data) < 4 || s.getGroupStatus(active.GUID) == 1 {
		return
	}
	status := uint32(0)
	if binary.LittleEndian.Uint32(data) != 0 {
		status = 2
	}
	s.setGroupStatus(active.GUID, status)
}

func (s *WorldServer) randomRoll(active realm.Character, data []byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, nil
	}
	minimum := binary.LittleEndian.Uint32(data)
	maximum := binary.LittleEndian.Uint32(data[4:])
	if maximum < minimum {
		return nil, nil
	}
	span := new(big.Int).SetUint64(uint64(maximum) - uint64(minimum) + 1)
	value, err := cryptorand.Int(cryptorand.Reader, span)
	if err != nil {
		return nil, err
	}
	body := append(encodeUint32(int64(minimum)), encodeUint32(int64(maximum))...)
	body = append(body, encodeUint32(int64(minimum+uint32(value.Uint64())))...)
	body = append(body, encodeGUID(active.GUID)...)
	return packet.Encode(packet.MSGRandomRoll, body)
}

func (s *WorldServer) setSelection(active realm.Character, data []byte) {
	if len(data) >= 8 {
		s.setPlayerSelection(active.GUID, binary.LittleEndian.Uint64(data))
	}
}

func (s *WorldServer) setTarget(active realm.Character, data []byte) {
	if len(data) >= 8 {
		s.setPlayerTarget(active.GUID, binary.LittleEndian.Uint64(data))
	}
}
