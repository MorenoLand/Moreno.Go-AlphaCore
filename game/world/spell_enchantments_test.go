package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestEnchantSpellEffects(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO SpellItemEnchantment (ID, Effect_1, EffectPointsMin_1, EffectArg_1, Name_enUS) VALUES (9, 2, 5, 0, 'Test Enchant')`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Enchanter", Health: 20})
	if err != nil {
		t.Fatal(err)
	}
	item, err := characters.CreateInventoryItem(guid, guid, 23, 0, 100, 1)
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases)}
	caster := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Enchanter", Health: 20}
	server.applySpellEffects(&spellCast{caster: caster, target: spellTarget{ItemGUID: uint64(item.GUID)}, spell: dbc.Spell{ID: 100, Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectEnchantPermanent), MiscValue: 9}}}})
	stored, found, err := characters.ItemByGUID(guid, item.GUID)
	if err != nil || !found {
		t.Fatalf("stored item=%#v found=%v err=%v", stored, found, err)
	}
	values := itemEnchantments(stored.Enchantments)
	if values[0].ID != 9 || values[0].Duration != -1 || values[0].Charges != 0 {
		t.Fatalf("permanent enchantments=%#v", values)
	}
	server.applySpellEffects(&spellCast{caster: caster, target: spellTarget{ItemGUID: uint64(item.GUID)}, spell: dbc.Spell{ID: 3408, Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectEnchantTemporary), MiscValue: 9}}}})
	stored, found, err = characters.ItemByGUID(guid, item.GUID)
	if err != nil || !found {
		t.Fatalf("temporary item=%#v found=%v err=%v", stored, found, err)
	}
	values = itemEnchantments(stored.Enchantments)
	if values[1].ID != 9 || values[1].Duration != 0 || values[1].Charges != 1 {
		t.Fatalf("temporary enchantments=%#v", values)
	}
}
