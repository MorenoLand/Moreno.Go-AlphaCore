package world

import (
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestAttackStartAndStop(t *testing.T) {
	server := &WorldServer{}
	active, target := realm.Character{GUID: 1, Map: 0, Health: 1}, realm.Character{GUID: 2, Map: 0, Health: 1}
	server.registerPlayer(active)
	server.registerPlayer(target)
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(target.GUID))
	responses, err := server.attack(active, data)
	if err != nil || len(responses) != 1 {
		t.Fatalf("attack responses=%d err=%v", len(responses), err)
	}
	start, err := packet.Parse(responses[0])
	if err != nil || start.Opcode != packet.SMSGAttackStart || server.combatTarget(active.GUID) != uint64(target.GUID) {
		t.Fatalf("start=%#v target=%d err=%v", start, server.combatTarget(active.GUID), err)
	}
	responses, err = server.attackStop(active)
	if err != nil || len(responses) != 2 {
		t.Fatalf("stop responses=%d err=%v", len(responses), err)
	}
	stop, err := packet.Parse(responses[0])
	if err != nil || stop.Opcode != packet.SMSGAttackStop || server.combatTarget(active.GUID) != 0 {
		t.Fatalf("stop=%#v target=%d err=%v", stop, server.combatTarget(active.GUID), err)
	}
	cancel, err := packet.Parse(responses[1])
	if err != nil || cancel.Opcode != packet.SMSGCancelCombat {
		t.Fatalf("cancel=%#v err=%v", cancel, err)
	}
}
