package world

import (
	"encoding/binary"
	"math"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) binderActivate(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 || s.Characters == nil || s.WorldData == nil {
		return nil, nil
	}
	binderGUID := binary.LittleEndian.Uint64(data)
	spawn, _, found, err := s.binderAt(active, binderGUID)
	if err != nil || !found {
		return nil, err
	}
	bind, found, err := s.Characters.Deathbind(active.GUID)
	if err != nil {
		return nil, err
	}
	if found && bind.CreatureBinderGUID == int64(uint32(binderGUID)) {
		response, err := packet.Encode(packet.SMSGPlayerBindError, nil)
		if err != nil {
			return nil, err
		}
		return [][]byte{response}, nil
	}
	bind = realm.Deathbind{PlayerGUID: active.GUID, CreatureBinderGUID: int64(uint32(binderGUID)), Map: spawn.Map, Zone: active.Zone, X: spawn.PositionX, Y: spawn.PositionY, Z: spawn.PositionZ}
	if err := s.Characters.SaveDeathbind(bind); err != nil {
		return nil, err
	}
	point, err := deathbindPointPacket(bind)
	if err != nil {
		return nil, err
	}
	bound, err := packet.Encode(packet.SMSGPlayerBound, encodeGUID(int64(binderGUID)))
	if err != nil {
		return nil, err
	}
	return [][]byte{point, bound}, nil
}

func (s *WorldServer) setDeathBindPoint(active realm.Character) ([]byte, error) {
	if s.Characters == nil {
		return nil, nil
	}
	bind, found, err := s.Characters.Deathbind(active.GUID)
	if err != nil || !found {
		return nil, err
	}
	return deathbindPointPacket(bind)
}

func (s *WorldServer) deathBindZone(active realm.Character) ([]byte, error) {
	if s.Characters == nil {
		return nil, nil
	}
	bind, found, err := s.Characters.Deathbind(active.GUID)
	if err != nil || !found {
		return nil, err
	}
	body := append(encodeUint32(bind.Map), encodeUint32(bind.Zone)...)
	return packet.Encode(packet.SMSGBindZoneReply, body)
}

func (s *WorldServer) repop(active *realm.Character) ([]byte, error) {
	if active == nil || active.Health > 0 || s.Characters == nil {
		return nil, nil
	}
	bind, found, err := s.Characters.Deathbind(active.GUID)
	if err != nil || !found {
		return nil, err
	}
	return s.teleportPlayer(active, bind.Map, bind.X, bind.Y, bind.Z, active.Orientation)
}

func (s *WorldServer) binderAt(active realm.Character, guid uint64) (worlddb.CreatureSpawn, worlddb.CreatureTemplate, bool, error) {
	spawn, creature, found, err := s.creatureAt(active, guid, 10)
	if err != nil || !found || creature.NPCFlags&0x10 == 0 {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, err
	}
	return spawn, creature, true, nil
}

func (s *WorldServer) creatureAt(active realm.Character, guid uint64, distance float32) (worlddb.CreatureSpawn, worlddb.CreatureTemplate, bool, error) {
	spawn, found, err := s.WorldData.CreatureSpawnByID(int64(uint32(guid)))
	if err != nil || !found || spawn.Map != active.Map {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, err
	}
	dx, dy, dz := spawn.PositionX-active.PositionX, spawn.PositionY-active.PositionY, spawn.PositionZ-active.PositionZ
	if dx*dx+dy*dy+dz*dz > distance*distance {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, nil
	}
	creature, found, err := s.WorldData.CreatureTemplate(spawn.Entry)
	if err != nil || !found {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, err
	}
	return spawn, creature, true, nil
}

func deathbindPointPacket(bind realm.Deathbind) ([]byte, error) {
	body := make([]byte, 16)
	binary.LittleEndian.PutUint32(body, math.Float32bits(bind.X))
	binary.LittleEndian.PutUint32(body[4:], math.Float32bits(bind.Y))
	binary.LittleEndian.PutUint32(body[8:], math.Float32bits(bind.Z))
	binary.LittleEndian.PutUint32(body[12:], uint32(bind.Map))
	return packet.Encode(packet.SMSGBindPointUpdate, body)
}
