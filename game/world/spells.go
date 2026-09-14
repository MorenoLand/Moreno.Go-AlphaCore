package world

import (
	"encoding/binary"
	"math"
	"math/rand"
	"sync"
	"time"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

type spellRegistry struct {
	mu        sync.Mutex
	casts     map[int64]*spellCast
	cooldowns map[int64]map[int64]time.Time
}

type spellVector struct{ X, Y, Z float32 }

type spellTarget struct {
	UnitGUID uint64
	ItemGUID uint64
	Source   *spellVector
	Dest     *spellVector
}

type spellCast struct {
	caster     realm.Character
	target     spellTarget
	spell      dbc.Spell
	targetMask packet.SpellTargetMask
	castTime   int64
	powerCost  int64
	sourceItem *realm.InventoryItem
	source     worlddb.ItemTemplate
	sourceSlot int
	started    time.Time
	timer      *time.Timer
}

func (s *WorldServer) castSpellPacket(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 6 {
		return nil, nil
	}
	spellID := int64(binary.LittleEndian.Uint32(data))
	targetMask := packet.SpellTargetMask(binary.LittleEndian.Uint16(data[4:]))
	target, ok := s.spellTarget(active, targetMask, data[6:])
	if !ok {
		return s.castFailure(active, spellID, packet.SpellFailedBadTargets)
	}
	return s.startSpellCast(active, spellID, target, targetMask)
}

func (s *WorldServer) useItemPacket(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 5 || s.Characters == nil || s.WorldData == nil {
		return nil, nil
	}
	item, found, err := s.Characters.ItemAt(active.GUID, inventoryBag(data[0]), int64(data[1]))
	if err != nil || !found {
		return nil, err
	}
	template, found, err := s.WorldData.ItemTemplate(item.ItemTemplate)
	if err != nil || !found {
		return nil, err
	}
	slot := int(data[2])
	if slot >= len(template.Spells) {
		return nil, nil
	}
	itemSpell := template.Spells[slot]
	if itemSpell.ID <= 0 || itemSpell.Trigger != 0 {
		return nil, nil
	}
	targetMask := packet.SpellTargetMask(binary.LittleEndian.Uint16(data[3:]))
	target, ok := s.spellTarget(active, targetMask, data[5:])
	if !ok {
		return nil, nil
	}
	return s.startSpellCastWithItem(active, itemSpell.ID, target, targetMask, &item, slot, template)
}

func (s *WorldServer) spellTarget(active realm.Character, mask packet.SpellTargetMask, data []byte) (spellTarget, bool) {
	target := spellTarget{}
	offset := 0
	if mask&(packet.SpellTargetUnit|packet.SpellTargetGameObject) != 0 {
		if len(data) < offset+8 {
			return spellTarget{}, false
		}
		target.UnitGUID = binary.LittleEndian.Uint64(data[offset:])
		offset += 8
	}
	if mask&packet.SpellTargetItemMask != 0 {
		if len(data) < offset+8 {
			return spellTarget{}, false
		}
		target.ItemGUID = binary.LittleEndian.Uint64(data[offset:])
		offset += 8
	}
	if mask&packet.SpellTargetSource != 0 {
		if len(data) < offset+12 {
			return spellTarget{}, false
		}
		target.Source = &spellVector{math.Float32frombits(binary.LittleEndian.Uint32(data[offset:])), math.Float32frombits(binary.LittleEndian.Uint32(data[offset+4:])), math.Float32frombits(binary.LittleEndian.Uint32(data[offset+8:]))}
		offset += 12
	}
	if mask&packet.SpellTargetDestination != 0 {
		if len(data) < offset+12 {
			return spellTarget{}, false
		}
		target.Dest = &spellVector{math.Float32frombits(binary.LittleEndian.Uint32(data[offset:])), math.Float32frombits(binary.LittleEndian.Uint32(data[offset+4:])), math.Float32frombits(binary.LittleEndian.Uint32(data[offset+8:]))}
	}
	if mask&packet.SpellTargetUnit != 0 {
		if target.UnitGUID == 0 {
			return spellTarget{}, false
		}
		player, found := s.playerByGUID(int64(target.UnitGUID))
		if !found || player.Map != active.Map {
			return spellTarget{}, false
		}
		return target, true
	}
	if mask&packet.SpellTargetItemMask != 0 && mask&packet.SpellTargetTradeItem == 0 && target.ItemGUID == 0 {
		return spellTarget{}, false
	}
	if mask&(packet.SpellTargetTerrain|packet.SpellTargetItemMask|packet.SpellTargetGameObject) == 0 {
		target.UnitGUID = uint64(active.GUID)
	}
	return target, true
}

func (s *WorldServer) startSpellCast(active realm.Character, spellID int64, target spellTarget, targetMask packet.SpellTargetMask) ([][]byte, error) {
	return s.startSpellCastWithItem(active, spellID, target, targetMask, nil, -1, worlddb.ItemTemplate{})
}

func (s *WorldServer) startSpellCastWithItem(active realm.Character, spellID int64, target spellTarget, targetMask packet.SpellTargetMask, sourceItem *realm.InventoryItem, sourceSlot int, source worlddb.ItemTemplate) ([][]byte, error) {
	if s.DBC == nil {
		return s.castFailure(active, spellID, packet.SpellFailedUnavailable)
	}
	spell, found, err := s.DBC.Spell(spellID)
	if err != nil || !found {
		return s.castFailure(active, spellID, packet.SpellFailedUnavailable)
	}
	if target.UnitGUID == 0 && targetMask == packet.SpellTargetSelf {
		target.UnitGUID = uint64(active.GUID)
	}
	if target.UnitGUID != 0 && target.UnitGUID != uint64(active.GUID) {
		player, found := s.playerByGUID(int64(target.UnitGUID))
		if !found || player.Map != active.Map {
			return s.castFailure(active, spellID, packet.SpellFailedBadTargets)
		}
	}
	if s.Characters != nil && sourceItem == nil {
		known, err := s.Characters.Spells(active.GUID)
		if err != nil {
			return nil, err
		}
		learned := false
		for _, value := range known {
			if value.ID == spellID && value.Active {
				learned = true
				break
			}
		}
		if !learned {
			return s.castFailure(active, spellID, packet.SpellFailedNotKnown)
		}
	}
	if active.Health <= 0 && packet.SpellAttributes(spell.Attributes)&packet.SpellAttributeAllowDead == 0 {
		return s.castFailure(active, spellID, packet.SpellFailedCasterDead)
	}
	if s.spellOnCooldown(active.GUID, spellID) {
		return s.castFailure(active, spellID, packet.SpellFailedNotReady)
	}
	powerCost := spell.ManaCost + spell.ManaCostPerLevel*int64(active.Level)
	if powerCost > playerPower(active, spell.PowerType) {
		return s.castFailure(active, spellID, packet.SpellFailedNoPower)
	}
	if spellRange, rangeFound, rangeErr := s.DBC.SpellRange(spell.RangeIndex); rangeErr != nil {
		return nil, rangeErr
	} else if rangeFound && target.UnitGUID != 0 && target.UnitGUID != uint64(active.GUID) {
		player, _ := s.playerByGUID(int64(target.UnitGUID))
		dx, dy, dz := active.PositionX-player.PositionX, active.PositionY-player.PositionY, active.PositionZ-player.PositionZ
		if spellRange.RangeMax > 0 && dx*dx+dy*dy+dz*dz > spellRange.RangeMax*spellRange.RangeMax {
			return s.castFailure(active, spellID, packet.SpellFailedOutOfRange)
		}
	}
	castTime := int64(0)
	if value, castFound, castErr := s.DBC.SpellCastTime(spell.CastingTimeIndex); castErr != nil {
		return nil, castErr
	} else if castFound {
		castTime = value.Base + value.PerLevel*int64(active.Level)
		if castTime < value.Minimum {
			castTime = value.Minimum
		}
	}
	cast := &spellCast{caster: active, target: target, spell: spell, targetMask: targetMask, castTime: castTime, powerCost: powerCost, sourceItem: sourceItem, sourceSlot: sourceSlot, source: source, started: time.Now()}
	if castTime <= 0 {
		s.performSpellCast(cast)
		return nil, nil
	}
	s.spells.mu.Lock()
	if s.spells.casts == nil {
		s.spells.casts = make(map[int64]*spellCast)
	}
	if previous := s.spells.casts[active.GUID]; previous != nil {
		if previous.timer != nil {
			previous.timer.Stop()
		}
		delete(s.spells.casts, active.GUID)
	}
	s.spells.casts[active.GUID] = cast
	cast.timer = time.AfterFunc(time.Duration(castTime)*time.Millisecond, func() { s.finishSpellCast(active.GUID, cast) })
	s.spells.mu.Unlock()
	start, err := spellStartPacket(cast)
	if err != nil {
		return nil, err
	}
	s.sendSpell(active, start)
	return nil, nil
}

func (s *WorldServer) finishSpellCast(guid int64, cast *spellCast) {
	s.spells.mu.Lock()
	if s.spells.casts[guid] != cast {
		s.spells.mu.Unlock()
		return
	}
	delete(s.spells.casts, guid)
	s.spells.mu.Unlock()
	s.performSpellCast(cast)
}

func (s *WorldServer) performSpellCast(cast *spellCast) {
	if cast.powerCost > 0 {
		_ = s.changePlayerPower(&cast.caster, cast.spell.PowerType, -cast.powerCost)
	}
	s.setSpellCooldown(cast.caster.GUID, cast.spell)
	result, err := spellCastResult(cast.spell.ID, packet.SpellNoError)
	if err == nil {
		s.sendSpell(cast.caster, result)
	}
	goPacket, err := spellGoPacket(cast)
	if err == nil {
		s.sendSpell(cast.caster, goPacket)
	}
	s.applySpellEffects(cast)
	s.consumeItemSpell(cast)
}

func (s *WorldServer) consumeItemSpell(cast *spellCast) {
	if cast.sourceItem == nil || s.Characters == nil {
		return
	}
	stat := cast.source.Spells[cast.sourceSlot]
	charges := cast.sourceItem.SpellCharges[cast.sourceSlot]
	hadCharges := charges != 0
	if hadCharges && stat.Charges != -1 {
		if charges > 0 {
			charges--
		} else {
			charges++
		}
		if s.Characters.UpdateItemSpellCharges(cast.sourceItem.GUID, cast.sourceItem.Owner, int64(cast.sourceSlot), charges) != nil {
			return
		}
		guid := uint64(cast.sourceItem.GUID) | 0x4000000000000000
		if update, err := packet.EncodeFieldUpdate(guid, 14+cast.sourceSlot, uint32(charges)); err == nil {
			s.sendSpell(cast.caster, update)
		}
	}
	if hadCharges && (charges == 0 || stat.Charges == -1) {
		if err := s.Characters.DeleteItem(cast.sourceItem.GUID, cast.sourceItem.Owner); err == nil {
			if update, err := packet.Encode(packet.SMSGDestroyObject, encodeGUID(int64(uint64(cast.sourceItem.GUID)|0x4000000000000000))); err == nil {
				s.sendSpell(cast.caster, update)
			}
		}
	}
}

func (s *WorldServer) applySpellEffects(cast *spellCast) {
	target, found := s.spellPlayerTarget(cast)
	if !found {
		return
	}
	effectiveLevel := int64(cast.caster.Level) - cast.spell.BaseLevel
	if effectiveLevel < 0 {
		effectiveLevel = 0
	}
	for index, effect := range cast.spell.Effects {
		points := spellEffectPoints(effect, effectiveLevel)
		switch packet.SpellEffect(effect.Type) {
		case packet.SpellEffectSchoolDamage:
			_ = s.changePlayerHealth(&target, -points)
		case packet.SpellEffectPowerBurn:
			amount := minPower(playerPower(target, effect.MiscValue), points)
			if amount > 0 {
				_ = s.changePlayerPower(&target, effect.MiscValue, -amount)
				_ = s.changePlayerHealth(&target, -amount)
			}
		case packet.SpellEffectHeal:
			_ = s.changePlayerHealth(&target, points)
		case packet.SpellEffectHealthLeech:
			if s.changePlayerHealth(&target, -points) == nil {
				caster := cast.caster
				_ = s.changePlayerHealth(&caster, points)
			}
		case packet.SpellEffectEnergize:
			_ = s.changePlayerPower(&target, effect.MiscValue, points)
		case packet.SpellEffectPowerDrain:
			amount := minPower(playerPower(target, effect.MiscValue), points)
			if amount > 0 {
				_ = s.changePlayerPower(&target, effect.MiscValue, -amount)
				caster := cast.caster
				_ = s.changePlayerPower(&caster, effect.MiscValue, amount)
			}
		case packet.SpellEffectApplyAura, packet.SpellEffectApplyAreaAura:
			s.applyAura(cast, target, index, effect)
		}
	}
}

func (s *WorldServer) spellPlayerTarget(cast *spellCast) (realm.Character, bool) {
	if cast.target.UnitGUID == uint64(cast.caster.GUID) {
		return cast.caster, true
	}
	player, found := s.playerByGUID(int64(cast.target.UnitGUID))
	return player, found
}

func spellEffectPoints(effect dbc.SpellEffect, level int64) int64 {
	dice := effect.BaseDice + effect.DicePerLevel*level
	scaled := float64(effect.RealPointsPerLevel) * float64(level)
	minimum := effect.BasePoints + dice + int64(math.Trunc(scaled))
	maximum := effect.BasePoints + effect.DieSides*dice + int64(math.Ceil(scaled))
	if effect.DieSides == 0 || minimum == maximum {
		return minimum
	}
	if minimum > maximum {
		minimum, maximum = maximum, minimum
	}
	return minimum + rand.Int63n(maximum-minimum+1)
}

func minPower(current, amount int64) int64 {
	if amount < 0 {
		return 0
	}
	if amount > current {
		return current
	}
	return amount
}

func (s *WorldServer) changePlayerHealth(player *realm.Character, delta int64) error {
	player.Health += delta
	if player.Health < 0 {
		player.Health = 0
	}
	if s.Characters != nil {
		if err := s.Characters.UpdateHealth(player.GUID, player.AccountID, player.RealmID, player.Health); err != nil {
			return err
		}
	}
	s.updatePlayer(*player)
	for _, field := range []int{22, 27} {
		if update, err := packet.EncodeFieldUpdate(uint64(player.GUID), field, uint32(player.Health)); err == nil {
			s.sendSpell(*player, update)
		}
	}
	return nil
}

func (s *WorldServer) changePlayerPower(player *realm.Character, powerType, delta int64) error {
	power := playerPower(*player, powerType) + delta
	if power < 0 {
		power = 0
	}
	setPlayerPower(player, powerType, power)
	if s.Characters != nil {
		if err := s.Characters.UpdatePower(player.GUID, player.AccountID, player.RealmID, powerType, power); err != nil {
			return err
		}
	}
	s.updatePlayer(*player)
	for _, field := range []int{23 + int(powerType), 28 + int(powerType)} {
		if update, err := packet.EncodeFieldUpdate(uint64(player.GUID), field, uint32(power)); err == nil {
			s.sendSpell(*player, update)
		}
	}
	return nil
}

func playerPower(player realm.Character, powerType int64) int64 {
	switch powerType {
	case 1:
		return player.Power2
	case 2:
		return player.Power3
	case 3:
		return player.Power4
	case 4:
		return player.Power5
	default:
		return player.Power1
	}
}

func setPlayerPower(player *realm.Character, powerType, value int64) {
	switch powerType {
	case 1:
		player.Power2 = value
	case 2:
		player.Power3 = value
	case 3:
		player.Power4 = value
	case 4:
		player.Power5 = value
	default:
		player.Power1 = value
	}
}

func (s *WorldServer) sendSpell(caster realm.Character, data []byte) {
	s.broadcastPlayer(caster, data)
	s.sendPlayer(caster.GUID, data)
}

func (s *WorldServer) spellOnCooldown(guid, spellID int64) bool {
	now := time.Now()
	s.spells.mu.Lock()
	defer s.spells.mu.Unlock()
	if s.spells.cooldowns == nil || s.spells.cooldowns[guid] == nil {
		return false
	}
	when, found := s.spells.cooldowns[guid][spellID]
	if found && now.Before(when) {
		return true
	}
	delete(s.spells.cooldowns[guid], spellID)
	return false
}

func (s *WorldServer) setSpellCooldown(guid int64, spell dbc.Spell) {
	cooldown := spell.RecoveryTime
	if spell.CategoryRecoveryTime > cooldown {
		cooldown = spell.CategoryRecoveryTime
	}
	if cooldown <= 0 {
		return
	}
	s.spells.mu.Lock()
	if s.spells.cooldowns == nil {
		s.spells.cooldowns = make(map[int64]map[int64]time.Time)
	}
	if s.spells.cooldowns[guid] == nil {
		s.spells.cooldowns[guid] = make(map[int64]time.Time)
	}
	s.spells.cooldowns[guid][spell.ID] = time.Now().Add(time.Duration(cooldown) * time.Millisecond)
	s.spells.mu.Unlock()
}

func (s *WorldServer) cancelSpell(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 4 {
		return nil, nil
	}
	id := int64(binary.LittleEndian.Uint32(data))
	s.spells.mu.Lock()
	cast := s.spells.casts[active.GUID]
	if cast == nil || cast.spell.ID != id {
		s.spells.mu.Unlock()
		return nil, nil
	}
	delete(s.spells.casts, active.GUID)
	if cast.timer != nil {
		cast.timer.Stop()
	}
	s.spells.mu.Unlock()
	result, err := spellCastResult(id, packet.SpellFailedInterrupted)
	if err != nil {
		return nil, err
	}
	return [][]byte{result}, nil
}

func (s *WorldServer) castFailure(active realm.Character, spellID int64, reason packet.SpellCastResult) ([][]byte, error) {
	result, err := spellCastResult(spellID, reason)
	if err != nil {
		return nil, err
	}
	return [][]byte{result}, nil
}

func spellCastResult(spellID int64, result packet.SpellCastResult) ([]byte, error) {
	data := encodeUint32(spellID)
	if result == packet.SpellNoError {
		data = append(data, byte(packet.SpellCastSuccess))
	} else {
		data = append(data, byte(packet.SpellCastFailed), byte(result))
	}
	return packet.Encode(packet.SMSGCastResult, data)
}

func spellStartPacket(cast *spellCast) ([]byte, error) {
	source := cast.caster.GUID
	if cast.sourceItem != nil {
		source = cast.sourceItem.GUID
	}
	data := append(encodeGUID(source), encodeGUID(cast.caster.GUID)...)
	data = append(data, encodeUint32(cast.spell.ID)...)
	data = append(data, 0, 0)
	data = append(data, encodeInt32(cast.castTime)...)
	data = append(data, byte(cast.targetMask), byte(cast.targetMask>>8))
	if cast.targetMask != packet.SpellTargetSelf {
		data = append(data, spellTargetData(cast.caster, cast.targetMask, cast.target)...)
	}
	return packet.Encode(packet.SMSGSpellStart, data)
}

func spellGoPacket(cast *spellCast) ([]byte, error) {
	source := cast.caster.GUID
	if cast.sourceItem != nil {
		source = cast.sourceItem.GUID
	}
	data := append(encodeGUID(source), encodeGUID(cast.caster.GUID)...)
	data = append(data, encodeUint32(cast.spell.ID)...)
	data = append(data, 0, 0)
	hit := cast.target.UnitGUID
	if hit == 0 && cast.targetMask == packet.SpellTargetSelf {
		hit = uint64(cast.caster.GUID)
	}
	if hit != 0 {
		data = append(data, 1)
		data = append(data, encodeUint64(hit)...)
	} else {
		data = append(data, 0)
	}
	data = append(data, 0, byte(cast.targetMask), byte(cast.targetMask>>8))
	if cast.targetMask != packet.SpellTargetSelf {
		data = append(data, spellTargetData(cast.caster, cast.targetMask, cast.target)...)
	}
	return packet.Encode(packet.SMSGSpellGo, data)
}

func spellTargetData(caster realm.Character, mask packet.SpellTargetMask, target spellTarget) []byte {
	data := make([]byte, 0, 64)
	if mask&(packet.SpellTargetUnit|packet.SpellTargetGameObject) != 0 {
		guid := target.UnitGUID
		if guid == 0 {
			guid = uint64(caster.GUID)
		}
		data = append(data, encodeUint64(guid)...)
	}
	if mask&packet.SpellTargetItemMask != 0 {
		data = append(data, encodeUint64(target.ItemGUID)...)
	}
	vector := target.Source
	if vector == nil {
		vector = &spellVector{caster.PositionX, caster.PositionY, caster.PositionZ}
	}
	if mask&packet.SpellTargetSource != 0 {
		data = append(data, encodeFloat(vector.X)...)
		data = append(data, encodeFloat(vector.Y)...)
		data = append(data, encodeFloat(vector.Z)...)
	}
	if mask&packet.SpellTargetDestination != 0 {
		vector = target.Dest
		if vector == nil {
			vector = &spellVector{caster.PositionX, caster.PositionY, caster.PositionZ}
		}
		data = append(data, encodeFloat(vector.X)...)
		data = append(data, encodeFloat(vector.Y)...)
		data = append(data, encodeFloat(vector.Z)...)
	}
	if mask&packet.SpellTargetString != 0 {
		data = append(data, make([]byte, 128)...)
	}
	return data
}

func encodeInt32(value int64) []byte {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(value))
	return data
}

func encodeFloat(value float32) []byte {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, math.Float32bits(value))
	return data
}
