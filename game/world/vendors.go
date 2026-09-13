package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const maxShopDistance float32 = 5.5555553

func (s *WorldServer) listInventory(active realm.Character, data []byte) ([][]byte, error) {
	if s.WorldData == nil || len(data) < 8 {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	if guid == 0 {
		return nil, nil
	}
	_, creature, found, err := s.vendorAt(active, guid)
	if err != nil || !found {
		return nil, err
	}
	items, err := s.WorldData.VendorItems(creature.Entry, creature.VendorID > 0)
	if err != nil {
		return nil, err
	}
	templates := make([]worlddb.ItemTemplate, 0, len(items))
	itemData := make([]byte, 0, len(items)*28)
	for index, vendorItem := range items {
		item, found, err := s.WorldData.ItemTemplate(vendorItem.Item)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		itemData = append(itemData, encodeVendorValue(int64(index+1))...)
		itemData = append(itemData, encodeVendorValue(item.Entry)...)
		itemData = append(itemData, encodeVendorValue(item.DisplayID)...)
		maxCount := vendorItem.MaxCount
		if maxCount <= 0 {
			maxCount = 0xffffffff
		}
		itemData = append(itemData, encodeVendorValue(maxCount)...)
		itemData = append(itemData, encodeVendorValue(item.BuyPrice)...)
		itemData = append(itemData, encodeVendorValue(item.MaxDurability)...)
		itemData = append(itemData, encodeVendorValue(item.BuyCount)...)
		templates = append(templates, item)
	}
	if len(templates) == 0 {
		itemData = []byte{0}
	}
	body := append(encodeGUID(int64(guid)), byte(len(templates)))
	body = append(body, itemData...)
	list, err := packet.Encode(packet.SMSGListInventory, body)
	if err != nil {
		return nil, err
	}
	queries, err := itemQueryPackets(templates)
	if err != nil {
		return nil, err
	}
	return append(queries, list), nil
}

func encodeVendorValue(value int64) []byte {
	return encodeUint32(value)
}

func (s *WorldServer) vendorAt(active realm.Character, guid uint64) (worlddb.CreatureSpawn, worlddb.CreatureTemplate, bool, error) {
	if s.WorldData == nil || guid == 0 {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, nil
	}
	spawn, found, err := s.WorldData.CreatureSpawnByID(int64(uint32(guid)))
	if err != nil || !found || spawn.Map != active.Map {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, err
	}
	dx, dy, dz := spawn.PositionX-active.PositionX, spawn.PositionY-active.PositionY, spawn.PositionZ-active.PositionZ
	if dx*dx+dy*dy+dz*dz >= maxShopDistance*maxShopDistance {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, nil
	}
	creature, found, err := s.WorldData.CreatureTemplate(spawn.Entry)
	if err != nil || !found || creature.NPCFlags&1 == 0 {
		return worlddb.CreatureSpawn{}, worlddb.CreatureTemplate{}, false, err
	}
	return spawn, creature, true, nil
}

func (s *WorldServer) buyItem(active *realm.Character, data []byte, inSlot bool) ([][]byte, error) {
	minimum := 14
	if inSlot {
		minimum = 22
	}
	if active == nil || len(data) < minimum || s.Characters == nil {
		return nil, nil
	}
	vendorGUID, itemEntry := binary.LittleEndian.Uint64(data), binary.LittleEndian.Uint32(data[8:])
	countOffset := 12
	if inSlot {
		countOffset = 21
	}
	count := int64(data[countOffset])
	if count <= 0 {
		count = 1
	}
	_, creature, found, err := s.creatureAt(*active, vendorGUID, maxShopDistance)
	if err != nil {
		return nil, err
	}
	if !found {
		return s.buyFailure(*active, itemEntry, vendorGUID, 5, count)
	}
	if creature.NPCFlags&1 == 0 {
		return s.buyFailure(*active, itemEntry, vendorGUID, 11, count)
	}
	items, err := s.WorldData.VendorItems(creature.Entry, creature.VendorID > 0)
	if err != nil {
		return nil, err
	}
	listed := false
	for _, item := range items {
		if item.Item == int64(itemEntry) {
			listed = true
			break
		}
	}
	if !listed {
		return s.buyFailure(*active, itemEntry, vendorGUID, 11, count)
	}
	template, found, err := s.WorldData.ItemTemplate(int64(itemEntry))
	if err != nil {
		return nil, err
	}
	if !found {
		return s.buyFailure(*active, itemEntry, vendorGUID, 11, count)
	}
	realCount := count
	if template.BuyCount != 1 {
		realCount = template.BuyCount
	}
	totalCost := template.BuyPrice * count
	if active.Money < totalCost {
		return s.buyFailure(*active, itemEntry, vendorGUID, 2, byteCount(realCount))
	}
	bag, slot, bagGUID := int64(23), int64(-1), uint64(0)
	if inSlot {
		bagGUID = binary.LittleEndian.Uint64(data[12:20])
		slot = int64(data[20])
		if bagGUID != uint64(active.GUID) {
			return s.buyFailure(*active, itemEntry, vendorGUID, 11, byteCount(realCount))
		}
	} else {
		items, err := s.Characters.InventoryItems(active.GUID, 23, 23, 39)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if item.ItemTemplate == int64(itemEntry) && template.Stackable > 1 && item.StackCount+realCount <= template.Stackable {
				if err := s.Characters.UpdateItemStack(item.GUID, active.GUID, item.StackCount+realCount); err != nil {
					return nil, err
				}
				active.Money -= totalCost
				if err := s.Characters.UpdateMoney(active.GUID, active.AccountID, active.RealmID, active.Money); err != nil {
					return nil, err
				}
				s.updatePlayer(*active)
				return s.purchaseResponses(*active, item, int64(itemEntry), realCount, 23)
			}
		}
		slot, err = s.Characters.FirstEmptySlot(active.GUID, 23, 23, 39)
		if err != nil {
			return nil, err
		}
		if slot < 0 {
			return s.buyFailure(*active, itemEntry, vendorGUID, 8, byteCount(realCount))
		}
	}
	if _, found, err := s.Characters.ItemAt(active.GUID, bag, slot); err != nil {
		return nil, err
	} else if found {
		return s.buyFailure(*active, itemEntry, vendorGUID, 8, byteCount(realCount))
	}
	item, err := s.Characters.CreateInventoryItem(active.GUID, 0, bag, slot, int64(itemEntry), realCount)
	if err != nil {
		return nil, err
	}
	active.Money -= totalCost
	if err := s.Characters.UpdateMoney(active.GUID, active.AccountID, active.RealmID, active.Money); err != nil {
		return nil, err
	}
	s.updatePlayer(*active)
	return s.purchaseResponses(*active, item, int64(itemEntry), realCount, bag)
}

