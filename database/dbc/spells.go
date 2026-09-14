package dbc

import (
	"database/sql"
	"fmt"
)

type SpellEffect struct {
	Type, DieSides, BaseDice, DicePerLevel, BasePoints              int64
	ImplicitTargetA, ImplicitTargetB, RadiusIndex, Aura, AuraPeriod int64
	ChainTargets, ItemType, MiscValue, TriggerSpell                 int64
	RealPointsPerLevel                                              float32
}

type Spell struct {
	ID, School, Category, CastUI, Attributes, AttributesEx, ShapeshiftMask      int64
	Targets, TargetCreatureType, RequiresSpellFocus, CasterAuraState            int64
	TargetAuraState, CastingTimeIndex, RecoveryTime, CategoryRecoveryTime       int64
	InterruptFlags, AuraInterruptFlags, ChannelInterruptFlags, ProcFlags        int64
	ProcChance, ProcCharges, MaxLevel, BaseLevel, SpellLevel, DurationIndex     int64
	PowerType, ManaCost, ManaCostPerLevel, ManaPerSecond, ManaPerSecondPerLevel int64
	RangeIndex, ManaCostPct, StartRecoveryCategory, StartRecoveryTime           int64
	Speed                                                                       float32
	EquippedItemClass, EquippedItemSubclass, SpellVisualID                      int64
	Name                                                                        string
	Effects                                                                     [3]SpellEffect
}

type SpellCastTime struct {
	ID, Base, PerLevel, Minimum int64
}

type SpellDuration struct {
	ID, Duration, DurationPerLevel, MaxDuration int64
}

type SpellRange struct {
	ID, Flags          int64
	RangeMin, RangeMax float32
}

type SpellRadius struct {
	ID             int64
	Radius         float32
	RadiusPerLevel float32
	RadiusMax      float32
}

const spellColumns = `ID, School, Category, CastUI, Attributes, AttributesEx, ShapeshiftMask, Targets, TargetCreatureType, RequiresSpellFocus, CasterAuraState, TargetAuraState, CastingTimeIndex, RecoveryTime, CategoryRecoveryTime, InterruptFlags, AuraInterruptFlags, ChannelInterruptFlags, ProcFlags, ProcChance, ProcCharges, MaxLevel, BaseLevel, SpellLevel, DurationIndex, PowerType, ManaCost, ManaCostPerLevel, ManaPerSecond, ManaPerSecondPerLevel, RangeIndex, Speed, ManaCostPct, StartRecoveryCategory, StartRecoveryTime, EquippedItemClass, EquippedItemSubclass, Effect_1, Effect_2, Effect_3, EffectDieSides_1, EffectDieSides_2, EffectDieSides_3, EffectBaseDice_1, EffectBaseDice_2, EffectBaseDice_3, EffectDicePerLevel_1, EffectDicePerLevel_2, EffectDicePerLevel_3, EffectRealPointsPerLevel_1, EffectRealPointsPerLevel_2, EffectRealPointsPerLevel_3, EffectBasePoints_1, EffectBasePoints_2, EffectBasePoints_3, ImplicitTargetA_1, ImplicitTargetA_2, ImplicitTargetA_3, ImplicitTargetB_1, ImplicitTargetB_2, ImplicitTargetB_3, EffectRadiusIndex_1, EffectRadiusIndex_2, EffectRadiusIndex_3, EffectAura_1, EffectAura_2, EffectAura_3, EffectAuraPeriod_1, EffectAuraPeriod_2, EffectAuraPeriod_3, EffectChainTargets_1, EffectChainTargets_2, EffectChainTargets_3, EffectItemType_1, EffectItemType_2, EffectItemType_3, EffectMiscValue_1, EffectMiscValue_2, EffectMiscValue_3, EffectTriggerSpell_1, EffectTriggerSpell_2, EffectTriggerSpell_3, SpellVisualID, COALESCE(Name_enUS, '')`

