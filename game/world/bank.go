package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const (
	bankSlotFailedTooMany uint32 = 0
	bankSlotInsufficient  uint32 = 1
	bankSlotNotBanker     uint32 = 2
	bankSlotOK            uint32 = 3
)

func (s *WorldServer) buyBankSlot(active *realm.Character, data []byte) ([][]byte, error) {
	if active == nil || len(data) < 8 || active.Health <= 0 || s.DBC == nil || s.WorldData == nil || s.Characters == nil {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	_, creature, found, err := s.creatureAt(*active, guid, maxShopDistance)
	result := bankSlotOK
	if err != nil {
		return nil, err
	}
	if !found || creature.NPCFlags&0x20 == 0 {
		result = bankSlotNotBanker
	}
	var cost int64
	if result == bankSlotOK {
		cost, found, err = s.DBC.BankSlotCost(int64(active.Bankslots) + 1)
		if err != nil {
			return nil, err
		}
		if !found {
			result = bankSlotFailedTooMany
		} else if active.Money < cost {
			result = bankSlotInsufficient
		}
	}
	if result != bankSlotOK {
		response, err := packet.Encode(packet.SMSGBuyBankSlotResult, encodeUint32(int64(result)))
		if err != nil {
			return nil, err
		}
		return [][]byte{response}, nil
	}
	active.Bankslots++
	active.Money -= cost
	if err := s.Characters.UpdateBankslots(active.GUID, active.AccountID, active.RealmID, active.Bankslots); err != nil {
		return nil, err
	}
	if err := s.Characters.UpdateMoney(active.GUID, active.AccountID, active.RealmID, active.Money); err != nil {
		return nil, err
	}
	s.updatePlayer(*active)
	bankSlots, err := packet.EncodeFieldUpdate(uint64(active.GUID), 526, uint32(active.Bankslots)<<16)
	if err != nil {
		return nil, err
	}
	money, err := packet.EncodeFieldUpdate(uint64(active.GUID), 55, uint32(active.Money))
	if err != nil {
		return nil, err
	}
	return [][]byte{bankSlots, money}, nil
}
