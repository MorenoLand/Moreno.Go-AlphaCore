package world

import (
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) completeSpellQuest(cast *spellCast, target realm.Character, effect dbc.SpellEffect) {
	player := target
	if player.GUID == 0 {
		player = cast.caster
	}
	if s.Characters == nil || s.WorldData == nil || player.GUID == 0 {
		return
	}
	quest, found, err := s.WorldData.QuestTemplate(effect.MiscValue)
	if err != nil || !found {
		return
	}
	state, found, err := s.Characters.QuestState(player.GUID, quest.Entry)
	if err != nil || !found || state.Rewarded || (state.State != questAccepted && state.State != questReward) {
		return
	}
	state.State = questReward
	state.Explored = true
	if err := s.Characters.SaveQuestState(state); err != nil {
		return
	}
	if update, err := packet.Encode(packet.SMSGQuestUpdateComplete, encodeUint32(quest.Entry)); err == nil {
		s.sendPlayer(player.GUID, update)
	}
}
