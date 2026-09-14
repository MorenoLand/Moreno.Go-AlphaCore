package world

import (
	"sync"
	"time"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

type playerRegistry struct {
	mu             sync.RWMutex
	players        map[int64]realm.Character
	groupStatus    map[int64]uint32
	selection      map[int64]uint64
	target         map[int64]uint64
	standState     map[int64]uint32
	weaponMode     map[int64]uint32
	combatTarget   map[int64]uint64
	pvpSource      map[int64]pvpLocation
	maxHealth      map[int64]int64
	maxHealthKnown map[int64]bool
	maxPower       map[int64][5]int64
	mountDisplay   map[int64]uint32
	comboPoints    map[int64]int64
	comboTarget    map[int64]uint64
	extraAttacks   map[int64]int64
	passiveSkills  map[int64]uint32
	displayIDs     map[int64]int64
	unitFlags      map[int64]uint32
	godMode        map[int64]bool
	beastMaster    map[int64]bool
	sanctuary      map[int64]time.Time
	speeds         map[int64][4]float32
	connections    map[int64]*playerConnection
}

func (s *WorldServer) registerPlayer(character realm.Character) {
	s.players.mu.Lock()
	if s.players.players == nil {
		s.players.players = make(map[int64]realm.Character)
		s.players.groupStatus = make(map[int64]uint32)
		s.players.selection = make(map[int64]uint64)
		s.players.target = make(map[int64]uint64)
		s.players.standState = make(map[int64]uint32)
		s.players.weaponMode = make(map[int64]uint32)
		s.players.combatTarget = make(map[int64]uint64)
		s.players.pvpSource = make(map[int64]pvpLocation)
		s.players.maxHealth = make(map[int64]int64)
		s.players.maxHealthKnown = make(map[int64]bool)
		s.players.maxPower = make(map[int64][5]int64)
		s.players.mountDisplay = make(map[int64]uint32)
		s.players.comboPoints = make(map[int64]int64)
		s.players.comboTarget = make(map[int64]uint64)
		s.players.extraAttacks = make(map[int64]int64)
		s.players.passiveSkills = make(map[int64]uint32)
		s.players.displayIDs = make(map[int64]int64)
		s.players.unitFlags = make(map[int64]uint32)
		s.players.godMode = make(map[int64]bool)
		s.players.beastMaster = make(map[int64]bool)
		s.players.sanctuary = make(map[int64]time.Time)
		s.players.speeds = make(map[int64][4]float32)
		s.players.connections = make(map[int64]*playerConnection)
	}
	if s.players.godMode == nil {
		s.players.godMode = make(map[int64]bool)
	}
	if s.players.beastMaster == nil {
		s.players.beastMaster = make(map[int64]bool)
	}
	if s.players.sanctuary == nil {
		s.players.sanctuary = make(map[int64]time.Time)
	}
	if s.players.maxPower == nil {
		s.players.maxPower = make(map[int64][5]int64)
	}
	if s.players.extraAttacks == nil {
		s.players.extraAttacks = make(map[int64]int64)
	}
	if s.players.mountDisplay == nil {
		s.players.mountDisplay = make(map[int64]uint32)
	}
	if s.players.maxHealth == nil {
		s.players.maxHealth = make(map[int64]int64)
	}
	if s.players.maxHealthKnown == nil {
		s.players.maxHealthKnown = make(map[int64]bool)
	}
	if s.players.speeds == nil {
		s.players.speeds = make(map[int64][4]float32)
	}
	s.players.players[character.GUID] = character
	if _, found := s.players.maxHealth[character.GUID]; !found {
		value := maxInt64(character.Health, 1)
		known := s.WorldData == nil
		if s.WorldData != nil {
			if stats, statsFound, err := s.WorldData.ClassStats(character.Class, character.Level); err == nil && statsFound {
				if stats.BaseHealth > value {
					value = stats.BaseHealth
				}
				if stats.BaseHealth > 0 {
					known = true
				}
			}
		}
		s.players.maxHealth[character.GUID] = value
		s.players.maxHealthKnown[character.GUID] = known
	}
	if _, found := s.players.unitFlags[character.GUID]; !found {
		s.players.unitFlags[character.GUID] = unitFlagPlayer
	}
	if _, found := s.players.maxPower[character.GUID]; !found {
		values := [5]int64{1000, 100, 100, 100, 0}
		if s.WorldData != nil {
			if stats, statsFound, err := s.WorldData.ClassStats(character.Class, character.Level); err == nil && statsFound && stats.BaseMana > 0 {
				values[0] = stats.BaseMana
			}
		}
		for index, value := range []int64{character.Power1, character.Power2, character.Power3, character.Power4, character.Power5} {
			if value > values[index] {
				values[index] = value
			}
		}
		s.players.maxPower[character.GUID] = values
	}
	if _, found := s.players.speeds[character.GUID]; !found {
		s.players.speeds[character.GUID] = defaultPlayerSpeeds()
	}
	s.players.mu.Unlock()
}

func (s *WorldServer) updatePlayer(character realm.Character) {
	s.players.mu.Lock()
	if s.players.players != nil {
		s.players.players[character.GUID] = character
		if character.Health > s.players.maxHealth[character.GUID] {
			s.players.maxHealth[character.GUID] = character.Health
		}
	}
	s.players.mu.Unlock()
}

func (s *WorldServer) unregisterPlayer(guid int64) {
	s.endDuelForPlayer(guid)
	s.unloadPets(guid)
	s.players.mu.Lock()
	delete(s.players.players, guid)
	delete(s.players.groupStatus, guid)
	delete(s.players.selection, guid)
	delete(s.players.target, guid)
	delete(s.players.standState, guid)
	delete(s.players.weaponMode, guid)
	delete(s.players.combatTarget, guid)
	delete(s.players.pvpSource, guid)
	delete(s.players.maxHealth, guid)
	delete(s.players.maxHealthKnown, guid)
	delete(s.players.maxPower, guid)
	delete(s.players.mountDisplay, guid)
	delete(s.players.comboPoints, guid)
	delete(s.players.comboTarget, guid)
	delete(s.players.passiveSkills, guid)
	delete(s.players.displayIDs, guid)
	delete(s.players.unitFlags, guid)
	delete(s.players.godMode, guid)
	delete(s.players.beastMaster, guid)
	delete(s.players.sanctuary, guid)
	delete(s.players.speeds, guid)
	delete(s.players.extraAttacks, guid)
	s.lootMu.Lock()
	delete(s.lootSelections, guid)
	delete(s.lootSelectionTypes, guid)
	for _, loot := range s.loots {
		delete(loot.active, guid)
	}
	s.lootMu.Unlock()
	delete(s.players.connections, guid)
	s.players.mu.Unlock()
}

func (s *WorldServer) playerMaxHealth(guid int64) int64 {
	s.players.mu.RLock()
	value := s.players.maxHealth[guid]
	s.players.mu.RUnlock()
	return maxInt64(value, 1)
}

func (s *WorldServer) playerMaxHealthKnown(guid int64) bool {
	s.players.mu.RLock()
	known := s.players.maxHealthKnown[guid]
	s.players.mu.RUnlock()
	return known
}

func (s *WorldServer) playerMaxPower(guid, powerType int64) int64 {
	s.players.mu.RLock()
	values := s.players.maxPower[guid]
	s.players.mu.RUnlock()
	if powerType < 0 || powerType >= int64(len(values)) || values[powerType] <= 0 {
		if powerType == 0 {
			return 1000
		}
		return 100
	}
	return values[powerType]
}

func (s *WorldServer) setPlayerMaxPower(guid, powerType, value int64) {
	if powerType < 0 || powerType >= 5 {
		return
	}
	s.players.mu.Lock()
	if s.players.maxPower == nil {
		s.players.maxPower = make(map[int64][5]int64)
	}
	values := s.players.maxPower[guid]
	values[powerType] = maxInt64(value, 1)
	s.players.maxPower[guid] = values
	s.players.mu.Unlock()
}

func (s *WorldServer) playerSpeeds(guid int64) [4]float32 {
	s.players.mu.RLock()
	values, found := s.players.speeds[guid]
	s.players.mu.RUnlock()
	if !found {
		return defaultPlayerSpeeds()
	}
	return values
}

func (s *WorldServer) setPlayerMaxHealth(guid, value int64) {
	s.players.mu.Lock()
	if s.players.maxHealth == nil {
		s.players.maxHealth = make(map[int64]int64)
	}
	s.players.maxHealth[guid] = maxInt64(value, 1)
	if s.players.maxHealthKnown == nil {
		s.players.maxHealthKnown = make(map[int64]bool)
	}
	s.players.maxHealthKnown[guid] = true
	s.players.mu.Unlock()
}

func (s *WorldServer) setGodMode(guid int64, enabled bool) {
	s.players.mu.Lock()
	if s.players.godMode == nil {
		s.players.godMode = make(map[int64]bool)
	}
	s.players.godMode[guid] = enabled
	s.players.mu.Unlock()
}

func (s *WorldServer) isGodMode(guid int64) bool {
	s.players.mu.RLock()
	enabled := s.players.godMode[guid]
	s.players.mu.RUnlock()
	return enabled
}

func (s *WorldServer) setBeastMaster(guid int64, enabled bool) {
	s.players.mu.Lock()
	if s.players.beastMaster == nil {
		s.players.beastMaster = make(map[int64]bool)
	}
	s.players.beastMaster[guid] = enabled
	s.players.mu.Unlock()
}

func (s *WorldServer) isBeastMaster(guid int64) bool {
	s.players.mu.RLock()
	enabled := s.players.beastMaster[guid]
	s.players.mu.RUnlock()
	return enabled
}

func (s *WorldServer) setSanctuary(guid int64, duration time.Duration) {
	s.players.mu.Lock()
	if s.players.sanctuary == nil {
		s.players.sanctuary = make(map[int64]time.Time)
	}
	if duration > 0 {
		s.players.sanctuary[guid] = time.Now().Add(duration)
	} else {
		delete(s.players.sanctuary, guid)
	}
	s.players.mu.Unlock()
}

func (s *WorldServer) isSanctuary(guid int64) bool {
	s.players.mu.Lock()
	until, found := s.players.sanctuary[guid]
	if found && !time.Now().Before(until) {
		delete(s.players.sanctuary, guid)
		found = false
	}
	s.players.mu.Unlock()
	return found
}

func (s *WorldServer) setUnitFlags(character realm.Character, flags uint32) {
	s.players.mu.Lock()
	if s.players.unitFlags == nil {
		s.players.unitFlags = make(map[int64]uint32)
	}
	s.players.unitFlags[character.GUID] = flags
	s.players.mu.Unlock()
	if update, err := packet.EncodeFieldUpdate(uint64(character.GUID), 54, flags); err == nil {
		s.sendSpell(character, update)
	}
}

func (s *WorldServer) unitFlags(guid int64) uint32 {
	s.players.mu.RLock()
	flags := s.players.unitFlags[guid]
	s.players.mu.RUnlock()
	if flags == 0 {
		return unitFlagPlayer
	}
	return flags
}

func (s *WorldServer) setCombatTarget(guid int64, target uint64) {
	s.players.mu.Lock()
	if s.players.combatTarget == nil {
		s.players.combatTarget = make(map[int64]uint64)
	}
	s.players.combatTarget[guid] = target
	s.players.mu.Unlock()
}

func (s *WorldServer) combatTarget(guid int64) uint64 {
	s.players.mu.RLock()
	target := s.players.combatTarget[guid]
	s.players.mu.RUnlock()
	return target
}

func (s *WorldServer) getGroupStatus(guid int64) uint32 {
	s.players.mu.RLock()
	status := s.players.groupStatus[guid]
	s.players.mu.RUnlock()
	return status
}

func (s *WorldServer) setGroupStatus(guid int64, status uint32) {
	s.players.mu.Lock()
	if s.players.groupStatus == nil {
		s.players.groupStatus = make(map[int64]uint32)
	}
	s.players.groupStatus[guid] = status
	s.players.mu.Unlock()
}

func (s *WorldServer) setPlayerSelection(guid int64, value uint64) {
	s.players.mu.Lock()
	if s.players.selection == nil {
		s.players.selection = make(map[int64]uint64)
	}
	s.players.selection[guid] = value
	s.players.mu.Unlock()
}

func (s *WorldServer) setPlayerTarget(guid int64, value uint64) {
	s.players.mu.Lock()
	if s.players.target == nil {
		s.players.target = make(map[int64]uint64)
	}
	s.players.target[guid] = value
	s.players.mu.Unlock()
}

func (s *WorldServer) setStandState(guid int64, value uint32) {
	s.players.mu.Lock()
	if s.players.standState == nil {
		s.players.standState = make(map[int64]uint32)
	}
	s.players.standState[guid] = value
	s.players.mu.Unlock()
}

func (s *WorldServer) setWeaponMode(guid int64, value uint32) {
	s.players.mu.Lock()
	if s.players.weaponMode == nil {
		s.players.weaponMode = make(map[int64]uint32)
	}
	s.players.weaponMode[guid] = value
	s.players.mu.Unlock()
}

func (s *WorldServer) bytes1(guid int64) uint32 {
	s.players.mu.RLock()
	mode, stand := s.players.weaponMode[guid], s.players.standState[guid]
	s.players.mu.RUnlock()
	if mode == 0 {
		mode = 1
	}
	return mode<<24 | stand
}

func (s *WorldServer) bytes1Update(active realm.Character) ([]byte, error) {
	update, err := packet.EncodeFieldUpdate(uint64(active.GUID), 172, s.bytes1(active.GUID))
	if err != nil {
		return nil, err
	}
	s.broadcastPlayer(active, update)
	return update, nil
}

func (s *WorldServer) onlinePlayers() []realm.Character {
	s.players.mu.RLock()
	players := make([]realm.Character, 0, len(s.players.players))
	for _, player := range s.players.players {
		players = append(players, player)
	}
	s.players.mu.RUnlock()
	return players
}
