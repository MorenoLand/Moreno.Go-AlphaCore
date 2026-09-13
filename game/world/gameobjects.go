package world

import (
	"encoding/binary"
	"math"
	"time"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const (
	gameObjectViewDistance     float32 = 100
	gameObjectInteractDistance float32 = 6
	gameObjectChairDistance    float32 = 3
	gameObjectFishingDistance  float32 = 21
	gameObjectTypeGeneric      int64   = 5
	gameObjectTypeDoor         int64   = 0
	gameObjectTypeButton       int64   = 1
	gameObjectTypeChair        int64   = 7
	gameObjectTypeFishingNode  int64   = 17
	gameObjectFlagInUse        int64   = 1
	gameObjectFlagNoInteract   int64   = 16
	gameObjectStateReady       int64   = 1
)

type gameObjectState struct {
	state, flags int64
	cooldown     time.Time
}

func (s *WorldServer) gameObjectAt(active realm.Character, guid uint64, distance float32) (worlddb.GameObjectSpawn, worlddb.GameObjectTemplate, bool, error) {
	if s.WorldData == nil || guid == 0 {
		return worlddb.GameObjectSpawn{}, worlddb.GameObjectTemplate{}, false, nil
	}
	spawn, found, err := s.WorldData.GameObjectSpawnByID(int64(uint32(guid)))
	if err != nil || !found || spawn.Map != active.Map {
		return worlddb.GameObjectSpawn{}, worlddb.GameObjectTemplate{}, false, err
	}
	dx, dy, dz := spawn.PositionX-active.PositionX, spawn.PositionY-active.PositionY, spawn.PositionZ-active.PositionZ
	if dx*dx+dy*dy+dz*dz > distance*distance {
		return worlddb.GameObjectSpawn{}, worlddb.GameObjectTemplate{}, false, nil
	}
	template, found, err := s.WorldData.GameObjectTemplate(spawn.Entry)
	if err != nil || !found {
		return worlddb.GameObjectSpawn{}, worlddb.GameObjectTemplate{}, false, err
	}
	return spawn, template, true, nil
}

func (s *WorldServer) gameObjectStateFor(guid uint64, spawn worlddb.GameObjectSpawn) gameObjectState {
	s.gameObjectMu.Lock()
	defer s.gameObjectMu.Unlock()
	if s.gameObjects == nil {
		s.gameObjects = make(map[uint64]gameObjectState)
	}
	state, found := s.gameObjects[guid]
	if !found {
		state = gameObjectState{state: spawn.State, flags: spawn.Flags}
		s.gameObjects[guid] = state
	}
	return state
}

func (s *WorldServer) setGameObjectState(guid uint64, state gameObjectState) {
	s.gameObjectMu.Lock()
	if s.gameObjects == nil {
		s.gameObjects = make(map[uint64]gameObjectState)
	}
	s.gameObjects[guid] = state
	s.gameObjectMu.Unlock()
}

func gameObjectUseDistance(template worlddb.GameObjectTemplate) float32 {
	if template.Type == gameObjectTypeChair {
		return gameObjectChairDistance
	}
	if template.Type == gameObjectTypeFishingNode {
		return gameObjectFishingDistance
	}
	return gameObjectInteractDistance
}

func gameObjectUpdate(guid uint64, field int, value uint32) ([]byte, error) {
	return packet.EncodeFieldUpdate(guid, field, value)
}

func (s *WorldServer) gameObjectUse(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 || s.WorldData == nil {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	spawn, template, found, err := s.gameObjectAt(active, guid, gameObjectViewDistance)
	if err != nil || !found || template.Type == gameObjectTypeGeneric {
		return nil, err
	}
	state := s.gameObjectStateFor(guid, spawn)
	if (state.flags|template.Flags)&gameObjectFlagNoInteract != 0 {
		return nil, nil
	}
	dx, dy, dz := spawn.PositionX-active.PositionX, spawn.PositionY-active.PositionY, spawn.PositionZ-active.PositionZ
	distance := gameObjectUseDistance(template)
	if dx*dx+dy*dy+dz*dz > distance*distance {
		return nil, nil
	}
	if template.Type == gameObjectTypeDoor || template.Type == gameObjectTypeButton {
		now := time.Now()
		if now.Before(state.cooldown) {
			return nil, nil
		}
		if !state.cooldown.IsZero() {
			state.cooldown = time.Time{}
			state.flags &^= gameObjectFlagInUse
			state.state = gameObjectStateReady
		}
		state.flags |= gameObjectFlagInUse
		if state.state == gameObjectStateReady {
			state.state = 0
		} else {
			state.state = gameObjectStateReady
		}
		if len(template.Data) > 2 && template.Data[2] > 0 {
			state.cooldown = now.Add(time.Duration(template.Data[2]) * time.Second / 65536)
		}
		s.setGameObjectState(guid, state)
		flags, err := gameObjectUpdate(guid, 7, uint32(state.flags|template.Flags))
		if err != nil {
			return nil, err
		}
		stateUpdate, err := gameObjectUpdate(guid, 12, uint32(state.state))
		if err != nil {
			return nil, err
		}
		s.broadcastPlayer(active, flags)
		s.broadcastPlayer(active, stateUpdate)
		return [][]byte{flags, stateUpdate}, nil
	}
	dynamic, err := gameObjectUpdate(guid, 18, 0)
	if err != nil {
		return nil, err
	}
	s.broadcastPlayer(active, dynamic)
	return [][]byte{dynamic}, nil
}

func (s *WorldServer) nearbyGameObjectPackets(player realm.Character) ([][]byte, error) {
	spawns, err := s.WorldData.GameObjectSpawns(player.Map, player.PositionX, player.PositionY, player.PositionZ, creatureViewDistance)
	if err != nil {
		return nil, err
	}
	packets := make([][]byte, 0, len(spawns))
	for _, spawn := range spawns {
		template, found, err := s.WorldData.GameObjectTemplate(spawn.Entry)
		if err != nil {
			return nil, err
		}
		if !found || template.DisplayID == 0 {
			continue
		}
		guid := uint64(spawn.SpawnID) | 0xf110000000000000
		fields := make([]uint32, 20)
		packet.SetUint64(fields, 0, guid)
		fields[2] = 33
		fields[3] = uint32(template.Entry)
		fields[4] = math.Float32bits(template.Scale)
		fields[6] = uint32(template.DisplayID)
		fields[7] = uint32(template.Flags | spawn.Flags)
		fields[8] = math.Float32bits(spawn.Rotation0)
		fields[9] = math.Float32bits(spawn.Rotation1)
		rotation2, rotation3 := spawn.Rotation2, spawn.Rotation3
		if rotation2 == 0 && rotation3 == 0 {
			rotation2 = float32(math.Sin(float64(spawn.Orientation) / 2))
			rotation3 = float32(math.Cos(float64(spawn.Orientation) / 2))
		}
		fields[10] = math.Float32bits(rotation2)
		fields[11] = math.Float32bits(rotation3)
		fields[12] = uint32(spawn.State)
		fields[13] = uint32(spawn.AnimProgress)
		fields[14] = math.Float32bits(spawn.PositionX)
		fields[15] = math.Float32bits(spawn.PositionY)
		fields[16] = math.Float32bits(spawn.PositionZ)
		fields[17] = math.Float32bits(spawn.Orientation)
		fields[19] = uint32(template.Faction)
		createPacket, err := packet.EncodeGameObjectCreate(guid, fields, packet.Movement{X: spawn.PositionX, Y: spawn.PositionY, Z: spawn.PositionZ, O: spawn.Orientation})
		if err != nil {
			return nil, err
		}
		packets = append(packets, createPacket)
	}
	return packets, nil
}
