package packet

import (
	"encoding/binary"
	"testing"
)

func TestFieldUpdatePacket(t *testing.T) {
	data, err := EncodeFieldUpdate(7, 172, 0x01000003)
	if err != nil {
		t.Fatal(err)
	}
	message, err := Parse(data)
	if err != nil || message.Opcode != SMSGUpdateObject {
		t.Fatalf("message=%#v err=%v", message, err)
	}
	if len(message.Data) != 42 || binary.LittleEndian.Uint32(message.Data) != 1 || message.Data[4] != UpdatePartial || binary.LittleEndian.Uint64(message.Data[5:]) != 7 || message.Data[13] != 6 || binary.LittleEndian.Uint32(message.Data[14+4*5:]) != 1<<12 || binary.LittleEndian.Uint32(message.Data[38:]) != 0x01000003 {
		t.Fatalf("data=%x", message.Data)
	}
}
