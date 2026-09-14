package world

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const (
	maxPlayerLevel         uint32 = 60
	maxGMLevel             uint32 = 100
	unitFlagDebugCombatLog uint32 = 0x02000000
	chatMessageSystem      byte   = 0x09
	maxPlayerMoney         int64  = 2147483647
)

func (s *WorldServer) gmCheat(active *realm.Character, opcode packet.Opcode, data []byte, gmLevel int) ([][]byte, error) {
	if active == nil || gmLevel <= 0 {
		return nil, nil
	}
	switch opcode {
	case packet.CMSGCheatSetMoney:
		if len(data) < 4 {
			return nil, nil
		}
		return s.cheatMoney(active, int64(binary.LittleEndian.Uint32(data)))
	case packet.CMSGLevelCheat:
		if len(data) < 4 {
			return nil, nil
		}
		return s.cheatLevel(active, binary.LittleEndian.Uint32(data), gmLevel > 0)
	case packet.CMSGLevelUpCheat:
		return s.cheatLevel(active, uint32(active.Level)+1, gmLevel > 0)
	case packet.CMSGPetLevelCheat:
		return nil, s.petLevelCheat(*active, data)
	case packet.CMSGLearnSpell:
		if len(data) < 4 {
			return nil, nil
		}
		return s.learnSpell(*active, int64(binary.LittleEndian.Uint32(data)))
	case packet.CMSGCreateItem:
		if len(data) < 4 {
			return nil, nil
		}
		return s.cheatCreateItem(active, int64(binary.LittleEndian.Uint32(data)))
	case packet.CMSGCreateMonster:
		if len(data) < 4 {
			return nil, nil
		}
		return nil, s.createCreature(*active, int64(binary.LittleEndian.Uint32(data)))
	case packet.CMSGDestroyMonster:
		if gmLevel < 2 || len(data) < 8 {
			return nil, nil
		}
		return nil, s.destroyCreature(*active, binary.LittleEndian.Uint64(data))
	case packet.CMSGGodMode:
		if len(data) < 1 {
			return nil, nil
		}
		return s.cheatGodMode(*active, data[0] >= 1)
	case packet.CMSGCooldownCheat:
		return s.cheatCooldowns(active.GUID)
	case packet.CMSGRecharge:
		return s.cheatRecharge(active)
	case packet.CMSGEnableDebugCombatLogging:
		if len(data) < 4 {
			return nil, nil
		}
		flags := s.unitFlags(active.GUID)
		if binary.LittleEndian.Uint32(data) != 0 {
			flags |= unitFlagDebugCombatLog
		} else {
			flags &^= unitFlagDebugCombatLog
		}
		s.setUnitFlags(*active, flags)
	case packet.CMSGBeastMaster:
		if len(data) < 1 {
			return nil, nil
		}
		enabled := data[0] >= 1
		s.setBeastMaster(active.GUID, enabled)
		if enabled {
			s.setSanctuary(active.GUID, 3*time.Second)
		}
		message, err := messageChatPacket(chatMessageSystem, 0, active.GUID, fmt.Sprintf("Beastmaster %s", map[bool]string{true: "enabled", false: "disabled"}[enabled]))
		if err != nil {
			return nil, err
		}
		return [][]byte{message}, nil
	case packet.CMSGTriggerCinematicCheat:
		if len(data) < 4 {
			return nil, nil
		}
		return s.cheatCinematic(*active, int64(binary.LittleEndian.Uint32(data)))
	case packet.CMSGTeleportToPlayer:
		return s.cheatGoPlayer(active, data)
	case packet.MSGGMSummon:
		return s.cheatSummon(active, data)
	}
	return nil, nil
}

func (s *WorldServer) cheatMoney(active *realm.Character, amount int64) ([][]byte, error) {
	if active.Money > maxPlayerMoney-amount {
		active.Money = maxPlayerMoney
	} else if active.Money+amount < 0 {
		active.Money = 0
	} else {
		active.Money += amount
	}
	if s.Characters != nil {
		if err := s.Characters.UpdateMoney(active.GUID, active.AccountID, active.RealmID, active.Money); err != nil {
			return nil, err
		}
	}
	s.updatePlayer(*active)
	update, err := s.playerFieldUpdate(*active, 55, uint32(active.Money))
	if err != nil {
		return nil, err
	}
	return [][]byte{update}, nil
}

