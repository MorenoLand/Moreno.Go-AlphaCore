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
	caster         realm.Character
	target         spellTarget
	spell          dbc.Spell
	targetMask     packet.SpellTargetMask
	castTime       int64
	powerCost      int64
	targets        []realm.Character
	effectTargets  map[int][]realm.Character
	targetCreature *creatureState
	sourceItem     *realm.InventoryItem
	source         worlddb.ItemTemplate
	sourceSlot     int
	triggered      bool
	started        time.Time
	timer          *time.Timer
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
	if template.Bonding == 3 && item.Flags&itemDynBound == 0 {
		item.Flags |= itemDynBound
		if err := s.Characters.UpdateItemFlags(item.GUID, item.Owner, item.Flags); err != nil {
			return nil, err
		}
		if update, err := packet.EncodeFieldUpdate(uint64(item.GUID)|0x4000000000000000, itemFieldFlag, encodedItemFlags(template, item.Flags)); err == nil {
			s.sendPlayer(active.GUID, update)
		}
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
		if found {
			if player.Map != active.Map {
				return spellTarget{}, false
			}
		} else if _, creatureFound, err := s.creatureStateAt(active, target.UnitGUID, creatureViewDistance); err != nil || !creatureFound {
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
		if found {
			if player.Map != active.Map {
				return s.castFailure(active, spellID, packet.SpellFailedBadTargets)
			}
		} else if _, found, err := s.creatureStateAt(active, target.UnitGUID, creatureViewDistance); err != nil {
			return nil, err
		} else if !found {
			return s.castFailure(active, spellID, packet.SpellFailedBadTargets)
		}
	}
	if result := s.validateSpellTarget(active, spell, target, targetMask); result != packet.SpellNoError {
		return s.castFailure(active, spellID, result)
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
	if sourceItem == nil {
		if s.hasAuraType(active.GUID, packet.AuraModStun) {
			return s.castFailure(active, spellID, packet.SpellFailedStunned)
		}
		if s.hasAuraType(active.GUID, packet.AuraModSilence) {
			return s.castFailure(active, spellID, packet.SpellFailedSilenced)
		}
		if s.hasAuraType(active.GUID, packet.AuraModPacify) && spell.School == 0 && packet.SpellAttributes(spell.Attributes)&packet.SpellAttributeAbility != 0 {
			return s.castFailure(active, spellID, packet.SpellFailedPacified)
		}
	}
	if s.spellOnCooldown(active.GUID, spellID) {
		return s.castFailure(active, spellID, packet.SpellFailedNotReady)
	}
	powerCost := spell.ManaCost + spell.ManaCostPerLevel*int64(active.Level)
	if packet.SpellAttributesEx(spell.AttributesEx)&packet.SpellAttributeExDrainAllPower != 0 {
		powerCost = playerPower(active, spell.PowerType)
	}
	if powerCost > playerPower(active, spell.PowerType) {
		return s.castFailure(active, spellID, packet.SpellFailedNoPower)
	}
	if spellRange, rangeFound, rangeErr := s.DBC.SpellRange(spell.RangeIndex); rangeErr != nil {
		return nil, rangeErr
	} else if rangeFound && target.UnitGUID != 0 && target.UnitGUID != uint64(active.GUID) {
		player, playerFound := s.playerByGUID(int64(target.UnitGUID))
		position := realm.Character{}
		if playerFound {
			position = player
		} else if state, found, stateErr := s.creatureStateAt(active, target.UnitGUID, creatureViewDistance); stateErr == nil && found {
			position = realm.Character{PositionX: state.Spawn.PositionX, PositionY: state.Spawn.PositionY, PositionZ: state.Spawn.PositionZ}
		}
		dx, dy, dz := active.PositionX-position.PositionX, active.PositionY-position.PositionY, active.PositionZ-position.PositionZ
		if spellRange.RangeMax > 0 && dx*dx+dy*dy+dz*dz > spellRange.RangeMax*spellRange.RangeMax {
			return s.castFailure(active, spellID, packet.SpellFailedOutOfRange)
		}
		if spellRange.RangeMin > 0 && dx*dx+dy*dy+dz*dz < spellRange.RangeMin*spellRange.RangeMin {
			return s.castFailure(active, spellID, packet.SpellFailedTooClose)
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
	var targetCreature *creatureState
	if target.UnitGUID != 0 && target.UnitGUID != uint64(active.GUID) {
		if _, found := s.playerByGUID(int64(target.UnitGUID)); !found {
			targetCreature, _, err = s.creatureStateAt(active, target.UnitGUID, creatureViewDistance)
			if err != nil {
				return nil, err
			}
		}
	}
	cast := &spellCast{caster: active, target: target, spell: spell, targetMask: targetMask, castTime: castTime, powerCost: powerCost, targets: s.spellTargets(active, spell, target), effectTargets: s.spellEffectTargetsAll(active, spell, target), targetCreature: targetCreature, sourceItem: sourceItem, sourceSlot: sourceSlot, source: source, started: time.Now()}
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

func (s *WorldServer) validateSpellTarget(active realm.Character, spell dbc.Spell, target spellTarget, mask packet.SpellTargetMask) packet.SpellCastResult {
	attributes := packet.SpellAttributes(spell.Attributes)
	attributesEx := packet.SpellAttributesEx(spell.AttributesEx)
	if s.combatTarget(active.GUID) != 0 && attributes&packet.SpellAttributeCantCombat != 0 {
		return packet.SpellFailedAffectingCombat
	}
	if target.UnitGUID == 0 {
		if spell.Targets&int64(packet.SpellTargetItem) != 0 && target.ItemGUID == 0 {
			return packet.SpellFailedBadTargets
		}
		return packet.SpellNoError
	}
	if target.UnitGUID == uint64(active.GUID) {
		if attributesEx&packet.SpellAttributeExCantTargetSelf != 0 {
			return packet.SpellFailedBadTargets
		}
		return packet.SpellNoError
	}
	targetPlayer, targetFound := s.playerByGUID(int64(target.UnitGUID))
	if !targetFound {
		if state, found, _ := s.creatureStateAt(active, target.UnitGUID, creatureViewDistance); found {
			if state.Health <= 0 && spell.Targets&int64(packet.SpellTargetDead) == 0 && !spellHasEffect(spell, packet.SpellEffectResurrect) {
				return packet.SpellFailedTargetsDead
			}
			if state.Health > 0 && spellHasEffect(spell, packet.SpellEffectResurrect) {
				return packet.SpellFailedTargetNotDead
			}
			return packet.SpellNoError
		}
		return packet.SpellFailedBadTargets
	}
	if targetPlayer.Health <= 0 && spell.Targets&int64(packet.SpellTargetDead) == 0 && !spellHasEffect(spell, packet.SpellEffectResurrect) {
		return packet.SpellFailedTargetsDead
	}
	if targetPlayer.Health > 0 && spellHasEffect(spell, packet.SpellEffectResurrect) {
		return packet.SpellFailedTargetNotDead
	}
	team, targetTeam, err := s.teams(active, targetPlayer)
	if err != nil || team == 0 || targetTeam == 0 {
		return packet.SpellNoError
	}
	if spellHarmful(spell) && team == targetTeam {
		return packet.SpellFailedTargetFriendly
	}
	if !spellHarmful(spell) && team != targetTeam && mask&packet.SpellTargetDead == 0 {
		return packet.SpellFailedTargetEnemy
	}
	return packet.SpellNoError
}

func spellHasEffect(spell dbc.Spell, effect packet.SpellEffect) bool {
	for _, value := range spell.Effects {
		if packet.SpellEffect(value.Type) == effect {
			return true
		}
	}
	return false
}

func spellHarmful(spell dbc.Spell) bool {
	if packet.SpellAttributes(spell.Attributes)&packet.SpellAttributeAuraDebuff != 0 {
		return true
	}
	for _, effect := range spell.Effects {
		switch packet.SpellEffect(effect.Type) {
		case packet.SpellEffectInstantKill, packet.SpellEffectSchoolDamage, packet.SpellEffectPowerBurn, packet.SpellEffectHealthLeech, packet.SpellEffectPowerDrain:
			return true
		}
	}
	return false
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
	result, err := spellCastResult(cast.spell.ID, packet.SpellNoError)
	if err == nil {
		s.sendSpell(cast.caster, result)
	}
	goPacket, err := spellGoPacket(cast)
	if err == nil {
		s.sendSpell(cast.caster, goPacket)
	}
	s.setSpellCooldown(cast.caster, cast.spell)
	if cast.powerCost > 0 {
		_ = s.changePlayerPower(&cast.caster, cast.spell.PowerType, -cast.powerCost)
	}
	s.consumeItemSpell(cast)
	s.applySpellEffects(cast)
}

func (s *WorldServer) triggerSpell(caster realm.Character, spellID int64, target spellTarget, targetMask packet.SpellTargetMask) {
	spell, found, err := s.DBC.Spell(spellID)
	if err != nil || !found {
		return
	}
	cast := &spellCast{caster: caster, target: target, spell: spell, targetMask: targetMask, targets: s.spellTargets(caster, spell, target), effectTargets: s.spellEffectTargetsAll(caster, spell, target), triggered: true, started: time.Now()}
	s.performSpellCast(cast)
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
	effectiveLevel := int64(cast.caster.Level) - cast.spell.BaseLevel
	if effectiveLevel < 0 {
		effectiveLevel = 0
	}
	for index, effect := range cast.spell.Effects {
		targets := cast.effectTargets[index]
		if len(targets) == 0 && cast.targetCreature != nil {
			s.applyCreatureSpellEffect(cast, cast.targetCreature, effect, spellEffectPoints(effect, effectiveLevel))
		}
		for _, target := range targets {
			points := spellEffectPoints(effect, effectiveLevel)
			switch packet.SpellEffect(effect.Type) {
			case packet.SpellEffectInstantKill:
				s.sendSpellDamage(cast.caster, target.GUID, target.Health, cast.spell.ID)
				_ = s.changePlayerHealth(&target, -target.Health)
			case packet.SpellEffectSchoolDamage, packet.SpellEffectWeaponDamage, packet.SpellEffectWeaponDamagePlus:
				s.sendSpellDamage(cast.caster, target.GUID, minPower(target.Health, points), cast.spell.ID)
				_ = s.changePlayerHealth(&target, -points)
			case packet.SpellEffectPowerBurn:
				amount := minPower(playerPower(target, effect.MiscValue), points)
				if amount > 0 {
					_ = s.changePlayerPower(&target, effect.MiscValue, -amount)
					s.sendSpellDamage(cast.caster, target.GUID, amount, cast.spell.ID)
					_ = s.changePlayerHealth(&target, -amount)
				}
			case packet.SpellEffectHeal:
				_ = s.changePlayerHealth(&target, points)
			case packet.SpellEffectHealMaxHealth:
				_ = s.changePlayerHealth(&target, s.playerMaxHealth(cast.caster.GUID)-target.Health)
			case packet.SpellEffectHealthLeech:
				if s.changePlayerHealth(&target, -points) == nil {
					s.sendSpellDamage(cast.caster, target.GUID, minPower(target.Health+points, points), cast.spell.ID)
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
			case packet.SpellEffectTriggerSpell:
				targetMask := packet.SpellTargetSelf
				if target.GUID != cast.caster.GUID {
					targetMask = packet.SpellTargetUnit
				}
				s.triggerSpell(cast.caster, effect.TriggerSpell, spellTarget{UnitGUID: uint64(target.GUID)}, targetMask)
			case packet.SpellEffectLearnSpell:
				if s.Characters == nil || effect.TriggerSpell <= 0 {
					continue
				}
				if err := s.Characters.AddSpell(target.GUID, effect.TriggerSpell); err != nil {
					continue
				}
				learned := append(encodeUint16(effect.TriggerSpell), encodeUint16(0)...)
				if update, err := packet.Encode(packet.SMSGLearnedSpell, learned); err == nil {
					s.sendPlayer(target.GUID, update)
				}
			case packet.SpellEffectCreateItem:
				s.createSpellItem(cast, target, effect, points)
			case packet.SpellEffectResurrect:
				if target.Health <= 0 {
					s.requestResurrection(cast, target, points)
				}
			case packet.SpellEffectBind:
				s.bindSpellTarget(cast, target)
			case packet.SpellEffectStuck:
				s.stuckSpellTarget(target)
			}
		}
	}
}

func (s *WorldServer) bindSpellTarget(cast *spellCast, target realm.Character) {
	if s.Characters == nil || target.GUID == 0 {
		return
	}
	bind := realm.Deathbind{PlayerGUID: target.GUID, Map: target.Map, Zone: target.Zone, X: target.PositionX, Y: target.PositionY, Z: target.PositionZ}
	if err := s.Characters.SaveDeathbind(bind); err != nil {
		return
	}
	if point, err := deathbindPointPacket(bind); err == nil {
		s.sendPlayer(target.GUID, point)
	}
	if bound, err := packet.Encode(packet.SMSGPlayerBound, encodeGUID(cast.caster.GUID)); err == nil {
		s.sendPlayer(target.GUID, bound)
	}
}

func (s *WorldServer) stuckSpellTarget(target realm.Character) {
	if s.Characters == nil {
		return
	}
	bind, found, err := s.Characters.Deathbind(target.GUID)
	if err != nil || !found {
		return
	}
	teleport := target
	if response, err := s.teleportPlayer(&teleport, bind.Map, bind.X, bind.Y, bind.Z, target.Orientation); err == nil {
		s.sendPlayer(target.GUID, response)
	}
}

func (s *WorldServer) applyCreatureSpellEffect(cast *spellCast, target *creatureState, effect dbc.SpellEffect, points int64) {
	switch packet.SpellEffect(effect.Type) {
	case packet.SpellEffectSchoolDamage:
		s.sendSpellDamage(cast.caster, int64(target.GUID), minPower(target.Health, points), cast.spell.ID)
		s.changeCreatureHealth(target, -points)
	case packet.SpellEffectInstantKill:
		s.sendSpellDamage(cast.caster, int64(target.GUID), target.Health, cast.spell.ID)
		s.changeCreatureHealth(target, -target.Health)
	case packet.SpellEffectWeaponDamage, packet.SpellEffectWeaponDamagePlus:
		s.sendSpellDamage(cast.caster, int64(target.GUID), minPower(target.Health, points), cast.spell.ID)
		s.changeCreatureHealth(target, -points)
	case packet.SpellEffectPowerBurn:
		amount := minPower(target.Mana, points)
		if amount > 0 {
			s.changeCreaturePower(target, effect.MiscValue, -amount)
			s.changeCreatureHealth(target, -amount)
		}
	case packet.SpellEffectHeal:
		s.changeCreatureHealth(target, points)
	case packet.SpellEffectHealMaxHealth:
		s.changeCreatureHealth(target, target.MaxHealth-target.Health)
	case packet.SpellEffectPowerDrain:
		amount := minPower(target.Mana, points)
		if amount > 0 {
			s.changeCreaturePower(target, effect.MiscValue, -amount)
			caster := cast.caster
			_ = s.changePlayerPower(&caster, effect.MiscValue, amount)
		}
	}
}

func (s *WorldServer) spellTargets(caster realm.Character, spell dbc.Spell, initial spellTarget) []realm.Character {
	targets := make([]realm.Character, 0, 4)
	for _, effect := range spell.Effects {
		for _, target := range s.spellEffectTargets(caster, spell, initial, effect) {
			duplicate := false
			for _, current := range targets {
				if current.GUID == target.GUID {
					duplicate = true
					break
				}
			}
			if !duplicate {
				targets = append(targets, target)
			}
		}
	}
	return targets
}

func (s *WorldServer) spellEffectTargetsAll(caster realm.Character, spell dbc.Spell, initial spellTarget) map[int][]realm.Character {
	targets := make(map[int][]realm.Character, len(spell.Effects))
	for index, effect := range spell.Effects {
		targets[index] = s.spellEffectTargets(caster, spell, initial, effect)
	}
	return targets
}

func (s *WorldServer) spellEffectTargets(caster realm.Character, spell dbc.Spell, initial spellTarget, effect dbc.SpellEffect) []realm.Character {
	initialTargets := make([]realm.Character, 0, 1)
	if initial.UnitGUID != 0 {
		if target, found := s.playerByGUID(int64(initial.UnitGUID)); found {
			initialTargets = append(initialTargets, target)
		} else if initial.UnitGUID == uint64(caster.GUID) {
			initialTargets = append(initialTargets, caster)
		}
	}
	mode := spellAreaMode(effect.ImplicitTargetA, effect.ImplicitTargetB)
	if mode == 0 || s.DBC == nil {
		return initialTargets
	}
	radiusEntry, found, err := s.DBC.SpellRadius(effect.RadiusIndex)
	if err != nil || !found {
		return nil
	}
	radius := radiusEntry.Radius + radiusEntry.RadiusPerLevel*float32(maxSpellLevel(caster.Level, spell.BaseLevel))
	if radiusEntry.RadiusMax > 0 && radius > radiusEntry.RadiusMax {
		radius = radiusEntry.RadiusMax
	}
	if radius <= 0 {
		return nil
	}
	center := caster
	if (effect.ImplicitTargetA == int64(packet.SpellImplicitAllEnemyInstant) || effect.ImplicitTargetB == int64(packet.SpellImplicitAllEnemyInstant)) && initial.UnitGUID != 0 {
		if value, found := s.playerByGUID(int64(initial.UnitGUID)); found {
			center = value
		}
	}
	targets := make([]realm.Character, 0, 4)
	for _, player := range s.onlinePlayers() {
		if player.Map != caster.Map {
			continue
		}
		dx, dy, dz := player.PositionX-center.PositionX, player.PositionY-center.PositionY, player.PositionZ-center.PositionZ
		if dx*dx+dy*dy+dz*dz <= radius*radius && spellAreaTarget(s, caster, player, mode) {
			targets = append(targets, player)
		}
	}
	return targets
}

func spellAreaMode(first, second int64) byte {
	if first == int64(packet.SpellImplicitAllEnemyInArea) || first == int64(packet.SpellImplicitAllEnemyInstant) || second == int64(packet.SpellImplicitAllEnemyInArea) || second == int64(packet.SpellImplicitAllEnemyInstant) {
		return 2
	}
	if first == int64(packet.SpellImplicitAllFriendlyAround) || first == int64(packet.SpellImplicitAllFriendlyInArea) || first == int64(packet.SpellImplicitAllParty) || first == int64(packet.SpellImplicitAroundCasterParty) || second == int64(packet.SpellImplicitAllFriendlyAround) || second == int64(packet.SpellImplicitAllFriendlyInArea) || second == int64(packet.SpellImplicitAllParty) {
		return 1
	}
	if first == int64(packet.SpellImplicitAllAroundCaster) {
		return 3
	}
	return 0
}

func spellAreaTarget(s *WorldServer, caster, target realm.Character, mode byte) bool {
	if mode == 3 {
		return true
	}
	if target.GUID == caster.GUID {
		return mode == 1
	}
	first, second, err := s.teams(caster, target)
	if err != nil {
		return false
	}
	if first == 0 && second == 0 {
		return mode == 2
	}
	if mode == 1 {
		return first == second
	}
	return first != second
}

func maxSpellLevel(level uint8, base int64) int64 {
	value := int64(level) - base
	if value < 0 {
		return 0
	}
	return value
}

func (s *WorldServer) createSpellItem(cast *spellCast, target realm.Character, effect dbc.SpellEffect, amount int64) {
	if s.Characters == nil || s.WorldData == nil || target.GUID == 0 {
		return
	}
	template, found, err := s.WorldData.ItemTemplate(effect.ItemType)
	if err != nil || !found {
		return
	}
	if amount < 1 {
		amount = 1
	}
	if template.Stackable > 0 && amount > template.Stackable {
		amount = template.Stackable
	}
	slot, err := s.Characters.FirstEmptySlot(target.GUID, 23, 23, 39)
	if err != nil || slot < 0 {
		return
	}
	item, err := s.Characters.CreateInventoryItem(target.GUID, cast.caster.GUID, 23, slot, effect.ItemType, amount)
	if err != nil {
		return
	}
	created, err := packet.EncodeItemCreate(uint64(item.GUID)|0x4000000000000000, uint32(item.ItemTemplate), uint64(item.Owner), uint64(item.Creator), uint32(item.StackCount), 0, 0, item.SpellCharges, packet.Movement{X: target.PositionX, Y: target.PositionY, Z: target.PositionZ, O: target.Orientation})
	if err == nil {
		s.sendPlayer(target.GUID, created)
	}
	push, err := itemPushResult(item, effect.ItemType, 23)
	if err == nil {
		s.sendPlayer(target.GUID, push)
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

func (s *WorldServer) sendSpellDamage(caster realm.Character, target, amount, spellID int64) {
	data := append(encodeUint64(uint64(target)), encodeUint32(amount)...)
	data = append(data, encodeUint32(0)...)
	data = append(data, encodeInt32(0)...)
	data = append(data, encodeUint32(spellID)...)
	data = append(data, encodeUint64(uint64(caster.GUID))...)
	if update, err := packet.Encode(packet.SMSGDamageDone, data); err == nil {
		s.sendSpell(caster, update)
	}
}

func (s *WorldServer) hasAuraType(guid int64, auraType packet.AuraType) bool {
	s.auras.mu.Lock()
	defer s.auras.mu.Unlock()
	for _, aura := range s.auras.active[guid] {
		if packet.AuraType(aura.effect.Aura) == auraType {
			return true
		}
	}
	return false
}

func (s *WorldServer) interruptMovement(active realm.Character, moved, turned bool) {
	if moved || turned {
		s.interruptAuras(active, moved, turned)
	}
	if !moved {
		return
	}
	s.spells.mu.Lock()
	cast := s.spells.casts[active.GUID]
	if cast == nil || cast.spell.InterruptFlags&packet.SpellInterruptMovement == 0 {
		s.spells.mu.Unlock()
		return
	}
	delete(s.spells.casts, active.GUID)
	if cast.timer != nil {
		cast.timer.Stop()
	}
	s.spells.mu.Unlock()
	result, err := spellCastResult(cast.spell.ID, packet.SpellFailedInterrupted)
	if err == nil {
		s.sendPlayer(active.GUID, result)
	}
	if failure, err := packet.Encode(packet.SMSGSpellFailure, append(encodeGUID(active.GUID), append(encodeUint32(cast.spell.ID), byte(packet.SpellFailedInterrupted))...)); err == nil {
		s.broadcastPlayer(active, failure)
		s.sendPlayer(active.GUID, failure)
	}
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

func (s *WorldServer) setSpellCooldown(caster realm.Character, spell dbc.Spell) {
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
	if s.spells.cooldowns[caster.GUID] == nil {
		s.spells.cooldowns[caster.GUID] = make(map[int64]time.Time)
	}
	expires := time.Now().Add(time.Duration(cooldown) * time.Millisecond)
	s.spells.cooldowns[caster.GUID][spell.ID] = expires
	s.spells.mu.Unlock()
	data := append(encodeUint32(spell.ID), encodeGUID(caster.GUID)...)
	data = append(data, encodeUint16(cooldown)...)
	if update, err := packet.Encode(packet.SMSGSpellCooldown, data); err == nil {
		s.sendPlayer(caster.GUID, update)
	}
	time.AfterFunc(time.Duration(cooldown)*time.Millisecond, func() {
		s.spells.mu.Lock()
		current, found := s.spells.cooldowns[caster.GUID][spell.ID]
		if found && !current.After(time.Now()) {
			delete(s.spells.cooldowns[caster.GUID], spell.ID)
		}
		s.spells.mu.Unlock()
		if found && !current.After(time.Now()) {
			clear := append(encodeUint32(spell.ID), encodeGUID(caster.GUID)...)
			if update, err := packet.Encode(packet.SMSGClearCooldown, clear); err == nil {
				s.sendPlayer(caster.GUID, update)
			}
		}
	})
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
	failure, err := packet.Encode(packet.SMSGSpellFailure, append(encodeGUID(active.GUID), append(encodeUint32(id), byte(packet.SpellFailedInterrupted))...))
	if err != nil {
		return nil, err
	}
	s.broadcastPlayer(active, failure)
	s.sendPlayer(active.GUID, failure)
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
	targets := cast.targets
	if len(targets) == 0 {
		if cast.target.UnitGUID != 0 {
			targets = []realm.Character{{GUID: int64(cast.target.UnitGUID)}}
		} else {
			targets = []realm.Character{cast.caster}
		}
	}
	data = append(data, byte(len(targets)))
	for _, target := range targets {
		data = append(data, encodeUint64(uint64(target.GUID))...)
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

func encodeUint16(value int64) []byte {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, uint16(value))
	return data
}

func encodeFloat(value float32) []byte {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, math.Float32bits(value))
	return data
}
