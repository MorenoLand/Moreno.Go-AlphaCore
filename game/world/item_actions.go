package world

import (
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const (
	itemDynBound       int64 = 1
	itemDynUnlocked    int64 = 4
	itemDynWrapped     int64 = 8
	itemTypeBag        int64 = 18
	itemFieldEntry           = 3
	itemFieldCreator         = 10
	itemFieldCreatorHi       = 11
	itemFieldStack           = 12
	itemFieldFlag            = 19
)

var giftEntries = map[int64]int64{5014: 5015, 5042: 5043, 5047: 5045, 5048: 5044, 5049: 5046}

func encodedItemFlags(template worlddb.ItemTemplate, dynamic int64) uint32 {
	return uint32(dynamic&0xffff) | uint32(template.Flags&0xffff)<<16
}

func itemUpdates(itemGUID uint64, template worlddb.ItemTemplate, creator, dynamic int64) ([][]byte, error) {
	guid := itemGUID | 0x4000000000000000
	updates := []struct {
		field int
		value uint32
	}{{itemFieldEntry, uint32(template.Entry)}, {itemFieldCreator, uint32(uint64(creator))}, {itemFieldCreatorHi, uint32(uint64(creator) >> 32)}, {itemFieldFlag, encodedItemFlags(template, dynamic)}}
	packets := make([][]byte, 0, len(updates))
	for _, update := range updates {
		response, err := packet.EncodeFieldUpdate(guid, update.field, update.value)
		if err != nil {
			return nil, err
		}
		packets = append(packets, response)
	}
	return packets, nil
}

func (s *WorldServer) openItem(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 2 || s.Characters == nil || s.WorldData == nil {
		return nil, nil
	}
	item, found, err := s.Characters.ItemAt(active.GUID, inventoryBag(data[0]), int64(data[1]))
	if err != nil {
		return nil, err
	}
	if !found {
		return inventoryResponse(active, bagItemNotFound)
	}
	if active.Health <= 0 {
		return inventoryResponse(active, bagNotWhileDead)
	}
	template, found, err := s.WorldData.ItemTemplate(item.ItemTemplate)
	if err != nil {
		return nil, err
	}
	if !found {
		return inventoryResponse(active, bagUnknownItem, item)
	}
	if item.Flags&itemDynUnlocked == 0 && template.LockID > 0 {
		if s.DBC == nil {
			return inventoryResponse(active, bagItemLocked, item)
		}
		lock, lockFound, err := s.DBC.LockByID(template.LockID)
		if err != nil {
			return nil, err
		}
		if !lockFound || lock.Skill1 != 0 || lock.Skill2 != 0 || lock.Skill3 != 0 || lock.Skill4 != 0 {
			return inventoryResponse(active, bagItemLocked, item)
		}
	}
	if item.Flags&itemDynWrapped == 0 {
		if template.Flags&4 != 0 {
			return s.sendLoot(active, uint64(item.GUID)|0x4000000000000000, lootItemSource, item.ItemTemplate, worlddb.GameObjectTemplate{}, 0)
		}
		return nil, nil
	}
	gift, found, err := s.Characters.GiftByItemGUID(item.GUID)
	if err != nil {
		return nil, err
	}
	if !found {
		if err := s.Characters.DeleteItem(item.GUID, active.GUID); err != nil {
			return nil, err
		}
		destroy, err := packet.Encode(packet.SMSGDestroyObject, encodeGUID(int64(uint64(item.GUID)|0x4000000000000000)))
		if err != nil {
			return nil, err
		}
		failure, err := inventoryResponse(active, bagUnknownItem, item)
		if err != nil {
			return nil, err
		}
		return append([][]byte{destroy}, failure...), nil
	}
	original, found, err := s.WorldData.ItemTemplate(gift.Entry)
	if err != nil {
		return nil, err
	}
	if !found {
		return inventoryResponse(active, bagError)
	}
	if err := s.Characters.UnwrapItem(item, gift, original.Entry, gift.Creator); err != nil {
		return nil, err
	}
	queries, err := itemQueryPackets([]worlddb.ItemTemplate{original})
	if err != nil {
		return nil, err
	}
	updates, err := itemUpdates(uint64(item.GUID), original, gift.Creator, item.Flags&^itemDynWrapped)
	if err != nil {
		return nil, err
	}
	return append(queries, updates...), nil
}

func (s *WorldServer) wrapItem(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 4 || s.Characters == nil || s.WorldData == nil {
		return nil, nil
	}
	wrapperBag, wrapperSlot, targetBag, targetSlot := inventoryBag(data[0]), int64(data[1]), inventoryBag(data[2]), int64(data[3])
	wrapper, found, err := s.Characters.ItemAt(active.GUID, wrapperBag, wrapperSlot)
	if err != nil {
		return nil, err
	}
	if !found {
		return inventoryResponse(active, bagItemNotFound)
	}
	target, found, err := s.Characters.ItemAt(active.GUID, targetBag, targetSlot)
	if err != nil {
		return nil, err
	}
	if !found {
		return inventoryResponse(active, bagItemNotFound)
	}
	targetTemplate, found, err := s.WorldData.ItemTemplate(target.ItemTemplate)
	if err != nil {
		return nil, err
	}
	if !found {
		return inventoryResponse(active, bagError)
	}
	if target.StackCount > 1 {
		return inventoryResponse(active, bagCantWrapStackable)
	}
	if target.Bag == 23 && target.Slot < 19 {
		return inventoryResponse(active, bagCantWrapEquipped)
	}
	if target.Flags&itemDynWrapped != 0 {
		return inventoryResponse(active, bagCantWrapWrapped)
	}
	if targetTemplate.InventoryType == itemTypeBag {
		return inventoryResponse(active, bagCantWrapBags)
	}
	if target.Flags&itemDynBound != 0 {
		return inventoryResponse(active, bagCantWrapBound)
	}
	if targetTemplate.MaxCount > 0 {
		return inventoryResponse(active, bagCantWrapUnique)
	}
	wrappedEntry, found := giftEntries[wrapper.ItemTemplate]
	if !found {
		return inventoryResponse(active, bagError)
	}
	wrappedTemplate, found, err := s.WorldData.ItemTemplate(wrappedEntry)
	if err != nil {
		return nil, err
	}
	if !found {
		return inventoryResponse(active, bagError)
	}
	if err := s.Characters.WrapItem(target, wrappedEntry, active.GUID); err != nil {
		return nil, err
	}
	queries, err := itemQueryPackets([]worlddb.ItemTemplate{wrappedTemplate})
	if err != nil {
		return nil, err
	}
	updates, err := itemUpdates(uint64(target.GUID), wrappedTemplate, active.GUID, target.Flags|itemDynWrapped)
	if err != nil {
		return nil, err
	}
	wrapperResponses, err := s.consumeWrapper(active, wrapper)
	if err != nil {
		return nil, err
	}
	return append(append(queries, updates...), wrapperResponses...), nil
}

func (s *WorldServer) consumeWrapper(active realm.Character, wrapper realm.InventoryItem) ([][]byte, error) {
	if wrapper.StackCount > 1 {
		count := wrapper.StackCount - 1
		if err := s.Characters.UpdateItemStack(wrapper.GUID, active.GUID, count); err != nil {
			return nil, err
		}
		update, err := packet.EncodeFieldUpdate(uint64(wrapper.GUID)|0x4000000000000000, itemFieldStack, uint32(count))
		if err != nil {
			return nil, err
		}
		return [][]byte{update}, nil
	}
	if err := s.Characters.DeleteItem(wrapper.GUID, active.GUID); err != nil {
		return nil, err
	}
	destroy, err := packet.Encode(packet.SMSGDestroyObject, encodeGUID(int64(uint64(wrapper.GUID)|0x4000000000000000)))
	if err != nil {
		return nil, err
	}
	return [][]byte{destroy}, nil
}

func inventoryResponse(active realm.Character, code byte, items ...realm.InventoryItem) ([][]byte, error) {
	response, err := encodeInventoryFailure(active, code, items...)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}
