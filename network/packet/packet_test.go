package packet

import (
	"bytes"
	"testing"
)

func TestPacketRoundTrip(t *testing.T) {
	data, err := Encode(CMSGCharEnum, []byte{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	packet, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if packet.Opcode != CMSGCharEnum || !bytes.Equal(packet.Data, []byte{1, 2, 3}) || packet.Size != 3 {
		t.Fatalf("unexpected packet: %#v", packet)
	}
}

func TestUpdatePacketCompression(t *testing.T) {
	data, err := Encode(SMSGUpdateObject, bytes.Repeat([]byte{7}, 128))
	if err != nil {
		t.Fatal(err)
	}
	packet, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if packet.Opcode != SMSGCompressedUpdateObject || len(packet.Data) < 5 {
		t.Fatalf("unexpected compressed packet: %#v", packet)
	}
}

func TestStringReadWrite(t *testing.T) {
	data, err := StringBytes("Alpha")
	if err != nil {
		t.Fatal(err)
	}
	value, err := ReadString(data, 0, 0)
	if err != nil || value != "Alpha" {
		t.Fatalf("value=%q err=%v", value, err)
	}
}
