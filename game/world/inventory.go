package world

import (
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const (
	bagCantStack     byte = 18
	bagNotEquippable byte = 19
	bagItemNotFound  byte = 22
	bagItemTooFew    byte = 25
	bagSplitFailed   byte = 26
	bagLootGone      byte = 47
	bagInventoryFull byte = 48
)

func inventoryBag(value byte) int64 {
	if value == 0xff {
		return 23
	}
	return int64(value)
}

func (s *WorldServer) destroyItem(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 6 {
		return nil, nil
	}
	item, found, err := s.Characters.ItemAt(active.GUID, inventoryBag(data[0]), int64(data[1]))
	if err != nil || !found {
		return nil, err
	}
	if err := s.Characters.DeleteItem(item.GUID, active.GUID); err != nil {
		return nil, err
	}
	guid := int64(uint64(item.GUID) | 0x4000000000000000)
	response, err := packet.Encode(packet.SMSGDestroyObject, encodeGUID(guid))
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) splitItem(active realm.Character, data []byte) ([]byte, error) {
	if len(data) < 5 {
		return nil, nil
	}
	source, found, err := s.Characters.ItemAt(active.GUID, inventoryBag(data[0]), int64(data[1]))
	if err != nil {
		return nil, err
	}
	if !found {
		return s.inventoryFailure(active, bagItemNotFound, source)
	}
	count := int64(data[4])
	if count <= 0 {
		return s.inventoryFailure(active, bagSplitFailed, source)
	}
	if source.StackCount < count {
		return s.inventoryFailure(active, bagItemTooFew, source)
	}
	destBag, destSlot := inventoryBag(data[2]), int64(data[3])
	destination, found, err := s.Characters.ItemAt(active.GUID, destBag, destSlot)
	if err != nil {
		return nil, err
	}
	if found {
		if destination.ItemTemplate != source.ItemTemplate {
			return s.inventoryFailure(active, bagCantStack, source, destination)
		}
		item, templateFound, err := s.WorldData.ItemTemplate(source.ItemTemplate)
		if err != nil {
			return nil, err
		}
		if !templateFound || destination.StackCount+count > item.Stackable {
			return s.inventoryFailure(active, bagCantStack, source, destination)
		}
		if err := s.Characters.UpdateItemStack(destination.GUID, active.GUID, destination.StackCount+count); err != nil {
			return nil, err
		}
		return nil, s.Characters.UpdateItemStack(source.GUID, active.GUID, source.StackCount-count)
	}
	return nil, s.Characters.SplitItem(source, destBag, destSlot, count)
}

func (s *WorldServer) autostoreItem(active realm.Character, data []byte) ([]byte, error) {
	if len(data) < 3 {
		return nil, nil
	}
	sourceBag, sourceSlot, destBag := inventoryBag(data[0]), int64(data[1]), inventoryBag(data[2])
	source, found, err := s.Characters.ItemAt(active.GUID, sourceBag, sourceSlot)
	if err != nil || !found {
		return nil, err
	}
	start, end := int64(0), int64(16)
	if destBag == 23 {
		start, end = 23, 39
	}
	destSlot, err := s.Characters.FirstEmptySlot(active.GUID, destBag, start, end)
	if err != nil {
		return nil, err
	}
	if destSlot < 0 {
		return s.inventoryFailure(active, bagInventoryFull, source)
	}
	if sourceBag == destBag && destSlot >= source.Slot {
		return s.inventoryFailure(active, bagLootGone, source)
	}
	return nil, s.Characters.UpdateItemLocation(source.GUID, active.GUID, destBag, destSlot)
}

func (s *WorldServer) autoequipItem(active realm.Character, data []byte) ([]byte, error) {
	if len(data) < 2 {
		return nil, nil
	}
	sourceBag, sourceSlot := inventoryBag(data[0]), int64(data[1])
	source, found, err := s.Characters.ItemAt(active.GUID, sourceBag, sourceSlot)
	if err != nil || !found {
		return nil, err
	}
	template, found, err := s.WorldData.ItemTemplate(source.ItemTemplate)
	if err != nil || !found {
		return nil, err
	}
	destSlot := equipmentSlot(template.InventoryType)
	if destSlot < 0 {
		return s.inventoryFailure(active, bagNotEquippable, source)
	}
	return nil, s.Characters.SwapItems(active.GUID, sourceBag, sourceSlot, 23, destSlot)
}

func (s *WorldServer) readItem(active realm.Character, data []byte) ([]byte, error) {
	if len(data) < 2 {
		return nil, nil
	}
	item, found, err := s.Characters.ItemAt(active.GUID, inventoryBag(data[0]), int64(data[1]))
	if err != nil {
		return nil, err
	}
	if !found {
		return s.inventoryFailure(active, bagItemNotFound)
	}
	template, found, err := s.WorldData.ItemTemplate(item.ItemTemplate)
	if err != nil {
		return nil, err
	}
	if !found || template.PageText == 0 {
		return s.inventoryFailure(active, bagItemNotFound, item)
	}
	guid := encodeGUID(int64(uint64(item.GUID) | 0x4000000000000000))
	if template.PageLanguage == 0 || template.PageLanguage == nativeLanguage(active.Race) {
		return packet.Encode(packet.SMSGReadItemOK, guid)
	}
	body := append(guid, 0)
	return packet.Encode(packet.SMSGReadItemFailed, body)
}

func (s *WorldServer) swapInventory(active realm.Character, data []byte) error {
	if len(data) < 2 {
		return nil
	}
	return s.Characters.SwapItems(active.GUID, 23, int64(data[0]), 23, int64(data[1]))
}

func (s *WorldServer) swapItems(active realm.Character, data []byte) error {
	if len(data) < 4 {
		return nil
	}
	return s.Characters.SwapItems(active.GUID, inventoryBag(data[2]), int64(data[3]), inventoryBag(data[0]), int64(data[1]))
}

func (s *WorldServer) inventoryFailure(active realm.Character, code byte, items ...realm.InventoryItem) ([]byte, error) {
	body := []byte{code}
	first, second := active.GUID, active.GUID
	if len(items) > 0 && items[0].GUID > 0 {
		first = int64(uint64(items[0].GUID) | 0x4000000000000000)
	}
	if len(items) > 1 && items[1].GUID > 0 {
		second = int64(uint64(items[1].GUID) | 0x4000000000000000)
	}
	body = append(body, encodeGUID(first)...)
	body = append(body, encodeGUID(second)...)
	body = append(body, 0)
	return packet.Encode(packet.SMSGInventoryChangeFailure, body)
}

func equipmentSlot(inventoryType int64) int64 {
	slots := map[int64]int64{1: 0, 2: 1, 3: 2, 4: 3, 5: 4, 6: 5, 7: 6, 8: 7, 9: 8, 10: 9, 11: 10, 12: 12, 13: 15, 14: 16, 15: 17, 16: 14, 17: 15, 18: 19, 19: 18, 20: 4, 21: 15, 22: 16, 23: 16, 25: 17, 26: 17}
	if slot, found := slots[inventoryType]; found {
		return slot
	}
	return -1
}

func nativeLanguage(race uint8) int64 {
	return map[uint8]int64{1: 7, 2: 1, 3: 6, 4: 2, 5: 7, 6: 3, 7: 13, 8: 14}[race]
}
