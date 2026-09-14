package world

import (
	"bytes"
	"encoding/binary"
	"math"
	"time"

	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const (
	petSummonSpellID  int64 = 883
	petCommandFlag          = uint32(0x01000000)
	petCommandStay    int64 = 0
	petCommandFollow  int64 = 1
	petCommandAttack  int64 = 2
	petCommandDismiss int64 = 3
	petActionBarSize        = 10
	petFollowDistance       = 2
	petFollowAngle          = math.Pi / 2
)

type petRecord struct {
	data   realm.Pet
	spells []int64
}

type activePetState struct {
	GUID     uint64
	Index    int
	Creature *creatureState
}

type petManagerState struct {
	owner     realm.Character
	permanent []petRecord
	active    *activePetState
}

func (s *WorldServer) loadPets(owner realm.Character) error {
	manager := &petManagerState{owner: owner}
	if s.Characters != nil {
		pets, err := s.Characters.Pets(owner.GUID)
		if err != nil {
			return err
		}
		for _, pet := range pets {
			spells, err := s.Characters.PetSpells(owner.GUID, pet.ID)
			if err != nil {
				return err
			}
			manager.permanent = append(manager.permanent, petRecord{data: pet, spells: spells})
		}
	}
	s.petMu.Lock()
	if s.pets == nil {
		s.pets = make(map[int64]*petManagerState)
	}
	s.pets[owner.GUID] = manager
	s.petMu.Unlock()
	return nil
}

func (s *WorldServer) ensurePetManager(owner realm.Character) *petManagerState {
	s.petMu.Lock()
	defer s.petMu.Unlock()
	if s.pets == nil {
		s.pets = make(map[int64]*petManagerState)
	}
	manager := s.pets[owner.GUID]
	if manager == nil {
		manager = &petManagerState{owner: owner}
		s.pets[owner.GUID] = manager
	} else {
		manager.owner = owner
	}
	return manager
}

func (s *WorldServer) initialPetPackets(owner realm.Character) ([][]byte, uint64, error) {
	state, record, found, err := s.activatePet(owner, -1)
	if err != nil || !found || state == nil {
		return nil, 0, err
	}
	create, err := s.petCreatePacket(*state)
	if err != nil {
		return nil, 0, err
	}
	spells, err := petSpellsPacket(state.GUID, record, false)
	if err != nil {
		return nil, 0, err
	}
	return [][]byte{create, spells}, state.GUID, nil
}

func (s *WorldServer) activatePet(owner realm.Character, index int) (*creatureState, petRecord, bool, error) {
	manager := s.ensurePetManager(owner)
	s.petMu.Lock()
	if manager.active != nil {
		active := manager.active
		record := manager.permanent[active.Index]
		state := active.Creature
		s.petMu.Unlock()
		return state, record, true, nil
	}
	if index < 0 {
		for petIndex, record := range manager.permanent {
			if record.data.Active {
				index = petIndex
				break
			}
		}
	}
	if index < 0 || index >= len(manager.permanent) {
		s.petMu.Unlock()
		return nil, petRecord{}, false, nil
	}
	record := manager.permanent[index]
	s.petMu.Unlock()
	state, err := s.buildPetState(owner, record)
	if err != nil || state == nil {
		return nil, record, false, err
	}
	s.petMu.Lock()
	if manager.active != nil {
		active := manager.active
		existing := active.Creature
		existingRecord := manager.permanent[active.Index]
		s.petMu.Unlock()
		s.removeCreature(state.GUID)
		return existing, existingRecord, true, nil
	}
	manager.active = &activePetState{GUID: state.GUID, Index: index, Creature: state}
	s.petMu.Unlock()
	return state, record, true, nil
}

func (s *WorldServer) buildPetState(owner realm.Character, record petRecord) (*creatureState, error) {
	if s.WorldData == nil {
		return nil, nil
	}
	template, found, err := s.WorldData.CreatureTemplate(record.data.CreatureID)
	if err != nil || !found {
		return nil, err
	}
	level := record.data.Level
	if level < 1 {
		level = int64(owner.Level)
	}
	stats, err := s.petCreatureStats(record.data.CreatureID, template.UnitClass, level)
	if err != nil {
		return nil, err
	}
	if s.DBC != nil {
		if race, raceFound, raceErr := s.DBC.Race(owner.Race); raceErr != nil {
			return nil, raceErr
		} else if raceFound {
			template.Faction = race.FactionID
		}
	}
	angle := float64(owner.Orientation) + petFollowAngle
	spawn := worlddb.CreatureSpawn{Entry: record.data.CreatureID, Map: owner.Map, PositionX: owner.PositionX + float32(math.Cos(angle)*petFollowDistance), PositionY: owner.PositionY + float32(math.Sin(angle)*petFollowDistance), PositionZ: owner.PositionZ, Orientation: owner.Orientation, HealthPercent: 100, ManaPercent: 100}
	state := s.newCreatureState(spawn, template, stats, level)
	if record.data.Health > 0 {
		state.Health = record.data.Health
	}
	if state.Health < 1 {
		state.Health = 1
	}
	if state.MaxHealth < state.Health {
		state.MaxHealth = state.Health
	}
	if record.data.Mana > 0 {
		state.Mana = record.data.Mana
	}
	state.GUID = s.nextCreatureGUID()
	state.OwnerGUID = uint64(owner.GUID)
	state.CreatedBySpell = record.data.CreatedBySpell
	state.PetID = record.data.ID
	state.PetNameTimestamp = record.data.RenameTime
	state.PetExperience = record.data.XP
	state.PetNextExperience = petExperienceForLevel(level)
	state.Pet = true
	s.setCreatureState(state)
	return &state, nil
}

func (s *WorldServer) petCreatureStats(entry, class, level int64) (worlddb.CreatureClassLevelStats, error) {
	if stats, found, err := s.WorldData.PetLevelStats(entry, level); err != nil {
		return worlddb.CreatureClassLevelStats{}, err
	} else if found {
		return worlddb.CreatureClassLevelStats{Health: stats.Health, BaseHealth: stats.Health, Mana: stats.Mana, BaseMana: stats.Mana, Strength: stats.Strength, Agility: stats.Agility, Stamina: stats.Stamina, Intellect: stats.Intellect, Spirit: stats.Spirit, Armor: stats.Armor}, nil
	}
	if stats, found, err := s.WorldData.CreatureClassLevelStats(class, level); err != nil {
		return worlddb.CreatureClassLevelStats{}, err
	} else if found {
		return stats, nil
	}
	return worlddb.CreatureClassLevelStats{}, nil
}

func (s *WorldServer) petCreatePacket(state creatureState) ([]byte, error) {
	scale, boundingRadius, combatReach := state.Template.Scale, float32(0), float32(0)
	if info, found, err := s.WorldData.CreatureModelInfo(state.Template.DisplayID1); err != nil {
		return nil, err
	} else if found {
		boundingRadius, combatReach = info.BoundingRadius, info.CombatReach
	}
	if s.DBC != nil {
		if display, found, err := s.DBC.CreatureDisplayInfo(state.Template.DisplayID1); err != nil {
			return nil, err
		} else if found && scale <= 0 {
			scale = display.ModelScale
		}
	}
	return s.creatureCreatePacket(state, scale, boundingRadius, combatReach)
}

func (s *WorldServer) activePetSnapshot(ownerGUID int64, petGUID uint64) (petRecord, activePetState, bool) {
	s.petMu.Lock()
	manager := s.pets[ownerGUID]
	if manager == nil || manager.active == nil || petGUID != 0 && manager.active.GUID != petGUID {
		s.petMu.Unlock()
		return petRecord{}, activePetState{}, false
	}
	active := *manager.active
	record := manager.permanent[active.Index]
	record.spells = append([]int64(nil), record.spells...)
	s.petMu.Unlock()
	return record, active, true
}

func (s *WorldServer) findActivePet(petGUID uint64) (int64, petRecord, activePetState, bool) {
	s.petMu.Lock()
	defer s.petMu.Unlock()
	for ownerGUID, manager := range s.pets {
		if manager.active == nil || manager.active.GUID != petGUID {
			continue
		}
		active := *manager.active
		record := manager.permanent[active.Index]
		record.spells = append([]int64(nil), record.spells...)
		return ownerGUID, record, active, true
	}
	return 0, petRecord{}, activePetState{}, false
}

func (s *WorldServer) updateActivePet(ownerGUID int64, petGUID uint64, update func(*realm.Pet)) (petRecord, activePetState, bool, error) {
	s.petMu.Lock()
	manager := s.pets[ownerGUID]
	if manager == nil || manager.active == nil || manager.active.GUID != petGUID {
		s.petMu.Unlock()
		return petRecord{}, activePetState{}, false, nil
	}
	index := manager.active.Index
	update(&manager.permanent[index].data)
	record := manager.permanent[index]
	record.spells = append([]int64(nil), record.spells...)
	active := *manager.active
	s.petMu.Unlock()
	if s.Characters != nil {
		if err := s.Characters.UpdatePet(record.data); err != nil {
			return record, active, true, err
		}
	}
	return record, active, true, nil
}

func (s *WorldServer) summonPermanentPet(owner realm.Character, spellID, creatureID int64) {
	manager := s.ensurePetManager(owner)
	index := -1
	s.petMu.Lock()
	if manager.active != nil {
		s.petMu.Unlock()
		return
	}
	if spellID == petSummonSpellID {
		if len(manager.permanent) > 0 {
			index, creatureID = 0, manager.permanent[0].data.CreatureID
		}
	} else {
		for petIndex, record := range manager.permanent {
			if record.data.CreatureID == creatureID {
				index = petIndex
				break
			}
		}
	}
	s.petMu.Unlock()
	if index < 0 && creatureID > 0 && s.WorldData != nil {
		if template, found, err := s.WorldData.CreatureTemplate(creatureID); err == nil && found {
			pet := realm.Pet{OwnerGUID: owner.GUID, CreatureID: creatureID, CreatedBySpell: spellID, Level: int64(owner.Level), Name: template.Name, Active: true, ActionBar: defaultPetActionBar()}
			if s.Characters != nil {
				id, err := s.Characters.CreatePet(pet)
				if err != nil {
					return
				}
				pet.ID = id
			}
			s.petMu.Lock()
			manager.permanent = append(manager.permanent, petRecord{data: pet})
			index = len(manager.permanent) - 1
			s.petMu.Unlock()
		}
	}
	if index < 0 {
		return
	}
	s.petMu.Lock()
	manager.permanent[index].data.Active = true
	pet := manager.permanent[index].data
	s.petMu.Unlock()
	if s.Characters != nil {
		_ = s.Characters.UpdatePet(pet)
	}
	state, record, found, err := s.activatePet(owner, index)
	if err != nil || !found {
		return
	}
	if create, err := s.petCreatePacket(*state); err == nil {
		s.sendSpell(owner, create)
	}
	if info, err := petSpellsPacket(state.GUID, record, false); err == nil {
		s.sendPlayer(owner.GUID, info)
	}
}

func (s *WorldServer) petAction(owner realm.Character, data []byte) error {
	if len(data) < 20 {
		return nil
	}
	petGUID := binary.LittleEndian.Uint64(data)
	action, target := binary.LittleEndian.Uint32(data[8:]), binary.LittleEndian.Uint64(data[12:])
	record, active, found := s.activePetSnapshot(owner.GUID, petGUID)
	if !found {
		return nil
	}
	actionID := int64(action & 0xffff)
	if actionID > petCommandDismiss {
		for _, spell := range record.spells {
			if spell == actionID {
				return nil
			}
		}
		return nil
	}
	if action&petCommandFlag != 0 {
		if actionID == petCommandDismiss {
			return s.detachPet(owner.GUID, active.GUID, true)
		}
		_, _, _, err := s.updateActivePet(owner.GUID, active.GUID, func(pet *realm.Pet) { pet.CommandState = actionID })
		if err == nil {
			s.sendPetSpells(owner.GUID)
		}
		return err
	}
	if actionID > petCommandAttack {
		return nil
	}
	_ = target
	_, _, _, err := s.updateActivePet(owner.GUID, active.GUID, func(pet *realm.Pet) { pet.ReactState = actionID })
	if err == nil {
		s.sendPetSpells(owner.GUID)
	}
	return err
}

func (s *WorldServer) petSetAction(owner realm.Character, data []byte) error {
	if len(data) < 16 {
		return nil
	}
	petGUID := binary.LittleEndian.Uint64(data)
	count := 1
	if len(data) == 24 {
		count = 2
	}
	record, active, found := s.activePetSnapshot(owner.GUID, petGUID)
	if !found {
		return nil
	}
	for index := 0; index < count; index++ {
		offset := 8 + index*8
		slot, action := int(binary.LittleEndian.Uint32(data[offset:])), binary.LittleEndian.Uint32(data[offset+4:])
		if slot < 0 || slot >= petActionBarSize {
			continue
		}
		actionID := int64(action & 0xffff)
		if actionID > petCommandDismiss && !petHasSpell(record, actionID) {
			continue
		}
		if actionID > petCommandDismiss && action != 0 {
			action |= 0x80 << 24
		}
		_, _, _, err := s.updateActivePet(owner.GUID, active.GUID, func(pet *realm.Pet) { pet.ActionBar[slot] = int64(action) })
		if err != nil {
			return err
		}
	}
	s.sendPetSpells(owner.GUID)
	return nil
}

func petHasSpell(record petRecord, spell int64) bool {
	for _, known := range record.spells {
		if known == spell {
			return true
		}
	}
	return false
}

func (s *WorldServer) petAbandon(owner realm.Character, data []byte) error {
	if len(data) < 8 {
		return nil
	}
	petGUID := binary.LittleEndian.Uint64(data)
	_, record, _, found := s.findActivePet(petGUID)
	if !found || record.data.OwnerGUID != owner.GUID {
		return nil
	}
	if err := s.detachPet(owner.GUID, petGUID, true); err != nil {
		return err
	}
	s.petMu.Lock()
	manager := s.pets[owner.GUID]
	index := -1
	if manager != nil {
		for petIndex, value := range manager.permanent {
			if value.data.ID == record.data.ID {
				index = petIndex
				break
			}
		}
		if index >= 0 {
			manager.permanent = append(manager.permanent[:index], manager.permanent[index+1:]...)
		}
	}
	s.petMu.Unlock()
	if s.Characters != nil {
		return s.Characters.DeletePet(record.data.ID)
	}
	return nil
}

func (s *WorldServer) petRename(owner realm.Character, data []byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, nil
	}
	petGUID := binary.LittleEndian.Uint64(data)
	name, err := packet.ReadString(data, 8, 0)
	if err != nil {
		return nil, nil
	}
	if !validName(name) {
		return packet.Encode(packet.SMSGPetNameInvalid, nil)
	}
	record, active, found := s.activePetSnapshot(owner.GUID, petGUID)
	if !found || record.data.CreatedBySpell != petSummonSpellID || record.data.RenameTime != 0 {
		return nil, nil
	}
	renameTime := time.Now().Unix()
	_, _, _, err = s.updateActivePet(owner.GUID, active.GUID, func(pet *realm.Pet) { pet.Name, pet.RenameTime = name, renameTime })
	if err != nil {
		return nil, err
	}
	s.updatePetCreatureField(active.GUID, 174, uint32(renameTime))
	return nil, nil
}

