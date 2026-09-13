package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) bankerActivate(active realm.Character, data []byte) ([]byte, error) {
	return s.npcActivation(active, data, 0x20, packet.SMSGShowBank)
}

func (s *WorldServer) tabardVendorActivate(active realm.Character, data []byte) ([]byte, error) {
	return s.npcActivation(active, data, 0x40, packet.MSGTabardVendorActivate)
}

func (s *WorldServer) npcActivation(active realm.Character, data []byte, flag int64, opcode packet.Opcode) ([]byte, error) {
	if len(data) < 8 {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	_, creature, found, err := s.questGiverAt(active, guid)
	if err != nil || !found || creature.NPCFlags&flag == 0 {
		return nil, err
	}
	return packet.Encode(opcode, encodeGUID(int64(guid)))
}
