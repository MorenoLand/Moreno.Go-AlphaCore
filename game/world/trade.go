package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const (
	tradeSlotCount              = 6
	tradePlayerBusy     uint32  = 0
	tradeProposed       uint32  = 1
	tradeInitiated      uint32  = 2
	tradeCancelled      uint32  = 3
	tradeAccepted       uint32  = 4
	tradeAlreadyTrading uint32  = 5
	tradePlayerNotFound uint32  = 6
	tradeStateChanged   uint32  = 7
	tradeComplete       uint32  = 8
	tradeTooFarAway     uint32  = 10
	tradeWrongFaction   uint32  = 11
	tradeDead           uint32  = 13
	tradePlayerIgnored  uint32  = 15
	tradeDistance       float32 = 11.111111
	tradeBound          int64   = 1
)

type tradeState struct {
	otherGUID int64
	accepted  bool
	money     int64
	items     [tradeSlotCount]int64
}

func tradeStatusPacket(status uint32, otherGUID int64) ([]byte, error) {
	body := encodeUint32(int64(status))
	if status == tradeProposed {
		body = append(body, encodeUint64(uint64(otherGUID))...)
	}
	return packet.Encode(packet.SMSGTradeStatus, body)
}

func (s *WorldServer) tradePairLocked(guid int64) (*tradeState, *tradeState, bool) {
	state := s.trades[guid]
	if state == nil {
		return nil, nil, false
	}
	other := s.trades[state.otherGUID]
	if other == nil || other.otherGUID != guid {
		delete(s.trades, guid)
		delete(s.trades, state.otherGUID)
		return nil, nil, false
	}
	return state, other, true
}

func (s *WorldServer) tradeStatusToPair(active realm.Character, status uint32) ([][]byte, error) {
	s.tradeMu.Lock()
	state, _, found := s.tradePairLocked(active.GUID)
	if found {
		otherGUID := state.otherGUID
		s.tradeMu.Unlock()
		response, err := tradeStatusPacket(status, 0)
		if err != nil {
			return nil, err
		}
		s.sendPlayer(otherGUID, response)
		return [][]byte{response}, nil
	}
	s.tradeMu.Unlock()
	return nil, nil
}

func (s *WorldServer) initiateTrade(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 {
		return nil, nil
	}
	targetGUID := int64(binary.LittleEndian.Uint64(data))
	if targetGUID <= 0 {
		return nil, nil
	}
	status := tradeProposed
	if active.Health <= 0 {
		status = tradeDead
	}
	if status == tradeProposed {
		s.tradeMu.Lock()
		_, _, found := s.tradePairLocked(active.GUID)
		s.tradeMu.Unlock()
		if found {
			status = tradeAlreadyTrading
		}
	}
	target, found := s.playerByGUID(targetGUID)
	if status == tradeProposed {
		switch {
		case !found:
			status = tradePlayerNotFound
		case target.GUID == active.GUID:
			status = tradePlayerBusy
		case target.Health <= 0:
			status = tradeDead
		case target.Map != active.Map:
			status = tradeTooFarAway
		default:
			dx, dy, dz := target.PositionX-active.PositionX, target.PositionY-active.PositionY, target.PositionZ-active.PositionZ
			if dx*dx+dy*dy+dz*dz > tradeDistance*tradeDistance {
				status = tradeTooFarAway
			}
		}
	}
	if status == tradeProposed && s.Characters != nil {
		social, err := s.Characters.Social(target.GUID)
		if err != nil {
			return nil, err
		}
		for _, entry := range social {
			if entry.OtherGUID == active.GUID && entry.Ignore {
				status = tradePlayerIgnored
				break
			}
		}
	}
	if status == tradeProposed {
		if first, second, err := s.teams(active, target); err != nil {
			return nil, err
		} else if first != 0 && second != 0 && first != second {
			status = tradeWrongFaction
		}
	}
	if status != tradeProposed {
		response, err := tradeStatusPacket(status, 0)
		if err != nil {
			return nil, err
		}
		return [][]byte{response}, nil
	}
	s.tradeMu.Lock()
	if _, _, found := s.tradePairLocked(active.GUID); found {
		s.tradeMu.Unlock()
		response, err := tradeStatusPacket(tradeAlreadyTrading, 0)
		return [][]byte{response}, err
	}
	if s.trades == nil {
		s.trades = make(map[int64]*tradeState)
	}
	s.trades[active.GUID] = &tradeState{otherGUID: target.GUID}
	s.trades[target.GUID] = &tradeState{otherGUID: active.GUID}
	s.tradeMu.Unlock()
	response, err := tradeStatusPacket(tradeProposed, target.GUID)
	if err != nil {
		return nil, err
	}
	request, err := tradeStatusPacket(tradeProposed, active.GUID)
	if err != nil {
		return nil, err
	}
	s.sendPlayer(target.GUID, request)
	return [][]byte{response}, nil
}

