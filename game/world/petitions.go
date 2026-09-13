package world

import (
	"encoding/binary"
	"strings"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const (
	petitionCharterEntry   int64  = 5863
	petitionCharterDisplay int64  = 9199
	petitionCharterCost    int64  = 1000
	petitionMaxSignatures         = 9
	petitionNPCFlag        int64  = 0x80
	petitionSuccess        uint32 = 0
	petitionAlreadySigned  uint32 = 1
	petitionAlreadyInGuild uint32 = 2
	petitionCharterCreator uint32 = 3
	petitionNotEnoughSigs  uint32 = 4
	petitionUnknown        uint32 = 5
)

func petitionItemGUID(guid uint64) int64 {
	return int64(guid &^ uint64(0x4000000000000000))
}

func (s *WorldServer) petitionNPC(active realm.Character, guid uint64) (bool, error) {
	_, creature, found, err := s.creatureAt(active, guid, maxShopDistance)
	return found && creature.NPCFlags&petitionNPCFlag != 0, err
}

func (s *WorldServer) petitionShowList(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 || active.Health <= 0 || s.WorldData == nil {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	valid, err := s.petitionNPC(active, guid)
	if err != nil || !valid {
		return nil, err
	}
	body := append(encodeGUID(int64(guid)), 1)
	for _, value := range []int64{1, petitionCharterEntry, petitionCharterDisplay, petitionCharterCost, 1} {
		body = append(body, encodeUint32(value)...)
	}
	response, err := packet.Encode(packet.SMSGPetitionShowlist, body)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) petitionBuy(active *realm.Character, data []byte) ([][]byte, error) {
	if active == nil || len(data) < 8 || active.Health <= 0 || s.Characters == nil || s.WorldData == nil {
		return nil, nil
	}
	npcGUID := binary.LittleEndian.Uint64(data)
	offset := 8
	if len(data) >= 20 {
		offset = 20
	}
	name, err := packet.ReadString(data, offset, 0)
	if err != nil {
		return nil, nil
	}
	_, creature, found, err := s.creatureAt(*active, npcGUID, maxShopDistance)
	if err != nil {
		return nil, err
	}
	if !found {
		return s.buyFailure(*active, uint32(petitionCharterEntry), npcGUID, 5, 1)
	}
	if creature.NPCFlags&petitionNPCFlag == 0 {
		return s.buyFailure(*active, uint32(petitionCharterEntry), npcGUID, 11, 1)
	}
	if s.guilds.forPlayer(active.GUID) != nil {
		return s.guildCommand(*active, guildCreateCommand, "", guildAlready)
	}
	name = strings.TrimSpace(name)
	if !validGuildName(name) {
		return s.guildCommand(*active, guildCreateCommand, "", guildNameInvalid)
	}
	if existing, found, err := s.Characters.PetitionByName(name); err != nil {
		return nil, err
	} else if found || existing.ID > 0 {
		return s.guildCommand(*active, guildCreateCommand, name, guildNameExists)
	}
	if petition, found, err := s.Characters.PetitionByOwner(active.GUID); err != nil {
		return nil, err
	} else if found || petition.ID > 0 {
		return s.buyFailure(*active, uint32(petitionCharterEntry), npcGUID, 8, 1)
	}
	count, err := s.Characters.ItemCount(active.GUID, petitionCharterEntry)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return s.buyFailure(*active, uint32(petitionCharterEntry), npcGUID, 8, 1)
	}
	if active.Money < petitionCharterCost {
		return s.buyFailure(*active, uint32(petitionCharterEntry), npcGUID, 2, 1)
	}
	slot, err := s.Characters.FirstEmptySlot(active.GUID, 23, 23, 39)
	if err != nil {
		return nil, err
	}
	if slot < 0 {
		return nil, nil
	}
	template, found, err := s.WorldData.ItemTemplate(petitionCharterEntry)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	item, err := s.Characters.CreateInventoryItem(active.GUID, 0, 23, slot, petitionCharterEntry, 1)
	if err != nil {
		return nil, err
	}
	if _, err := s.Characters.CreatePetition(active.GUID, item.GUID, name); err != nil {
		return nil, err
	}
	active.Money -= petitionCharterCost
	if err := s.Characters.UpdateMoney(active.GUID, active.AccountID, active.RealmID, active.Money); err != nil {
		return nil, err
	}
	s.updatePlayer(*active)
	queries, err := itemQueryPackets([]worlddb.ItemTemplate{template})
	if err != nil {
		return nil, err
	}
	created, err := packet.EncodeItemCreate(uint64(item.GUID)|0x4000000000000000, uint32(item.ItemTemplate), uint64(item.Owner), 0, 1, 0, encodedItemFlags(template, item.Flags), item.SpellCharges, packet.Movement{X: active.PositionX, Y: active.PositionY, Z: active.PositionZ, O: active.Orientation})
	if err != nil {
		return nil, err
	}
	push, err := itemPushResult(item, petitionCharterEntry, 23)
	if err != nil {
		return nil, err
	}
	money, err := packet.EncodeFieldUpdate(uint64(active.GUID), 55, uint32(active.Money))
	if err != nil {
		return nil, err
	}
	return append(queries, created, push, money), nil
}

func (s *WorldServer) petitionSignatures(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 || s.Characters == nil {
		return nil, nil
	}
	itemGUID := petitionItemGUID(binary.LittleEndian.Uint64(data))
	item, found, err := s.Characters.ItemByGUID(active.GUID, itemGUID)
	if err != nil || !found {
		return nil, err
	}
	petition, found, err := s.Characters.PetitionByItemGUID(item.GUID)
	if err != nil || !found || petition.OwnerGUID != active.GUID {
		return nil, err
	}
	signers, err := s.Characters.PetitionSigners(petition.ID)
	if err != nil {
		return nil, err
	}
	body := append(encodeUint64(uint64(petition.ItemGUID)), encodeUint64(uint64(petition.OwnerGUID))...)
	body = append(body, encodeUint32(petition.ID)...)
	body = append(body, byte(len(signers)))
	for _, signer := range signers {
		body = append(body, encodeUint64(uint64(signer))...)
		body = append(body, encodeUint32(0)...)
	}
	response, err := packet.Encode(packet.SMSGPetitionShowSignatures, body)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) petitionQuery(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 12 || s.Characters == nil {
		return nil, nil
	}
	id, itemGUID := int64(binary.LittleEndian.Uint32(data)), petitionItemGUID(binary.LittleEndian.Uint64(data[4:]))
	if id <= 0 || itemGUID <= 0 {
		return nil, nil
	}
	petition, found, err := s.Characters.PetitionByItemGUID(itemGUID)
	if err != nil || !found || petition.ID != id {
		return nil, err
	}
	name, err := packet.StringBytes(petition.Name)
	if err != nil {
		return nil, err
	}
	body := append(encodeUint32(petition.ID), encodeUint64(uint64(petition.OwnerGUID))...)
	body = append(body, name...)
	body = append(body, 0)
	for _, value := range []int64{1, petitionMaxSignatures, petitionMaxSignatures, 0, 0, 0, 0, 0} {
		body = append(body, encodeUint32(value)...)
	}
	body = append(body, 0, 0)
	for range 4 {
		body = append(body, encodeUint32(0)...)
	}
	response, err := packet.Encode(packet.SMSGPetitionQueryResponse, body)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func petitionResult(opcode packet.Opcode, result uint32) ([]byte, error) {
	return packet.Encode(opcode, encodeUint32(int64(result)))
}

func (s *WorldServer) petitionSign(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 || s.Characters == nil {
		return nil, nil
	}
	itemGUID := petitionItemGUID(binary.LittleEndian.Uint64(data))
	petition, found, err := s.Characters.PetitionByItemGUID(itemGUID)
	if err != nil {
		return nil, err
	}
	if !found {
		response, responseErr := petitionResult(packet.SMSGPetitionSignResults, petitionUnknown)
		return [][]byte{response}, responseErr
	}
	result := petitionSuccess
	if petition.OwnerGUID == active.GUID {
		result = petitionCharterCreator
	} else if s.guilds.forPlayer(active.GUID) != nil {
		result = petitionAlreadyInGuild
	} else if invited := s.guilds.pending(active.GUID); invited != nil {
		return s.guildCommand(active, guildInviteCommand, active.Name, guildAlready)
	} else if signed, signErr := s.Characters.PetitionSignerExists(petition.ID, active.GUID); signErr != nil {
		return nil, signErr
	} else if signed {
		result = petitionAlreadySigned
	} else if signers, signErr := s.Characters.PetitionSigners(petition.ID); signErr != nil {
		return nil, signErr
	} else if len(signers) >= petitionMaxSignatures {
		return nil, nil
	} else {
		owner, ownerFound, ownerErr := s.Characters.CharacterByGUID(petition.OwnerGUID)
		if ownerErr != nil {
			return nil, ownerErr
		}
		if ownerFound {
			first, second, teamErr := s.teams(owner, active)
			if teamErr != nil {
				return nil, teamErr
			}
			if first != 0 && second != 0 && first != second {
				return s.guildCommand(active, guildCreateCommand, "", guildNotAllied)
			}
		}
		if err := s.Characters.AddPetitionSigner(petition.ID, active.GUID); err != nil {
			return nil, err
		}
	}
	response, err := petitionResult(packet.SMSGPetitionSignResults, result)
	if err != nil {
		return nil, err
	}
	if result == petitionSuccess {
		if _, online := s.playerByGUID(petition.OwnerGUID); online {
			s.sendPlayer(petition.OwnerGUID, response)
		}
	}
	return [][]byte{response}, nil
}

func (s *WorldServer) petitionOffer(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 16 || s.Characters == nil {
		return nil, nil
	}
	itemGUID, targetGUID := petitionItemGUID(binary.LittleEndian.Uint64(data)), int64(binary.LittleEndian.Uint64(data[8:]))
	item, found, err := s.Characters.ItemByGUID(active.GUID, itemGUID)
	if err != nil || !found {
		response, responseErr := petitionResult(packet.SMSGPetitionSignResults, petitionUnknown)
		return [][]byte{response}, responseErr
	}
	petition, found, err := s.Characters.PetitionByItemGUID(item.GUID)
	if err != nil || !found || petition.OwnerGUID != active.GUID {
		response, responseErr := petitionResult(packet.SMSGPetitionSignResults, petitionUnknown)
		return [][]byte{response}, responseErr
	}
	target, found := s.playerByGUID(targetGUID)
	if !found {
		return s.guildCommand(active, guildInviteCommand, "Player", guildPlayerMissing)
	}
	first, second, err := s.teams(active, target)
	if err != nil {
		return nil, err
	}
	if first != 0 && second != 0 && first != second {
		return s.guildCommand(active, guildInviteCommand, "", guildNotAllied)
	}
	if s.guilds.forPlayer(target.GUID) != nil {
		return s.guildCommand(active, guildInviteCommand, target.Name, guildAlready)
	}
	if s.guilds.pending(target.GUID) != nil {
		return s.guildCommand(active, guildInviteCommand, target.Name, guildAlready)
	}
	signatures, err := s.petitionSignatures(active, encodeUint64(uint64(item.GUID)))
	if err != nil {
		return nil, err
	}
	if len(signatures) > 0 {
		s.sendPlayer(target.GUID, signatures[0])
	}
	return nil, nil
}

func (s *WorldServer) petitionTurnIn(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 || s.Characters == nil {
		return nil, nil
	}
	itemGUID := petitionItemGUID(binary.LittleEndian.Uint64(data))
	item, found, err := s.Characters.ItemByGUID(active.GUID, itemGUID)
	if err != nil || !found {
		response, responseErr := petitionResult(packet.SMSGTurnInPetitionResults, petitionUnknown)
		return [][]byte{response}, responseErr
	}
	petition, found, err := s.Characters.PetitionByItemGUID(item.GUID)
	if err != nil || !found {
		response, responseErr := petitionResult(packet.SMSGTurnInPetitionResults, petitionUnknown)
		return [][]byte{response}, responseErr
	}
	if s.guilds.forPlayer(active.GUID) != nil {
		response, responseErr := petitionResult(packet.SMSGTurnInPetitionResults, petitionAlreadyInGuild)
		return [][]byte{response}, responseErr
	}
	if petition.OwnerGUID != active.GUID {
		return nil, nil
	}
	signers, err := s.Characters.PetitionSigners(petition.ID)
	if err != nil {
		return nil, err
	}
	if len(signers) < petitionMaxSignatures {
		response, responseErr := petitionResult(packet.SMSGTurnInPetitionResults, petitionNotEnoughSigs)
		return [][]byte{response}, responseErr
	}
	if existing, found, err := s.Characters.GuildByName(petition.Name); err != nil {
		return nil, err
	} else if found || existing.ID > 0 {
		return s.guildCommand(active, guildCreateCommand, petition.Name, guildNameExists)
	}
	guild, err := s.Characters.CreateGuild(petition.Name, "", active.GUID)
	if err != nil {
		return nil, err
	}
	if err := s.Characters.AddGuildMember(guild.ID, active.GUID, 0); err != nil {
		return nil, err
	}
	state := s.guilds.created(guild, active.GUID)
	for _, signer := range signers {
		if err := s.Characters.AddGuildMember(guild.ID, signer, 4); err != nil {
			return nil, err
		}
		s.guilds.addPlayer(state, signer, 4)
	}
	if err := s.Characters.DeletePetition(petition.ID); err != nil {
		return nil, err
	}
	if err := s.Characters.DeleteItem(item.GUID, active.GUID); err != nil {
		return nil, err
	}
	result, err := petitionResult(packet.SMSGTurnInPetitionResults, petitionSuccess)
	if err != nil {
		return nil, err
	}
	destroy, err := packet.Encode(packet.SMSGDestroyObject, encodeGUID(int64(uint64(item.GUID)|0x4000000000000000)))
	if err != nil {
		return nil, err
	}
	return [][]byte{result, destroy}, nil
}
