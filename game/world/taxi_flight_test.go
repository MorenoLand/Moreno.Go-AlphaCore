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

func TestActivateTaxiAndFinishFlight(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Traveler", Map: 0, Money: 100})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO TaxiNodes (ID, ContinentID, X, Y, Z) VALUES (1, 0, 0, 0, 0), (2, 0, 50, 0, 0); INSERT INTO TaxiPath (ID, FromTaxiNode, ToTaxiNode, Cost) VALUES (1, 1, 2, 10); INSERT INTO TaxiPathNode (ID, PathID, NodeIndex, ContinentID, LocX, LocY, LocZ) VALUES (1, 1, 0, 0, 0, 0, 0), (2, 1, 1, 0, 50, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, npc_flags) VALUES (100, 20, 'Flight Master', 4); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z) VALUES (10, 100, 0, 0, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Traveler", Map: 0, Money: 100, Taximask: taxiMaskString(3), Health: 1}
	data := make([]byte, 16)
	binary.LittleEndian.PutUint64(data, 0xf13000000000000a)
	binary.LittleEndian.PutUint32(data[8:], 1)
	binary.LittleEndian.PutUint32(data[12:], 2)
	responses, err := server.activateTaxi(&active, data)
	if err != nil || len(responses) != 2 || active.Money != 90 || active.TaxiPath == "" {
		t.Fatalf("activate responses=%d money=%d path=%q err=%v", len(responses), active.Money, active.TaxiPath, err)
	}
	reply, err := packet.Parse(responses[0])
	if err != nil || reply.Opcode != packet.SMSGActivateTaxiReply || binary.LittleEndian.Uint32(reply.Data) != taxiOK {
		t.Fatalf("reply=%#v err=%v", reply, err)
	}
	move, err := packet.Parse(responses[1])
	if err != nil || move.Opcode != packet.SMSGMonsterMove || binary.LittleEndian.Uint64(move.Data) != uint64(guid) || binary.LittleEndian.Uint32(move.Data[25:]) != 0x200 || binary.LittleEndian.Uint32(move.Data[33:]) != 2 {
		t.Fatalf("move=%#v err=%v", move, err)
	}
	movement := make([]byte, 48)
	binary.LittleEndian.PutUint32(movement[24:], math.Float32bits(50))
	if err := server.updateMovement(&active, packet.MSGMoveHeartbeat, movement); err != nil {
		t.Fatal(err)
	}
	if active.TaxiPath != "" {
		t.Fatalf("flight path remains %q", active.TaxiPath)
	}
	stored, found, err := characters.CharacterByGUID(guid)
	if err != nil || !found || stored.TaxiPath != "" || stored.PositionX != 50 {
		t.Fatalf("stored=%#v found=%v err=%v", stored, found, err)
	}
}
