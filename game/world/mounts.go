package world

import (
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const (
	unitFlagMountIcon         uint32 = 0x00001000
	unitFlagMount             uint32 = 0x00002000
	mountResultNotMounted            = 3
	mountResultAlreadyMounted        = 2
	dismountResultNotMounted         = 1
)

func (s *WorldServer) mountDisplayID(guid int64) uint32 {
	s.players.mu.RLock()
	display := s.players.mountDisplay[guid]
	s.players.mu.RUnlock()
	return display
}

func (s *WorldServer) summonMount(target realm.Character, creatureEntry int64) {
	if s.mountDisplayID(target.GUID) != 0 {
		s.unmountPlayer(target, false, true)
		return
	}
	if s.WorldData == nil {
		return
	}
	template, found, err := s.WorldData.CreatureTemplate(creatureEntry)
	if err != nil || !found {
		return
	}
	s.mountPlayer(target, uint32(template.DisplayID1), false)
}

func (s *WorldServer) mountPlayer(target realm.Character, display uint32, report bool) bool {
	if display == 0 || s.DBC == nil {
		if report {
			s.sendMountResult(target.GUID, mountResultNotMounted)
		}
		return false
	}
	if _, found, err := s.DBC.CreatureDisplayInfo(int64(display)); err != nil || !found {
		if report {
			s.sendMountResult(target.GUID, mountResultNotMounted)
		}
		return false
	}
	if s.mountDisplayID(target.GUID) != 0 {
		if report {
			s.sendMountResult(target.GUID, mountResultAlreadyMounted)
		}
		return false
	}
	if record, active, found := s.activePetSnapshot(target.GUID, 0); found && record.data.Active {
		_ = s.detachPet(target.GUID, active.GUID, true)
	}
	s.players.mu.Lock()
	if s.players.mountDisplay == nil {
		s.players.mountDisplay = make(map[int64]uint32)
	}
	s.players.mountDisplay[target.GUID] = display
	s.players.mu.Unlock()
	s.setUnitFlags(target, s.unitFlags(target.GUID)|unitFlagMountIcon|unitFlagMount)
	s.sendMountField(target, display)
	if report {
		s.sendMountResult(target.GUID, 0)
	}
	return true
}

func (s *WorldServer) unmountPlayer(target realm.Character, fromAura, report bool) bool {
	if s.mountDisplayID(target.GUID) == 0 {
		if report && !fromAura {
			s.sendDismountResult(target.GUID, dismountResultNotMounted)
		}
		return false
	}
	s.players.mu.Lock()
	delete(s.players.mountDisplay, target.GUID)
	s.players.mu.Unlock()
	if pure, err := packet.Encode(packet.SMSGPureMountCancelled, encodeGUID(target.GUID)); err == nil {
		s.broadcastPlayer(target, pure)
		s.sendPlayer(target.GUID, pure)
	}
	s.setUnitFlags(target, s.unitFlags(target.GUID)&^(unitFlagMountIcon|unitFlagMount))
	s.sendMountField(target, 0)
	return true
}

func (s *WorldServer) sendMountField(target realm.Character, display uint32) {
	if update, err := packet.EncodeFieldUpdate(uint64(target.GUID), 152, display); err == nil {
		s.broadcastPlayer(target, update)
		s.sendPlayer(target.GUID, update)
	}
}

func (s *WorldServer) sendMountResult(guid int64, result uint32) {
	if response, err := packet.Encode(packet.SMSGMountResult, encodeUint32(int64(result))); err == nil {
		s.sendPlayer(guid, response)
	}
}

func (s *WorldServer) sendDismountResult(guid int64, result uint32) {
	if response, err := packet.Encode(packet.SMSGDismountResult, encodeUint32(int64(result))); err == nil {
		s.sendPlayer(guid, response)
	}
}
