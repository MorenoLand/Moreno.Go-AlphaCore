package world

import (
	"math"
	"time"
)

func (s *WorldServer) distractSpell(cast *spellCast, target *creatureState) {
	if cast == nil || target == nil || target.Health <= 0 {
		return
	}
	duration := s.spellDurationMillis(cast)
	if duration <= 0 {
		return
	}
	now := time.Now()
	s.creatures.mu.Lock()
	state := target
	if current := s.creatures.active[target.GUID]; current != nil {
		state = current
	}
	if state.CombatTarget != 0 || state.DistractedUntil.After(now) {
		s.creatures.mu.Unlock()
		return
	}
	x, y := state.Spawn.PositionX, state.Spawn.PositionY
	if cast.target.Dest != nil {
		x, y = cast.target.Dest.X, cast.target.Dest.Y
	}
	state.DistractedAngle = float32(math.Atan2(float64(y-state.Spawn.PositionY), float64(x-state.Spawn.PositionX)))
	state.DistractedUntil = now.Add(time.Duration(duration) * time.Millisecond)
	state.Spawn.Orientation = state.DistractedAngle
	s.creatures.mu.Unlock()
}
