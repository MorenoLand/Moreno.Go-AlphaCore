package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
)

func TestCreatureTemplateSpellIDs(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, name, spell_id1, spell_id2, spell_id3, spell_id4, loot_id, gold_min, gold_max) VALUES (100, 'Innate Spells', 11, 12, 13, 14, 500, 2, 3)`); err != nil {
		t.Fatal(err)
	}
	creature, found, err := NewStore(databases).CreatureTemplate(100)
	if err != nil || !found || creature.SpellIDs != [4]int64{11, 12, 13, 14} || creature.LootID != 500 || creature.GoldMin != 2 || creature.GoldMax != 3 {
		t.Fatalf("creature=%#v found=%v err=%v", creature, found, err)
	}
}
