package world

import (
	"context"
	"encoding/binary"
	"math"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestAreaTriggerTeleport(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO AreaTrigger (ID, ContinentID, X, Y, Z, Radius) VALUES (1, 0, 10, 20, 30, 5), (2, 0, 100, 100, 100, 5); INSERT INTO Map (ID) VALUES (0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO areatrigger_teleport (id, name, target_map, target_position_x, target_position_y, target_position_z, target_orientation) VALUES (1, 'Test', 0, 40, 50, 60, 0.75)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: realm.NewStore(databases), DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: 1, AccountID: 1, RealmID: 1, Name: "Player", Map: 0, PositionX: 10, PositionY: 20, PositionZ: 30}
	response, err := server.areaTrigger(&active, []byte{1, 0, 0, 0})
	if err != nil {
		t.Fatal(err)
	}
	message, err := packet.Parse(response)
	if err != nil || message.Opcode != packet.SMSGNewWorld || len(message.Data) != 17 || message.Data[0] != 0 || math.Float32frombits(binary.LittleEndian.Uint32(message.Data[1:])) != 40 || math.Float32frombits(binary.LittleEndian.Uint32(message.Data[5:])) != 50 || math.Float32frombits(binary.LittleEndian.Uint32(message.Data[9:])) != 60 || math.Float32frombits(binary.LittleEndian.Uint32(message.Data[13:])) != 0.75 {
		t.Fatalf("teleport=%#v err=%v", message, err)
	}
	if active.PositionX != 40 || active.PositionY != 50 || active.PositionZ != 60 || active.Orientation != 0.75 {
		t.Fatalf("active=%#v", active)
	}
	response, err = server.worldTeleport(&active, append(make([]byte, 4), append([]byte{0, 0, 0, 0}, make([]byte, 16)...)...), 0)
	if err != nil || response != nil {
		t.Fatalf("non-gm teleport=%x err=%v", response, err)
	}
}