func (s *WorldServer) cheatLevel(active *realm.Character, level uint32, gm bool) ([][]byte, error) {
	maxLevel := maxPlayerLevel
	if gm {
		maxLevel = maxGMLevel
	}
	if level == 0 || level > maxLevel || level == uint32(active.Level) {
		return nil, nil
	}
	oldLevel := uint32(active.Level)
	oldHealth, oldPower := active.Health, playerPower(*active, int64(playerPowerType(active.Class)))
	levelingUp := level > oldLevel
	levelCount := level - oldLevel
	if !levelingUp {
		levelCount = oldLevel - level
	}
	var talent, skill int64
	for index := uint32(0); index < levelCount; index++ {
		value := oldLevel + index + 1
		if !levelingUp {
			value = oldLevel - index
		}
		talent += talentPointsGain(value)
		skill += skillPointsGain(value)
	}
	if levelingUp {
		active.Talentpoints += talent
		active.Skillpoints += skill
	} else {
		active.Talentpoints = maxInt64(active.Talentpoints-talent, 0)
		active.Skillpoints = maxInt64(active.Skillpoints-skill, 0)
	}
	active.Level = uint8(level)
	active.Leveltime = 0
	powerType := int64(playerPowerType(active.Class))
	statsFound := false
	if s.WorldData != nil {
		stats, found, err := s.WorldData.ClassStats(active.Class, active.Level)
		if err != nil {
			return nil, err
		}
		statsFound = found
		if found {
			active.Health = maxInt64(stats.BaseHealth, 1)
			if powerType == 0 {
				active.Power1 = maxInt64(stats.BaseMana, 0)
				if stats.BaseMana > 0 {
					s.setPlayerMaxPower(active.GUID, powerType, stats.BaseMana)
				}
			} else {
				s.setPlayerMaxPower(active.GUID, powerType, maxPowerValue(powerType))
				setPlayerPower(active, powerType, maxPowerValue(powerType))
			}
			s.setPlayerMaxHealth(active.GUID, active.Health)
		}
	}
	if s.Characters != nil {
		if err := s.Characters.UpdateLevelState(active.GUID, active.AccountID, active.RealmID, active.Level, active.Talentpoints, active.Skillpoints); err != nil {
			return nil, err
		}
		if statsFound && active.Health != oldHealth {
			if err := s.Characters.UpdateHealth(active.GUID, active.AccountID, active.RealmID, active.Health); err != nil {
				return nil, err
			}
		}
		if statsFound && playerPower(*active, powerType) != oldPower {
			if err := s.Characters.UpdatePower(active.GUID, active.AccountID, active.RealmID, powerType, playerPower(*active, powerType)); err != nil {
				return nil, err
			}
		}
	}
	s.updatePlayer(*active)
	updates := []struct {
		field int
		value uint32
	}{{32, uint32(active.Level)}, {333, uint32(xpToLevel(active.Level))}, {623, uint32(active.Talentpoints)}, {624, uint32(active.Skillpoints)}}
	if statsFound {
		updates = append(updates, struct {
			field int
			value uint32
		}{22, uint32(active.Health)}, struct {
			field int
			value uint32
		}{27, uint32(active.Health)})
		updates = append(updates, struct {
			field int
			value uint32
		}{23 + int(powerType), uint32(playerPower(*active, powerType))}, struct {
			field int
			value uint32
		}{28 + int(powerType), uint32(s.playerMaxPower(active.GUID, powerType))})
	}
	responses := make([][]byte, 0, len(updates)+1)
	for _, update := range updates {
		value, err := s.playerFieldUpdate(*active, update.field, update.value)
		if err != nil {
			return nil, err
		}
		responses = append(responses, value)
	}
	if levelingUp {
		body := append(encodeUint32(int64(active.Level)), encodeUint32(active.Health-oldHealth)...)
		mana := int64(0)
		if powerType == 0 {
			mana = playerPower(*active, powerType) - oldPower
		}
		body = append(body, encodeUint32(mana)...)
		levelup, err := packet.Encode(packet.SMSGLevelupInfo, body)
		if err != nil {
			return nil, err
		}
		responses = append(responses, levelup)
	}
	return responses, nil
}

func talentPointsGain(level uint32) int64 { return 10 + int64(level/10)*5 }

func skillPointsGain(level uint32) int64 {
	if level%10 == 0 {
		return 2
	}
	return 1
}