func (s *WorldServer) beginTrade(active realm.Character) ([][]byte, error) {
	return s.tradeStatusToPair(active, tradeInitiated)
}

func (s *WorldServer) tradeChanged(active realm.Character, state tradeState) ([][]byte, error) {
	changed, err := tradeStatusPacket(tradeStateChanged, 0)
	if err != nil {
		return nil, err
	}
	extended, err := s.tradeExtendedPacket(active.GUID, state, false)
	if err != nil {
		return nil, err
	}
	otherExtended, err := s.tradeExtendedPacket(active.GUID, state, true)
	if err != nil {
		return nil, err
	}
	s.tradeMu.Lock()
	otherGUID := state.otherGUID
	s.tradeMu.Unlock()
	s.sendPlayer(otherGUID, changed)
	s.sendPlayer(otherGUID, otherExtended)
	return [][]byte{changed, extended}, nil
}

func (s *WorldServer) tradeExtendedPacket(ownerGUID int64, state tradeState, target bool) ([]byte, error) {
	body := []byte{0}
	if target {
		body[0] = 1
	}
	body = append(body, encodeUint32(tradeSlotCount)...)
	body = append(body, encodeUint32(state.money)...)
	body = append(body, encodeUint32(0)...)
	body = append(body, encodeUint32(-1)...)
	for slot, itemGUID := range state.items {
		body = append(body, byte(slot))
		item, found, err := s.Characters.ItemByGUID(ownerGUID, itemGUID)
		if err != nil {
			return nil, err
		}
		entry, display, count, creator := int64(0), int64(0), int64(0), int64(0)
		if found {
			count, creator = item.StackCount, item.Creator
			if template, templateFound, templateErr := s.WorldData.ItemTemplate(item.ItemTemplate); templateErr != nil {
				return nil, templateErr
			} else if templateFound {
				entry, display = template.Entry, template.DisplayID
			}
		}
		for _, value := range []int64{entry, display, count, 0} {
			body = append(body, encodeUint32(value)...)
		}
		body = append(body, encodeUint64(uint64(creator))...)
	}
	return packet.Encode(packet.SMSGTradeStatusExtended, body)
}

func (s *WorldServer) setTradeItem(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 3 || s.Characters == nil {
		return nil, nil
	}
	tradeSlot, bag, slot := int(data[0]), inventoryBag(data[1]), int64(data[2])
	s.tradeMu.Lock()
	state, other, valid := s.tradePairLocked(active.GUID)
	if valid {
		if tradeSlot >= tradeSlotCount {
			s.tradeMu.Unlock()
			return s.cancelTrade(active)
		}
		if bag >= 39 || bag == 23 && slot >= 39 {
			s.tradeMu.Unlock()
			return s.cancelTrade(active)
		}
		for index, itemGUID := range state.items {
			if index != tradeSlot && itemGUID != 0 {
				if item, found, err := s.Characters.ItemAt(active.GUID, bag, slot); err != nil {
					s.tradeMu.Unlock()
					return nil, err
				} else if found && item.GUID == itemGUID {
					s.tradeMu.Unlock()
					return s.cancelTrade(active)
				}
			}
		}
	}
	s.tradeMu.Unlock()
	item, found, err := s.Characters.ItemAt(active.GUID, bag, slot)
	if err != nil || !found {
		return nil, err
	}
	s.tradeMu.Lock()
	state, other, valid = s.tradePairLocked(active.GUID)
	if !valid {
		s.tradeMu.Unlock()
		return nil, nil
	}
	for index, itemGUID := range state.items {
		if index != tradeSlot && itemGUID == item.GUID {
			s.tradeMu.Unlock()
			return s.cancelTrade(active)
		}
	}
	state.items[tradeSlot] = item.GUID
	state.accepted, other.accepted = false, false
	copyState := *state
	s.tradeMu.Unlock()
	return s.tradeChanged(active, copyState)
}

func (s *WorldServer) clearTradeItem(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 1 {
		return nil, nil
	}
	s.tradeMu.Lock()
	state, other, valid := s.tradePairLocked(active.GUID)
	if !valid {
		s.tradeMu.Unlock()
		return nil, nil
	}
	if int(data[0]) >= tradeSlotCount {
		s.tradeMu.Unlock()
		return s.cancelTrade(active)
	}
	state.items[data[0]] = 0
	state.accepted, other.accepted = false, false
	copyState := *state
	s.tradeMu.Unlock()
	return s.tradeChanged(active, copyState)
}

