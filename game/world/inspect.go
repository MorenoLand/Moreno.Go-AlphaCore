package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) inspect(active realm.Character, data []byte) error {
	if len(data) < 8 {
		return nil
	}
	targetGUID := int64(binary.LittleEndian.Uint64(data))
	target, found := s.playerByGUID(targetGUID)
	if !found || target.Map != active.Map {
		return nil
	}
	dx, dy, dz := target.PositionX-active.PositionX, target.PositionY-active.PositionY, target.PositionZ-active.PositionZ
	if dx*dx+dy*dy+dz*dz > maxShopDistance*maxShopDistance {
		return nil
	}
	s.setPlayerSelection(active.GUID, uint64(targetGUID))
	response, err := packet.Encode(packet.SMSGInspect, encodeGUID(active.GUID))
	if err != nil {
		return err
	}
	s.sendPlayer(targetGUID, response)
	return nil
}