func (s *WorldServer) petNameQuery(_ realm.Character, data []byte) ([]byte, error) {
	if len(data) < 12 {
		return nil, nil
	}
	petID, petGUID := binary.LittleEndian.Uint32(data), binary.LittleEndian.Uint64(data[4:])
	_, record, _, found := s.findActivePet(petGUID)
	if !found {
		return nil, nil
	}
	name, err := packet.StringBytes(record.data.Name)
	if err != nil {
		return nil, err
	}
	body := append(encodeUint32(int64(petID)), name...)
	body = append(body, encodeUint32(record.data.RenameTime)...)
	return packet.Encode(packet.SMSGPetNameQueryResponse, body)
}

func (s *WorldServer) petLevelCheat(owner realm.Character, data []byte) error {
	if len(data) < 4 {
		return nil
	}
	level := int64(binary.LittleEndian.Uint32(data))
	if level < 1 || level > 100 {
		return nil
	}
	record, active, found := s.activePetSnapshot(owner.GUID, 0)
	if !found {
		return nil
	}
	stats, err := s.petCreatureStats(record.data.CreatureID, active.Creature.Template.UnitClass, level)
	if err != nil {
		return err
	}
	record, active, _, err = s.updateActivePet(owner.GUID, active.GUID, func(pet *realm.Pet) {
		pet.Level, pet.XP, pet.Health, pet.Mana = level, 0, stats.Health, stats.Mana
	})
	if err != nil {
		return err
	}
	s.creatures.mu.Lock()
	if state := s.creatures.active[active.GUID]; state != nil {
		state.Level, state.Health, state.MaxHealth, state.Mana = level, stats.Health, stats.Health, stats.Mana
	}
	s.creatures.mu.Unlock()
	s.updatePetCreatureField(active.GUID, 32, uint32(level))
	s.updatePetCreatureField(active.GUID, 175, 0)
	s.updatePetCreatureField(active.GUID, 176, uint32(petExperienceForLevel(level)))
	s.updatePetCreatureField(active.GUID, 22, uint32(stats.Health))
	s.updatePetCreatureField(active.GUID, 27, uint32(stats.Health))
	s.updatePetCreatureField(active.GUID, 23, uint32(stats.Mana))
	s.updatePetCreatureField(active.GUID, 28, uint32(stats.Mana))
	_ = record
	return nil
}

