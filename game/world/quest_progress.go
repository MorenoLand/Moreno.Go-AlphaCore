package world

import (
	"encoding/binary"
	"fmt"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const (
	questAccepted int64 = 2
	questReward   int64 = 3
)

func (s *WorldServer) questAccept(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 12 || s.Characters == nil {
		return nil, nil
	}
	giverGUID, questID := binary.LittleEndian.Uint64(data), int64(binary.LittleEndian.Uint32(data[8:]))
	entry, gameObject, found, err := s.questGiverEntry(active, giverGUID)
	if err != nil || !found {
		return nil, err
	}
	if !s.questRelated(entry, questID, gameObject, false) {
		return nil, nil
	}
	quest, found, err := s.WorldData.QuestTemplate(questID)
	if err != nil || !found {
		return nil, err
	}
	if quest.MinLevel > 0 && int64(active.Level) < quest.MinLevel {
		return s.questInvalid(1)
	}
	state, found, err := s.Characters.QuestState(active.GUID, questID)
	if err != nil {
		return nil, err
	}
	if found && !state.Rewarded {
		return s.questInvalid(13)
	}
	if err := s.Characters.SaveQuestState(realm.QuestState{GUID: active.GUID, Quest: questID, State: questAccepted}); err != nil {
		return nil, err
	}
	return s.questQuery(encodeUint32(questID))
}

func (s *WorldServer) questComplete(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 12 || s.Characters == nil {
		return nil, nil
	}
	giverGUID, questID := binary.LittleEndian.Uint64(data), int64(binary.LittleEndian.Uint32(data[8:]))
	entry, gameObject, found, err := s.questGiverEntry(active, giverGUID)
	if err != nil || !found {
		return nil, err
	}
	if !s.questRelated(entry, questID, gameObject, true) {
		return nil, nil
	}
	quest, found, err := s.WorldData.QuestTemplate(questID)
	if err != nil || !found {
		return nil, err
	}
	state, found, err := s.Characters.QuestState(active.GUID, questID)
	if err != nil {
		return nil, err
	}
	if !found {
		return s.questDetailsResponses(giverGUID, quest)
	}
	if state.State != questAccepted && state.State != questReward {
		return s.questDetailsResponses(giverGUID, quest)
	}
	complete, err := s.questRequirementsMet(active, quest)
	if err != nil {
		return nil, err
	}
	if !complete {
		return s.questRequestResponses(giverGUID, quest, false)
	}
	state.State = questReward
	if err := s.Characters.SaveQuestState(state); err != nil {
		return nil, err
	}
	return s.questOfferResponses(giverGUID, quest)
}

func (s *WorldServer) questRequestReward(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 12 || s.Characters == nil {
		return nil, nil
	}
	giverGUID, questID := binary.LittleEndian.Uint64(data), int64(binary.LittleEndian.Uint32(data[8:]))
	entry, gameObject, found, err := s.questGiverEntry(active, giverGUID)
	if err != nil || !found || !s.questRelated(entry, questID, gameObject, true) {
		return nil, err
	}
	quest, found, err := s.WorldData.QuestTemplate(questID)
	if err != nil || !found {
		return nil, err
	}
	state, found, err := s.Characters.QuestState(active.GUID, questID)
	if err != nil || !found {
		return nil, err
	}
	if state.State == questAccepted {
		complete, err := s.questRequirementsMet(active, quest)
		if err != nil || !complete {
			return nil, err
		}
		state.State = questReward
		if err := s.Characters.SaveQuestState(state); err != nil {
			return nil, err
		}
	}
	if state.State != questReward {
		return nil, nil
	}
	return s.questOfferResponses(giverGUID, quest)
}

func (s *WorldServer) questChooseReward(active *realm.Character, data []byte) ([][]byte, error) {
	if active == nil || len(data) < 16 || s.Characters == nil {
		return nil, nil
	}
	giverGUID, questID := binary.LittleEndian.Uint64(data), int64(binary.LittleEndian.Uint32(data[8:]))
	choice := int(binary.LittleEndian.Uint32(data[12:]))
	quest, found, err := s.WorldData.QuestTemplate(questID)
	if err != nil || !found {
		return nil, err
	}
	state, found, err := s.Characters.QuestState(active.GUID, questID)
	if err != nil || !found || state.State != questReward || state.Rewarded {
		return nil, err
	}
	entry, gameObject, giverFound, err := s.questGiverEntry(*active, giverGUID)
	if err != nil || !giverFound || !s.questRelated(entry, questID, gameObject, true) {
		return nil, err
	}
	choices, choiceCounts := questItems(quest.RewChoiceItemIDs[:], quest.RewChoiceItemCounts[:])
	rewards, rewardCounts := questItems(quest.RewItemIDs[:], quest.RewItemCounts[:])
	if choice < 0 || choice >= len(choices) && len(choices) > 0 || len(choices) == 0 && choice != 0 {
		return nil, nil
	}
	if len(choices) > 0 {
		rewards = append([]int64{choices[choice]}, rewards...)
		rewardCounts = append([]int64{choiceCounts[choice]}, rewardCounts...)
	}
	for index, entry := range rewards {
		if err := s.addQuestReward(*active, entry, rewardCounts[index]); err != nil {
			return nil, err
		}
	}
	active.Money += quest.RewOrReqMoney
	if active.Money > 2147483647 {
		active.Money = 2147483647
	}
	if err := s.Characters.UpdateMoney(active.GUID, active.AccountID, active.RealmID, active.Money); err != nil {
		return nil, err
	}
	state.Rewarded = true
	if err := s.Characters.SaveQuestState(state); err != nil {
		return nil, err
	}
	s.updatePlayer(*active)
	body := append(encodeUint32(questID), make([]byte, 12)...)
	binary.LittleEndian.PutUint32(body[8:], uint32(quest.RewXP))
	binary.LittleEndian.PutUint32(body[12:], uint32(maxInt64(quest.RewOrReqMoney, -quest.RewOrReqMoney)))
	body = append(body, encodeUint32(int64(len(rewards)))...)
	for index, entry := range rewards {
		body = append(body, encodeUint32(entry)...)
		body = append(body, encodeUint32(rewardCounts[index])...)
	}
	complete, err := packet.Encode(packet.SMSGQuestGiverQuestComplete, body)
	if err != nil {
		return nil, err
	}
	money, err := packet.EncodeFieldUpdate(uint64(active.GUID), 55, uint32(active.Money))
	if err != nil {
		return nil, err
	}
	return [][]byte{complete, money}, nil
}

func (s *WorldServer) questRemove(active realm.Character, data []byte) error {
	if active == (realm.Character{}) || len(data) < 1 || s.Characters == nil {
		return nil
	}
	states, err := s.Characters.QuestStates(active.GUID)
	if err != nil || int(data[0]) >= len(states) {
		return err
	}
	return s.Characters.DeleteQuestState(active.GUID, states[data[0]].Quest)
}

func (s *WorldServer) questConfirmAccept(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 4 || s.Characters == nil {
		return nil, nil
	}
	questID := int64(binary.LittleEndian.Uint32(data))
	quest, found, err := s.WorldData.QuestTemplate(questID)
	if err != nil || !found {
		return nil, err
	}
	if _, found, err := s.Characters.QuestState(active.GUID, questID); err != nil || found {
		return nil, err
	}
	if err := s.Characters.SaveQuestState(realm.QuestState{GUID: active.GUID, Quest: questID, State: questAccepted}); err != nil {
		return nil, err
	}
	return s.questQuery(encodeUint32(quest.Entry))
}

func (s *WorldServer) questRequirementsMet(active realm.Character, quest worlddb.QuestTemplate) (bool, error) {
	for index, entry := range quest.ReqItemIDs {
		if entry == 0 {
			continue
		}
		count, err := s.Characters.ItemCount(active.GUID, entry)
		if err != nil || count < quest.ReqItemCounts[index] {
			return false, err
		}
	}
	return true, nil
}

func (s *WorldServer) addQuestReward(active realm.Character, entry, count int64) error {
	if count <= 0 {
		return nil
	}
	template, found, err := s.WorldData.ItemTemplate(entry)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("quest reward item %d not found", entry)
	}
	if template.Stackable > 1 {
		items, err := s.Characters.InventoryItems(active.GUID, 23, 23, 39)
		if err != nil {
			return err
		}
		for _, item := range items {
			if item.ItemTemplate == entry && item.StackCount+count <= template.Stackable {
				return s.Characters.UpdateItemStack(item.GUID, active.GUID, item.StackCount+count)
			}
		}
	}
	slot, err := s.Characters.FirstEmptySlot(active.GUID, 23, 23, 39)
	if err != nil {
		return err
	}
	if slot < 0 {
		return fmt.Errorf("inventory full for quest reward %d", entry)
	}
	_, err = s.Characters.CreateInventoryItem(active.GUID, 0, 23, slot, entry, count)
	return err
}

