package world

import (
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func requiresComboPoints(spell dbc.Spell) bool {
	attributes := packet.SpellAttributesEx(spell.AttributesEx)
	return attributes&(packet.SpellAttributeExRequireTargetCombo|packet.SpellAttributeExRequireCombo) != 0
}

func (s *WorldServer) comboState(guid int64) (int64, uint64) {
	s.players.mu.RLock()
	points, target := s.players.comboPoints[guid], s.players.comboTarget[guid]
	s.players.mu.RUnlock()
	return points, target
}

func (s *WorldServer) validComboTarget(active realm.Character, target spellTarget) bool {
	points, comboTarget := s.comboState(active.GUID)
	selected := target.UnitGUID
	if selected == 0 || selected == uint64(active.GUID) {
		selected = comboTarget
	}
	return points > 0 && comboTarget != 0 && selected == comboTarget
}

func (s *WorldServer) addComboPoints(guid int64, target uint64, alive bool, amount int64) {
	if !alive || target == 0 || amount <= 0 {
		return
	}
	points, currentTarget := s.comboState(guid)
	if currentTarget != target {
		points = amount
	} else {
		points += amount
	}
	if points > 5 {
		points = 5
	}
	s.setComboState(guid, points, target)
}

func (s *WorldServer) removeComboPoints(guid int64) {
	_, target := s.comboState(guid)
	if target == 0 {
		return
	}
	s.setComboState(guid, 0, 0)
}

func (s *WorldServer) setComboState(guid, points int64, target uint64) {
	s.players.mu.Lock()
	if s.players.comboPoints == nil {
		s.players.comboPoints = make(map[int64]int64)
	}
	if s.players.comboTarget == nil {
		s.players.comboTarget = make(map[int64]uint64)
	}
	s.players.comboPoints[guid] = points
	s.players.comboTarget[guid] = target
	player, found := s.players.players[guid]
	s.players.mu.Unlock()
	s.sendComboField(guid, player, found, 18, uint32(target))
	s.sendComboField(guid, player, found, 19, uint32(target>>32))
	s.sendComboField(guid, player, found, 182, uint32(points))
}

func (s *WorldServer) sendComboField(guid int64, player realm.Character, found bool, field int, value uint32) {
	update, err := packet.EncodeFieldUpdate(uint64(guid), field, value)
	if err != nil {
		return
	}
	if found {
		s.sendSpell(player, update)
	} else {
		s.sendPlayer(guid, update)
	}
}
