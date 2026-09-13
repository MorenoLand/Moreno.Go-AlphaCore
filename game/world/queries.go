package world

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) itemQuerySingle(data []byte) ([]byte, error) {
	if len(data) < 12 {
		return nil, nil
	}
	entry := int64(binary.LittleEndian.Uint32(data))
	if entry <= 0 {
		return nil, nil
	}
	item, found, err := s.WorldData.ItemTemplate(entry)
	if err != nil || !found {
		return nil, err
	}
	payload, err := itemQueryData(item)
	if err != nil {
		return nil, err
	}
	return packet.Encode(packet.SMSGItemQuerySingleResponse, payload)
}

func (s *WorldServer) itemQueryMultiple(data []byte) ([][]byte, error) {
	if len(data) < 4 {
		return nil, nil
	}
	count := binary.LittleEndian.Uint32(data)
	if count == 0 || uint64(count) > uint64((len(data)-4)/4) {
		return nil, nil
	}
	items := make([]worlddb.ItemTemplate, 0, count)
	for index := int(count) - 1; index >= 0; index-- {
		entry := int64(binary.LittleEndian.Uint32(data[4+index*4:]))
		item, found, err := s.WorldData.ItemTemplate(entry)
		if err != nil {
			return nil, err
		}
		if found {
			items = append(items, item)
		}
	}
	return itemQueryPackets(items)
}

