package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const (
	trainerServiceAvailable   uint32 = 0
	trainerServiceUnavailable uint32 = 1
	trainerServiceUsed        uint32 = 2
	trainerFailUnavailable    uint32 = 0
	trainerFailMoney          uint32 = 1
	trainerFailPoints         uint32 = 2
)

var trainerLockpickingSpells = map[int64]bool{1804: true, 6461: true, 6463: true}

func (s *WorldServer) trainerCanTrain(active realm.Character, creature worlddb.CreatureTemplate) bool {
	return creature.TrainerClass <= 0 || creature.TrainerClass == 4 || creature.TrainerClass == int64(active.Class)
}

func (s *WorldServer) trainerRequirementLevel(spell worlddb.TrainerSpell) (int64, error) {
	if spell.ReqLevel > 0 {
		return spell.ReqLevel, nil
	}
	level, _, err := s.DBC.SpellBaseLevel(spell.PlayerSpell)
	return level, err
}

func (s *WorldServer) trainerSpellStatus(active realm.Character, spell worlddb.TrainerSpell, known map[int64]bool) (uint32, error) {
	if known[spell.PlayerSpell] {
		return trainerServiceUsed, nil
	}
	requiredLevel, err := s.trainerRequirementLevel(spell)
	if err != nil {
		return trainerServiceUnavailable, err
	}
	if requiredLevel > int64(active.Level) {
		return trainerServiceUnavailable, nil
	}
	if spell.ReqSkill > 0 {
		value, found, err := s.Characters.SkillValue(active.GUID, spell.ReqSkill)
		if err != nil {
			return trainerServiceUnavailable, err
		}
		if !found || value < spell.ReqSkillValue {
			return trainerServiceUnavailable, nil
		}
	}
	for _, required := range []int64{spell.ReqSpell1, spell.ReqSpell2, spell.ReqSpell3} {
		if required > 0 && !known[required] {
			return trainerServiceUnavailable, nil
		}
	}
	if preceded, found, err := s.DBC.PrecededSpell(spell.PlayerSpell); err != nil {
		return trainerServiceUnavailable, err
	} else if found && preceded > 0 && !known[preceded] {
		return trainerServiceUnavailable, nil
	}
	return trainerServiceAvailable, nil
}

