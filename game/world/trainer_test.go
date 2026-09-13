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

func TestTrainerListAndBuy(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Student", Level: 1, Money: 200})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, BaseLevel) VALUES (500, 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, npc_flags, trainer_id, trainer_type) VALUES (100, 20, 'Trainer', 8, 10, 0); INSERT INTO trainer_template (template_entry, spell, playerspell, spellcost, reqlevel) VALUES (10, 600, 500, 100, 1); INSERT INTO npc_trainer_greeting (entry, content_default) VALUES (100, 'Train'); INSERT INTO spawns_creatures (spawn_id, spawn_entry1, map, position_x, position_y, position_z) VALUES (1, 100, 0, 0, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Student", Level: 1, Money: 200, Map: 0}
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, 0xf130000000000001)
	responses, err := server.trainerList(active, data)
	if err != nil || len(responses) != 1 {
		t.Fatalf("list responses=%d err=%v", len(responses), err)
	}
	list, err := packet.Parse(responses[0])
	if err != nil || list.Opcode != packet.SMSGTrainerList || len(list.Data) < 8+4+4+33 || binary.LittleEndian.Uint32(list.Data[12:]) != 1 || binary.LittleEndian.Uint32(list.Data[16:]) != 600 {
		t.Fatalf("list=%#v err=%v", list, err)
	}
	buy := make([]byte, 12)
	copy(buy, data)
	binary.LittleEndian.PutUint32(buy[8:], 600)
	responses, err = server.trainerBuy(&active, buy)
	if err != nil || len(responses) != 2 {
		t.Fatalf("buy responses=%d err=%v", len(responses), err)
	}
	succeeded, err := packet.Parse(responses[0])
	if err != nil || succeeded.Opcode != packet.SMSGTrainerBuySucceeded || binary.LittleEndian.Uint32(succeeded.Data[8:]) != 600 || active.Money != 100 {
		t.Fatalf("buy=%#v money=%d err=%v", succeeded, active.Money, err)
	}
	spells, err := characters.Spells(guid)
	if err != nil || len(spells) != 1 || spells[0].ID != 500 {
		t.Fatalf("spells=%#v err=%v", spells, err)
	}
}
