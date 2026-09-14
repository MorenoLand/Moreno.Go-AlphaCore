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

func TestPetRuntimeLifecycle(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Owner", Level: 4, Health: 100})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, faction, unit_class, beast_family) VALUES (123, 456, 'Wolf', 1, 1, 1); INSERT INTO pet_levelstats (creature_entry, level, hp, mana, armor) VALUES (123, 4, 80, 20, 10), (123, 8, 160, 40, 20)`); err != nil {
		t.Fatal(err)
	}
	pet := realm.Pet{OwnerGUID: guid, CreatureID: 123, CreatedBySpell: 883, Level: 4, Name: "Wolf", Active: true, ActionBar: defaultPetActionBar()}
	pet.ID, err = characters.CreatePet(pet)
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddPetSpell(guid, pet.ID, 42); err != nil {
		t.Fatal(err)
	}
	owner := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Owner", Level: 4, Health: 100}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	server.registerPlayer(owner)
	if err := server.loadPets(owner); err != nil {
		t.Fatal(err)
	}
	packets, petGUID, err := server.initialPetPackets(owner)
	if err != nil || len(packets) != 2 || petGUID == 0 {
		t.Fatalf("packets=%d pet=%d err=%v", len(packets), petGUID, err)
	}
	spellInfo, err := packet.Parse(packets[1])
	if err != nil || spellInfo.Opcode != packet.SMSGPetSpells || binary.LittleEndian.Uint64(spellInfo.Data) != petGUID || binary.LittleEndian.Uint32(spellInfo.Data[56:]) != 1 || binary.LittleEndian.Uint16(spellInfo.Data[60:]) != 42 {
		t.Fatalf("spell info=%#v err=%v", spellInfo, err)
	}
	action := make([]byte, 16)
	binary.LittleEndian.PutUint64(action, petGUID)
	binary.LittleEndian.PutUint32(action[8:], 3)
	binary.LittleEndian.PutUint32(action[12:], uint32(42))
	if err := server.petSetAction(owner, action); err != nil {
		t.Fatal(err)
	}
	loaded, found, err := characters.Pet(pet.ID)
	if err != nil || !found || loaded.ActionBar[3] != int64(uint32(42)|0x80<<24) {
		t.Fatalf("action bar=%#v found=%v err=%v", loaded.ActionBar, found, err)
	}
	rename := append(encodeUint64(petGUID), []byte("Rex\x00")...)
	if response, err := server.petRename(owner, rename); err != nil || response != nil {
		t.Fatalf("rename response=%x err=%v", response, err)
	}
	nameResponse, err := server.petNameQuery(owner, append(encodeUint32(9), encodeUint64(petGUID)...))
	if err != nil {
		t.Fatal(err)
	}
	namePacket, err := packet.Parse(nameResponse)
	if err != nil || namePacket.Opcode != packet.SMSGPetNameQueryResponse || binary.LittleEndian.Uint32(namePacket.Data) != 9 {
		t.Fatalf("name packet=%#v err=%v", namePacket, err)
	}
	name, err := packet.ReadString(namePacket.Data, 4, 0)
	if err != nil || name != "Rex" {
		t.Fatalf("name=%q err=%v", name, err)
	}
	if err := server.petLevelCheat(owner, encodeUint32(8)); err != nil {
		t.Fatal(err)
	}
	loaded, found, err = characters.Pet(pet.ID)
	if err != nil || !found || loaded.Level != 8 || loaded.XP != 0 {
		t.Fatalf("leveled pet=%#v found=%v err=%v", loaded, found, err)
	}
	command := make([]byte, 20)
	binary.LittleEndian.PutUint64(command, petGUID)
	binary.LittleEndian.PutUint32(command[8:], petCommandFlag|uint32(petCommandDismiss))
	if err := server.petAction(owner, command); err != nil {
		t.Fatal(err)
	}
	loaded, found, err = characters.Pet(pet.ID)
	ownerGUID, _, _, activeFound := server.findActivePet(petGUID)
	if err != nil || !found || loaded.Active || activeFound || ownerGUID != 0 {
		t.Fatalf("dismissed pet=%#v found=%v err=%v", loaded, found, err)
	}
}
