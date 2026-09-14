package world

import (
	"encoding/binary"
	"math"

	"Moreno.AlphaCore/database/realm"
)

func (s *WorldServer) worldTeleport(active *realm.Character, data []byte, gmLevel int) ([]byte, error) {
	if active == nil || gmLevel <= 0 || len(data) != 21 && len(data) != 24 {
		return nil, nil
	}
	mapID, offset := int64(data[4]), 5
	if len(data) == 24 {
		mapID, offset = int64(binary.LittleEndian.Uint32(data[4:8])), 8
	}
	return s.teleportPlayer(active, mapID, math.Float32frombits(binary.LittleEndian.Uint32(data[offset:])), math.Float32frombits(binary.LittleEndian.Uint32(data[offset+4:])), math.Float32frombits(binary.LittleEndian.Uint32(data[offset+8:])), math.Float32frombits(binary.LittleEndian.Uint32(data[offset+12:])))
}

func (s *WorldServer) areaTrigger(active *realm.Character, data []byte) ([]byte, error) {
	if active == nil || len(data) < 4 || s.DBC == nil || s.WorldData == nil {
		return nil, nil
	}
	trigger, found, err := s.DBC.AreaTriggerByID(int64(binary.LittleEndian.Uint32(data)))
	if err != nil || !found {
		return nil, err
	}
	dx, dy, dz := active.PositionX-trigger.X, active.PositionY-trigger.Y, active.PositionZ-trigger.Z
	radius := trigger.Radius * 2
	if dx*dx+dy*dy+dz*dz > radius*radius {
		return nil, nil
	}
	teleport, found, err := s.WorldData.AreaTriggerTeleport(trigger.ID)
	if err != nil || !found {
		return nil, err
	}
	if found, err := s.DBC.MapExists(teleport.TargetMap); err != nil || !found {
		return nil, err
	}
	return s.teleportPlayer(active, teleport.TargetMap, teleport.TargetPositionX, teleport.TargetPositionY, teleport.TargetPositionZ, teleport.TargetOrientation)
}

func (s *WorldServer) teleportPlayer(active *realm.Character, mapID int64, x, y, z, o float32) ([]byte, error) {
	active.Map, active.PositionX, active.PositionY, active.PositionZ, active.Orientation = mapID, x, y, z, o
	if s.Characters != nil {
		if err := s.Characters.UpdateLocation(active.GUID, active.AccountID, active.RealmID, mapID, x, y, z, o); err != nil {
			return nil, err
		}
	}
	s.updatePlayer(*active)
	return newWorldPacket(*active)
}