func (s *WorldServer) setTradeGold(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 4 {
		return nil, nil
	}
	money := int64(binary.LittleEndian.Uint32(data))
	if money > active.Money {
		money = active.Money
	}
	s.tradeMu.Lock()
	state, other, valid := s.tradePairLocked(active.GUID)
	if !valid {
		s.tradeMu.Unlock()
		return nil, nil
	}
	state.money = money
	state.accepted, other.accepted = false, false
	copyState := *state
	s.tradeMu.Unlock()
	return s.tradeChanged(active, copyState)
}

func (s *WorldServer) unacceptTrade(active realm.Character) ([][]byte, error) {
	s.tradeMu.Lock()
	state, other, valid := s.tradePairLocked(active.GUID)
	if !valid {
		s.tradeMu.Unlock()
		return nil, nil
	}
	state.accepted, other.accepted = false, false
	otherGUID := state.otherGUID
	s.tradeMu.Unlock()
	response, err := tradeStatusPacket(tradeStateChanged, 0)
	if err != nil {
		return nil, err
	}
	s.sendPlayer(otherGUID, response)
	return [][]byte{response}, nil
}

func (s *WorldServer) cancelTrade(active realm.Character) ([][]byte, error) {
	s.tradeMu.Lock()
	state, _, valid := s.tradePairLocked(active.GUID)
	if !valid {
		s.tradeMu.Unlock()
		return nil, nil
	}
	otherGUID := state.otherGUID
	delete(s.trades, active.GUID)
	delete(s.trades, otherGUID)
	s.tradeMu.Unlock()
	response, err := tradeStatusPacket(tradeCancelled, 0)
	if err != nil {
		return nil, err
	}
	s.sendPlayer(otherGUID, response)
	return [][]byte{response}, nil
}

func tradeSlots(items []realm.InventoryItem, count int) []int64 {
	occupied := make(map[int64]bool, len(items))
	for _, item := range items {
		occupied[item.Slot] = true
	}
	slots := make([]int64, 0, count)
	for slot := int64(23); slot < 39 && len(slots) < count; slot++ {
		if !occupied[slot] {
			occupied[slot] = true
			slots = append(slots, slot)
		}
	}
	return slots
}

