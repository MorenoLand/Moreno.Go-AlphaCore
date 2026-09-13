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

func TestTaxiDiscoveryAndNodeStatus(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Traveler", Map: 0})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO TaxiNodes (ID, ContinentID, X, Y, Z) VALUES (1, 0, 0, 0, 0), (2, 0, 50, 0, 0); INSERT INTO TaxiPath (ID, FromTaxiNode, ToTaxiNode, Cost) VALUES (1, 1, 2, 10)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, npc_flags) VALUES (100, 20, 'Flight Master', 4); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z) VALUES (10, 100, 0, 0, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Map: 0}
	flightMaster := make([]byte, 8)
	binary.LittleEndian.PutUint64(flightMaster, 0xf13000000000000a)
	responses, err := server.taxiQueryNodes(&active, flightMaster)
	if err != nil || len(responses) != 2 {
		t.Fatalf("discover responses=%d err=%v", len(responses), err)
	}
	status, err := packet.Parse(responses[1])
	if err != nil || status.Opcode != packet.SMSGTaxiNodeStatus || status.Data[8] != 1 {
		t.Fatalf("discover status=%#v err=%v", status, err)
	}
	active.Taximask = taxiMaskString(3)
	responses, err = server.taxiQueryNodes(&active, flightMaster)
	if err != nil || len(responses) != 1 {
		t.Fatalf("show responses=%d err=%v", len(responses), err)
	}
	show, err := packet.Parse(responses[0])
	if err != nil || show.Opcode != packet.SMSGShowTaxiNodes || len(show.Data) != 32 || binary.LittleEndian.Uint32(show.Data) != 1 || binary.LittleEndian.Uint64(show.Data[4:]) != binary.LittleEndian.Uint64(flightMaster) || binary.LittleEndian.Uint32(show.Data[12:]) != 1 || binary.LittleEndian.Uint64(show.Data[16:]) != 2 || binary.LittleEndian.Uint64(show.Data[24:]) != 3 {
		t.Fatalf("show=%#v err=%v", show, err)
	}
	statusData, err := server.taxiNodeStatus(active, flightMaster)
	if err != nil || statusData == nil {
		t.Fatal(err)
	}
	statusPacket, err := packet.Parse(statusData)
	if err != nil || statusPacket.Opcode != packet.SMSGTaxiNodeStatus || statusPacket.Data[8] != 0 {
		t.Fatalf("known status=%#v err=%v", statusPacket, err)
	}
}
