package dbc

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
)

func TestSpellData(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, BaseLevel, Effect_1, EffectDieSides_1, EffectBaseDice_1, EffectDicePerLevel_1, EffectRealPointsPerLevel_1, EffectBasePoints_1, ImplicitTargetA_1, ImplicitTargetB_1, EffectRadiusIndex_1, EffectAura_1, EffectAuraPeriod_1, EffectChainTargets_1, EffectItemType_1, EffectMiscValue_1, EffectTriggerSpell_1, SpellVisualID) VALUES (42, 7, 10, 2, 3, 4, 1.5, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)`); err != nil {
		t.Fatal(err)
	}
	spell, found, err := NewStore(databases).Spell(42)
	if err != nil || !found || spell.BaseLevel != 7 || spell.Effects[0].Type != 10 || spell.Effects[0].DieSides != 2 || spell.Effects[0].BaseDice != 3 || spell.Effects[0].DicePerLevel != 4 || spell.Effects[0].RealPointsPerLevel != 1.5 || spell.Effects[0].BasePoints != 5 || spell.Effects[0].ImplicitTargetA != 6 || spell.Effects[0].ImplicitTargetB != 7 || spell.Effects[0].RadiusIndex != 8 || spell.Effects[0].Aura != 9 || spell.Effects[0].AuraPeriod != 10 || spell.Effects[0].ChainTargets != 11 || spell.Effects[0].ItemType != 12 || spell.Effects[0].MiscValue != 13 || spell.Effects[0].TriggerSpell != 14 || spell.SpellVisualID != 15 || spell.Name != "" {
		t.Fatalf("spell=%#v found=%v err=%v", spell, found, err)
	}
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO SpellCastTimes (ID, Base, PerLevel, Minimum) VALUES (1, 100, 5, 50); INSERT INTO SpellDuration (ID, Duration, DurationPerLevel, MaxDuration) VALUES (2, 1000, 10, 2000); INSERT INTO SpellRange (ID, RangeMin, RangeMax, Flags) VALUES (3, 1, 30, 4); INSERT INTO SpellRadius (ID, Radius, RadiusPerLevel, RadiusMax) VALUES (4, 5, 1, 10)`); err != nil {
		t.Fatal(err)
	}
	castTime, castFound, err := NewStore(databases).SpellCastTime(1)
	if err != nil || !castFound || castTime.Base != 100 || castTime.PerLevel != 5 || castTime.Minimum != 50 {
		t.Fatalf("cast time=%#v found=%v err=%v", castTime, castFound, err)
	}
	duration, durationFound, err := NewStore(databases).SpellDuration(2)
	if err != nil || !durationFound || duration.Duration != 1000 || duration.DurationPerLevel != 10 || duration.MaxDuration != 2000 {
		t.Fatalf("duration=%#v found=%v err=%v", duration, durationFound, err)
	}
	rangeEntry, rangeFound, err := NewStore(databases).SpellRange(3)
	if err != nil || !rangeFound || rangeEntry.RangeMin != 1 || rangeEntry.RangeMax != 30 || rangeEntry.Flags != 4 {
		t.Fatalf("range=%#v found=%v err=%v", rangeEntry, rangeFound, err)
	}
	radius, radiusFound, err := NewStore(databases).SpellRadius(4)
	if err != nil || !radiusFound || radius.Radius != 5 || radius.RadiusPerLevel != 1 || radius.RadiusMax != 10 {
		t.Fatalf("radius=%#v found=%v err=%v", radius, radiusFound, err)
	}
}
