package world

import (
	"time"

	"Moreno.AlphaCore/database/dbc"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) sendEventSpell(cast *spellCast, effect dbc.SpellEffect) {
	if cast == nil || s.WorldData == nil {
		return
	}
	scripts, err := s.WorldData.EventScripts(effect.MiscValue)
	if err != nil {
		return
	}
	for _, script := range scripts {
		if script.Delay <= 0 {
			s.executeEventScript(cast, script)
		} else {
			time.AfterFunc(time.Duration(script.Delay)*time.Second, func() { s.executeEventScript(cast, script) })
		}
	}
}

func (s *WorldServer) executeEventScript(cast *spellCast, script worlddb.EventScript) {
	switch script.Command {
	case 10:
		s.summonEventCreature(cast, script)
	case 15:
		s.triggerSpell(cast.caster, script.DataLong[0], cast.target, cast.targetMask)
	case 17:
		s.createEventItem(cast, script)
	case 48:
		s.damageEventTarget(cast, script.DataLong[0], script.DataLong[1] != 0)
	}
}

func (s *WorldServer) summonEventCreature(cast *spellCast, script worlddb.EventScript) {
	if cast == nil || s.WorldData == nil || script.DataLong[0] <= 0 {
		return
	}
	template, found, err := s.WorldData.CreatureTemplate(script.DataLong[0])
	if err != nil || !found {
		return
	}
	level := int64(cast.caster.Level)
	if level < 1 {
		level = template.LevelMin
	}
	stats, found, err := s.WorldData.CreatureClassLevelStats(template.UnitClass, level)
	if err != nil || !found {
		return
	}
	spawn := worlddb.CreatureSpawn{Entry: template.Entry, Map: cast.caster.Map, PositionX: script.X, PositionY: script.Y, PositionZ: script.Z, Orientation: script.O, HealthPercent: 100, ManaPercent: 100}
	if spawn.PositionX == 0 && spawn.PositionY == 0 && spawn.PositionZ == 0 {
		spawn.PositionX, spawn.PositionY, spawn.PositionZ, spawn.Orientation = cast.caster.PositionX, cast.caster.PositionY, cast.caster.PositionZ, cast.caster.Orientation
	}
	state := s.newCreatureState(spawn, template, stats, level)
	state.GUID, state.OwnerGUID, state.CreatedBySpell = s.nextCreatureGUID(), uint64(cast.caster.GUID), cast.spell.ID
	s.setCreatureState(state)
	if script.DataLong[1] > 0 {
		state.Timer = time.AfterFunc(time.Duration(script.DataLong[1])*time.Millisecond, func() { s.despawnCreature(state.GUID) })
		s.creatures.mu.Lock()
		if current := s.creatures.active[state.GUID]; current != nil {
			current.Timer = state.Timer
		}
		s.creatures.mu.Unlock()
	}
	if create, err := s.creatureCreatePacket(state, template.Scale, 0, 0); err == nil {
		s.sendSpell(cast.caster, create)
	}
}

func (s *WorldServer) createEventItem(cast *spellCast, script worlddb.EventScript) {
	if cast == nil || s.Characters == nil || s.WorldData == nil || script.DataLong[0] <= 0 {
		return
	}
	amount := script.DataLong[1]
	if amount <= 0 {
		amount = 1
	}
	template, found, err := s.WorldData.ItemTemplate(script.DataLong[0])
	if err != nil || !found {
		return
	}
	slot, err := s.Characters.FirstEmptySlot(cast.caster.GUID, 23, 0, 39)
	if err != nil || slot < 0 {
		return
	}
	item, err := s.Characters.CreateInventoryItem(cast.caster.GUID, cast.caster.GUID, 23, slot, template.Entry, amount)
	if err != nil {
		return
	}
	if create, err := packet.EncodeItemCreate(uint64(item.GUID)|0x4000000000000000, uint32(item.ItemTemplate), uint64(item.Owner), uint64(item.Creator), uint32(item.StackCount), 0, encodedItemFlags(template, item.Flags), item.SpellCharges, packet.Movement{X: cast.caster.PositionX, Y: cast.caster.PositionY, Z: cast.caster.PositionZ, O: cast.caster.Orientation}); err == nil {
		s.sendPlayer(cast.caster.GUID, create)
	}
	if push, err := itemPushResult(item, template.Entry, 23); err == nil {
		s.sendPlayer(cast.caster.GUID, push)
	}
}

func (s *WorldServer) damageEventTarget(cast *spellCast, amount int64, percent bool) {
	if cast == nil || amount <= 0 {
		return
	}
	if percent {
		if target, found := s.playerByGUID(int64(cast.target.UnitGUID)); found {
			amount = target.Health * amount / 100
		} else if cast.targetCreature != nil {
			amount = cast.targetCreature.Health * amount / 100
		}
	}
	if target, found := s.playerByGUID(int64(cast.target.UnitGUID)); found {
		_ = s.changePlayerHealth(&target, -amount)
	} else if cast.targetCreature != nil {
		s.changeCreatureHealth(cast.targetCreature, -amount)
	}
}