func maxPowerValue(powerType int64) int64 {
	if powerType == 0 {
		return 1000
	}
	return 100
}

func (s *WorldServer) learnSpell(active realm.Character, spellID int64) ([][]byte, error) {
	if s.Characters == nil || s.DBC == nil || spellID <= 0 {
		return nil, nil
	}
	exists, err := s.DBC.SpellExists(spellID)
	if err != nil || !exists {
		return nil, err
	}
	allowed, err := s.DBC.SpellAllowedForRaceClass(spellID, active.Race, active.Class)
	if err != nil || !allowed {
		return nil, err
	}
	spells, err := s.Characters.Spells(active.GUID)
	if err != nil {
		return nil, err
	}
	for _, known := range spells {
		if known.ID == spellID {
			return nil, nil
		}
	}
	if err := s.Characters.AddSpell(active.GUID, spellID); err != nil {
		return nil, err
	}
	preceded, found, err := s.DBC.PrecededSpell(spellID)
	if err != nil {
		return nil, err
	}
	if found {
		for _, known := range spells {
			if known.ID == preceded && known.Active {
				if err := s.Characters.SetSpellActive(active.GUID, preceded, false); err != nil {
					return nil, err
				}
				data := append(encodeUint16(preceded), encodeUint16(spellID)...)
				response, err := packet.Encode(packet.SMSSupersededSpell, data)
				if err != nil {
					return nil, err
				}
				return [][]byte{response}, nil
			}
		}
	}
	data := append(encodeUint16(spellID), encodeUint16(0)...)
	response, err := packet.Encode(packet.SMSGLearnedSpell, data)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) cheatCreateItem(active *realm.Character, entry int64) ([][]byte, error) {
	if s.Characters == nil || s.WorldData == nil || entry <= 0 {
		return nil, nil
	}
	template, found, err := s.WorldData.ItemTemplate(entry)
	if err != nil || !found {
		return nil, err
	}
	if template.MaxCount > 0 {
		count, err := s.Characters.ItemCount(active.GUID, entry)
		if err != nil {
			return nil, err
		}
		if count >= template.MaxCount {
			return nil, nil
		}
	}
	items, err := s.Characters.InventoryItems(active.GUID, 23, 23, 39)
	if err != nil {
		return nil, err
	}
	if template.Stackable > 1 {
		for _, item := range items {
			if item.ItemTemplate != entry || item.StackCount >= template.Stackable {
				continue
			}
			item.StackCount++
			if err := s.Characters.UpdateItemStack(item.GUID, active.GUID, item.StackCount); err != nil {
				return nil, err
			}
			update, err := packet.EncodeFieldUpdate(uint64(item.GUID)|0x4000000000000000, itemFieldStack, uint32(item.StackCount))
			if err != nil {
				return nil, err
			}
			push, err := itemPushResult(item, entry, 23)
			if err != nil {
				return nil, err
			}
			return [][]byte{update, push}, nil
		}
	}
	slot, err := s.Characters.FirstEmptySlot(active.GUID, 23, 23, 39)
	if err != nil || slot < 0 {
		return nil, err
	}
	item, err := s.Characters.CreateInventoryItem(active.GUID, 0, 23, slot, entry, 1)
	if err != nil {
		return nil, err
	}
	if template.Bonding == 1 {
		item.Flags = itemDynBound
		if err := s.Characters.UpdateItemFlags(item.GUID, active.GUID, item.Flags); err != nil {
			return nil, err
		}
	}
	for index, spell := range template.Spells {
		if spell.Charges == 0 {
			continue
		}
		item.SpellCharges[index] = spell.Charges
		if err := s.Characters.UpdateItemSpellCharges(item.GUID, active.GUID, int64(index), spell.Charges); err != nil {
			return nil, err
		}
	}
	created, err := packet.EncodeItemCreate(uint64(item.GUID)|0x4000000000000000, uint32(item.ItemTemplate), uint64(item.Owner), uint64(item.Creator), uint32(item.StackCount), 0, encodedItemFlags(template, item.Flags), item.SpellCharges, packet.Movement{X: active.PositionX, Y: active.PositionY, Z: active.PositionZ, O: active.Orientation})
	if err != nil {
		return nil, err
	}
	push, err := itemPushResult(item, entry, 23)
	if err != nil {
		return nil, err
	}
	return [][]byte{created, push}, nil
}

