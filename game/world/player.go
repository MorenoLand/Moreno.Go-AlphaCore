package world

import (
	"encoding/binary"
	"math"
	"time"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const (
	charLoginFailed = 0x35
	unitFlagPlayer  = 0x08
	objectType      = 0x19
)

func (s *WorldServer) playerLogin(accountID int64, data []byte) ([]byte, error) {
	if len(data) < 8 {
		return packet.Encode(packet.SMSGCharacterLoginFailed, []byte{charLoginFailed})
	}
	character, found, err := s.Characters.Character(int64(binary.LittleEndian.Uint64(data)), accountID, 1)
	if err != nil {
		return nil, err
	}
	if !found {
		return packet.Encode(packet.SMSGCharacterLoginFailed, []byte{charLoginFailed})
	}
	loginTime := loginTimeSpeed()
	newWorld, err := newWorldPacket(character)
	if err != nil {
		return nil, err
	}
	return joinPackets(loginTime, newWorld), nil
}

func (s *WorldServer) initialPlayerPackets(character realm.Character) ([]byte, error) {
	spells, err := s.Characters.Spells(character.GUID)
	if err != nil {
		return nil, err
	}
	spellButtons, err := s.Characters.SpellButtons(character.GUID)
	if err != nil {
		return nil, err
	}
	buttons, err := s.Characters.Buttons(character.GUID)
	if err != nil {
		return nil, err
	}
	initialSpells := make([]byte, 3, 3+len(spells)*4+2)
	initialSpells[0] = 0
	binary.LittleEndian.PutUint16(initialSpells[1:], uint16(len(spells)))
	for _, spell := range spells {
		if !spell.Active {
			continue
		}
		value := make([]byte, 4)
		binary.LittleEndian.PutUint16(value, uint16(spell.ID))
		binary.LittleEndian.PutUint16(value[2:], uint16(spellButtons[spell.ID]))
		initialSpells = append(initialSpells, value...)
	}
	initialSpells = append(initialSpells, 0, 0)
	spellsPacket, err := packet.Encode(packet.SMSGInitialSpells, initialSpells)
	if err != nil {
		return nil, err
	}
	actionData := make([]byte, 120*4)
	for slot, action := range buttons {
		if slot < 0 || slot >= 120 {
			continue
		}
		binary.LittleEndian.PutUint32(actionData[slot*4:], uint32(action))
	}
	actionPacket, err := packet.Encode(packet.SMSGActionButtons, actionData)
	if err != nil {
		return nil, err
	}
	factionData := make([]byte, 4+64*5)
	binary.LittleEndian.PutUint32(factionData, 64)
	factionPacket, err := packet.Encode(packet.SMSGInitializeFactions, factionData)
	if err != nil {
		return nil, err
	}
	inventory, err := s.Characters.WorldInventory(character.GUID)
	if err != nil {
		return nil, err
	}
	items := make([]worlddb.ItemTemplate, 0, len(inventory))
	instances := make([]realm.InventoryItem, 0, len(inventory))
	for _, instance := range inventory {
		item, found, err := s.WorldData.ItemTemplate(instance.ItemTemplate)
		if err != nil {
			return nil, err
		}
		if found {
			items = append(items, item)
			instances = append(instances, instance)
		}
	}
	itemQueryPackets, err := itemQueryPackets(items)
	if err != nil {
		return nil, err
	}
	packets := [][]byte{factionPacket, spellsPacket, actionPacket}
	packets = append(packets, itemQueryPackets...)
	for index, item := range items {
		instance := instances[index]
		itemPacket, err := packet.EncodeItemCreate(uint64(instance.GUID)|0x4000000000000000, uint32(item.Entry), uint64(instance.Owner), uint64(instance.Creator), uint32(instance.StackCount), int32(instance.Duration), uint32(instance.Flags), instance.SpellCharges, packet.Movement{X: character.PositionX, Y: character.PositionY, Z: character.PositionZ, O: character.Orientation})
		if err != nil {
			return nil, err
		}
		packets = append(packets, itemPacket)
	}
	race := dbc.Race{ID: int64(character.Race)}
	if s.DBC != nil {
		if value, found, queryErr := s.DBC.Race(character.Race); queryErr != nil {
			return nil, queryErr
		} else if found {
			race = value
		}
	}
	values := buildPlayerFields(character, race, inventory)
	createPacket, err := packet.EncodePlayerCreate(uint64(character.GUID), values, packet.Movement{X: character.PositionX, Y: character.PositionY, Z: character.PositionZ, O: character.Orientation, WalkSpeed: 2.5, RunSpeed: 7, SwimSpeed: 4.722222, TurnRate: 3.141594})
	if err != nil {
		return nil, err
	}
	packets = append(packets, createPacket)
	creaturePackets, err := s.nearbyCreaturePackets(character)
	if err != nil {
		return nil, err
	}
	packets = append(packets, creaturePackets...)
	gameObjectPackets, err := s.nearbyGameObjectPackets(character)
	if err != nil {
		return nil, err
	}
	packets = append(packets, gameObjectPackets...)
	return joinPackets(packets...), nil
}

func buildPlayerFields(character realm.Character, race dbc.Race, inventory []realm.InventoryItem) []uint32 {
	values := make([]uint32, packet.PlayerFieldCount)
	packet.SetUint64(values, 0, uint64(character.GUID))
	values[2] = objectType
	packet.SetFloat(values, 4, playerScale(character))
	health := character.Health
	if health <= 0 {
		health = 1
	}
	values[22] = uint32(health)
	values[23] = uint32(character.Power1)
	values[24] = uint32(character.Power2)
	values[25] = uint32(character.Power3)
	values[26] = uint32(character.Power4)
	values[27] = uint32(health)
	values[28] = uint32(character.Power1)
	values[29] = 1000
	values[30] = 100
	values[31] = 100
	values[32] = uint32(character.Level)
	values[33] = uint32(race.FactionID)
	values[34] = byteValue(playerPowerType(character.Class), character.Gender, character.Class, character.Race)
	values[54] = unitFlagPlayer
	values[55] = uint32(character.Money)
	packet.SetFloat(values, 148, 0.388999998569489)
	packet.SetFloat(values, 149, playerCombatReach(character))
	values[151] = uint32(playerDisplayID(character, race))
	values[172] = byteValue(1, 0, 0, 0)
	values[328] = 0x89
	values[331] = byteValue(character.Haircolour, character.Hairstyle, character.Face, character.Skin)
	values[332] = uint32(character.XP)
	values[333] = uint32(xpToLevel(character.Level))
	values[526] = byteValue(0, character.Bankslots, character.Facialhair, uint8(character.ExtraFlags))
	values[623] = uint32(character.Talentpoints)
	values[624] = uint32(character.Skillpoints)
	for _, item := range inventory {
		field := -1
		if item.Slot >= 0 && item.Slot < 23 {
			field = 184 + int(item.Slot)*2
		} else if item.Slot >= 23 && item.Slot < 39 {
			field = 230 + int(item.Slot-23)*2
		}
		if field >= 0 {
			packet.SetUint64(values, field, uint64(item.GUID)|0x4000000000000000)
		}
	}
	return values
}

func newWorldPacket(character realm.Character) ([]byte, error) {
	data := make([]byte, 17)
	data[0] = byte(character.Map)
	binary.LittleEndian.PutUint32(data[1:], math.Float32bits(character.PositionX))
	binary.LittleEndian.PutUint32(data[5:], math.Float32bits(character.PositionY))
	binary.LittleEndian.PutUint32(data[9:], math.Float32bits(character.PositionZ))
	binary.LittleEndian.PutUint32(data[13:], math.Float32bits(character.Orientation))
	return packet.Encode(packet.SMSGNewWorld, data)
}

func loginTimeSpeed() []byte {
	now := time.Now()
	year := now.Year() - 2000
	month := int(now.Month()) - 1
	day := now.Day() - 1
	timeBits := now.Minute() | (now.Hour() << 6) | (int(now.Weekday()) << 11) | (day << 14) | (month << 20) | (year << 24)
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data, uint32(timeBits))
	binary.LittleEndian.PutUint32(data[4:], math.Float32bits(0.016666668))
	result, _ := packet.Encode(packet.SMSGLoginSetTimeSpeed, data)
	return result
}

