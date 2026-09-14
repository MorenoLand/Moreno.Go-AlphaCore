package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestPVPPortRoundTrip(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Map (ID, PVP) VALUES (0, 0), (1, 1), (2, 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO worldports (entry, x, y, z, o, map, name) VALUES (1, 10, 11, 12, 13, 1, 'PvPZone01'), (2, 20, 21, 22, 23, 2, 'PvPZone02')`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: realm.NewStore(databases), DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := &realm.Character{GUID: 1, AccountID: 1, RealmID: 1, Map: 0, PositionX: 1, PositionY: 2, PositionZ: 3, Orientation: 4}
	response, err := server.pvpPort(active)
	if err != nil || response == nil || (active.Map != 1 && active.Map != 2) {
		t.Fatalf("enter map=%d response=%x err=%v", active.Map, response, err)
	}
	parsed, err := packet.Parse(response)
	if err != nil || parsed.Opcode != packet.SMSGNewWorld || parsed.Data[0] != byte(active.Map) {
		t.Fatalf("enter packet=%#v err=%v", parsed, err)
	}
	response, err = server.pvpPort(active)
	if err != nil || response == nil || active.Map != 0 || active.PositionX != 1 || active.PositionY != 2 || active.PositionZ != 3 || active.Orientation != 4 {
		t.Fatalf("return map=%d position=(%f,%f,%f,%f) response=%x err=%v", active.Map, active.PositionX, active.PositionY, active.PositionZ, active.Orientation, response, err)
	}
}
