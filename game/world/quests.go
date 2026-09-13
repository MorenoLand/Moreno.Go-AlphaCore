package world

import (
	"encoding/binary"
	"sort"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const (
	questGiverNone    uint32 = 0
	questGiverTrivial uint32 = 1
	questGiverReward  uint32 = 3
	questGiverQuest   uint32 = 4
)

type questMenuEntry struct {
	Quest  worlddb.QuestTemplate
	Status uint32
}

func (s *WorldServer) questGiverAt(active realm.Character, guid uint64) (worlddb.CreatureSpawn, worlddb.CreatureTemplate, bool, error) {
	if s.WorldData == nil || guid == 0 {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, nil
	}
	spawn, found, err := s.WorldData.CreatureSpawnByID(int64(uint32(guid)))
	if err != nil || !found || spawn.Map != active.Map {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, err
	}
	dx, dy, dz := spawn.PositionX-active.PositionX, spawn.PositionY-active.PositionY, spawn.PositionZ-active.PositionZ
	if dx*dx+dy*dy+dz*dz > maxShopDistance*maxShopDistance {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, nil
	}
	creature, found, err := s.WorldData.CreatureTemplate(spawn.Entry)
	if err != nil || !found {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, err
	}
	return spawn, creature, true, nil
}

func (s *WorldServer) questGiverStatus(active realm.Character, data []byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	entry, gameObject, found, err := s.questGiverEntry(active, guid)
	if err != nil || !found {
		return nil, err
	}
	status := questGiverNone
	starters, err := s.questRelations(entry, gameObject, false)
	if err != nil {
		return nil, err
	}
	finishers, err := s.questRelations(entry, gameObject, true)
	if err != nil {
		return nil, err
	}
	if len(starters) > 0 {
		status = questGiverQuest
	}
	if s.Characters != nil {
		for _, relation := range finishers {
			state, found, err := s.Characters.QuestState(active.GUID, relation.Quest)
			if err != nil {
				return nil, err
			}
			if found && (state.Rewarded || state.State == 3) {
				status = questGiverReward
				break
			}
			if found && state.State == 2 {
				status = questGiverQuest
			}
		}
	}
	body := append(encodeGUID(int64(guid)), encodeUint32(int64(status))...)
	return packet.Encode(packet.SMSGQuestGiverStatus, body)
}

func (s *WorldServer) questGiverHello(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	entry, gameObject, found, err := s.questGiverEntry(active, guid)
	if err != nil || !found {
		return nil, err
	}
	starters, err := s.questRelations(entry, gameObject, false)
	if err != nil {
		return nil, err
	}
	finishers, err := s.questRelations(entry, gameObject, true)
	if err != nil {
		return nil, err
	}
	entries := make(map[int64]questMenuEntry, len(starters)+len(finishers))
	for _, relation := range starters {
		quest, found, err := s.WorldData.QuestTemplate(relation.Quest)
		if err != nil {
			return nil, err
		}
		if found {
			entries[quest.Entry] = questMenuEntry{Quest: quest, Status: questGiverNone}
		}
	}
	if s.Characters != nil {
		for _, relation := range finishers {
			state, found, err := s.Characters.QuestState(active.GUID, relation.Quest)
			if err != nil {
				return nil, err
			}
			if !found {
				continue
			}
			quest, found, err := s.WorldData.QuestTemplate(relation.Quest)
			if err != nil {
				return nil, err
			}
			if found && state.State == 2 {
				entries[quest.Entry] = questMenuEntry{Quest: quest, Status: questGiverQuest}
			} else if found && (state.Rewarded || state.State == 3) {
				entries[quest.Entry] = questMenuEntry{Quest: quest, Status: questGiverReward}
			}
		}
	}
	keys := make([]int64, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	greeting, found, err := s.WorldData.QuestGreeting(entry)
	if err != nil {
		return nil, err
	}
	if len(keys) == 1 && (!found || greeting.Content == "") {
		return s.questDetailsResponses(guid, entries[keys[0]].Quest)
	}
	menu := make([]questMenuEntry, 0, len(keys))
	for _, key := range keys {
		menu = append(menu, entries[key])
	}
	return [][]byte{questListPacket(guid, greeting.Content, greeting.EmoteID, menu)}, nil
}

func (s *WorldServer) questGiverQuery(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 12 {
		return nil, nil
	}
	guid, questID := binary.LittleEndian.Uint64(data), int64(binary.LittleEndian.Uint32(data[8:]))
	entry, gameObject, found, err := s.questGiverEntry(active, guid)
	if err != nil || !found {
		return nil, err
	}
	starters, err := s.questRelations(entry, gameObject, false)
	if err != nil {
		return nil, err
	}
	finishers, err := s.questRelations(entry, gameObject, true)
	if err != nil {
		return nil, err
	}
	related := false
	for _, relation := range append(starters, finishers...) {
		if relation.Quest == questID {
			related = true
			break
		}
	}
	if !related {
		return nil, nil
	}
	quest, found, err := s.WorldData.QuestTemplate(questID)
	if err != nil || !found {
		return nil, err
	}
	return s.questDetailsResponses(guid, quest)
}

func (s *WorldServer) questDetailsResponses(guid uint64, quest worlddb.QuestTemplate) ([][]byte, error) {
	body, items, err := s.questDetailsData(guid, quest)
	if err != nil {
		return nil, err
	}
	details, err := packet.Encode(packet.SMSGQuestGiverQuestDetails, body)
	if err != nil {
		return nil, err
	}
	queries, err := itemQueryPackets(items)
	if err != nil {
		return nil, err
	}
	return append(queries, details), nil
}

func (s *WorldServer) questDetailsData(guid uint64, quest worlddb.QuestTemplate) ([]byte, []worlddb.ItemTemplate, error) {
	title, err := packet.StringBytes(quest.Title)
	if err != nil {
		return nil, nil, err
	}
	details, err := packet.StringBytes(quest.Details)
	if err != nil {
		return nil, nil, err
	}
	objectives, err := packet.StringBytes(quest.Objectives)
	if err != nil {
		return nil, nil, err
	}
	body := append(encodeGUID(int64(guid)), encodeUint32(quest.Entry)...)
	body = append(body, title...)
	body = append(body, details...)
	body = append(body, objectives...)
	body = append(body, encodeUint32(1)...)
	choiceIDs, choiceCounts := questItems(quest.RewChoiceItemIDs[:], quest.RewChoiceItemCounts[:])
	rewardIDs, rewardCounts := questItems(quest.RewItemIDs[:], quest.RewItemCounts[:])
	items := make([]worlddb.ItemTemplate, 0, len(choiceIDs)+len(rewardIDs))
	body = append(body, encodeUint32(int64(len(choiceIDs)))...)
	body, items, err = s.appendQuestItemData(body, choiceIDs, choiceCounts, items)
	if err != nil {
		return nil, nil, err
	}
	body = append(body, encodeUint32(int64(len(rewardIDs)))...)
	body, items, err = s.appendQuestItemData(body, rewardIDs, rewardCounts, items)
	if err != nil {
		return nil, nil, err
	}
	body = append(body, encodeUint32(quest.RewOrReqMoney)...)
	body = append(body, encodeUint32(4)...)
	for index := 0; index < 4; index++ {
		body = append(body, encodeUint32(quest.DetailsEmotes[index])...)
		body = append(body, encodeUint32(quest.DetailsEmoteDelays[index])...)
	}
	return body, items, nil
}

func (s *WorldServer) appendQuestItemData(body []byte, entries, counts []int64, items []worlddb.ItemTemplate) ([]byte, []worlddb.ItemTemplate, error) {
	for index, entry := range entries {
		displayID := int64(0)
		if s.WorldData != nil {
			item, found, err := s.WorldData.ItemTemplate(entry)
			if err != nil {
				return nil, nil, err
			}
			if found {
				displayID = item.DisplayID
				items = append(items, item)
			}
		}
		body = append(body, encodeUint32(entry)...)
		body = append(body, encodeUint32(counts[index])...)
		body = append(body, encodeUint32(displayID)...)
	}
	return body, items, nil
}

func questItems(entries, counts []int64) ([]int64, []int64) {
	resultEntries, resultCounts := make([]int64, 0, len(entries)), make([]int64, 0, len(counts))
	for index, entry := range entries {
		if entry != 0 {
			resultEntries = append(resultEntries, entry)
			resultCounts = append(resultCounts, counts[index])
		}
	}
	return resultEntries, resultCounts
}

func questListPacket(guid uint64, greeting string, emote int64, entries []questMenuEntry) []byte {
	message, _ := packet.StringBytes(greeting)
	if len(message) > 256 {
		message = append(append([]byte(nil), message[:255]...), 0)
	}
	body := append(encodeGUID(int64(guid)), message...)
	body = append(body, make([]byte, 8)...)
	body = append(body, encodeUint32(emote)...)
	body = append(body, byte(len(entries)))
	for _, entry := range entries {
		body = append(body, encodeUint32(entry.Quest.Entry)...)
		body = append(body, encodeUint32(int64(entry.Status))...)
		body = append(body, encodeUint32(entry.Quest.QuestLevel)...)
		title, _ := packet.StringBytes(entry.Quest.Title)
		body = append(body, title...)
	}
	result, _ := packet.Encode(packet.SMSGQuestGiverQuestList, body)
	return result
}
