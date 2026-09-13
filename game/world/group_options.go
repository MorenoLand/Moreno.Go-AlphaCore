package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func (s *WorldServer) groupLootMethod(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 12 {
		return nil, nil
	}
	group := s.groupFor(active.GUID)
	if group == nil {
		return s.partyLeaveResult(active, "", partyNotInGroup)
	}
	if group.LeaderGUID != active.GUID {
		return s.partyResult(active, "", partyNotLeader)
	}
	group.LootMethod = byte(binary.LittleEndian.Uint32(data))
	group.LootMaster = 0
	lootMaster := int64(binary.LittleEndian.Uint64(data[4:12]))
	if lootMaster > 0 {
		if _, found := s.playerByGUID(lootMaster); found {
			group.LootMaster = lootMaster
		} else if _, found, err := s.Characters.CharacterByGUID(lootMaster); err != nil {
			return nil, err
		} else if found {
			group.LootMaster = lootMaster
		}
	}
	if group.ID > 0 {
		if err := s.Characters.UpdateGroupLoot(group.ID, int64(group.LootMethod), group.LootMaster); err != nil {
			return nil, err
		}
	}
	return nil, s.sendGroupLists(group)
}

func (s *WorldServer) minimapPing(active realm.Character, data []byte) error {
	if len(data) < 8 {
		return nil
	}
	group := s.groupFor(active.GUID)
	if group == nil {
		return nil
	}
	body := append(encodeGUID(active.GUID), data[:8]...)
	ping, err := packet.Encode(packet.MSGMinimapPing, body)
	if err != nil {
		return err
	}
	for _, member := range group.Members {
		s.sendPlayer(member, ping)
	}
	return nil
}
