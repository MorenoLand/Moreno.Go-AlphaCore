package world

import (
	"time"

	"Moreno.AlphaCore/database/dbc"
)

const creatureThreatNotToLeaveCombat float64 = 1e-4

func (s *WorldServer) addExtraAttacks(guid int64, amount int64, alive bool) {
	if guid == 0 || amount <= 0 || !alive {
		return
	}
	s.players.mu.Lock()
	if s.players.extraAttacks == nil {
		s.players.extraAttacks = make(map[int64]int64)
	}
	s.players.extraAttacks[guid] += amount
	s.players.mu.Unlock()
}

func (s *WorldServer) extraAttacks(guid int64) int64 {
	s.players.mu.RLock()
	amount := s.players.extraAttacks[guid]
	s.players.mu.RUnlock()
	return amount
}

func (s *WorldServer) addCreatureExtraAttacks(target *creatureState, amount int64) {
	if target == nil || target.Health <= 0 || amount <= 0 {
		return
	}
	s.creatures.mu.Lock()
	state := target
	if current := s.creatures.active[target.GUID]; current != nil {
		state = current
	}
	state.ExtraAttacks += amount
	s.creatures.mu.Unlock()
}

func spellEffectSimplePoints(effect dbc.SpellEffect) int64 {
	return effect.BasePoints + effect.BaseDice
}

func (s *WorldServer) addCreatureThreat(cast *spellCast, target *creatureState, amount float64, pull bool) {
	if cast == nil || target == nil || cast.caster.GUID == 0 || cast.caster.Health <= 0 || target.Health <= 0 {
		return
	}
	if amount == 0 {
		amount = creatureThreatNotToLeaveCombat
	}
	s.creatures.mu.Lock()
	state := target
	if current := s.creatures.active[target.GUID]; current != nil {
		state = current
	}
	if state.Threat == nil {
		state.Threat = make(map[uint64]float64)
	}
	state.Threat[uint64(cast.caster.GUID)] += amount
	if pull {
		state.PullUntil = time.Now().Add(3 * time.Second)
	}
	state.CombatTarget = maxCreatureThreatTarget(state.Threat)
	combatTarget := state.CombatTarget
	s.creatures.mu.Unlock()
	if combatTarget == uint64(cast.caster.GUID) {
		s.setCombatTarget(cast.caster.GUID, target.GUID)
	}
}

func maxCreatureThreatTarget(threat map[uint64]float64) uint64 {
	var target uint64
	var value float64
	for guid, amount := range threat {
		if amount > value || target == 0 {
			target, value = guid, amount
		}
	}
	return target
}