func joinPackets(packets ...[]byte) []byte {
	length := 0
	for _, packet := range packets {
		length += len(packet)
	}
	result := make([]byte, 0, length)
	for _, packet := range packets {
		result = append(result, packet...)
	}
	return result
}

func byteValue(first, second, third, fourth uint8) uint32 {
	return uint32(first)<<24 | uint32(second)<<16 | uint32(third)<<8 | uint32(fourth)
}

func playerPowerType(class uint8) uint8 {
	if class == 3 {
		return 2
	}
	if class == 4 {
		return 3
	}
	return 0
}

func playerDisplayID(character realm.Character, race dbc.Race) int64 {
	if character.Gender == 0 {
		return race.MaleDisplayID
	}
	return race.FemaleDisplayID
}

func playerScale(character realm.Character) float32 {
	if character.Race == 6 {
		if character.Gender == 0 {
			return 1.35
		}
		return 1.25
	}
	if character.Race == 7 {
		return 1.15
	}
	return 1
}

func playerCombatReach(character realm.Character) float32 {
	if character.Race == 6 {
		if character.Gender == 0 {
			return 4.05
		}
		return 3.75
	}
	if character.Race == 7 {
		return 1.725
	}
	return 1.5
}

func xpToLevel(level uint8) int64 {
	value := int(level)
	diff := 0
	if value > 31 {
		diff = 5 * (value - 30)
	} else if value == 31 {
		diff = 6
	} else if value == 30 {
		diff = 3
	} else if value == 29 {
		diff = 1
	}
	return int64(math.Round(float64((8*value+diff)*(45+5*value)+1)/100) * 100)
}