func itemQueryPackets(items []worlddb.ItemTemplate) ([][]byte, error) {
	packets := make([][]byte, 0, 1)
	queryData := make([]byte, 0)
	written := 0
	flush := func() error {
		if written == 0 {
			return nil
		}
		body := append([]byte{byte(written)}, queryData...)
		response, err := packet.Encode(packet.SMSGItemQueryMultipleResponse, body)
		if err != nil {
			return err
		}
		packets = append(packets, response)
		queryData = queryData[:0]
		written = 0
		return nil
	}
	for _, item := range items {
		itemData, err := itemQueryData(item)
		if err != nil {
			return nil, err
		}
		if written == 255 || len(queryData)+len(itemData)+packet.HeaderSize+4 > packet.MaxPacketSize {
			if err := flush(); err != nil {
				return nil, err
			}
		}
		if len(itemData)+packet.HeaderSize+4 > packet.MaxPacketSize {
			return nil, fmt.Errorf("item query response is too large")
		}
		queryData = append(queryData, itemData...)
		written++
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return packets, nil
}

func itemQueryData(item worlddb.ItemTemplate) ([]byte, error) {
	name, err := packet.StringBytes(item.Name)
	if err != nil {
		return nil, err
	}
	description, err := packet.StringBytes(item.Description)
	if err != nil {
		return nil, err
	}
	data := make([]byte, 0, 512)
	put := func(value int64) {
		encoded := make([]byte, 4)
		binary.LittleEndian.PutUint32(encoded, uint32(value))
		data = append(data, encoded...)
	}
	putFloat := func(value float32) { put(int64(int32(value))) }
	put(item.Entry)
	put(item.Class)
	put(item.Subclass)
	data = append(data, name...)
	data = append(data, 0, 0, 0)
	for _, value := range []int64{item.DisplayID, item.Quality, item.Flags, item.BuyPrice, item.SellPrice, item.InventoryType, item.AllowableClass, item.AllowableRace, item.ItemLevel, item.RequiredLevel, item.RequiredSkill, item.RequiredSkillRank, item.MaxCount, item.Stackable, item.ContainerSlots} {
		put(value)
	}
	for _, stat := range item.Stats {
		put(stat.Type)
		put(stat.Value)
	}
	for _, damage := range item.Damages {
		putFloat(damage.Minimum)
		putFloat(damage.Maximum)
		put(damage.Type)
	}
	for _, value := range []int64{item.Armor, item.HolyRes, item.FireRes, item.NatureRes, item.FrostRes, item.ShadowRes, item.Delay, item.AmmoType, 0} {
		put(value)
	}
	for _, spell := range item.Spells {
		for _, value := range []int64{spell.ID, spell.Trigger, spell.Charges, spell.Cooldown, spell.Category, spell.CategoryCooldown} {
			put(value)
		}
	}
	put(item.Bonding)
	data = append(data, description...)
	for _, value := range []int64{item.PageText, item.PageLanguage, item.PageMaterial, item.StartQuest, item.LockID, item.Material, item.Sheath} {
		put(value)
	}
	return data, nil
}

func (s *WorldServer) pageTextQuery(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 12 {
		return nil, nil
	}
	pageID := int64(binary.LittleEndian.Uint32(data))
	responses := make([][]byte, 0, 1)
	first := true
	for first || pageID > 0 {
		first = false
		page, found, err := s.WorldData.PageText(pageID)
		if err != nil {
			return nil, err
		}
		text := "Item page missing."
		next := int64(0)
		if found {
			text = formatGameText(page.Text, active)
			next = page.NextPage
		}
		encodedText, err := packet.StringBytes(text)
		if err != nil {
			return nil, err
		}
		body := make([]byte, 4, 8+len(encodedText))
		binary.LittleEndian.PutUint32(body, uint32(pageID))
		body = append(body, encodedText...)
		encoded := make([]byte, 4)
		binary.LittleEndian.PutUint32(encoded, uint32(next))
		body = append(body, encoded...)
		response, err := packet.Encode(packet.SMSGPageTextQueryResponse, body)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
		if !found {
			break
		}
		pageID = next
	}
	return responses, nil
}

func (s *WorldServer) creatureQuery(data []byte) ([]byte, error) {
	if len(data) < 12 || binary.LittleEndian.Uint64(data[4:]) == 0 {
		return nil, nil
	}
	creature, found, err := s.WorldData.CreatureTemplate(int64(binary.LittleEndian.Uint32(data)))
	if err != nil || !found {
		return nil, err
	}
	return creatureQueryPacket(creature)
}

func creatureQueryPacket(creature worlddb.CreatureTemplate) ([]byte, error) {
	name, err := packet.StringBytes(creature.Name)
	if err != nil {
		return nil, err
	}
	subname, err := packet.StringBytes(creature.Subname)
	if err != nil {
		return nil, err
	}
	body := make([]byte, 0, 20+len(name)+len(subname))
	put := func(value int64) {
		encoded := make([]byte, 4)
		binary.LittleEndian.PutUint32(encoded, uint32(value))
		body = append(body, encoded...)
	}
	put(creature.Entry)
	body = append(body, name...)
	body = append(body, 0, 0, 0)
	body = append(body, subname...)
	for _, value := range []int64{creature.StaticFlags, creature.Type, creature.BeastFamily} {
		put(value)
	}
	return packet.Encode(packet.SMSGCreatureQueryResponse, body)
}

func (s *WorldServer) gameObjectQuery(data []byte) ([]byte, error) {
	if len(data) < 12 || binary.LittleEndian.Uint64(data[4:]) == 0 {
		return nil, nil
	}
	gameObject, found, err := s.WorldData.GameObjectTemplate(int64(binary.LittleEndian.Uint32(data)))
	if err != nil || !found {
		return nil, err
	}
	return gameObjectQueryPacket(gameObject)
}

func gameObjectQueryPacket(gameObject worlddb.GameObjectTemplate) ([]byte, error) {
	name, err := packet.StringBytes(gameObject.Name)
	if err != nil {
		return nil, err
	}
	body := make([]byte, 0, 64+len(name))
	put := func(value int64) {
		encoded := make([]byte, 4)
		binary.LittleEndian.PutUint32(encoded, uint32(value))
		body = append(body, encoded...)
	}
	for _, value := range []int64{gameObject.Entry, gameObject.Type, gameObject.DisplayID} {
		put(value)
	}
	body = append(body, name...)
	body = append(body, 0, 0, 0)
	for _, value := range gameObject.Data {
		put(value)
	}
	return packet.Encode(packet.SMSGGameObjectQueryResponse, body)
}

func (s *WorldServer) questQuery(data []byte) ([][]byte, error) {
	if len(data) < 4 {
		return nil, nil
	}
	quest, found, err := s.WorldData.QuestTemplate(int64(binary.LittleEndian.Uint32(data)))
	if err != nil || !found {
		return nil, err
	}
	body := make([]byte, 0, 512)
	put := func(value int64) {
		encoded := make([]byte, 4)
		binary.LittleEndian.PutUint32(encoded, uint32(value))
		body = append(body, encoded...)
	}
	putFloat := func(value float32) {
		encoded := make([]byte, 4)
		binary.LittleEndian.PutUint32(encoded, math.Float32bits(value))
		body = append(body, encoded...)
	}
	for _, value := range []int64{quest.Entry, quest.Method, quest.QuestLevel, quest.ZoneOrSort, quest.Type, quest.NextQuestInChain, quest.RewOrReqMoney, quest.SrcItemID} {
		put(value)
	}
	for index := range quest.RewItemIDs {
		put(quest.RewItemIDs[index])
		put(quest.RewItemCounts[index])
	}
	for index := range quest.RewChoiceItemIDs {
		put(quest.RewChoiceItemIDs[index])
		put(quest.RewChoiceItemCounts[index])
	}
	put(quest.PointMapID)
	putFloat(quest.PointX)
	putFloat(quest.PointY)
	put(quest.PointOpt)
	for _, text := range []string{quest.Title, quest.Details, quest.Objectives, quest.EndText} {
		encoded, err := packet.StringBytes(text)
		if err != nil {
			return nil, err
		}
		body = append(body, encoded...)
	}
	responses := make([][]byte, 0, 2)
	for index := range quest.ReqCreatureOrGOIDs {
		objectID := quest.ReqCreatureOrGOIDs[index]
		name := []byte{0}
		if objectID != 0 {
			var objectPacket []byte
			if objectID < 0 {
				object, found, err := s.WorldData.GameObjectTemplate(-objectID)
				if err != nil {
					return nil, err
				}
				if !found {
					return nil, fmt.Errorf("required gameobject template %d not found", -objectID)
				}
				name, err = packet.StringBytes(object.Name)
				if err != nil {
					return nil, err
				}
				objectPacket, err = gameObjectQueryPacket(object)
				if err != nil {
					return nil, err
				}
			} else {
				object, found, err := s.WorldData.CreatureTemplate(objectID)
				if err != nil {
					return nil, err
				}
				if !found {
					return nil, fmt.Errorf("required creature template %d not found", objectID)
				}
				name, err = packet.StringBytes(object.Name)
				if err != nil {
					return nil, err
				}
				objectPacket, err = creatureQueryPacket(object)
				if err != nil {
					return nil, err
				}
			}
			responses = append(responses, objectPacket)
		}
		encodedID := objectID
		if encodedID < 0 {
			encodedID = -encodedID | int64(0x80000000)
		}
		for _, value := range []int64{encodedID, quest.ReqCreatureOrGOCounts[index], quest.ReqItemIDs[index], quest.ReqItemCounts[index]} {
			put(value)
		}
		body = append(body, name...)
	}
	response, err := packet.Encode(packet.SMSGQuestQueryResponse, body)
	if err != nil {
		return nil, err
	}
	return append(responses, response), nil
}

func formatGameText(text string, character realm.Character) string {
	for _, replacement := range []struct{ old, value string }{{"$B", "\n"}, {"$b", "\n"}, {"$N", character.Name}, {"$n", character.Name}, {"$R", raceText(character.Race)}, {"$r", strings.ToLower(raceText(character.Race))}, {"$C", classText(character.Class)}, {"$c", strings.ToLower(classText(character.Class))}} {
		text = strings.ReplaceAll(text, replacement.old, replacement.value)
	}
	for start := strings.Index(text, "$g"); start >= 0; {
		rest := text[start+2:]
		end := strings.IndexByte(rest, ';')
		if end < 0 {
			break
		}
		choices := strings.Split(rest[:end], ":")
		replacement := ""
		if int(character.Gender) < len(choices) {
			replacement = strings.TrimSpace(choices[character.Gender])
		}
		text = text[:start] + replacement + rest[end+1:]
		start = strings.Index(text, "$g")
	}
	return text
}

func raceText(race uint8) string {
	return map[uint8]string{1: "Human", 2: "Orc", 3: "Dwarf", 4: "Night Elf", 5: "Undead", 6: "Tauren", 7: "Gnome", 8: "Troll"}[race]
}

func classText(class uint8) string {
	return map[uint8]string{1: "Warrior", 2: "Paladin", 3: "Hunter", 4: "Rogue", 5: "Priest", 7: "Shaman", 8: "Mage", 9: "Warlock", 11: "Druid"}[class]
}
