package world

import (
	"time"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) dispelSpellAuras(cast *spellCast, target realm.Character, count int64) {
	if count <= 0 {
		return
	}
	friendly := true
	if target.GUID != cast.caster.GUID {
		team, targetTeam, err := s.teams(cast.caster, target)
		if err == nil && team != 0 && targetTeam != 0 {
			friendly = team == targetTeam
		}
	}
	s.auras.mu.Lock()
	remove := make([]int, 0)
	for slot, aura := range s.auras.active[target.GUID] {
		if aura.harmful != friendly {
			continue
		}
		remove = append(remove, slot)
		if int64(len(remove)) == count {
			break
		}
	}
	s.auras.mu.Unlock()
	for _, slot := range remove {
		s.removeAura(target, slot)
	}
}

func (s *WorldServer) interruptSpellTarget(target realm.Character, penalty int64) {
	if target.Health <= 0 {
		return
	}
	s.spells.mu.Lock()
	cast := s.spells.casts[target.GUID]
	if cast == nil {
		s.spells.mu.Unlock()
		return
	}
	delete(s.spells.casts, target.GUID)
	if cast.timer != nil {
		cast.timer.Stop()
	}
	s.spells.mu.Unlock()
	result, err := spellCastResult(cast.spell.ID, packet.SpellFailedInterrupted)
	if err == nil {
		s.sendPlayer(target.GUID, result)
	}
	failure, err := packet.Encode(packet.SMSGSpellFailure, append(encodeGUID(target.GUID), append(encodeUint32(cast.spell.ID), byte(packet.SpellFailedInterrupted))...))
	if err == nil {
		s.sendPlayer(target.GUID, failure)
	}
	if penalty > 0 {
		s.setSpellCooldownPenalty(target, cast.spell.ID, penalty)
	}
}

func (s *WorldServer) setSpellCooldownPenalty(target realm.Character, spellID, penalty int64) {
	expires := time.Now().Add(time.Duration(penalty) * time.Millisecond)
	s.spells.mu.Lock()
	if s.spells.cooldowns == nil {
		s.spells.cooldowns = make(map[int64]map[int64]time.Time)
	}
	if s.spells.cooldowns[target.GUID] == nil {
		s.spells.cooldowns[target.GUID] = make(map[int64]time.Time)
	}
	s.spells.cooldowns[target.GUID][spellID] = expires
	s.spells.mu.Unlock()
	data := append(encodeUint32(spellID), encodeGUID(target.GUID)...)
	data = append(data, encodeUint16(penalty)...)
	if update, err := packet.Encode(packet.SMSGSpellCooldown, data); err == nil {
		s.sendPlayer(target.GUID, update)
	}
	time.AfterFunc(time.Duration(penalty)*time.Millisecond, func() {
		s.spells.mu.Lock()
		current, found := s.spells.cooldowns[target.GUID][spellID]
		if found && current.Equal(expires) {
			delete(s.spells.cooldowns[target.GUID], spellID)
		}
		s.spells.mu.Unlock()
		if found && current.Equal(expires) {
			clear := append(encodeUint32(spellID), encodeGUID(target.GUID)...)
			if update, err := packet.Encode(packet.SMSGClearCooldown, clear); err == nil {
				s.sendPlayer(target.GUID, update)
			}
		}
	})
}

func (s *WorldServer) spellDurationMillis(cast *spellCast) int64 {
	if s.DBC == nil || cast.spell.DurationIndex <= 0 {
		return 0
	}
	duration, found, err := s.DBC.SpellDuration(cast.spell.DurationIndex)
	if err != nil || !found {
		return 0
	}
	level := cast.effectLevel
	if !cast.ranked {
		level = int64(cast.caster.Level) - cast.spell.BaseLevel
	}
	if level < 0 {
		level = 0
	}
	value := duration.Duration + duration.DurationPerLevel*level
	if duration.MaxDuration > 0 && value > duration.MaxDuration {
		value = duration.MaxDuration
	}
	return value
}

func (s *WorldServer) teleportSpellTarget(cast *spellCast, target realm.Character) {
	if cast.target.Dest == nil {
		return
	}
	teleport := target
	response, err := s.teleportPlayer(&teleport, cast.caster.Map, cast.target.Dest.X, cast.target.Dest.Y, cast.target.Dest.Z, target.Orientation)
	if err == nil {
		s.sendPlayer(target.GUID, response)
	}
}

func (s *WorldServer) summonSpellTarget(cast *spellCast, target realm.Character) {
	if target.GUID == 0 || target.GUID == cast.caster.GUID {
		return
	}
	if _, found := s.playerByGUID(target.GUID); !found {
		return
	}
	teleport := target
	response, err := s.teleportPlayer(&teleport, cast.caster.Map, cast.caster.PositionX, cast.caster.PositionY, cast.caster.PositionZ, cast.caster.Orientation)
	if err == nil {
		s.sendPlayer(target.GUID, response)
	}
}
