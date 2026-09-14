package world

import (
	"encoding/binary"
	"sync"
	"time"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const (
	unitAuraField      = 56
	unitAuraFlagsField = 112
	positiveAuraStart  = 0
	harmfulAuraStart   = 32
	visibleAuraEnd     = 56
)

type auraRegistry struct {
	mu     sync.Mutex
	active map[int64]map[int]*auraState
}

type auraState struct {
	spellID, casterID int64
	slot, effectIndex int
	duration          int64
	passive, harmful  bool
	cancelable        bool
	points            int64
	target            realm.Character
	effect            dbc.SpellEffect
	interruptFlags    int64
	timer, periodic   *time.Timer
}

func (s *WorldServer) applyAura(cast *spellCast, target realm.Character, effectIndex int, effect dbc.SpellEffect) {
	if target.GUID == 0 || s.DBC == nil {
		return
	}
	duration := int64(-1)
	if value, found, err := s.DBC.SpellDuration(cast.spell.DurationIndex); err == nil && found {
		duration = value.Duration
		level := int64(cast.caster.Level) - cast.spell.BaseLevel
		if level < 0 {
			level = 0
		}
		if duration >= 0 {
			duration += value.DurationPerLevel * level
			if value.MaxDuration > 0 && duration > value.MaxDuration {
				duration = value.MaxDuration
			}
		}
	}
	passive := packet.SpellAttributes(cast.spell.Attributes)&packet.SpellAttributePassive != 0
	harmful := packet.SpellAttributes(cast.spell.Attributes)&packet.SpellAttributeAuraDebuff != 0
	period := effect.AuraPeriod
	if period == 0 {
		switch packet.AuraType(effect.Aura) {
		case packet.AuraPeriodicDamage, packet.AuraPeriodicHeal, packet.AuraPeriodicTriggerSpell, packet.AuraPeriodicEnergize, packet.AuraPeriodicLeech, packet.AuraPeriodicManaFunnel, packet.AuraPeriodicManaLeech:
			period = 5000
		}
	}
	aura := &auraState{spellID: cast.spell.ID, casterID: cast.caster.GUID, slot: -1, effectIndex: effectIndex, duration: duration, passive: passive, harmful: harmful, cancelable: !harmful && packet.SpellAttributes(cast.spell.Attributes)&packet.SpellAttributeCantCancel == 0, points: spellEffectPoints(effect, maxSpellLevel(cast.caster.Level, cast.spell.BaseLevel)), target: target, effect: effect, interruptFlags: cast.spell.AuraInterruptFlags}
	s.auras.mu.Lock()
	if s.auras.active == nil {
		s.auras.active = make(map[int64]map[int]*auraState)
	}
	if s.auras.active[target.GUID] == nil {
		s.auras.active[target.GUID] = make(map[int]*auraState)
	}
	for slot, current := range s.auras.active[target.GUID] {
		if current.spellID == aura.spellID {
			aura.slot = slot
			if current.timer != nil {
				current.timer.Stop()
			}
			break
		}
	}
	if aura.slot < 0 && !aura.passive {
		start := positiveAuraStart
		if aura.harmful {
			start = harmfulAuraStart
		}
		aura.slot = start
		for slot := start; slot < visibleAuraEnd; slot++ {
			if _, exists := s.auras.active[target.GUID][slot]; !exists {
				aura.slot = slot
				break
			}
		}
	}
	s.auras.active[target.GUID][aura.slot] = aura
	s.auras.mu.Unlock()
	s.auraEffectChange(target, aura, false)
	if aura.passive {
		return
	}
	if duration > 0 {
		aura.timer = time.AfterFunc(time.Duration(duration)*time.Millisecond, func() { s.removeAura(target, aura.slot) })
	}
	if period > 0 {
		aura.periodic = time.AfterFunc(time.Duration(period)*time.Millisecond, func() { s.tickAura(target.GUID, aura.slot, aura, period) })
	}
	s.writeAura(target, aura, false)
	s.refreshAuraUnitFlags(target)
}

func (s *WorldServer) interruptAuras(target realm.Character, moved, turned bool) {
	s.auras.mu.Lock()
	remove := make([]int, 0)
	for slot, aura := range s.auras.active[target.GUID] {
		if aura.passive {
			continue
		}
		if moved && aura.interruptFlags&packet.SpellAuraInterruptMovement != 0 || turned && aura.interruptFlags&packet.SpellAuraInterruptTurning != 0 {
			remove = append(remove, slot)
		}
	}
	s.auras.mu.Unlock()
	for _, slot := range remove {
		s.removeAura(target, slot)
	}
}

func (s *WorldServer) tickAura(guid int64, slot int, aura *auraState, period int64) {
	s.auras.mu.Lock()
	if s.auras.active[guid][slot] != aura {
		s.auras.mu.Unlock()
		return
	}
	target, found := s.playerByGUID(guid)
	if !found {
		target = aura.target
	}
	s.auras.mu.Unlock()
	points := spellEffectPoints(aura.effect, 0)
	switch packet.AuraType(aura.effect.Aura) {
	case packet.AuraPeriodicDamage:
		_ = s.changePlayerHealth(&target, -points)
	case packet.AuraPeriodicHeal:
		_ = s.changePlayerHealth(&target, points)
	case packet.AuraPeriodicEnergize:
		_ = s.changePlayerPower(&target, aura.effect.MiscValue, points)
	case packet.AuraPeriodicManaLeech:
		amount := minPower(playerPower(target, aura.effect.MiscValue), points)
		if amount > 0 {
			_ = s.changePlayerPower(&target, aura.effect.MiscValue, -amount)
			caster, casterFound := s.playerByGUID(aura.casterID)
			if casterFound {
				_ = s.changePlayerPower(&caster, aura.effect.MiscValue, amount)
			}
		}
	case packet.AuraPeriodicTriggerSpell:
		caster, casterFound := s.playerByGUID(aura.casterID)
		if casterFound {
			s.triggerSpell(caster, aura.effect.TriggerSpell, spellTarget{UnitGUID: uint64(target.GUID)}, packet.SpellTargetUnit)
		}
	case packet.AuraPeriodicLeech:
		if s.changePlayerHealth(&target, -points) == nil {
			caster, casterFound := s.playerByGUID(aura.casterID)
			if casterFound {
				_ = s.changePlayerHealth(&caster, points)
			}
		}
	}
	s.auras.mu.Lock()
	if s.auras.active[guid][slot] == aura {
		aura.periodic = time.AfterFunc(time.Duration(period)*time.Millisecond, func() { s.tickAura(guid, slot, aura, period) })
	}
	s.auras.mu.Unlock()
}

func (s *WorldServer) cancelAura(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 4 {
		return nil, nil
	}
	spellID := int64(binary.LittleEndian.Uint32(data))
	s.auras.mu.Lock()
	auras := s.auras.active[active.GUID]
	remove := make([]*auraState, 0)
	for _, aura := range auras {
		if aura.spellID != spellID {
			continue
		}
		if aura.passive || !aura.cancelable {
			s.auras.mu.Unlock()
			return nil, nil
		}
		remove = append(remove, aura)
	}
	s.auras.mu.Unlock()
	for _, aura := range remove {
		s.removeAura(active, aura.slot)
	}
	return nil, nil
}

func (s *WorldServer) removeAura(target realm.Character, slot int) {
	s.auras.mu.Lock()
	aura := s.auras.active[target.GUID][slot]
	if aura == nil {
		s.auras.mu.Unlock()
		return
	}
	if aura.timer != nil {
		aura.timer.Stop()
	}
	if aura.periodic != nil {
		aura.periodic.Stop()
	}
	delete(s.auras.active[target.GUID], slot)
	if len(s.auras.active[target.GUID]) == 0 {
		delete(s.auras.active, target.GUID)
	}
	s.auras.mu.Unlock()
	s.auraEffectChange(target, aura, true)
	if aura.passive {
		return
	}
	s.writeAura(target, aura, true)
	s.refreshAuraUnitFlags(target)
}

func (s *WorldServer) refreshAuraUnitFlags(target realm.Character) {
	flags := uint32(unitFlagPlayer)
	s.auras.mu.Lock()
	for _, aura := range s.auras.active[target.GUID] {
		switch packet.AuraType(aura.effect.Aura) {
		case packet.AuraModStealth:
			flags |= 0x00008000
		case packet.AuraModPacify:
			flags |= 0x00020000
		case packet.AuraModDisarm:
			flags |= 0x00200000
		case packet.AuraModConfuse:
			flags |= 0x00400000
		case packet.AuraModFear:
			flags |= 0x00800000
		}
	}
	s.auras.mu.Unlock()
	s.setUnitFlags(target, flags)
}

func (s *WorldServer) writeAura(target realm.Character, aura *auraState, clear bool) {
	duration := int64(0)
	if !clear {
		duration = aura.duration
	}
	data := []byte{byte(aura.slot), 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(data[1:], uint32(duration))
	if update, err := packet.Encode(packet.SMSGUpdateAuraDuration, data); err == nil {
		s.sendPlayer(target.GUID, update)
	}
	value := int64(0)
	if !clear {
		value = aura.spellID
	}
	if update, err := packet.EncodeFieldUpdate(uint64(target.GUID), unitAuraField+aura.slot, uint32(value)); err == nil {
		s.sendSpell(target, update)
	}
	flags := s.auraFlags(target.GUID)
	for index, value := range flags {
		if update, err := packet.EncodeFieldUpdate(uint64(target.GUID), unitAuraFlagsField+index, value); err == nil {
			s.sendSpell(target, update)
		}
	}
}

func (s *WorldServer) auraFlags(guid int64) [7]uint32 {
	var flags [7]uint32
	s.auras.mu.Lock()
	for slot, aura := range s.auras.active[guid] {
		if aura.passive || slot < 0 || slot >= visibleAuraEnd {
			continue
		}
		value := uint32(packet.AuraFlagEffect0) >> aura.effectIndex
		if aura.cancelable {
			value |= uint32(packet.AuraFlagCancelable)
		}
		flags[slot/8] |= value << uint((slot&7)*4)
	}
	s.auras.mu.Unlock()
	return flags
}