func (s *WorldServer) questRelated(entry, quest int64, gameObject, finisher bool) bool {
	relations, err := s.questRelations(entry, gameObject, finisher)
	if err != nil {
		return false
	}
	for _, relation := range relations {
		if relation.Quest == quest {
			return true
		}
	}
	return false
}

func (s *WorldServer) questInvalid(reason uint32) ([][]byte, error) {
	response, err := packet.Encode(packet.SMSGQuestGiverQuestInvalid, encodeUint32(int64(reason)))
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) questRequestResponses(guid uint64, quest worlddb.QuestTemplate, complete bool) ([][]byte, error) {
	title, err := packet.StringBytes(quest.Title)
	if err != nil {
		return nil, err
	}
	text, err := packet.StringBytes(quest.RequestItemsText)
	if err != nil {
		return nil, err
	}
	body := append(encodeGUID(int64(guid)), encodeUint32(quest.Entry)...)
	body = append(body, title...)
	body = append(body, text...)
	body = append(body, encodeUint32(0)...)
	emote := quest.IncompleteEmote
	if complete {
		emote = quest.CompleteEmote
	}
	body = append(body, encodeUint32(emote)...)
	body = append(body, encodeUint32(0)...)
	entries, counts := questItems(quest.ReqItemIDs[:], quest.ReqItemCounts[:])
	body = append(body, encodeUint32(int64(len(entries)))...)
	items := make([]worlddb.ItemTemplate, 0, len(entries))
	body, items, err = s.appendQuestItemData(body, entries, counts, items)
	if err != nil {
		return nil, err
	}
	body = append(body, encodeUint32(1)...)
	body = append(body, encodeUint32(int64(boolUint32(complete)))...)
	body = append(body, encodeUint32(1)...)
	response, err := packet.Encode(packet.SMSGQuestGiverRequestItems, body)
	if err != nil {
		return nil, err
	}
	queries, err := itemQueryPackets(items)
	if err != nil {
		return nil, err
	}
	return append(queries, response), nil
}

