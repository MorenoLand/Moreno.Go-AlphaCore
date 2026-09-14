package realm

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
)

func TestPetStore(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	store := NewStore(databases)
	owner, err := store.Create(Character{AccountID: 1, RealmID: 1, Name: "Owner"})
	if err != nil {
		t.Fatal(err)
	}
	pet := Pet{OwnerGUID: owner, CreatureID: 123, CreatedBySpell: 883, Level: 4, XP: 12, ReactState: 1, CommandState: 2, Name: "Wolf", RenameTime: 34, Health: 50, Mana: 20, Active: true}
	pet.ActionBar[0], pet.ActionBar[9] = 0x07000002, 0x8100002a
	pet.ID, err = store.CreatePet(pet)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AddPetSpell(owner, pet.ID, 42); err != nil {
		t.Fatal(err)
	}
	loaded, found, err := store.Pet(pet.ID)
	if err != nil || !found || loaded.OwnerGUID != owner || loaded.CreatureID != pet.CreatureID || loaded.Name != pet.Name || loaded.RenameTime != pet.RenameTime || !loaded.Active || loaded.ActionBar != pet.ActionBar {
		t.Fatalf("loaded=%#v found=%v err=%v", loaded, found, err)
	}
	spells, err := store.PetSpells(owner, pet.ID)
	if err != nil || len(spells) != 1 || spells[0] != 42 {
		t.Fatalf("spells=%#v err=%v", spells, err)
	}
	loaded.Name, loaded.Active = "Dire Wolf", false
	if err := store.UpdatePet(loaded); err != nil {
		t.Fatal(err)
	}
	updated, found, err := store.Pet(pet.ID)
	if err != nil || !found || updated.Name != "Dire Wolf" || updated.Active {
		t.Fatalf("updated=%#v found=%v err=%v", updated, found, err)
	}
	if err := store.DeletePet(pet.ID); err != nil {
		t.Fatal(err)
	}
	spells, err = store.PetSpells(owner, pet.ID)
	if err != nil || len(spells) != 0 {
		t.Fatalf("spells after delete=%#v err=%v", spells, err)
	}
}
