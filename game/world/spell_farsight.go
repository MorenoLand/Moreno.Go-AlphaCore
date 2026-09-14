package world

import (
	"time"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const farsightDynamicType int64 = 2

type dynamicObject struct {
	guid, caster uint64
	spell        int64
	mapID        int64
	x, y, z, o   float32
	radius       float32
	timer        *time.Timer
}

func (s *WorldServer) farsightSpell(cast *spellCast, effect dbc.SpellEffect) {
	if cast == nil || cast.caster.GUID == 0 || cast.caster.Health <= 0 {
		return
	}
	x, y, z, o := cast.caster.PositionX, cast.caster.PositionY, cast.caster.PositionZ, cast.caster.Orientation
	if cast.target.Dest != nil {
		x, y, z = cast.target.Dest.X, cast.target.Dest.Y, cast.target.Dest.Z
	} else if target, found := s.playerByGUID(int64(cast.target.UnitGUID)); found {
		x, y, z, o = target.PositionX, target.PositionY, target.PositionZ, target.Orientation
	} else if target := cast.targetCreature; target != nil {
		x, y, z, o = target.Spawn.PositionX, target.Spawn.PositionY, target.Spawn.PositionZ, target.Spawn.Orientation
	}
	radius := float32(0)
	if s.DBC != nil {
		if value, found, err := s.DBC.SpellRadius(effect.RadiusIndex); err == nil && found {
			radius = value.Radius + value.RadiusPerLevel*float32(maxSpellLevel(cast.caster.Level, cast.spell.BaseLevel))
			if value.RadiusMax > 0 && radius > value.RadiusMax {
				radius = value.RadiusMax
			}
		}
	}
	s.dynamicMu.Lock()
	if s.dynamicObjects == nil {
		s.dynamicObjects = make(map[uint64]dynamicObject)
	}
	s.nextDynamicObject++
	guid := uint64(0xf100000000000000) | s.nextDynamicObject
	object := dynamicObject{guid: guid, caster: uint64(cast.caster.GUID), spell: cast.spell.ID, mapID: cast.caster.Map, x: x, y: y, z: z, o: o, radius: radius}
	s.dynamicObjects[guid] = object
	duration := s.spellDurationMillis(cast)
	if duration > 0 {
		object.timer = time.AfterFunc(time.Duration(duration)*time.Millisecond, func() { s.destroyDynamicObject(guid) })
		s.dynamicObjects[guid] = object
	}
	s.dynamicMu.Unlock()
	fields := packet.DynamicObjectFields(guid, uint64(cast.caster.GUID), cast.spell.ID, farsightDynamicType, radius, x, y, z, o)
	if create, err := packet.EncodeDynamicObjectCreate(guid, fields, packet.Movement{X: x, Y: y, Z: z, O: o}); err == nil {
		s.sendSpell(cast.caster, create)
	}
	s.setFarSight(cast.caster, guid)
}

func (s *WorldServer) setFarSight(player realm.Character, guid uint64) {
	s.players.mu.Lock()
	if s.players.farSight == nil {
		s.players.farSight = make(map[int64]uint64)
	}
	s.players.farSight[player.GUID] = guid
	s.players.mu.Unlock()
	for field, value := range map[int]uint32{318: uint32(guid), 319: uint32(guid >> 32)} {
		if update, err := packet.EncodeFieldUpdate(uint64(player.GUID), field, value); err == nil {
			s.sendPlayer(player.GUID, update)
		}
	}
}

func (s *WorldServer) farSight(guid int64) uint64 {
	s.players.mu.RLock()
	value := s.players.farSight[guid]
	s.players.mu.RUnlock()
	return value
}

func (s *WorldServer) destroyDynamicObject(guid uint64) {
	s.dynamicMu.Lock()
	object, found := s.dynamicObjects[guid]
	if found {
		delete(s.dynamicObjects, guid)
		if object.timer != nil {
			object.timer.Stop()
		}
	}
	s.dynamicMu.Unlock()
	if !found {
		return
	}
	if s.farSight(int64(object.caster)) == guid {
		player, found := s.playerByGUID(int64(object.caster))
		if !found {
			player = realm.Character{GUID: int64(object.caster)}
		}
		s.setFarSight(player, 0)
	}
	if destroy, err := packet.Encode(packet.SMSGDestroyObject, encodeGUID(int64(guid))); err == nil {
		player, found := s.playerByGUID(int64(object.caster))
		if !found {
			player = realm.Character{GUID: int64(object.caster), Map: object.mapID}
		}
		s.sendSpell(player, destroy)
	}
}