func (s *WorldServer) questOfferResponses(guid uint64, quest worlddb.QuestTemplate) ([][]byte, error) {
	title, err := packet.StringBytes(quest.Title)
	if err != nil {
		return nil, err
	}
	text, err := packet.StringBytes(quest.OfferRewardText)
	if err != nil {
		return nil, err
	}
	body := append(encodeGUID(int64(guid)), encodeUint32(quest.Entry)...)
	body = append(body, title...)
	body = append(body, text...)
	body = append(body, encodeUint32(1)...)
	body = append(body, encodeUint32(4)...)
	for index := range quest.OfferRewardEmotes {
		body = append(body, encodeUint32(quest.OfferRewardEmoteDelays[index])...)
		body = append(body, encodeUint32(quest.OfferRewardEmotes[index])...)
	}
	choices, choiceCounts := questItems(quest.RewChoiceItemIDs[:], quest.RewChoiceItemCounts[:])
	rewards, rewardCounts := questItems(quest.RewItemIDs[:], quest.RewItemCounts[:])
	body = append(body, encodeUint32(int64(len(choices)))...)
	items := make([]worlddb.ItemTemplate, 0, len(choices)+len(rewards))
	body, items, err = s.appendQuestItemData(body, choices, choiceCounts, items)
	if err != nil {
		return nil, err
	}
	body = append(body, encodeUint32(int64(len(rewards)))...)
	body, items, err = s.appendQuestItemData(body, rewards, rewardCounts, items)
	if err != nil {
		return nil, err
	}
	body = append(body, encodeUint32(maxInt64(quest.RewOrReqMoney, -quest.RewOrReqMoney))...)
	response, err := packet.Encode(packet.SMSGQuestGiverOfferReward, body)
	if err != nil {
		return nil, err
	}
	queries, err := itemQueryPackets(items)
	if err != nil {
		return nil, err
	}
	return append(queries, response), nil
}

func boolUint32(value bool) uint32 {
	if value {
		return 1
	}
	return 0
}

func maxInt64(first, second int64) int64 {
	if first > second {
		return first
	}
	return second
}
