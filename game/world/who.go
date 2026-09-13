package world

import (
	"encoding/binary"
	"strings"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func readCString(data []byte, start int) (string, int, error) {
	value, err := packet.ReadString(data, start, 0)
	if err != nil {
		return "", 0, err
	}
	return value, start + len(value) + 1, nil
}

func (s *WorldServer) who(active realm.Character, data []byte) ([]byte, error) {
	if len(data) < 26 {
		return nil, nil
	}
	levelMin := binary.LittleEndian.Uint32(data)
	levelMax := binary.LittleEndian.Uint32(data[4:])
	name, offset, err := readCString(data, 8)
	if err != nil {
		return nil, nil
	}
	guild, offset, err := readCString(data, offset)
	if err != nil || len(data) < offset+12 {
		return nil, nil
	}
	raceMask := binary.LittleEndian.Uint32(data[offset:])
	classMask := binary.LittleEndian.Uint32(data[offset+4:])
	zoneCount := binary.LittleEndian.Uint32(data[offset+8:])
	if zoneCount > 10 {
		return nil, nil
	}
	offset += 12
	zones := make([]uint32, zoneCount)
	for index := range zones {
		if len(data) < offset+4 {
			return nil, nil
		}
		zones[index] = binary.LittleEndian.Uint32(data[offset:])
		if zones[index] == 0 {
			zones[index] = uint32(active.Zone)
		}
		offset += 4
	}
	if len(data) < offset+4 {
		return nil, nil
	}
	userStringCount := binary.LittleEndian.Uint32(data[offset:])
	if userStringCount > 4 {
		return nil, nil
	}
	offset += 4
	userStrings := make([]string, userStringCount)
	for index := range userStrings {
		userStrings[index], offset, err = readCString(data, offset)
		if err != nil {
			return nil, nil
		}
	}
	online := s.onlinePlayers()
	result := make([]byte, 8)
	matched := uint32(0)
	for _, player := range online {
		if matched == 49 {
			continue
		}
		matches, err := s.whoMatches(player, name, guild, levelMin, levelMax, raceMask, classMask, zones, userStrings)
		if err != nil {
			return nil, err
		}
		if !matches {
			continue
		}
		playerName, err := packet.StringBytes(player.Name)
		if err != nil {
			return nil, err
		}
		guildName, err := packet.StringBytes("")
		if err != nil {
			return nil, err
		}
		result = append(result, playerName...)
		result = append(result, guildName...)
		for _, value := range []uint32{uint32(player.Level), 0, uint32(player.Race), uint32(player.Zone), 0} {
			encoded := make([]byte, 4)
			binary.LittleEndian.PutUint32(encoded, value)
			result = append(result, encoded...)
		}
		matched++
	}
	binary.LittleEndian.PutUint32(result, matched)
	count := matched
	if len(online) > 49 {
		count = uint32(len(online))
	}
	binary.LittleEndian.PutUint32(result[4:], count)
	return packet.Encode(packet.SMSGWho, result)
}

func (s *WorldServer) whoMatches(player realm.Character, name, guild string, levelMin, levelMax, raceMask, classMask uint32, zones []uint32, userStrings []string) (bool, error) {
	if uint32(player.Level) < levelMin || uint32(player.Level) > levelMax || (name != "" && !strings.Contains(strings.ToLower(player.Name), strings.ToLower(name))) {
		return false, nil
	}
	if player.Race == 0 || raceMask != 0xffffffff && (raceMask&(1<<uint(player.Race-1))) != (1<<uint(player.Race-1)) {
		return false, nil
	}
	if player.Class == 0 || classMask != 0xffffffff && (classMask&(1<<uint(player.Class-1))) != (1<<uint(player.Class-1)) {
		return false, nil
	}
	if len(zones) > 0 {
		area, found, err := s.DBC.AreaByIDAndMap(player.Zone, player.Map)
		if err != nil {
			return false, err
		}
		if !found {
			return false, nil
		}
		areas := []int64{area.ID}
		if area.ParentAreaNum > 0 {
			parent, found, err := s.DBC.AreaByAreaNumber(area.ParentAreaNum, player.Map)
			if err != nil {
				return false, err
			}
			if found {
				areas = append(areas, parent.ID)
			}
		}
		found = false
		for _, zone := range zones {
			for _, current := range areas {
				if zone == uint32(current) {
					found = true
				}
			}
		}
		if !found {
			return false, nil
		}
	}
	for _, value := range userStrings {
		if strings.Contains(strings.ToLower(player.Name), strings.ToLower(value)) {
			continue
		}
		return false, nil
	}
	return true, nil
}
