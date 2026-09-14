package world

import (
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestDebugAIStatePlayerPacket(t *testing.T) {
	server := &WorldServer{}
	active := realm.Character{GUID: 1, Name: "Requester", Health: 1, Map: 0, PositionX: 1, PositionY: 2, PositionZ: 3}
	target := realm.Character{GUID: 2, Name: "Target", Health: 1, Map: 0, PositionX: 4, PositionY: 6, PositionZ: 3}
	server.registerPlayer(active)
	server.registerPlayer(target)
	response, err := server.debugAIState(active, encodeUint64(uint64(target.GUID)))
	if err != nil {
		t.Fatal(err)
	}
	message, err := packet.Parse(response)
	if err != nil || message.Opcode != packet.SMSGDebugAIState || len(message.Data) < 12 || binary.LittleEndian.Uint64(message.Data) != uint64(target.GUID) || binary.LittleEndian.Uint32(message.Data[8:]) != 3 {
		t.Fatalf("message=%#v err=%v", message, err)
	}
	if value, err := packet.ReadString(message.Data, 12, 0); err != nil || value == "" {
		t.Fatalf("first debug line=%q err=%v", value, err)
	}
}