func (s *Store) Spell(id int64) (Spell, bool, error) {
	var spell Spell
	values := []any{&spell.ID, &spell.School, &spell.Category, &spell.CastUI, &spell.Attributes, &spell.AttributesEx, &spell.ShapeshiftMask, &spell.Targets, &spell.TargetCreatureType, &spell.RequiresSpellFocus, &spell.CasterAuraState, &spell.TargetAuraState, &spell.CastingTimeIndex, &spell.RecoveryTime, &spell.CategoryRecoveryTime, &spell.InterruptFlags, &spell.AuraInterruptFlags, &spell.ChannelInterruptFlags, &spell.ProcFlags, &spell.ProcChance, &spell.ProcCharges, &spell.MaxLevel, &spell.BaseLevel, &spell.SpellLevel, &spell.DurationIndex, &spell.PowerType, &spell.ManaCost, &spell.ManaCostPerLevel, &spell.ManaPerSecond, &spell.ManaPerSecondPerLevel, &spell.RangeIndex, &spell.Speed, &spell.ManaCostPct, &spell.StartRecoveryCategory, &spell.StartRecoveryTime, &spell.EquippedItemClass, &spell.EquippedItemSubclass}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.Type)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.DieSides)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.BaseDice)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.DicePerLevel)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.RealPointsPerLevel)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.BasePoints)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.ImplicitTargetA)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.ImplicitTargetB)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.RadiusIndex)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.Aura)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.AuraPeriod)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.ChainTargets)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.ItemType)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.MiscValue)
	}
	for index := range spell.Effects {
		effect := &spell.Effects[index]
		values = append(values, &effect.TriggerSpell)
	}
	values = append(values, &spell.SpellVisualID, &spell.Name)
	err := s.db.QueryRow(`SELECT `+spellColumns+` FROM Spell WHERE ID = ?`, id).Scan(values...)
	if err == sql.ErrNoRows {
		return Spell{}, false, nil
	}
	if err != nil {
		return Spell{}, false, fmt.Errorf("query spell: %w", err)
	}
	return spell, true, nil
}

func (s *Store) SpellCastTime(id int64) (SpellCastTime, bool, error) {
	var value SpellCastTime
	err := s.db.QueryRow(`SELECT ID, Base, PerLevel, Minimum FROM SpellCastTimes WHERE ID = ?`, id).Scan(&value.ID, &value.Base, &value.PerLevel, &value.Minimum)
	if err == sql.ErrNoRows {
		return SpellCastTime{}, false, nil
	}
	if err != nil {
		return SpellCastTime{}, false, fmt.Errorf("query spell cast time: %w", err)
	}
	return value, true, nil
}

func (s *Store) SpellDuration(id int64) (SpellDuration, bool, error) {
	var value SpellDuration
	err := s.db.QueryRow(`SELECT ID, Duration, DurationPerLevel, MaxDuration FROM SpellDuration WHERE ID = ?`, id).Scan(&value.ID, &value.Duration, &value.DurationPerLevel, &value.MaxDuration)
	if err == sql.ErrNoRows {
		return SpellDuration{}, false, nil
	}
	if err != nil {
		return SpellDuration{}, false, fmt.Errorf("query spell duration: %w", err)
	}
	return value, true, nil
}

func (s *Store) SpellRange(id int64) (SpellRange, bool, error) {
	var value SpellRange
	err := s.db.QueryRow(`SELECT ID, RangeMin, RangeMax, Flags FROM SpellRange WHERE ID = ?`, id).Scan(&value.ID, &value.RangeMin, &value.RangeMax, &value.Flags)
	if err == sql.ErrNoRows {
		return SpellRange{}, false, nil
	}
	if err != nil {
		return SpellRange{}, false, fmt.Errorf("query spell range: %w", err)
	}
	return value, true, nil
}

func (s *Store) SpellRadius(id int64) (SpellRadius, bool, error) {
	var value SpellRadius
	err := s.db.QueryRow(`SELECT ID, Radius, RadiusPerLevel, RadiusMax FROM SpellRadius WHERE ID = ?`, id).Scan(&value.ID, &value.Radius, &value.RadiusPerLevel, &value.RadiusMax)
	if err == sql.ErrNoRows {
		return SpellRadius{}, false, nil
	}
	if err != nil {
		return SpellRadius{}, false, fmt.Errorf("query spell radius: %w", err)
	}
	return value, true, nil
}
