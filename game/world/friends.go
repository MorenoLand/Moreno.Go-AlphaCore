package world

import (
	"encoding/binary"
	"fmt"
	"strings"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const (
	friendListFull     byte = 0x01
	friendOnline       byte = 0x02
	friendOffline      byte = 0x03
	friendNotFound     byte = 0x04
	friendRemoved      byte = 0x05
	friendAddedOnline  byte = 0x06
	friendAddedOffline byte = 0x07
	friendAlready      byte = 0x08
	friendSelf         byte = 0x09
	friendEnemy        byte = 0x0a
	ignoreListFull     byte = 0x0b
	ignoreSelf         byte = 0x0c
	ignoreNotFound     byte = 0x0d
	ignoreAlready      byte = 0x0e
	ignoreAdded        byte = 0x0f
	ignoreRemoved      byte = 0x10
)

func (s *WorldServer) friendList(active realm.Character) ([][]byte, error) {
	social, err := s.Characters.Social(active.GUID)
	if err != nil {
		return nil, err
	}
	friendCount, ignoreCount := socialCount(social, false), socialCount(social, true)
	if friendCount > 255 || ignoreCount > 255 {
		return nil, fmt.Errorf("social list exceeds protocol limit")
	}
	friends, ignores := []byte{byte(friendCount)}, []byte{byte(ignoreCount)}
	responses := make([][]byte, 0, 2)
	for _, entry := range social {
		if entry.Ignore {
			ignores = append(ignores, encodeGUID(entry.OtherGUID)...)
			continue
		}
		player, online := s.playerByGUID(entry.OtherGUID)
		if !online {
			friends = append(friends, encodeGUID(entry.OtherGUID)...)
			friends = append(friends, friendOffline)
			continue
		}
		namePacket, err := s.nameQuery(encodeGUID(player.GUID))
		if err != nil {
			return nil, err
		}
		if namePacket != nil {
			responses = append(responses, namePacket)
		}
		zone, err := s.parentZone(player)
		if err != nil {
			return nil, err
		}
		friends = append(friends, encodeGUID(player.GUID)...)
		friends = append(friends, friendOnline)
		for _, value := range []int64{zone, int64(player.Level), int64(player.Class)} {
			friends = append(friends, encodeUint32(value)...)
		}
	}
	friendsPacket, err := packet.Encode(packet.SMSGFriendList, friends)
	if err != nil {
		return nil, err
	}
	ignoresPacket, err := packet.Encode(packet.SMSGIgnoreList, ignores)
	if err != nil {
		return nil, err
	}
	return append(append(responses, friendsPacket), ignoresPacket), nil
}

func (s *WorldServer) friendAdd(active realm.Character, data []byte, ignored bool) ([]byte, error) {
	if len(data) < 1 {
		return nil, nil
	}
	name, err := packet.ReadString(data, 0, 0)
	if err != nil {
		return nil, nil
	}
	name = strings.TrimSpace(name)
	target, isOnline := s.playerByName(name)
	found := isOnline
	if !found {
		target, found, err = s.Characters.CharacterByName(name)
		if err != nil {
			return nil, err
		}
	}
	status := friendNotFound
	if ignored {
		status = ignoreNotFound
	}
	if !found {
		return friendStatusPacket(status, 0)
	}
	targetGUID := target.GUID
	if ignored {
		status = ignoreAdded
	} else if isOnline {
		status = friendAddedOnline
	} else {
		status = friendAddedOffline
	}
	social, err := s.Characters.Social(active.GUID)
	if err != nil {
		return nil, err
	}
	if ignored {
		if active.GUID == targetGUID {
			status = ignoreSelf
		} else if hasSocial(social, targetGUID, true) {
			status = ignoreAlready
		} else if socialCount(social, true) >= 25 {
			status = ignoreListFull
		}
	} else if activeTeam, targetTeam, teamErr := s.teams(active, target); teamErr != nil {
		return nil, teamErr
	} else if activeTeam != targetTeam {
		status = friendEnemy
	} else if active.GUID == targetGUID {
		status = friendSelf
	} else if hasSocial(social, targetGUID, false) {
		status = friendAlready
	} else if socialCount(social, false) >= 50 {
		status = friendListFull
	}
	if status == friendAddedOnline || status == friendAddedOffline || status == ignoreAdded {
		if err := s.Characters.AddSocial(active.GUID, targetGUID, ignored); err != nil {
			return nil, err
		}
	}
	return friendStatusPacket(status, targetGUID)
}

func (s *WorldServer) friendDelete(active realm.Character, data []byte, ignored bool) ([]byte, error) {
	if len(data) != 8 {
		return nil, nil
	}
	targetGUID := int64(binary.LittleEndian.Uint64(data))
	social, err := s.Characters.Social(active.GUID)
	if err != nil {
		return nil, err
	}
	status := friendNotFound
	if ignored {
		status = ignoreNotFound
	}
	if hasSocial(social, targetGUID, ignored) {
		if err := s.Characters.DeleteSocial(active.GUID, targetGUID, ignored); err != nil {
			return nil, err
		}
		status = friendRemoved
		if ignored {
			status = ignoreRemoved
		}
	}
	return friendStatusPacket(status, targetGUID)
}

func friendStatusPacket(status byte, guid int64) ([]byte, error) {
	return packet.Encode(packet.SMSGFriendStatus, append([]byte{status}, encodeGUID(guid)...))
}

func (s *WorldServer) playerByGUID(guid int64) (realm.Character, bool) {
	for _, player := range s.onlinePlayers() {
		if player.GUID == guid {
			return player, true
		}
	}
	return realm.Character{}, false
}

func (s *WorldServer) playerByName(name string) (realm.Character, bool) {
	for _, player := range s.onlinePlayers() {
		if strings.EqualFold(player.Name, name) {
			return player, true
		}
	}
	return realm.Character{}, false
}

func (s *WorldServer) teams(first, second realm.Character) (int64, int64, error) {
	if s.DBC == nil {
		return 0, 0, nil
	}
	firstRace, firstFound, err := s.DBC.Race(first.Race)
	if err != nil {
		return 0, 0, err
	}
	secondRace, secondFound, err := s.DBC.Race(second.Race)
	if err != nil {
		return 0, 0, err
	}
	firstTeam, secondTeam := int64(0), int64(0)
	if firstFound {
		firstTeam = firstRace.BaseLanguage
	}
	if secondFound {
		secondTeam = secondRace.BaseLanguage
	}
	return firstTeam, secondTeam, nil
}

func (s *WorldServer) parentZone(player realm.Character) (int64, error) {
	if s.DBC == nil {
		return player.Zone, nil
	}
	area, found, err := s.DBC.AreaByIDAndMap(player.Zone, player.Map)
	if err != nil || !found || area.ParentAreaNum <= 0 {
		return player.Zone, err
	}
	parent, found, err := s.DBC.AreaByAreaNumber(area.ParentAreaNum, player.Map)
	if err != nil || !found {
		return player.Zone, err
	}
	return parent.ID, nil
}

func hasSocial(entries []realm.Social, guid int64, ignored bool) bool {
	for _, entry := range entries {
		if entry.OtherGUID == guid && entry.Ignore == ignored {
			return true
		}
	}
	return false
}

func socialCount(entries []realm.Social, ignored bool) int {
	count := 0
	for _, entry := range entries {
		if entry.Ignore == ignored {
			count++
		}
	}
	return count
}

func encodeGUID(guid int64) []byte {
	encoded := make([]byte, 8)
	binary.LittleEndian.PutUint64(encoded, uint64(guid))
	return encoded
}

func encodeUint32(value int64) []byte {
	encoded := make([]byte, 4)
	binary.LittleEndian.PutUint32(encoded, uint32(value))
	return encoded
}