func (s *WorldServer) cheatGodMode(active realm.Character, enabled bool) ([][]byte, error) {
	s.setGodMode(active.GUID, enabled)
	mode, err := packet.Encode(packet.SMSGGodMode, []byte{boolByte(enabled)})
	if err != nil {
		return nil, err
	}
	message, err := messageChatPacket(chatMessageSystem, 0, active.GUID, fmt.Sprintf("Godmode %s", map[bool]string{true: "enabled", false: "disabled"}[enabled]))
	if err != nil {
		return nil, err
	}
	return [][]byte{mode, message}, nil
}

func boolByte(value bool) byte {
	if value {
		return 1
	}
	return 0
}

func (s *WorldServer) cheatCooldowns(guid int64) ([][]byte, error) {
	s.spells.mu.Lock()
	delete(s.spells.cooldowns, guid)
	s.spells.mu.Unlock()
	response, err := packet.Encode(packet.SMSGCooldownCheat, encodeGUID(guid))
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) cheatRecharge(active *realm.Character) ([][]byte, error) {
	powerType := int64(playerPowerType(active.Class))
	setPlayerPower(active, powerType, s.playerMaxPower(active.GUID, powerType))
	if s.Characters != nil {
		if err := s.Characters.UpdatePower(active.GUID, active.AccountID, active.RealmID, powerType, playerPower(*active, powerType)); err != nil {
			return nil, err
		}
	}
	s.updatePlayer(*active)
	update, err := s.playerFieldUpdate(*active, 23+int(powerType), uint32(playerPower(*active, powerType)))
	if err != nil {
		return nil, err
	}
	return [][]byte{update}, nil
}

func (s *WorldServer) cheatCinematic(active realm.Character, id int64) ([][]byte, error) {
	if s.DBC == nil || id <= 0 {
		return nil, nil
	}
	found, err := s.DBC.CinematicSequenceExists(id)
	if err != nil || !found {
		return nil, err
	}
	response, err := packet.Encode(packet.SMSGTriggerCinematic, encodeUint32(id))
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) cheatGoPlayer(active *realm.Character, data []byte) ([][]byte, error) {
	name, err := packet.ReadString(data, 0, 0)
	if err != nil || strings.TrimSpace(name) == "" {
		return nil, nil
	}
	target, found := s.playerByName(strings.TrimSpace(name))
	if !found && s.Characters != nil {
		target, found, err = s.Characters.CharacterByName(strings.TrimSpace(name))
		if err != nil {
			return nil, err
		}
	}
	if found {
		response, err := s.teleportPlayer(active, target.Map, target.PositionX, target.PositionY, target.PositionZ, target.Orientation)
		if err != nil {
			return nil, err
		}
		return [][]byte{response}, nil
	}
	if s.WorldData == nil {
		return nil, nil
	}
	port, found, err := s.WorldData.WorldportByName(strings.TrimSpace(name))
	if err != nil || !found {
		return nil, err
	}
	response, err := s.teleportPlayer(active, port.Map, port.X, port.Y, port.Z, port.O)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) cheatSummon(active *realm.Character, data []byte) ([][]byte, error) {
	name, err := packet.ReadString(data, 0, 0)
	if err != nil || strings.TrimSpace(name) == "" {
		return nil, nil
	}
	name = strings.TrimSpace(name)
	target, online := s.playerByName(name)
	if online {
		response, err := s.teleportPlayer(&target, active.Map, active.PositionX, active.PositionY, active.PositionZ, active.Orientation)
		if err != nil {
			return nil, err
		}
		s.sendPlayer(target.GUID, response)
		return nil, nil
	}
	if s.Characters == nil {
		return nil, nil
	}
	target, found, err := s.Characters.CharacterByName(name)
	if err != nil || !found {
		return nil, err
	}
	if err := s.Characters.UpdateLocation(target.GUID, target.AccountID, target.RealmID, active.Map, active.PositionX, active.PositionY, active.PositionZ, active.Orientation); err != nil {
		return nil, err
	}
	return nil, s.Characters.UpdateZone(target.GUID, target.AccountID, target.RealmID, active.Zone)
}

func (s *WorldServer) playerFieldUpdate(active realm.Character, field int, value uint32) ([]byte, error) {
	update, err := packet.EncodeFieldUpdate(uint64(active.GUID), field, value)
	if err != nil {
		return nil, err
	}
	s.broadcastPlayer(active, update)
	return update, nil
}
