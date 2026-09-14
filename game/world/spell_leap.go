package world

import (
	"math"

	"Moreno.AlphaCore/database/realm"
)

const (
	leapCombatReach          float32 = 1.5
	leapMeleeRangeOffset     float32 = 1.3333334
	leapDistanceExpandFactor float32 = 1.05
	leapDistanceNormalize    float32 = 0.9
	leapCombatDistanceFactor float32 = 0.6
)

func (s *WorldServer) leapSpell(cast *spellCast) {
	if cast == nil || cast.caster.GUID == 0 || cast.caster.Health <= 0 {
		return
	}
	caster := cast.caster
	if cast.target.Dest != nil {
		s.teleportLeaper(&caster, cast.target.Dest.X, cast.target.Dest.Y, cast.target.Dest.Z, caster.Orientation)
		return
	}
	if cast.target.UnitGUID == 0 || cast.target.UnitGUID == uint64(caster.GUID) {
		return
	}
	var x, y, z float32
	targetReach := leapCombatReach
	if target, found := s.playerByGUID(int64(cast.target.UnitGUID)); found {
		if target.Map != caster.Map {
			return
		}
		x, y, z = target.PositionX, target.PositionY, target.PositionZ
	} else if target := cast.targetCreature; target != nil {
		if target.Spawn.Map != caster.Map {
			return
		}
		x, y, z = target.Spawn.PositionX, target.Spawn.PositionY, target.Spawn.PositionZ
		if s.WorldData != nil {
			if model, found, err := s.WorldData.CreatureModelInfo(target.Template.DisplayID1); err == nil && found && model.CombatReach > 0 {
				targetReach = model.CombatReach
			}
		}
	} else {
		return
	}
	dx, dy, dz := x-caster.PositionX, y-caster.PositionY, z-caster.PositionZ
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
	if distance <= 0 {
		return
	}
	combatDistance := (leapCombatReach + targetReach + leapMeleeRangeOffset) * leapDistanceExpandFactor * leapDistanceNormalize * leapCombatDistanceFactor
	if distance < combatDistance {
		combatDistance = distance
	}
	factor := (distance - combatDistance) / distance
	s.teleportLeaper(&caster, caster.PositionX+dx*factor, caster.PositionY+dy*factor, caster.PositionZ+dz*factor, float32(math.Atan2(float64(dy), float64(dx))))
}

func (s *WorldServer) teleportLeaper(caster *realm.Character, x, y, z, orientation float32) {
	response, err := s.teleportPlayer(caster, caster.Map, x, y, z, orientation)
	if err == nil {
		s.sendPlayer(caster.GUID, response)
	}
}
