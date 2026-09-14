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

func TestCharacterListIncludesActivePet(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Owner"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, beast_family) VALUES (123, 456, 'Wolf', 7)`); err != nil {
		t.Fatal(err)
	}
	if _, err := characters.CreatePet(realm.Pet{OwnerGUID: guid, CreatureID: 123, Level: 8, Active: true}); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	response, err := server.characterList(1)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := packet.Parse(response)
	if err != nil || parsed.Opcode != packet.SMSGCharEnum || parsed.Data[0] != 1 {
		t.Fatalf("response=%#v err=%v", parsed, err)
	}
	name, err := packet.ReadString(parsed.Data, 9, 0)
	if err != nil || name != "Owner" {
		t.Fatalf("name=%q err=%v", name, err)
	}
	petStart := 1 + 8 + len(name) + 1 + 9 + 8 + 12 + 4
	if binary.LittleEndian.Uint32(parsed.Data[petStart:]) != 456 || binary.LittleEndian.Uint32(parsed.Data[petStart+4:]) != 8 || binary.LittleEndian.Uint32(parsed.Data[petStart+8:]) != 7 {
		t.Fatalf("pet info=%x", parsed.Data[petStart:petStart+12])
	}
}