func (s *WorldServer) tradeOwnedItems(ownerGUID int64, state tradeState) ([]realm.InventoryItem, error) {
	items := make([]realm.InventoryItem, 0, tradeSlotCount)
	seen := make(map[int64]bool)
	for _, itemGUID := range state.items {
		if itemGUID == 0 || seen[itemGUID] {
			continue
		}
		seen[itemGUID] = true
		item, found, err := s.Characters.ItemByGUID(ownerGUID, itemGUID)
		if err != nil {
			return nil, err
		}
		if !found || item.Flags&tradeBound != 0 || item.Bag >= 39 || item.Bag == 23 && item.Slot >= 39 {
			return nil, nil
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *WorldServer) tradeCreatePackets(ownerGUID int64, items []realm.InventoryItem, character realm.Character) ([][]byte, [][]byte, error) {
	queries := make([][]byte, 0)
	creates := make([][]byte, 0, len(items))
	for _, item := range items {
		template, found, err := s.WorldData.ItemTemplate(item.ItemTemplate)
		if err != nil {
			return nil, nil, err
		}
		if !found {
			continue
		}
		query, err := itemQueryPackets([]worlddb.ItemTemplate{template})
		if err != nil {
			return nil, nil, err
		}
		queries = append(queries, query...)
		create, err := packet.EncodeItemCreate(uint64(item.GUID)|0x4000000000000000, uint32(item.ItemTemplate), uint64(ownerGUID), uint64(item.Creator), uint32(item.StackCount), int32(item.Duration), encodedItemFlags(template, item.Flags), item.SpellCharges, packet.Movement{X: character.PositionX, Y: character.PositionY, Z: character.PositionZ, O: character.Orientation})
		if err != nil {
			return nil, nil, err
		}
		creates = append(creates, create)
	}
	return queries, creates, nil
}

func tradeDestroyPackets(items []realm.InventoryItem) ([][]byte, error) {
	packets := make([][]byte, 0, len(items))
	for _, item := range items {
		response, err := packet.Encode(packet.SMSGDestroyObject, encodeGUID(int64(uint64(item.GUID)|0x4000000000000000)))
		if err != nil {
			return nil, err
		}
		packets = append(packets, response)
	}
	return packets, nil
}

func (s *WorldServer) acceptTrade(active realm.Character) ([][]byte, error) {
	s.tradeMu.Lock()
	state, other, valid := s.tradePairLocked(active.GUID)
	if !valid {
		s.tradeMu.Unlock()
		response, err := tradeStatusPacket(tradeCancelled, 0)
		return [][]byte{response}, err
	}
	state.accepted = true
	if !other.accepted {
		otherGUID := state.otherGUID
		s.tradeMu.Unlock()
		response, err := tradeStatusPacket(tradeAccepted, 0)
		if err != nil {
			return nil, err
		}
		s.sendPlayer(otherGUID, response)
		return nil, nil
	}
	first, second := *state, *other
	s.tradeMu.Unlock()
	otherPlayer, found := s.playerByGUID(first.otherGUID)
	if !found || s.Characters == nil || s.WorldData == nil {
		return s.cancelTrade(active)
	}
	if active.Money+second.money > 2147483647 || otherPlayer.Money+first.money > 2147483647 || first.money > active.Money || second.money > otherPlayer.Money {
		return s.cancelTrade(active)
	}
	firstItems, err := s.tradeOwnedItems(active.GUID, first)
	if err != nil || firstItems == nil {
		return s.cancelTrade(active)
	}
	secondItems, err := s.tradeOwnedItems(otherPlayer.GUID, second)
	if err != nil || secondItems == nil {
		return s.cancelTrade(active)
	}
	activeInventory, err := s.Characters.InventoryItems(active.GUID, 23, 23, 39)
	if err != nil {
		return nil, err
	}
	otherInventory, err := s.Characters.InventoryItems(otherPlayer.GUID, 23, 23, 39)
	if err != nil {
		return nil, err
	}
	firstSlots, secondSlots := tradeSlots(activeInventory, len(secondItems)), tradeSlots(otherInventory, len(firstItems))
	if len(firstSlots) != len(secondItems) || len(secondSlots) != len(firstItems) {
		return s.cancelTrade(active)
	}
	activeItems := append([]realm.InventoryItem(nil), secondItems...)
	otherItems := append([]realm.InventoryItem(nil), firstItems...)
	for index := range secondItems {
		if err := s.Characters.TransferItem(secondItems[index].GUID, otherPlayer.GUID, active.GUID, 23, firstSlots[index]); err != nil {
			return nil, err
		}
		activeItems[index].Owner, activeItems[index].Bag, activeItems[index].Slot = active.GUID, 23, firstSlots[index]
	}
	for index := range firstItems {
		if err := s.Characters.TransferItem(firstItems[index].GUID, active.GUID, otherPlayer.GUID, 23, secondSlots[index]); err != nil {
			return nil, err
		}
		otherItems[index].Owner, otherItems[index].Bag, otherItems[index].Slot = otherPlayer.GUID, 23, secondSlots[index]
	}
	active.Money = active.Money - first.money + second.money
	otherPlayer.Money = otherPlayer.Money - second.money + first.money
	if err := s.Characters.UpdateMoney(active.GUID, active.AccountID, active.RealmID, active.Money); err != nil {
		return nil, err
	}
	if err := s.Characters.UpdateMoney(otherPlayer.GUID, otherPlayer.AccountID, otherPlayer.RealmID, otherPlayer.Money); err != nil {
		return nil, err
	}
	s.updatePlayer(active)
	s.updatePlayer(otherPlayer)
	s.tradeMu.Lock()
	delete(s.trades, active.GUID)
	delete(s.trades, otherPlayer.GUID)
	s.tradeMu.Unlock()
	queries, creates, err := s.tradeCreatePackets(active.GUID, activeItems, active)
	if err != nil {
		return nil, err
	}
	otherQueries, otherCreates, err := s.tradeCreatePackets(otherPlayer.GUID, otherItems, otherPlayer)
	if err != nil {
		return nil, err
	}
	destroys, err := tradeDestroyPackets(firstItems)
	if err != nil {
		return nil, err
	}
	otherDestroys, err := tradeDestroyPackets(secondItems)
	if err != nil {
		return nil, err
	}
	complete, err := tradeStatusPacket(tradeComplete, 0)
	if err != nil {
		return nil, err
	}
	otherResponses := make([][]byte, 0, len(otherQueries)+len(otherCreates)+len(otherDestroys)+1)
	if otherQueries != nil {
		otherResponses = append(otherResponses, otherQueries...)
	}
	otherResponses = append(otherResponses, otherCreates...)
	otherResponses = append(otherResponses, otherDestroys...)
	otherResponses = append(otherResponses, complete)
	for _, response := range otherResponses {
		s.sendPlayer(otherPlayer.GUID, response)
	}
	responses := make([][]byte, 0, len(queries)+len(creates)+len(destroys)+1)
	if queries != nil {
		responses = append(responses, queries...)
	}
	responses = append(responses, creates...)
	responses = append(responses, destroys...)
	responses = append(responses, complete)
	return responses, nil
}
