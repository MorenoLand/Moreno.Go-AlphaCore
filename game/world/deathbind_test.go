package world

import (
	"context"
	"encoding/binary"
	"math"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestBinderAndDeathbind(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Player", Zone: 7, Health: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, npc_flags) VALUES (100, 20, 'Binder', 16); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z) VALUES (1, 100, 0, 1, 2, 3)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Player", Map: 0, Zone: 7, Health: 1}
	binderGUID := uint64(0xf130000000000001)
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, binderGUID)
	responses, err := server.binderActivate(active, data)
	if err != nil || len(responses) != 2 {
		t.Fatalf("binder responses=%d err=%v", len(responses), err)
	}
	point, err := packet.Parse(responses[0])
	if err != nil || point.Opcode != packet.SMSGBindPointUpdate || len(point.Data) != 16 || math.Float32frombits(binary.LittleEndian.Uint32(point.Data)) != 1 || math.Float32frombits(binary.LittleEndian.Uint32(point.Data[4:])) != 2 || math.Float32frombits(binary.LittleEndian.Uint32(point.Data[8:])) != 3 || binary.LittleEndian.Uint32(point.Data[12:]) != 0 {
		t.Fatalf("point=%#v err=%v", point, err)
	}
	bound, err := packet.Parse(responses[1])
	if err != nil || bound.Opcode != packet.SMSGPlayerBound || binary.LittleEndian.Uint64(bound.Data) != binderGUID {
		t.Fatalf("bound=%#v err=%v", bound, err)
	}
	zone, err := server.deathBindZone(active)
	if err != nil {
		t.Fatal(err)
	}
	zonePacket, err := packet.Parse(zone)
	if err != nil || zonePacket.Opcode != packet.SMSGBindZoneReply || binary.LittleEndian.Uint32(zonePacket.Data[4:]) != 7 {
		t.Fatalf("zone=%#v err=%v", zonePacket, err)
	}
	active.Health = 0
	world, err := server.repop(&active)
	if err != nil {
		t.Fatal(err)
	}
	worldPacket, err := packet.Parse(world)
	if err != nil || worldPacket.Opcode != packet.SMSGNewWorld || math.Float32frombits(binary.LittleEndian.Uint32(worldPacket.Data[1:])) != 1 || math.Float32frombits(binary.LittleEndian.Uint32(worldPacket.Data[5:])) != 2 || math.Float32frombits(binary.LittleEndian.Uint32(worldPacket.Data[9:])) != 3 {
		t.Fatalf("repop=%#v err=%v", worldPacket, err)
	}
}
