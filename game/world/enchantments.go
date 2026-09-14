package world

import (
	"strconv"
	"strings"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/network/packet"
)

const itemFieldEnchantment = 20

var enchantmentCharges = map[int64]int64{3408: 1}

func itemEnchantments(value string) [5]packet.ItemEnchantment {
	var enchantments [5]packet.ItemEnchantment
	values := strings.Split(value, ",")
	for index := range enchantments {
		base := index * 3
		if base+2 >= len(values) {
			break
		}
		enchantments[index].ID, _ = strconv.ParseInt(values[base], 10, 64)
		enchantments[index].Duration, _ = strconv.ParseInt(values[base+1], 10, 64)
		enchantments[index].Charges, _ = strconv.ParseInt(values[base+2], 10, 64)
	}
	return enchantments
}

func itemEnchantmentsString(enchantments [5]packet.ItemEnchantment) string {
	values := make([]string, 0, len(enchantments)*3)
	for _, enchantment := range enchantments {
		values = append(values, strconv.FormatInt(enchantment.ID, 10), strconv.FormatInt(enchantment.Duration, 10), strconv.FormatInt(enchantment.Charges, 10))
	}
	return strings.Join(values, ",")
}

func (s *WorldServer) enchantItem(cast *spellCast, effect dbc.SpellEffect, temporary bool) {
	if cast == nil || s.Characters == nil || s.DBC == nil || cast.target.ItemGUID == 0 {
		return
	}
	item, found, err := s.Characters.ItemByGUID(cast.caster.GUID, int64(cast.target.ItemGUID&0x3fffffffffffffff))
	if err != nil || !found {
		return
	}
	_, found, err = s.DBC.SpellItemEnchantment(effect.MiscValue)
	if err != nil || !found {
		return
	}
	slot, duration, charges := 0, int64(-1), int64(0)
	if temporary {
		slot, duration, charges = 1, 0, enchantmentCharges[cast.spell.ID]
		if charges == 0 {
			duration = 30 * 60
			if cast.spell.CastUI != 0 {
				duration = 60 * 60
			}
		}
	}
	values := itemEnchantments(item.Enchantments)
	values[slot] = packet.ItemEnchantment{ID: effect.MiscValue, Duration: duration, Charges: charges}
	if err := s.Characters.UpdateItemEnchantments(item.GUID, item.Owner, itemEnchantmentsString(values)); err != nil {
		return
	}
	guid := uint64(item.GUID) | 0x4000000000000000
	for index, value := range []int64{values[slot].ID, values[slot].Duration, values[slot].Charges} {
		if update, updateErr := packet.EncodeFieldUpdate(guid, itemFieldEnchantment+slot*3+index, uint32(value)); updateErr == nil {
			s.sendPlayer(cast.caster.GUID, update)
		}
	}
	if temporary && duration > 0 {
		data := append(encodeUint64(uint64(item.GUID)), encodeUint32(int64(slot))...)
		data = append(data, encodeUint32(duration)...)
		if update, updateErr := packet.Encode(packet.SMSGItemEnchantTimeUpdate, data); updateErr == nil {
			s.sendPlayer(cast.caster.GUID, update)
		}
	}
	data := append(encodeUint32(0), encodeUint64(uint64(cast.caster.GUID))...)
	data = append(data, encodeUint64(uint64(item.Owner))...)
	data = append(data, encodeUint32(effect.MiscValue)...)
	data = append(data, encodeUint32(item.ItemTemplate)...)
	if logPacket, logErr := packet.Encode(packet.SMSGEnchantmentLog, data); logErr == nil {
		s.sendSpell(cast.caster, logPacket)
	}
}
