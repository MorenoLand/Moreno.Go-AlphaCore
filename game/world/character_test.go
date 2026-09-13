package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	realmdb "Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestCharacterCreateListDelete(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO playercreateinfo (id, race, "class", map, zone, position_x, position_y, position_z, orientation) VALUES (1, 1, 1, 0, 12, -8949.95, -132.493, 83.5312, 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO player_classlevelstats ("class", level, basehp, basemana) VALUES (1, 1, 20, 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, "class", inventory_type) VALUES (38, 'Recruit Shirt', 4, 4), (25, 'Potion', 0, 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO playercreateinfo_item (id, race, "class", itemid, amount) VALUES (1, 1, 1, 38, 1), (2, 1, 1, 25, 1)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: realmdb.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	data := append([]byte("Testone\x00"), []byte{1, 1, 0, 0, 0, 0, 0, 0, 0}...)
	response, err := server.characterCreate(1, data)
	if err != nil {
		t.Fatal(err)
	}
	created, err := packet.Parse(response)
	if err != nil || created.Opcode != packet.SMSGCharCreate || len(created.Data) != 1 || created.Data[0] != 0x28 {
		t.Fatalf("created=%#v err=%v", created, err)
	}
	response, err = server.characterList(1)
	if err != nil {
		t.Fatal(err)
	}
	listed, err := packet.Parse(response)
	if err != nil || listed.Opcode != packet.SMSGCharEnum || len(listed.Data) < 2 || listed.Data[0] != 1 {
		t.Fatalf("listed=%#v err=%v", listed, err)
	}
	characters, err := server.Characters.Characters(1, 1)
	if err != nil || len(characters) != 1 {
		t.Fatalf("characters=%#v err=%v", characters, err)
	}
	shirt, found, err := server.Characters.ItemAt(characters[0].GUID, 23, 3)
	if err != nil || !found || shirt.ItemTemplate != 38 {
		t.Fatalf("shirt=%#v found=%v err=%v", shirt, found, err)
	}
	potion, found, err := server.Characters.ItemAt(characters[0].GUID, 23, 23)
	if err != nil || !found || potion.ItemTemplate != 25 || potion.StackCount != 4 {
		t.Fatalf("potion=%#v found=%v err=%v", potion, found, err)
	}
	guid := make([]byte, 8)
	binary.LittleEndian.PutUint64(guid, uint64(characters[0].GUID))
	response, err = server.characterDelete(1, guid)
	if err != nil {
		t.Fatal(err)
	}
	deleted, err := packet.Parse(response)
	if err != nil || deleted.Opcode != packet.SMSGCharDelete || len(deleted.Data) != 1 || deleted.Data[0] != 0x2e {
		t.Fatalf("deleted=%#v err=%v", deleted, err)
	}
}
