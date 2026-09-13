package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) playerMacro(active realm.Character, data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, nil
	}
	category := binary.LittleEndian.Uint32(data)
	if category > 0xd {
		return nil, nil
	}
	body := append(encodeGUID(active.GUID), data[:4]...)
	voice, err := packet.Encode(packet.SMSGPlayerMacro, body)
	if err != nil {
		return nil, err
	}
	s.broadcastPlayer(active, voice)
	return voice, nil
}

func (s *WorldServer) mountSpecialAnim(active realm.Character) error {
	voice, err := packet.Encode(packet.SMSGMountSpecialAnim, encodeGUID(active.GUID))
	if err != nil {
		return err
	}
	s.broadcastPlayer(active, voice)
	return nil
}
