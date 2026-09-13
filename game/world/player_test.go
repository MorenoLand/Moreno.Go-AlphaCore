package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestPlayerLoginPacketOrder(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO ChrRaces (ID, FactionID, MaleDisplayId, FemaleDisplayId, CreatureType) VALUES (1, 1, 49, 50, 7)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Testone", Race: 1, Class: 1, Level: 1, PositionX: 1, PositionY: 2, PositionZ: 3, Orientation: 4, Map: 0, Zone: 12, Health: 20})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(guid))
	result, err := server.playerLogin(1, data)
	if err != nil {
		t.Fatal(err)
	}
	packets := make([]packet.Packet, 0, 3)
	for len(result) > 0 {
		if len(result) < packet.HeaderSize {
			t.Fatalf("truncated packet stream: %d", len(result))
		}
		size, _, err := packet.ParseHeader(result[:packet.HeaderSize])
		if err != nil || len(result) < packet.HeaderSize+int(size) {
			t.Fatalf("invalid packet stream: size=%d err=%v", size, err)
		}
		message, err := packet.Parse(result[:packet.HeaderSize+int(size)])
		if err != nil {
			t.Fatal(err)
		}
		packets = append(packets, message)
		result = result[packet.HeaderSize+int(size):]
	}
	if len(packets) != 3 || packets[0].Opcode != packet.SMSGLoginSetTimeSpeed || packets[1].Opcode != packet.SMSGNewWorld || packets[2].Opcode != packet.SMSGCompressedUpdateObject {
		t.Fatalf("packet order: %#v", packets)
	}
}
