package world

import (
	"math"
	"time"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

const (
	duelStateRequested  uint32  = 0
	duelStateStarted    uint32  = 1
	duelStateFinished   uint32  = 3
	duelStatusInBounds  bool    = false
	duelStatusOutBounds bool    = true
	duelWinnerKnockout  byte    = 0
	duelWinnerRetreat   byte    = 1
	duelCanceled        byte    = 0
	duelFinished        byte    = 1
	duelBoundaryRadius  float32 = 50
	duelStartRadius     float32 = 10
	duelGraceTime               = 10 * time.Second
	duelArbiterGUIDHigh uint64  = 0xf110000000000000
)

type duelParticipant struct {
	player        realm.Character
	target        int64
	team          uint32
	accepted      bool
	outOfBounds   bool
	boundaryTimer *time.Timer
}

type duelState struct {
	guid         uint64
	mapID        int64
	x, y, z      float32
	orientation  float32
	phase        uint32
	participants map[int64]*duelParticipant
}

func (s *WorldServer) nextDuelGUID() uint64 {
	s.duelMu.Lock()
	s.nextDuel++
	guid := duelArbiterGUIDHigh | s.nextDuel
	s.duelMu.Unlock()
	return guid
}

func (s *WorldServer) duelForPlayer(guid int64) *duelState {
	s.duelMu.Lock()
	state := s.duels[guid]
	s.duelMu.Unlock()
	return state
}

func (s *WorldServer) duelTarget(guid int64) int64 {
	s.duelMu.Lock()
	state := s.duels[guid]
	target := int64(0)
	if state != nil && state.participants[guid] != nil {
		target = state.participants[guid].target
	}
	s.duelMu.Unlock()
	return target
}

func (s *WorldServer) requestDuel(cast *spellCast, target realm.Character, effect dbc.SpellEffect) {
	if cast == nil || target.GUID == 0 || target.GUID == cast.caster.GUID || s.WorldData == nil {
		return
	}
	if player, found := s.playerByGUID(target.GUID); found {
		target = player
	} else {
		return
	}
	if s.duelForPlayer(cast.caster.GUID) != nil {
		s.endDuelForPlayer(cast.caster.GUID)
	}
	if s.duelForPlayer(target.GUID) != nil {
		return
	}
	template, found, err := s.WorldData.GameObjectTemplate(effect.MiscValue)
	if err != nil || !found {
		return
	}
	state := &duelState{guid: s.nextDuelGUID(), mapID: cast.caster.Map, x: target.PositionX, y: target.PositionY, z: target.PositionZ, orientation: target.Orientation, phase: duelStateRequested, participants: map[int64]*duelParticipant{
		cast.caster.GUID: {player: cast.caster, target: target.GUID, team: 1},
		target.GUID:      {player: target, target: cast.caster.GUID, team: 2},
	}}
	s.duelMu.Lock()
	if s.duels == nil {
		s.duels = make(map[int64]*duelState)
	}
	if s.duels[cast.caster.GUID] != nil || s.duels[target.GUID] != nil {
		s.duelMu.Unlock()
		return
	}
	s.duels[cast.caster.GUID], s.duels[target.GUID] = state, state
	s.duelMu.Unlock()
	s.setDuelArbiter(cast.caster, state.guid)
	s.setDuelArbiter(target, state.guid)
	if create, err := duelArbiterPacket(state, template); err == nil {
		s.sendSpell(cast.caster, create)
	}
	data := append(encodeUint64(state.guid), encodeUint64(uint64(cast.caster.GUID))...)
	if requested, err := packet.Encode(packet.SMSGDuelRequested, data); err == nil {
		s.sendPlayer(cast.caster.GUID, requested)
		s.sendPlayer(target.GUID, requested)
	}
}

func duelArbiterPacket(state *duelState, template worlddb.GameObjectTemplate) ([]byte, error) {
	scale := template.Scale
	if scale <= 0 {
		scale = 1
	}
	fields := make([]uint32, 20)
	packet.SetUint64(fields, 0, state.guid)
	fields[2] = 33
	fields[3] = uint32(template.Entry)
	fields[4] = math.Float32bits(scale)
	fields[6] = uint32(template.DisplayID)
	fields[7] = uint32(template.Flags)
	fields[10] = math.Float32bits(float32(math.Sin(float64(state.orientation) / 2)))
	fields[11] = math.Float32bits(float32(math.Cos(float64(state.orientation) / 2)))
	fields[12] = 1
	fields[14] = math.Float32bits(state.x)
	fields[15] = math.Float32bits(state.y)
	fields[16] = math.Float32bits(state.z)
	fields[17] = math.Float32bits(0)
	fields[19] = uint32(template.Faction)
	return packet.EncodeGameObjectCreate(state.guid, fields, packet.Movement{X: state.x, Y: state.y, Z: state.z, O: state.orientation})
}

func (s *WorldServer) duelAccept(guid int64) {
	s.duelMu.Lock()
	state := s.duels[guid]
	if state == nil || state.phase != duelStateRequested || state.participants[guid] == nil {
		s.duelMu.Unlock()
		return
	}
	state.participants[guid].accepted = true
	first, second := state.participantsForDuel()
	if first == nil || second == nil || !first.accepted || !second.accepted {
		s.duelMu.Unlock()
		return
	}
	if !duelWithinStartDistance(state, first.player) || !duelWithinStartDistance(state, second.player) {
		s.duelMu.Unlock()
		s.endDuel(state, duelWinnerRetreat, duelCanceled, 0)
		return
	}
	firstPlayer, secondPlayer := first.player, second.player
	firstTeam, secondTeam := first.team, second.team
	state.phase = duelStateStarted
	s.duelMu.Unlock()
	s.setDuelTeam(firstPlayer, firstTeam)
	s.setDuelTeam(secondPlayer, secondTeam)
}

func (s *WorldServer) duelCancel(guid int64) {
	s.duelMu.Lock()
	state := s.duels[guid]
	participant := stateParticipant(state, guid)
	if participant == nil {
		s.duelMu.Unlock()
		return
	}
	complete := duelCanceled
	if state.phase == duelStateStarted {
		complete = duelFinished
	}
	winner := participant.target
	s.duelMu.Unlock()
	s.endDuel(state, duelWinnerRetreat, complete, winner)
}

func (s *WorldServer) endDuelForPlayer(guid int64) {
	state := s.duelForPlayer(guid)
	if state != nil {
		s.endDuel(state, duelWinnerRetreat, duelCanceled, 0)
	}
}

func (s *WorldServer) endDuel(state *duelState, winnerFlag, completeFlag byte, winner int64) {
	if state == nil {
		return
	}
	s.duelMu.Lock()
	if state.phase == duelStateFinished {
		s.duelMu.Unlock()
		return
	}
	state.phase = duelStateFinished
	participants := make([]realm.Character, 0, len(state.participants))
	var winnerPlayer, loserPlayer realm.Character
	hasWinner := false
	for guid, participant := range state.participants {
		participants = append(participants, participant.player)
		if s.duels[guid] == state {
			delete(s.duels, guid)
		}
		if participant.boundaryTimer != nil {
			participant.boundaryTimer.Stop()
		}
	}
	if completeFlag == duelFinished {
		if participant := stateParticipant(state, winner); participant != nil {
			if loser := stateParticipant(state, participant.target); loser != nil {
				winnerPlayer, loserPlayer, hasWinner = participant.player, loser.player, true
			}
		}
	}
	s.duelMu.Unlock()
	complete, err := packet.Encode(packet.SMSGDuelComplete, []byte{completeFlag})
	if err == nil {
		for _, participant := range participants {
			s.sendPlayer(participant.GUID, complete)
		}
	}
	if hasWinner {
		winnerName, winnerErr := packet.StringBytes(winnerPlayer.Name)
		loserName, loserErr := packet.StringBytes(loserPlayer.Name)
		if winnerErr == nil && loserErr == nil {
			data := append([]byte{winnerFlag}, winnerName...)
			data = append(data, loserName...)
			if result, resultErr := packet.Encode(packet.SMSGDuelWinner, data); resultErr == nil && len(participants) > 0 {
				s.sendSpell(participants[0], result)
			}
		}
	}
	if len(participants) > 0 {
		if destroy, destroyErr := packet.Encode(packet.SMSGDestroyObject, encodeGUID(int64(state.guid))); destroyErr == nil {
			s.sendSpell(participants[0], destroy)
		}
	}
	for _, participant := range participants {
		s.setDuelArbiter(participant, 0)
		s.setDuelTeam(participant, 0)
		s.setCombatTarget(participant.GUID, 0)
	}
}

func (s *WorldServer) duelMovement(player realm.Character) {
	s.duelMu.Lock()
	state := s.duels[player.GUID]
	participant := stateParticipant(state, player.GUID)
	if state == nil || participant == nil || state.phase != duelStateStarted {
		s.duelMu.Unlock()
		return
	}
	participant.player = player
	inBounds := player.Map == state.mapID && distance3(player.PositionX, player.PositionY, player.PositionZ, state.x, state.y, state.z) < duelBoundaryRadius
	if inBounds && participant.outOfBounds {
		participant.outOfBounds = duelStatusInBounds
		if participant.boundaryTimer != nil {
			participant.boundaryTimer.Stop()
			participant.boundaryTimer = nil
		}
		s.duelMu.Unlock()
		if inside, err := packet.Encode(packet.SMSGDuelInBounds, nil); err == nil {
			s.sendPlayer(player.GUID, inside)
		}
		return
	}
	if !inBounds && !participant.outOfBounds {
		participant.outOfBounds = duelStatusOutBounds
		participant.boundaryTimer = time.AfterFunc(duelGraceTime, func() { s.duelBoundaryExpired(player.GUID, state) })
		s.duelMu.Unlock()
		if outside, err := packet.Encode(packet.SMSGDuelOutOfBounds, encodeUint32(int64(duelGraceTime/time.Second))); err == nil {
			s.sendPlayer(player.GUID, outside)
		}
		return
	}
	s.duelMu.Unlock()
}

func (s *WorldServer) duelBoundaryExpired(guid int64, state *duelState) {
	s.duelMu.Lock()
	participant := stateParticipant(state, guid)
	valid := state != nil && state.phase == duelStateStarted && participant != nil && participant.outOfBounds
	winner := int64(0)
	if valid {
		winner = participant.target
	}
	s.duelMu.Unlock()
	if valid {
		s.endDuel(state, duelWinnerRetreat, duelFinished, winner)
	}
}

func (s *WorldServer) duelKnockout(attacker, target int64) {
	s.duelMu.Lock()
	state := s.duels[attacker]
	participant := stateParticipant(state, attacker)
	valid := state != nil && state.phase == duelStateStarted && participant != nil && participant.target == target
	s.duelMu.Unlock()
	if valid {
		s.endDuel(state, duelWinnerKnockout, duelFinished, attacker)
	}
}

func (s *WorldServer) setDuelArbiter(player realm.Character, guid uint64) {
	s.sendDuelField(player, 320, uint32(guid))
	s.sendDuelField(player, 321, uint32(guid>>32))
}

func (s *WorldServer) setDuelTeam(player realm.Character, team uint32) {
	s.sendDuelField(player, 622, team)
}

func (s *WorldServer) sendDuelField(player realm.Character, field int, value uint32) {
	update, err := packet.EncodeFieldUpdate(uint64(player.GUID), field, value)
	if err != nil {
		return
	}
	s.broadcastPlayer(player, update)
	s.sendPlayer(player.GUID, update)
}

func stateParticipant(state *duelState, guid int64) *duelParticipant {
	if state == nil {
		return nil
	}
	return state.participants[guid]
}

func (state *duelState) participantsForDuel() (*duelParticipant, *duelParticipant) {
	participants := make([]*duelParticipant, 0, 2)
	for _, participant := range state.participants {
		participants = append(participants, participant)
	}
	if len(participants) != 2 {
		return nil, nil
	}
	return participants[0], participants[1]
}

func duelWithinStartDistance(state *duelState, player realm.Character) bool {
	return player.Map == state.mapID && distance3(player.PositionX, player.PositionY, player.PositionZ, state.x, state.y, state.z) <= duelStartRadius
}

func duelDamageEffect(effect dbc.SpellEffect) bool {
	switch packet.SpellEffect(effect.Type) {
	case packet.SpellEffectInstantKill, packet.SpellEffectSchoolDamage, packet.SpellEffectWeaponDamage, packet.SpellEffectWeaponDamagePlus, packet.SpellEffectPowerBurn, packet.SpellEffectHealthLeech:
		return true
	default:
		return false
	}
}
