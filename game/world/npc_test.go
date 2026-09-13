package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestBankerAndTabardActivation(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, npc_flags) VALUES (100, 20, 'NPC', 96); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z) VALUES (1, 100, 0, 0, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{WorldData: worlddb.NewStore(databases)}
	active := realm.Character{Map: 0}
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, 0xf130000000000001)
	for _, test := range []struct {
		name   string
		call   func([]byte) ([]byte, error)
		opcode packet.Opcode
	}{
		{"bank", func(data []byte) ([]byte, error) { return server.bankerActivate(active, data) }, packet.SMSGShowBank},
		{"tabard", func(data []byte) ([]byte, error) { return server.tabardVendorActivate(active, data) }, packet.MSGTabardVendorActivate},
	} {
		response, err := test.call(data)
		message, parseErr := packet.Parse(response)
		if err != nil || parseErr != nil || message.Opcode != test.opcode || len(message.Data) != 8 || binary.LittleEndian.Uint64(message.Data) != binary.LittleEndian.Uint64(data) {
			t.Fatalf("%s=%#v err=%v parse=%v", test.name, message, err, parseErr)
		}
	}
}
