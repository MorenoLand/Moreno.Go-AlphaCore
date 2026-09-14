package world

import (
	"encoding/binary"
	"math"
	"testing"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestMovementInterruptsFlaggedCast(t *testing.T) {
	server := &WorldServer{}
	active := realm.Character{GUID: 1, Map: 0, Health: 10, PositionX: 1, PositionY: 2, PositionZ: 3}
	server.registerPlayer(active)
	server.spells.casts = map[int64]*spellCast{active.GUID: {caster: active, spell: dbc.Spell{ID: 42, InterruptFlags: packet.SpellInterruptMovement}}}
	data := make([]byte, 48)
	binary.LittleEndian.PutUint32(data[24:], math.Float32bits(4))
	binary.LittleEndian.PutUint32(data[28:], math.Float32bits(2))
	binary.LittleEndian.PutUint32(data[32:], math.Float32bits(3))
	binary.LittleEndian.PutUint32(data[36:], math.Float32bits(0))
	if err := server.updateMovement(&active, packet.MSGMoveHeartbeat, data); err != nil {
		t.Fatal(err)
	}
	server.spells.mu.Lock()
	_, found := server.spells.casts[active.GUID]
	server.spells.mu.Unlock()
	if found {
		t.Fatal("movement-interrupted cast remained")
	}
}