func (s *WorldServer) updatePetCreatureField(guid uint64, field int, value uint32) {
	update, err := packet.EncodeFieldUpdate(guid, field, value)
	if err != nil {
		return
	}
	ownerGUID, _, active, found := s.findActivePet(guid)
	if !found {
		return
	}
	owner, ownerFound := s.playerByGUID(ownerGUID)
	if ownerFound {
		s.broadcastPlayer(owner, update)
	}
	s.sendPlayer(ownerGUID, update)
	_ = active
}

func (s *WorldServer) detachPet(ownerGUID int64, petGUID uint64, clearActive bool) error {
	s.petMu.Lock()
	manager := s.pets[ownerGUID]
	if manager == nil || manager.active == nil || manager.active.GUID != petGUID {
		s.petMu.Unlock()
		return nil
	}
	active := *manager.active
	record := manager.permanent[active.Index]
	if clearActive {
		manager.permanent[active.Index].data.Active = false
		record.data.Active = false
	}
	manager.active = nil
	s.petMu.Unlock()
	if clearActive && s.Characters != nil {
		if err := s.Characters.UpdatePet(record.data); err != nil {
			return err
		}
	}
	s.removeCreature(active.GUID)
	if owner, found := s.playerByGUID(ownerGUID); found {
		if destroy, err := packet.Encode(packet.SMSGDestroyObject, encodeGUID(int64(active.GUID))); err == nil {
			s.sendSpell(owner, destroy)
		}
	}
	s.sendPetSpells(ownerGUID)
	return nil
}

