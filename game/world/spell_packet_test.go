package world

import (
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestSpellGoUsesPrimaryEffectTargets(t *testing.T) {
	cast := &spellCast{caster: realm.Character{GUID: 1}, spell: dbc.Spell{ID: 42}, targets: []realm.Character{{GUID: 2}, {GUID: 3}}, effectTargets: map[int][]realm.Character{0: {{GUID: 2}}, 1: {{GUID: 3}}}}
	encoded, err := spellGoPacket(cast)
	if err != nil {
		t.Fatal(err)
	}
	message, err := packet.Parse(encoded)
	if err != nil || message.Opcode != packet.SMSGSpellGo || len(message.Data) < 32 {
		t.Fatalf("packet=%#v err=%v", message, err)
	}
	if message.Data[22] != 1 || binary.LittleEndian.Uint64(message.Data[23:31]) != 2 || message.Data[31] != 0 {
		t.Fatalf("primary target data=%#v", message.Data)
	}
	cast.target = spellTarget{GameObjectGUID: 0xf110000000000001}
	cast.effectTargets = map[int][]realm.Character{0: {}}
	encoded, err = spellGoPacket(cast)
	if err != nil {
		t.Fatal(err)
	}
	message, err = packet.Parse(encoded)
	if err != nil || message.Data[22] != 1 || binary.LittleEndian.Uint64(message.Data[23:31]) != 0xf110000000000001 {
		t.Fatalf("object target data=%#v err=%v", message.Data, err)
	}
	cast.castFlags = packet.SpellCastFlagProc
	start, err := spellStartPacket(cast)
	if err != nil {
		t.Fatal(err)
	}
	startMessage, err := packet.Parse(start)
	if err != nil || binary.LittleEndian.Uint16(startMessage.Data[20:22]) != uint16(packet.SpellCastFlagProc) {
		t.Fatalf("start flags=%#v err=%v", startMessage, err)
	}
	encoded, err = spellGoPacket(cast)
	if err != nil {
		t.Fatal(err)
	}
	message, err = packet.Parse(encoded)
	if err != nil || binary.LittleEndian.Uint16(message.Data[20:22]) != 0 {
		t.Fatalf("go flags=%#v err=%v", message, err)
	}
}
