package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
)

func TestCreatureSpellQuery(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_spells (entry, name, spellId_1, probability_1, castTarget_1, delayInitialMin_1, delayInitialMax_1, delayRepeatMin_1, delayRepeatMax_1, scriptId_1, spellId_8) VALUES (400, 'Test List', 6016, 75, 1, 4, 14, 38, 42, 99, 7000)`); err != nil {
		t.Fatal(err)
	}
	spell, found, err := NewStore(databases).CreatureSpell(400)
	if err != nil || !found || spell.Name != "Test List" || spell.Spells[0].SpellID != 6016 || spell.Spells[0].Probability != 75 || spell.Spells[0].DelayRepeatMax != 42 || spell.Spells[0].ScriptID != 99 || spell.Spells[7].SpellID != 7000 {
		t.Fatalf("spell=%#v found=%v err=%v", spell, found, err)
	}
}