func (s *WorldServer) trainerList(active realm.Character, data []byte) ([][]byte, error) {
	if len(data) < 8 || s.DBC == nil || s.WorldData == nil {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	if guid == uint64(active.GUID) {
		return nil, nil
	}
	_, creature, found, err := s.creatureAt(active, guid, maxShopDistance)
	if err != nil || !found || creature.NPCFlags&0x8 == 0 || creature.TrainerID <= 0 || !s.trainerCanTrain(active, creature) {
		return nil, err
	}
	spells, err := s.WorldData.TrainerSpells(creature.TrainerID)
	if err != nil {
		return nil, err
	}
	learned, err := s.Characters.Spells(active.GUID)
	if err != nil {
		return nil, err
	}
	known := make(map[int64]bool, len(learned))
	for _, spell := range learned {
		known[spell.ID] = true
	}
	spellData := make([]byte, 0, len(spells)*36)
	count := 0
	for _, spell := range spells {
		if spell.PlayerSpell <= 0 {
			continue
		}
		if exists, err := s.DBC.SpellExists(spell.PlayerSpell); err != nil || !exists {
			if err != nil {
				return nil, err
			}
			continue
		}
		if creature.TrainerClass == 4 && int64(active.Class) != creature.TrainerClass && !trainerLockpickingSpells[spell.PlayerSpell] {
			continue
		}
		status, err := s.trainerSpellStatus(active, spell, known)
		if err != nil {
			return nil, err
		}
		if requiredLevel, levelErr := s.trainerRequirementLevel(spell); levelErr != nil {
			return nil, levelErr
		} else {
			spell.ReqLevel = requiredLevel
		}
		spellData = append(spellData, trainerSpellData(spell, status)...)
		count++
	}
	greeting := "Hello, $c!  Ready for some training?"
	if value, found, err := s.WorldData.TrainerGreeting(creature.Entry); err != nil {
		return nil, err
	} else if found {
		greeting = value.Content
	}
	greetingBytes, err := packet.StringBytes(greeting)
	if err != nil {
		return nil, err
	}
	body := append(encodeGUID(int64(guid)), encodeUint32(creature.TrainerType)...)
	body = append(body, encodeUint32(int64(count))...)
	body = append(body, spellData...)
	body = append(body, greetingBytes...)
	response, err := packet.Encode(packet.SMSGTrainerList, body)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}

func trainerSpellData(spell worlddb.TrainerSpell, status uint32) []byte {
	data := make([]byte, 0, 33)
	data = append(data, encodeUint32(spell.Spell)...)
	data = append(data, byte(status))
	data = append(data, encodeUint32(spell.SpellCost)...)
	data = append(data, byte(spell.TalentPointCost), byte(spell.SkillPointCost), byte(spell.ReqLevel))
	data = append(data, encodeUint32(spell.ReqSkill)...)
	data = append(data, encodeUint32(spell.ReqSkillValue)...)
	data = append(data, encodeUint32(0)...)
	data = append(data, encodeUint32(spell.ReqSpell1)...)
	data = append(data, encodeUint32(spell.ReqSpell2)...)
	data = append(data, encodeUint32(spell.ReqSpell3)...)
	return data
}

func (s *WorldServer) trainerBuy(active *realm.Character, data []byte) ([][]byte, error) {
	if active == nil || len(data) < 12 || s.DBC == nil || s.WorldData == nil || s.Characters == nil {
		return nil, nil
	}
	guid, trainingSpell := binary.LittleEndian.Uint64(data), int64(binary.LittleEndian.Uint32(data[8:]))
	_, creature, found, err := s.creatureAt(*active, guid, maxShopDistance)
	if err != nil || !found || creature.NPCFlags&0x8 == 0 || creature.TrainerID <= 0 || !s.trainerCanTrain(*active, creature) {
		return trainerBuyFailure(guid, trainingSpell, trainerFailUnavailable)
	}
	spells, err := s.WorldData.TrainerSpells(creature.TrainerID)
	if err != nil {
		return nil, err
	}
	var trainerSpell *worlddb.TrainerSpell
	for index := range spells {
		if spells[index].Spell == trainingSpell {
			trainerSpell = &spells[index]
			break
		}
	}
	if trainerSpell == nil || trainerSpell.PlayerSpell <= 0 {
		return trainerBuyFailure(guid, trainingSpell, trainerFailUnavailable)
	}
	if exists, err := s.DBC.SpellExists(trainerSpell.PlayerSpell); err != nil || !exists {
		return trainerBuyFailure(guid, trainingSpell, trainerFailUnavailable)
	}
	learned, err := s.Characters.Spells(active.GUID)
	if err != nil {
		return nil, err
	}
	for _, spell := range learned {
		if spell.ID == trainerSpell.PlayerSpell {
			return trainerBuyFailure(guid, trainingSpell, trainerFailUnavailable)
		}
	}
	known := make(map[int64]bool, len(learned))
	for _, spell := range learned {
		known[spell.ID] = true
	}
	status, err := s.trainerSpellStatus(*active, *trainerSpell, known)
	if err != nil {
		return nil, err
	}
	if status != trainerServiceAvailable {
		return trainerBuyFailure(guid, trainingSpell, trainerFailUnavailable)
	}
	if trainerSpell.SpellCost > active.Money {
		return trainerBuyFailure(guid, trainingSpell, trainerFailMoney)
	}
	if trainerSpell.SkillPointCost > active.Skillpoints {
		return trainerBuyFailure(guid, trainingSpell, trainerFailPoints)
	}
	if err := s.Characters.AddSpell(active.GUID, trainerSpell.PlayerSpell); err != nil {
		return nil, err
	}
	active.Money -= trainerSpell.SpellCost
	active.Skillpoints -= trainerSpell.SkillPointCost
	if err := s.Characters.UpdateMoney(active.GUID, active.AccountID, active.RealmID, active.Money); err != nil {
		return nil, err
	}
	if err := s.Characters.UpdateSkillpoints(active.GUID, active.AccountID, active.RealmID, active.Skillpoints); err != nil {
		return nil, err
	}
	s.updatePlayer(*active)
	succeeded, err := packet.Encode(packet.SMSGTrainerBuySucceeded, append(encodeGUID(int64(guid)), encodeUint32(trainingSpell)...))
	if err != nil {
		return nil, err
	}
	money, err := packet.EncodeFieldUpdate(uint64(active.GUID), 55, uint32(active.Money))
	if err != nil {
		return nil, err
	}
	return [][]byte{succeeded, money}, nil
}

func trainerBuyFailure(guid uint64, spell int64, reason uint32) ([][]byte, error) {
	body := append(encodeGUID(int64(guid)), encodeUint32(spell)...)
	body = append(body, encodeUint32(int64(reason))...)
	response, err := packet.Encode(packet.SMSGTrainerBuyFailed, body)
	if err != nil {
		return nil, err
	}
	return [][]byte{response}, nil
}
