package world

import (
	"encoding/binary"
	"math"
	"math/rand"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const (
	lootTypeNotAllowed   uint32 = 0
	lootTypeCorpse       uint32 = 1
	lootTypeFishing      uint32 = 3
	lootErrorDidntKill   byte   = 0
	lootErrorTooFar      byte   = 4
	lootErrorBadFacing   byte   = 5
	lootErrorLocked      byte   = 6
	lootErrorNotStanding byte   = 8
	lootErrorStunned     byte   = 9
	lootItemSource       byte   = 1
	lootGameObjectSource byte   = 2
	lootCreatureSource   byte   = 3
	lootPickpocketSource byte   = 4
	lootTypePicklock     uint32 = 2
)

type lootItem struct {
	entry, quantity int64
	claimed         bool
}

type lootKey struct {
	guid       uint64
	sourceType byte
}

type lootState struct {
	guid, sourceEntry uint64
	sourceType        byte
	money             int64
	items             []lootItem
	generated         bool
	active            map[int64]bool
}

func (s *WorldServer) lootStateLocked(guid uint64, sourceType byte, entry int64) *lootState {
	if s.loots == nil {
		s.loots = make(map[lootKey]*lootState)
	}
	key := lootKey{guid: guid, sourceType: sourceType}
	state := s.loots[key]
	if state == nil {
		state = &lootState{guid: guid, sourceType: sourceType, sourceEntry: uint64(entry), active: make(map[int64]bool)}
		s.loots[key] = state
	}
	return state
}

func lootGroupResult(templates []worlddb.LootTemplate) []worlddb.LootTemplate {
	groups := make(map[int64][]worlddb.LootTemplate)
	for _, item := range templates {
		groups[item.GroupID] = append(groups[item.GroupID], item)
	}
	result := make([]worlddb.LootTemplate, 0, len(templates))
	for groupID, items := range groups {
		rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
		equal := 0
		if groupID > 0 {
			for _, item := range items {
				if item.Chance == 0 {
					equal++
				}
			}
		}
		split := float64(0)
		if equal > 0 {
			split = 100 / float64(equal)
		}
		roll := rand.Float64() * 100
		for _, item := range items {
			chance := math.Abs(item.Chance)
			if chance == 0 {
				chance = split
			}
			if roll < chance {
				result = append(result, item)
				if groupID > 0 {
					break
				}
				roll = rand.Float64() * 100
			} else if groupID > 0 {
				roll -= chance
			} else {
				roll = rand.Float64() * 100
			}
		}
	}
	return result
}

func (s *WorldServer) generateLootTemplates(templates []worlddb.LootTemplate) ([]worlddb.LootTemplate, error) {
	selected := lootGroupResult(templates)
	result := make([]worlddb.LootTemplate, 0, len(selected))
	for _, item := range selected {
		if item.MinCountOrRef >= 0 {
			result = append(result, item)
			continue
		}
		references, err := s.WorldData.ReferenceLootTemplates(-item.MinCountOrRef)
		if err != nil {
			return nil, err
		}
		expanded, err := s.generateLootTemplates(references)
		if err != nil {
			return nil, err
		}
		result = append(result, expanded...)
	}
	return result, nil
}

func (s *WorldServer) lootTemplates(sourceType byte, entry int64, object worlddb.GameObjectTemplate, zone int64) ([]worlddb.LootTemplate, error) {
	switch sourceType {
	case lootItemSource:
		return s.WorldData.ItemLootTemplates(entry)
	case lootGameObjectSource:
		if object.Type == gameObjectTypeFishingNode {
			return s.WorldData.FishingLootTemplates(zone)
		}
		return s.WorldData.GameObjectLootTemplates(object.Data[1])
	case lootPickpocketSource:
		return s.WorldData.PickpocketLootTemplates(entry)
	default:
		creature, found, err := s.WorldData.CreatureTemplate(entry)
		if err != nil || !found || creature.LootID <= 0 {
			return nil, err
		}
		templates, err := s.WorldData.CreatureLootTemplates(creature.LootID)
		if err != nil || len(templates) == 0 || creature.SkinningLootID <= 0 {
			return templates, err
		}
		skinning, err := s.WorldData.SkinningLootTemplates(creature.SkinningLootID)
		if err != nil {
			return nil, err
		}
		maxGroup := int64(0)
		for _, item := range templates {
			if item.GroupID > maxGroup {
				maxGroup = item.GroupID
			}
		}
		for index := range skinning {
			skinning[index].GroupID = maxGroup + 1
		}
		return append(templates, skinning...), nil
	}
}

func (s *WorldServer) generateLoot(state *lootState, sourceType byte, entry int64, object worlddb.GameObjectTemplate, zone int64) error {
	templates, err := s.lootTemplates(sourceType, entry, object, zone)
	if err != nil {
		return err
	}
	selected, err := s.generateLootTemplates(templates)
	if err != nil {
		return err
	}
	state.items = make([]lootItem, 0, len(selected))
	for _, item := range selected {
		maxCount := item.MaxCount
		if maxCount < item.MinCountOrRef {
			maxCount = item.MinCountOrRef
		}
		quantity := item.MinCountOrRef
		if maxCount > quantity {
			quantity += int64(rand.Intn(int(maxCount - quantity + 1)))
		}
		state.items = append(state.items, lootItem{entry: item.Item, quantity: quantity})
	}
	if sourceType == lootCreatureSource {
		if creature, found, err := s.WorldData.CreatureTemplate(entry); err != nil {
			return err
		} else if found {
			if creature.GoldMax > creature.GoldMin {
				state.money = creature.GoldMin + int64(rand.Intn(int(creature.GoldMax-creature.GoldMin+1)))
			} else {
				state.money = creature.GoldMin
			}
		}
	} else if sourceType == lootGameObjectSource && object.MaxGold > object.MinGold {
		state.money = object.MinGold + int64(rand.Intn(int(object.MaxGold-object.MinGold+1)))
	} else if sourceType == lootGameObjectSource {
		state.money = object.MinGold
	}
	state.generated = true
	return nil
}

func (s *WorldServer) setLootSelection(guid int64, object uint64, sourceType byte) {
	s.lootMu.Lock()
	if s.lootSelections == nil {
		s.lootSelections = make(map[int64]uint64)
	}
	if s.lootSelectionTypes == nil {
		s.lootSelectionTypes = make(map[int64]byte)
	}
	s.lootSelections[guid] = object
	s.lootSelectionTypes[guid] = sourceType
	s.lootMu.Unlock()
}

func (s *WorldServer) lootSelection(guid int64) uint64 {
	s.lootMu.Lock()
	object := s.lootSelections[guid]
	s.lootMu.Unlock()
	return object
}

func (s *WorldServer) selectedLootStateLocked(guid int64) (uint64, *lootState) {
	object := s.lootSelections[guid]
	if object == 0 {
		return 0, nil
	}
	if state := s.loots[lootKey{guid: object, sourceType: s.lootSelectionTypes[guid]}]; state != nil {
		return object, state
	}
	for key, state := range s.loots {
		if key.guid == object && state.active[guid] {
			return object, state
		}
	}
	return object, nil
}

func (s *WorldServer) sendLoot(active realm.Character, guid uint64, sourceType byte, entry int64, object worlddb.GameObjectTemplate, zone int64) ([][]byte, error) {
	s.lootMu.Lock()
	state := s.lootStateLocked(guid, sourceType, entry)
	if !state.generated {
		if err := s.generateLoot(state, sourceType, entry, object, zone); err != nil {
			s.lootMu.Unlock()
			return nil, err
		}
	}
	state.active[active.GUID] = true
	items := append([]lootItem(nil), state.items...)
	money := state.money
	s.lootMu.Unlock()
	s.setLootSelection(active.GUID, guid, sourceType)
	itemData := make([]byte, 0, len(items)*17)
	templates := make([]worlddb.ItemTemplate, 0, len(items))
	count := 0
	for slot, item := range items {
		if item.claimed {
			continue
		}
		template, found, err := s.WorldData.ItemTemplate(item.entry)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		itemData = append(itemData, byte(slot))
		itemData = append(itemData, encodeUint32(item.entry)...)
		itemData = append(itemData, encodeUint32(item.quantity)...)
		itemData = append(itemData, encodeUint32(template.DisplayID)...)
		templates = append(templates, template)
		count++
	}
	lootType := lootTypeCorpse
	if sourceType == lootPickpocketSource {
		lootType = lootTypePicklock
	}
	body := append(encodeUint64(guid), encodeUint32(int64(lootType))...)
	if sourceType == lootGameObjectSource && object.Type == gameObjectTypeFishingNode {
		binary.LittleEndian.PutUint32(body[8:], lootTypeFishing)
	}
	body = append(body, encodeUint32(money)...)
	body = append(body, byte(count))
	body = append(body, itemData...)
	response, err := packet.Encode(packet.SMSGLootResponse, body)
	if err != nil {
		return nil, err
	}
	queries, err := itemQueryPackets(templates)
	if err != nil {
		return nil, err
	}
	return append(queries, response), nil
}

func lootErrorPacket(guid uint64, errorCode byte) ([]byte, error) {
	body := append(encodeUint64(guid), byte(lootTypeNotAllowed), errorCode)
	return packet.Encode(packet.SMSGLootResponse, body)
}

func (s *WorldServer) lootRequest(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 || active.Health <= 0 || s.WorldData == nil {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	high := guid & 0xffff000000000000
	if high == 0x4000000000000000 {
		item, found, err := s.Characters.ItemByGUID(active.GUID, petitionItemGUID(guid))
		if err != nil || !found {
			response, responseErr := lootErrorPacket(guid, lootErrorDidntKill)
			return [][]byte{response}, responseErr
		}
		template, found, err := s.WorldData.ItemTemplate(item.ItemTemplate)
		if err != nil || !found || template.Flags&4 == 0 {
			response, responseErr := lootErrorPacket(guid, lootErrorDidntKill)
			return [][]byte{response}, responseErr
		}
		return s.sendLoot(active, guid, lootItemSource, item.ItemTemplate, worlddb.GameObjectTemplate{}, 0)
	}
	if high == 0xf110000000000000 {
		spawn, object, found, err := s.gameObjectAt(active, guid, gameObjectViewDistance)
		if err != nil || !found {
			response, responseErr := lootErrorPacket(guid, lootErrorDidntKill)
			return [][]byte{response}, responseErr
		}
		dx, dy, dz := spawn.PositionX-active.PositionX, spawn.PositionY-active.PositionY, spawn.PositionZ-active.PositionZ
		if dx*dx+dy*dy+dz*dz > 25 {
			response, responseErr := lootErrorPacket(guid, lootErrorTooFar)
			return [][]byte{response}, responseErr
		}
		return s.sendLoot(active, guid, lootGameObjectSource, object.Entry, object, active.Zone)
	}
	if high == 0xf130000000000000 {
		spawn, creature, found, err := s.creatureAt(active, guid, creatureViewDistance)
		if err != nil || !found {
			response, responseErr := lootErrorPacket(guid, lootErrorDidntKill)
			return [][]byte{response}, responseErr
		}
		dx, dy, dz := spawn.PositionX-active.PositionX, spawn.PositionY-active.PositionY, spawn.PositionZ-active.PositionZ
		if dx*dx+dy*dy+dz*dz > 25 {
			response, responseErr := lootErrorPacket(guid, lootErrorTooFar)
			return [][]byte{response}, responseErr
		}
		return s.sendLoot(active, guid, lootCreatureSource, creature.Entry, worlddb.GameObjectTemplate{}, 0)
	}
	response, err := lootErrorPacket(guid, lootErrorDidntKill)
	return [][]byte{response}, err
}

func (s *WorldServer) lootItem(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 1 {
		return nil, nil
	}
	s.lootMu.Lock()
	_, state := s.selectedLootStateLocked(active.GUID)
	if state == nil || !state.active[active.GUID] || int(data[0]) >= len(state.items) || state.items[data[0]].claimed {
		s.lootMu.Unlock()
		return nil, nil
	}
	item := state.items[data[0]]
	s.lootMu.Unlock()
	template, found, err := s.WorldData.ItemTemplate(item.entry)
	if err != nil || !found {
		return nil, err
	}
	slot, err := s.Characters.FirstEmptySlot(active.GUID, 23, 23, 39)
	if err != nil {
		return nil, err
	}
	if slot < 0 {
		return inventoryResponse(active, bagInventoryFull)
	}
	created, err := s.Characters.CreateInventoryItem(active.GUID, 0, 23, slot, item.entry, item.quantity)
	if err != nil {
		return nil, err
	}
	s.lootMu.Lock()
	if state.items[data[0]].claimed {
		s.lootMu.Unlock()
		return nil, nil
	}
	state.items[data[0]].claimed = true
	s.lootMu.Unlock()
	progress, err := s.questItemProgress(active, item.entry, item.quantity)
	if err != nil {
		return nil, err
	}
	create, err := packet.EncodeItemCreate(uint64(created.GUID)|0x4000000000000000, uint32(created.ItemTemplate), uint64(active.GUID), 0, uint32(created.StackCount), 0, encodedItemFlags(template, created.Flags), created.SpellCharges, packet.Movement{X: active.PositionX, Y: active.PositionY, Z: active.PositionZ, O: active.Orientation})
	if err != nil {
		return nil, err
	}
	push, err := itemPushResult(created, item.entry, 23)
	if err != nil {
		return nil, err
	}
	removed, err := packet.Encode(packet.SMSGLootRemoved, []byte{data[0]})
	if err != nil {
		return nil, err
	}
	queries, err := itemQueryPackets([]worlddb.ItemTemplate{template})
	if err != nil {
		return nil, err
	}
	return append(append(append(queries, create, push), removed), progress...), nil
}

func (s *WorldServer) lootMoney(active *realm.Character) ([][]byte, error) {
	if active == nil || s.Characters == nil {
		return nil, nil
	}
	s.lootMu.Lock()
	_, state := s.selectedLootStateLocked(active.GUID)
	if state == nil || !state.active[active.GUID] || state.money <= 0 {
		s.lootMu.Unlock()
		return nil, nil
	}
	money := state.money
	state.money = 0
	s.lootMu.Unlock()
	active.Money += money
	if active.Money > 2147483647 {
		active.Money = 2147483647
	}
	if err := s.Characters.UpdateMoney(active.GUID, active.AccountID, active.RealmID, active.Money); err != nil {
		return nil, err
	}
	s.updatePlayer(*active)
	moneyUpdate, err := packet.EncodeFieldUpdate(uint64(active.GUID), 55, uint32(active.Money))
	if err != nil {
		return nil, err
	}
	clear, err := packet.Encode(packet.SMSGLootClearMoney, nil)
	if err != nil {
		return nil, err
	}
	return [][]byte{moneyUpdate, clear}, nil
}

func (s *WorldServer) lootRelease(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	s.lootMu.Lock()
	if s.lootSelections[active.GUID] != guid {
		s.lootMu.Unlock()
		return nil, nil
	}
	_, state := s.selectedLootStateLocked(active.GUID)
	noLoot := state != nil && !stateHasLoot(state)
	if state != nil {
		delete(state.active, active.GUID)
	}
	delete(s.lootSelections, active.GUID)
	delete(s.lootSelectionTypes, active.GUID)
	s.lootMu.Unlock()
	response, err := packet.Encode(packet.SMSGLootReleaseResponse, append(encodeUint64(guid), 1))
	if err != nil {
		return nil, err
	}
	responses := [][]byte{response}
	if state != nil && state.sourceType == lootGameObjectSource {
		gameObject := s.gameObjectStateFor(guid, worlddb.GameObjectSpawn{})
		gameObject.flags &^= gameObjectFlagInUse
		gameObject.state = gameObjectStateReady
		s.setGameObjectState(guid, gameObject)
		flags, err := gameObjectUpdate(guid, 7, uint32(gameObject.flags))
		if err != nil {
			return nil, err
		}
		stateUpdate, err := gameObjectUpdate(guid, 12, uint32(gameObject.state))
		if err != nil {
			return nil, err
		}
		s.broadcastPlayer(active, flags)
		s.broadcastPlayer(active, stateUpdate)
		responses = append(responses, flags, stateUpdate)
	}
	if state != nil && state.sourceType == lootItemSource && noLoot {
		if err := s.Characters.DeleteItem(int64(state.guid&0x3fffffffffffffff), active.GUID); err != nil {
			return nil, err
		}
	}
	return responses, nil
}

func stateHasLoot(state *lootState) bool {
	if state.money > 0 {
		return true
	}
	for _, item := range state.items {
		if !item.claimed {
			return true
		}
	}
	return false
}
