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
	spawn, found, err := s.WorldData.CreatureSpawnByID(int64(uint32(guid)))
	if err != nil || !found || spawn.Map != active.Map {
		return nil, err
	}
	dx, dy, dz := spawn.PositionX-active.PositionX, spawn.PositionY-active.PositionY, spawn.PositionZ-active.PositionZ
	if dx*dx+dy*dy+dz*dz >= maxShopDistance*maxShopDistance {
		return nil, nil
	}
	creature, found, err := s.WorldData.CreatureTemplate(spawn.Entry)
	if err != nil || !found || creature.NPCFlags&1 == 0 {
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
		itemData = append(itemData, encodeVendorValue(vendorItem.MaxCount)...)
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
