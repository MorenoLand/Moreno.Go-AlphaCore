package world

import (
	"context"
	"math"
	"testing"
	"time"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestDistractSpellEffect(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO SpellDuration (ID, Duration, DurationPerLevel, MaxDuration) VALUES (1, 5000, 0, 5000)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	caster := realm.Character{GUID: 1, Health: 100}
	target := &creatureState{GUID: 2, Health: 100, MaxHealth: 100, Spawn: worlddb.CreatureSpawn{PositionX: 5, PositionY: 5}}
	cast := &spellCast{caster: caster, target: spellTarget{Dest: &spellVector{X: 7, Y: 5}}, targetCreature: target, spell: dbc.Spell{DurationIndex: 1, Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectDistract)}}}}
	server.applySpellEffects(cast)
	if math.Abs(float64(target.DistractedAngle)) > 0.001 || !target.DistractedUntil.After(time.Now()) || target.Spawn.Orientation != target.DistractedAngle {
		t.Fatalf("distracted target=%#v", target)
	}
	angle, until := target.DistractedAngle, target.DistractedUntil
	target.CombatTarget = 3
	server.applySpellEffects(cast)
	if target.DistractedAngle != angle || !target.DistractedUntil.Equal(until) {
		t.Fatalf("combat target changed distraction=%#v", target)
	}
}
