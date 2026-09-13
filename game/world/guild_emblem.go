package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const guildEmblemCost int64 = 100000

func (s *WorldServer) guildSaveEmblem(active *realm.Character, data []byte) ([]byte, error) {
	if active == nil || len(data) < 20 {
		return nil, nil
	}
	result := uint32(0)
	state := s.guilds.forPlayer(active.GUID)
	if state == nil {
		result = 2
	} else if state.Members[active.GUID] != 0 {
		result = 3
	} else if active.Money <= guildEmblemCost {
		result = 4
	} else {
		values := [5]int64{int64(binary.LittleEndian.Uint32(data)), int64(binary.LittleEndian.Uint32(data[4:])), int64(binary.LittleEndian.Uint32(data[8:])), int64(binary.LittleEndian.Uint32(data[12:])), int64(binary.LittleEndian.Uint32(data[16:]))}
		if err := s.Characters.UpdateGuildEmblem(state.Guild.ID, values[0], values[1], values[2], values[3], values[4]); err != nil {
			return nil, err
		}
		if err := s.Characters.UpdateMoney(active.GUID, active.AccountID, active.RealmID, active.Money-guildEmblemCost); err != nil {
			return nil, err
		}
		state.Guild.EmblemStyle, state.Guild.EmblemColor, state.Guild.BorderStyle = values[0], values[1], values[2]
		state.Guild.BorderColor, state.Guild.BackgroundColor = values[3], values[4]
		active.Money -= guildEmblemCost
		s.updatePlayer(*active)
		if query, err := guildQueryState(state); err != nil {
			return nil, err
		} else {
			s.broadcastGuild(state, query, 0)
		}
	}
	return packet.Encode(packet.MSGSaveGuildEmblem, encodeUint32(int64(result)))
}
