package world

import (
	"sync"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

type creatureRegistry struct {
	mu     sync.Mutex
	active map[uint64]*creatureState
}

type creatureState struct {
	GUID                           uint64
	Spawn                          worlddb.CreatureSpawn
	Template                       worlddb.CreatureTemplate
	Level, Health, MaxHealth, Mana int64
}

func (s *WorldServer) creatureStateAt(active realm.Character, guid uint64, distance float32) (*creatureState, bool, error) {
	if s.WorldData == nil {
		return nil, false, nil
	}
	spawn, template, found, err := s.creatureAt(active, guid, distance)
	if err != nil || !found {
		return nil, false, err
	}
	s.creatures.mu.Lock()
	if s.creatures.active == nil {
		s.creatures.active = make(map[uint64]*creatureState)
	}
	if state := s.creatures.active[guid]; state != nil {
		s.creatures.mu.Unlock()
		return state, true, nil
	}
	level := template.LevelMin
	stats, found, err := s.WorldData.CreatureClassLevelStats(template.UnitClass, level)
	if err != nil || !found {
		s.creatures.mu.Unlock()
		return nil, false, err
	}
	health := int64(float64(stats.Health) * float64(template.HealthMultiplier) * float64(spawn.HealthPercent) / 100)
	if health < 1 {
		health = 1
	}
	mana := int64(float64(stats.Mana) * float64(template.ManaMultiplier) * float64(spawn.ManaPercent) / 100)
	state := &creatureState{GUID: guid, Spawn: spawn, Template: template, Level: level, Health: health, MaxHealth: health, Mana: mana}
	s.creatures.active[guid] = state
	s.creatures.mu.Unlock()
	return state, true, nil
}

func (s *WorldServer) setCreatureState(state creatureState) {
	s.creatures.mu.Lock()
	if s.creatures.active == nil {
		s.creatures.active = make(map[uint64]*creatureState)
	}
	if current := s.creatures.active[state.GUID]; current == nil || current.Health > current.MaxHealth {
		s.creatures.active[state.GUID] = &state
	}
	s.creatures.mu.Unlock()
}

func (s *WorldServer) creatureHealth(guid uint64) (int64, bool) {
	s.creatures.mu.Lock()
	state, found := s.creatures.active[guid]
	var health int64
	if found {
		health = state.Health
	}
	s.creatures.mu.Unlock()
	return health, found
}

func (s *WorldServer) changeCreatureHealth(state *creatureState, delta int64) {
	s.creatures.mu.Lock()
	current := s.creatures.active[state.GUID]
	if current == nil {
		s.creatures.mu.Unlock()
		return
	}
	current.Health += delta
	if current.Health < 0 {
		current.Health = 0
	}
	value := current.Health
	viewer := realm.Character{GUID: int64(current.GUID), Map: current.Spawn.Map, PositionX: current.Spawn.PositionX, PositionY: current.Spawn.PositionY, PositionZ: current.Spawn.PositionZ}
	s.creatures.mu.Unlock()
	for _, field := range []int{22, 27} {
		if update, err := packet.EncodeFieldUpdate(state.GUID, field, uint32(value)); err == nil {
			s.broadcastPlayer(viewer, update)
		}
	}
}

func (s *WorldServer) changeCreaturePower(state *creatureState, powerType, delta int64) {
	if powerType != 0 {
		return
	}
	s.creatures.mu.Lock()
	current := s.creatures.active[state.GUID]
	if current == nil {
		s.creatures.mu.Unlock()
		return
	}
	current.Mana += delta
	if current.Mana < 0 {
		current.Mana = 0
	}
	value := current.Mana
	viewer := realm.Character{GUID: int64(current.GUID), Map: current.Spawn.Map, PositionX: current.Spawn.PositionX, PositionY: current.Spawn.PositionY, PositionZ: current.Spawn.PositionZ}
	s.creatures.mu.Unlock()
	for _, field := range []int{23, 28} {
		if update, err := packet.EncodeFieldUpdate(state.GUID, field, uint32(value)); err == nil {
			s.broadcastPlayer(viewer, update)
		}
	}
}
