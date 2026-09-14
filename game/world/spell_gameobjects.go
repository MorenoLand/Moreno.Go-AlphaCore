package world

import (
	"math"
	"time"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) summonGameObject(cast *spellCast, effect dbc.SpellEffect, wild bool) {
	if cast == nil || s.WorldData == nil {
		return
	}
	template, found, err := s.WorldData.GameObjectTemplate(effect.MiscValue)
	if err != nil || !found {
		return
	}
	if !wild && s.DBC != nil {
		if race, found, err := s.DBC.Race(cast.caster.Race); err == nil && found {
			template.Faction = race.FactionID
		}
	}
	spawn := worlddb.GameObjectSpawn{Entry: template.Entry, Map: cast.caster.Map, PositionX: cast.caster.PositionX, PositionY: cast.caster.PositionY, PositionZ: cast.caster.PositionZ, Orientation: cast.caster.Orientation, State: gameObjectStateReady, Flags: template.Flags}
	if cast.target.Dest != nil {
		spawn.PositionX, spawn.PositionY, spawn.PositionZ = cast.target.Dest.X, cast.target.Dest.Y, cast.target.Dest.Z
	} else if cast.target.Source != nil {
		spawn.PositionX, spawn.PositionY, spawn.PositionZ = cast.target.Source.X, cast.target.Source.Y, cast.target.Source.Z
	} else if target, found := s.playerByGUID(int64(cast.target.UnitGUID)); found {
		spawn.PositionX, spawn.PositionY, spawn.PositionZ, spawn.Orientation = target.PositionX, target.PositionY, target.PositionZ, target.Orientation
	}
	guid := s.nextGameObjectGUID()
	instance := dynamicGameObject{spawn: spawn, template: template}
	s.gameObjectMu.Lock()
	if s.dynamicGameObjects == nil {
		s.dynamicGameObjects = make(map[uint64]dynamicGameObject)
	}
	s.dynamicGameObjects[guid] = instance
	s.gameObjectMu.Unlock()
	duration := s.spellDurationMillis(cast)
	if duration == 0 {
		duration = 120000
	}
	if duration > 0 {
		timer := time.AfterFunc(time.Duration(duration)*time.Millisecond, func() { s.destroyGameObject(guid) })
		s.gameObjectMu.Lock()
		instance.timer = timer
		if current, ok := s.dynamicGameObjects[guid]; ok {
			current.timer = timer
			s.dynamicGameObjects[guid] = current
		}
		s.gameObjectMu.Unlock()
	}
	if create, err := dynamicGameObjectPacket(guid, spawn, template); err == nil {
		s.sendSpell(cast.caster, create)
	}
}

func dynamicGameObjectPacket(guid uint64, spawn worlddb.GameObjectSpawn, template worlddb.GameObjectTemplate) ([]byte, error) {
	scale := template.Scale
	if scale <= 0 {
		scale = 1
	}
	fields := make([]uint32, 20)
	packet.SetUint64(fields, 0, guid)
	fields[2] = 33
	fields[3] = uint32(template.Entry)
	fields[4] = math.Float32bits(scale)
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
	return packet.EncodeGameObjectCreate(guid, fields, packet.Movement{X: spawn.PositionX, Y: spawn.PositionY, Z: spawn.PositionZ, O: spawn.Orientation})
}

func (s *WorldServer) destroyGameObject(guid uint64) {
	s.gameObjectMu.Lock()
	instance, found := s.dynamicGameObjects[guid]
	if found {
		delete(s.dynamicGameObjects, guid)
		if instance.timer != nil {
			instance.timer.Stop()
		}
	}
	s.gameObjectMu.Unlock()
	if !found {
		return
	}
	viewer := realm.Character{GUID: int64(guid), Map: instance.spawn.Map, PositionX: instance.spawn.PositionX, PositionY: instance.spawn.PositionY, PositionZ: instance.spawn.PositionZ}
	if destroy, err := packet.Encode(packet.SMSGDestroyObject, encodeGUID(int64(guid))); err == nil {
		s.broadcastPlayer(viewer, destroy)
	}
}

func (s *WorldServer) activateGameObject(cast *spellCast) {
	if cast == nil || cast.target.GameObjectGUID == 0 {
		return
	}
	data := append(encodeGUID(int64(cast.target.GameObjectGUID)), encodeUint32(0)...)
	if animation, err := packet.Encode(packet.SMSGGameObjectCustomAnim, data); err == nil {
		s.broadcastPlayer(cast.caster, animation)
	}
}