func (s *WorldServer) sellItem(active *realm.Character, data []byte) ([][]byte, error) {
	if active == nil || len(data) < 17 || s.Characters == nil {
		return nil, nil
	}
	vendorGUID, itemGUID := binary.LittleEndian.Uint64(data), binary.LittleEndian.Uint64(data[8:])
	_, creature, found, err := s.creatureAt(*active, vendorGUID, maxShopDistance)
	if err != nil {
		return nil, err
	}
	if !found {
		return s.sellFailure(*active, itemGUID, vendorGUID, 1)
	}
	if creature.NPCFlags&1 == 0 {
		return s.sellFailure(*active, itemGUID, vendorGUID, 2)
	}
	itemGUID &= 0x3fffffffffffffff
	item, found, err := s.Characters.ItemByGUID(active.GUID, int64(itemGUID))
	if err != nil {
		return nil, err
	}
	if !found {
		return s.sellFailure(*active, itemGUID, vendorGUID, 1)
	}
	template, found, err := s.WorldData.ItemTemplate(item.ItemTemplate)
	if err != nil {
		return nil, err
	}
	if !found || template.SellPrice == 0 {
		return s.sellFailure(*active, itemGUID, vendorGUID, 2)
	}
	amount := int64(data[16])
	if amount <= 0 {
		amount = item.StackCount
	}
	if amount > item.StackCount {
		return s.sellFailure(*active, itemGUID, vendorGUID, 1)
	}
	if amount == item.StackCount {
		if err := s.Characters.DeleteItem(item.GUID, active.GUID); err != nil {
			return nil, err
		}
	} else if err := s.Characters.UpdateItemStack(item.GUID, active.GUID, item.StackCount-amount); err != nil {
		return nil, err
	}
	active.Money += template.SellPrice * amount
	if active.Money > 2147483647 {
		active.Money = 2147483647
	}
	if err := s.Characters.UpdateMoney(active.GUID, active.AccountID, active.RealmID, active.Money); err != nil {
		return nil, err
	}
	s.updatePlayer(*active)
	responses := make([][]byte, 0, 2)
	if amount == item.StackCount {
		response, err := packet.Encode(packet.SMSGDestroyObject, encodeGUID(int64(uint64(item.GUID)|0x4000000000000000)))
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	money, err := packet.EncodeFieldUpdate(uint64(active.GUID), 55, uint32(active.Money))
	if err != nil {
		return nil, err
	}
	responses = append(responses, money)
	return responses, nil
}

func (s *WorldServer) purchaseResponses(active realm.Character, item realm.InventoryItem, entry, count, bag int64) ([][]byte, error) {
	responses := make([][]byte, 0, 3)
	if item.StackCount == count {
		created, err := packet.EncodeItemCreate(uint64(item.GUID)|0x4000000000000000, uint32(entry), uint64(active.GUID), 0, uint32(item.StackCount), 0, 0, item.SpellCharges, packet.Movement{X: active.PositionX, Y: active.PositionY, Z: active.PositionZ, O: active.Orientation})
		if err != nil {
			return nil, err
		}
		responses = append(responses, created)
	}
	push, err := itemPushResult(item, entry, bag)
	if err != nil {
		return nil, err
	}
	responses = append(responses, push)
	money, err := packet.EncodeFieldUpdate(uint64(active.GUID), 55, uint32(active.Money))
	if err != nil {
		return nil, err
	}
	return append(responses, money), nil
}

func itemPushResult(item realm.InventoryItem, entry, bag int64) ([]byte, error) {
	body := append(encodeGUID(int64(uint64(item.GUID)|0x4000000000000000)), encodeUint32(1)...)
	body = append(body, encodeUint32(1)...)
	if bag == 23 {
		body = append(body, 0xff)
	} else {
		body = append(body, byte(bag))
	}
	body = append(body, encodeUint32(entry)...)
	return packet.Encode(packet.SMSGItemPushResult, body)
}

func (s *WorldServer) buyFailure(active realm.Character, entry uint32, vendor uint64, result uint32, count int64) ([][]byte, error) {
	body := append(encodeGUID(int64(vendor)), encodeUint32(int64(entry))...)
	body = append(body, byte(count), byte(result))
	response, err := packet.Encode(packet.SMSGBuyFailed, body)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) sellFailure(active realm.Character, item, vendor uint64, result byte) ([][]byte, error) {
	body := append(encodeGUID(int64(vendor)), encodeGUID(int64(item))...)
	body = append(body, result)
	response, err := packet.Encode(packet.SMSGSellItem, body)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func byteCount(value int64) int64 {
	if value > 255 {
		return 255
	}
	return value
}