func (s *WorldServer) unloadPets(ownerGUID int64) {
	s.petMu.Lock()
	manager := s.pets[ownerGUID]
	var guid uint64
	if manager != nil && manager.active != nil {
		guid = manager.active.GUID
	}
	delete(s.pets, ownerGUID)
	s.petMu.Unlock()
	if guid != 0 {
		s.removeCreature(guid)
	}
}

func petExperienceForLevel(level int64) int64 {
	if level < 1 {
		level = 1
	}
	return xpToLevel(uint8(level)) / 4
}

func defaultPetActionBar() [10]int64 {
	return [10]int64{0x07000002, 0x07000001, 0x07000000, 0x06000002, 0x06000001, 0x06000000}
}

func petSpellsPacket(guid uint64, record petRecord, reset bool) ([]byte, error) {
	if reset {
		return packet.Encode(packet.SMSGPetSpells, encodeUint64(0))
	}
	bar := record.data.ActionBar
	zero := true
	for _, action := range bar {
		if action != 0 {
			zero = false
			break
		}
	}
	if zero {
		bar = defaultPetActionBar()
	}
	var body bytes.Buffer
	write := func(value any) error { return binary.Write(&body, binary.LittleEndian, value) }
	if err := write(guid); err != nil {
		return nil, err
	}
	if err := write(uint32(0)); err != nil {
		return nil, err
	}
	for _, value := range []byte{byte(record.data.ReactState), byte(record.data.CommandState), 0, 0} {
		if err := write(value); err != nil {
			return nil, err
		}
	}
	for _, action := range bar {
		if err := write(uint32(action)); err != nil {
			return nil, err
		}
	}
	if err := write(uint32(len(record.spells))); err != nil {
		return nil, err
	}
	for _, spell := range record.spells {
		if err := write(uint16(spell)); err != nil {
			return nil, err
		}
	}
	if err := write(byte(0)); err != nil {
		return nil, err
	}
	return packet.Encode(packet.SMSGPetSpells, body.Bytes())
}

func (s *WorldServer) sendPetSpells(ownerGUID int64) {
	record, active, found := s.activePetSnapshot(ownerGUID, 0)
	var data []byte
	var err error
	if found {
		data, err = petSpellsPacket(active.GUID, record, false)
	} else {
		data, err = petSpellsPacket(0, petRecord{}, true)
	}
	if err == nil {
		s.sendPlayer(ownerGUID, data)
	}
}
