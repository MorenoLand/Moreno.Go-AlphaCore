package world

import (
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) dummySpellEffect(cast *spellCast, target realm.Character) {
	if cast == nil {
		return
	}
	switch cast.spell.ID {
	case 6245:
		s.playDummyEmote(target, 78)
	case 6655:
		s.playDummyEmote(target, 17)
	case 6236:
		s.setPlayerDisplay(target, 2279)
	}
}

func (s *WorldServer) dummyCreatureSpellEffect(cast *spellCast, target *creatureState) {
	if cast == nil || target == nil {
		return
	}
	switch cast.spell.ID {
	case 6245:
		s.playCreatureEmote(target, 78)
	case 6655:
		s.playCreatureEmote(target, 17)
	case 6236:
		s.setCreatureDisplay(target, 2279)
	}
}

func (s *WorldServer) playDummyEmote(target realm.Character, emote uint32) {
	if target.GUID == 0 {
		return
	}
	data := append(encodeUint32(int64(emote)), encodeGUID(target.GUID)...)
	if message, err := packet.Encode(packet.SMSGEmote, data); err == nil {
		s.broadcastPlayer(target, message)
	}
}

func (s *WorldServer) playCreatureEmote(target *creatureState, emote uint32) {
	data := append(encodeUint32(int64(emote)), encodeGUID(int64(target.GUID))...)
	if message, err := packet.Encode(packet.SMSGEmote, data); err == nil {
		viewer := realm.Character{GUID: int64(target.GUID), Map: target.Spawn.Map, PositionX: target.Spawn.PositionX, PositionY: target.Spawn.PositionY, PositionZ: target.Spawn.PositionZ}
		s.broadcastPlayer(viewer, message)
	}
}

func (s *WorldServer) playerDisplayID(guid int64) (int64, bool) {
	s.players.mu.RLock()
	display, found := s.players.displayIDs[guid]
	s.players.mu.RUnlock()
	return display, found
}

func (s *WorldServer) setPlayerDisplay(target realm.Character, display int64) {
	if target.GUID == 0 || display <= 0 {
		return
	}
	s.players.mu.Lock()
	if s.players.displayIDs == nil {
		s.players.displayIDs = make(map[int64]int64)
	}
	s.players.displayIDs[target.GUID] = display
	s.players.mu.Unlock()
	if update, err := packet.EncodeFieldUpdate(uint64(target.GUID), 151, uint32(display)); err == nil {
		s.broadcastPlayer(target, update)
		s.sendPlayer(target.GUID, update)
	}
}

func (s *WorldServer) setCreatureDisplay(target *creatureState, display int64) {
	if display <= 0 {
		return
	}
	s.creatures.mu.Lock()
	if state := s.creatures.active[target.GUID]; state != nil {
		state.Template.DisplayID1 = display
	}
	s.creatures.mu.Unlock()
	if update, err := packet.EncodeFieldUpdate(target.GUID, 151, uint32(display)); err == nil {
		viewer := realm.Character{GUID: int64(target.GUID), Map: target.Spawn.Map, PositionX: target.Spawn.PositionX, PositionY: target.Spawn.PositionY, PositionZ: target.Spawn.PositionZ}
		s.broadcastPlayer(viewer, update)
	}
}
