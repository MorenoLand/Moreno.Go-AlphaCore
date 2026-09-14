package world

import (
	"context"
	"encoding/binary"
	"net"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

func TestCastSpellAppliesHealAndSendsPackets(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, BaseLevel, Effect_1, EffectBasePoints_1) VALUES (42, 1, 10, 5)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Race: 1, Class: 1, Level: 5, Map: 0, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(guid, 42); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Caster", Level: 5, Map: 0, Health: 10}
	server.registerPlayer(active)
	client, connection := net.Pipe()
	defer client.Close()
	server.attachPlayer(guid, connection)
	data := make([]byte, 6)
	binary.LittleEndian.PutUint32(data, 42)
	done := make(chan error, 1)
	go func() { _, err := server.castSpellPacket(active, data); done <- err }()
	result, err := sockets.ReadPacket(client)
	if err != nil || result.Opcode != packet.SMSGCastResult || len(result.Data) != 5 || binary.LittleEndian.Uint32(result.Data) != 42 || result.Data[4] != byte(packet.SpellCastSuccess) {
		t.Fatalf("cast result=%#v err=%v", result, err)
	}
	goPacket, err := sockets.ReadPacket(client)
	if err != nil || goPacket.Opcode != packet.SMSGSpellGo || len(goPacket.Data) < 31 {
		t.Fatalf("spell go=%#v err=%v", goPacket, err)
	}
	for range 2 {
		if _, err := sockets.ReadPacket(client); err != nil {
			t.Fatal(err)
		}
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	stored, found, err := characters.Character(guid, 1, 1)
	if err != nil || !found || stored.Health != 15 {
		t.Fatalf("stored=%#v found=%v err=%v", stored, found, err)
	}
}

func TestCastSpellRejectsUnknownSpell(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID) VALUES (42)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, Health: 10}
	data := make([]byte, 6)
	binary.LittleEndian.PutUint32(data, 42)
	responses, err := server.castSpellPacket(active, data)
	if err != nil || len(responses) != 1 {
		t.Fatalf("responses=%d err=%v", len(responses), err)
	}
	response, err := packet.Parse(responses[0])
	if err != nil || response.Opcode != packet.SMSGCastResult || len(response.Data) != 6 || response.Data[4] != byte(packet.SpellCastFailed) || response.Data[5] != byte(packet.SpellFailedNotKnown) {
		t.Fatalf("response=%#v err=%v", response, err)
	}
}

func TestUseItemConsumesChargedSpell(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, BaseLevel, Effect_1, EffectBasePoints_1) VALUES (100, 1, 10, 3)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, spellid_1, spelltrigger_1, spellcharges_1) VALUES (200, 'Healing Potion', 100, 0, 1)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Level: 5, Map: 0, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	item, err := characters.CreateInventoryItem(guid, 0, 23, 0, 200, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.Realm).Exec(`UPDATE character_inventory SET SpellCharges1 = 1 WHERE guid = ?`, item.GUID); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Caster", Level: 5, Map: 0, Health: 10}
	data := []byte{0xff, 0, 0, 0, 0}
	if responses, err := server.useItemPacket(active, data); err != nil || len(responses) != 0 {
		t.Fatalf("responses=%d err=%v", len(responses), err)
	}
	stored, found, err := characters.Character(guid, 1, 1)
	if err != nil || !found || stored.Health != 13 {
		t.Fatalf("stored=%#v found=%v err=%v", stored, found, err)
	}
	if _, found, err := characters.ItemAt(guid, 23, 0); err != nil || found {
		t.Fatalf("consumed item found=%v err=%v", found, err)
	}
}

func TestTriggeredSpellAppliesEffect(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, Effect_1, EffectTriggerSpell_1, EffectBasePoints_1) VALUES (42, 64, 43, 2), (43, 10, 2, 4)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(guid, 42); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, Health: 10}
	if responses, err := server.castSpellPacket(active, append(encodeUint32(42), 0, 0)); err != nil || len(responses) != 0 {
		t.Fatalf("responses=%d err=%v", len(responses), err)
	}
	stored, found, err := characters.Character(guid, 1, 1)
	if err != nil || !found || stored.Health != 14 {
		t.Fatalf("stored=%#v found=%v err=%v", stored, found, err)
	}
}

func TestSpellCreatesItem(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, Effect_1, EffectItemType_1, EffectBasePoints_1) VALUES (42, 24, 200, 2)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, stackable) VALUES (200, 'Created Item', 20)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(guid, 42); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, Health: 10}
	if responses, err := server.castSpellPacket(active, append(encodeUint32(42), 0, 0)); err != nil || len(responses) != 0 {
		t.Fatalf("responses=%d err=%v", len(responses), err)
	}
	item, found, err := characters.ItemAt(guid, 23, 23)
	if err != nil || !found || item.ItemTemplate != 200 || item.StackCount != 2 {
		t.Fatalf("item=%#v found=%v err=%v", item, found, err)
	}
}

func TestSpellCooldownExpiresOnWire(t *testing.T) {
	server := &WorldServer{}
	active := realm.Character{GUID: 1, AccountID: 1, RealmID: 1, Map: 0, Health: 10}
	server.registerPlayer(active)
	client, connection := net.Pipe()
	defer client.Close()
	server.attachPlayer(active.GUID, connection)
	done := make(chan struct{})
	go func() {
		server.setSpellCooldown(active, dbc.Spell{ID: 42, RecoveryTime: 20})
		close(done)
	}()
	cooldown, err := sockets.ReadPacket(client)
	if err != nil || cooldown.Opcode != packet.SMSGSpellCooldown || len(cooldown.Data) != 14 || binary.LittleEndian.Uint32(cooldown.Data) != 42 {
		t.Fatalf("cooldown=%#v err=%v", cooldown, err)
	}
	clear, err := sockets.ReadPacket(client)
	if err != nil || clear.Opcode != packet.SMSGClearCooldown || len(clear.Data) != 12 || binary.LittleEndian.Uint32(clear.Data) != 42 {
		t.Fatalf("clear=%#v err=%v", clear, err)
	}
	<-done
	if server.spellOnCooldown(active.GUID, 42) {
		t.Fatal("spell cooldown remained after expiry")
	}
}

func TestEnemyAreaSpellTargetsNearbyPlayer(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, Effect_1, EffectBasePoints_1, ImplicitTargetA_1, ImplicitTargetB_1, EffectRadiusIndex_1) VALUES (42, 2, 3, 22, 15, 1); INSERT INTO SpellRadius (ID, Radius, RadiusMax) VALUES (1, 10, 10)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	casterID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, PositionX: 0, PositionY: 0, PositionZ: 0, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	targetID, err := characters.Create(realm.Character{AccountID: 2, RealmID: 1, Name: "Target", Level: 1, Map: 0, PositionX: 3, PositionY: 0, PositionZ: 0, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(casterID, 42); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases)}
	caster := realm.Character{GUID: casterID, AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, Health: 10}
	target := realm.Character{GUID: targetID, AccountID: 2, RealmID: 1, Name: "Target", Level: 1, Map: 0, PositionX: 3, Health: 10}
	server.registerPlayer(caster)
	server.registerPlayer(target)
	if responses, err := server.castSpellPacket(caster, append(encodeUint32(42), 0, 0)); err != nil || len(responses) != 0 {
		t.Fatalf("responses=%d err=%v", len(responses), err)
	}
	storedCaster, _, _ := characters.Character(casterID, 1, 1)
	storedTarget, _, _ := characters.Character(targetID, 2, 1)
	if storedCaster.Health != 10 || storedTarget.Health != 7 {
		t.Fatalf("caster health=%d target health=%d", storedCaster.Health, storedTarget.Health)
	}
}

func TestMixedSpellEffectsUseSeparateTargets(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, Effect_1, EffectBasePoints_1, ImplicitTargetA_1, ImplicitTargetB_1, EffectRadiusIndex_1, Effect_2, EffectBasePoints_2, ImplicitTargetA_2) VALUES (42, 2, 3, 22, 15, 1, 10, 2, 1); INSERT INTO SpellRadius (ID, Radius, RadiusMax) VALUES (1, 10, 10)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	casterID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, PositionX: 0, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	targetID, err := characters.Create(realm.Character{AccountID: 2, RealmID: 1, Name: "Target", Level: 1, Map: 0, PositionX: 3, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(casterID, 42); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases)}
	caster := realm.Character{GUID: casterID, AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, Health: 10}
	target := realm.Character{GUID: targetID, AccountID: 2, RealmID: 1, Name: "Target", Level: 1, Map: 0, PositionX: 3, Health: 10}
	server.registerPlayer(caster)
	server.registerPlayer(target)
	if _, err := server.castSpellPacket(caster, append(encodeUint32(42), 0, 0)); err != nil {
		t.Fatal(err)
	}
	storedCaster, _, _ := characters.Character(casterID, 1, 1)
	storedTarget, _, _ := characters.Character(targetID, 2, 1)
	if storedCaster.Health != 12 || storedTarget.Health != 7 {
		t.Fatalf("caster health=%d target health=%d", storedCaster.Health, storedTarget.Health)
	}
}

func TestHarmfulSpellRejectsFriendlyTarget(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, Effect_1, EffectBasePoints_1) VALUES (42, 2, 3); INSERT INTO ChrRaces (ID, BaseLanguage) VALUES (1, 1), (2, 1)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	casterID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Race: 1, Level: 1, Map: 0, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	targetID, err := characters.Create(realm.Character{AccountID: 2, RealmID: 1, Name: "Target", Race: 2, Level: 1, Map: 0, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(casterID, 42); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases)}
	caster := realm.Character{GUID: casterID, AccountID: 1, RealmID: 1, Name: "Caster", Race: 1, Level: 1, Map: 0, Health: 10}
	target := realm.Character{GUID: targetID, AccountID: 2, RealmID: 1, Name: "Target", Race: 2, Level: 1, Map: 0, Health: 10}
	server.registerPlayer(caster)
	server.registerPlayer(target)
	data := append(encodeUint32(42), byte(packet.SpellTargetUnit), 0)
	data = append(data, encodeUint64(uint64(targetID))...)
	responses, err := server.castSpellPacket(caster, data)
	if err != nil || len(responses) != 1 {
		t.Fatalf("responses=%d err=%v", len(responses), err)
	}
	response, err := packet.Parse(responses[0])
	if err != nil || response.Opcode != packet.SMSGCastResult || len(response.Data) != 6 || response.Data[5] != byte(packet.SpellFailedTargetFriendly) {
		t.Fatalf("response=%#v err=%v", response, err)
	}
}

func TestSpellHealthEffectsUsePlayerVitals(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID, Effect_1, EffectBasePoints_1) VALUES (42, 67, 0), (43, 1, 0)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, Health: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(guid, 42); err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(guid, 43); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Caster", Level: 1, Map: 0, Health: 100}
	server.registerPlayer(active)
	active.Health = 20
	server.updatePlayer(active)
	if err := characters.UpdateHealth(guid, 1, 1, 20); err != nil {
		t.Fatal(err)
	}
	if _, err := server.castSpellPacket(active, append(encodeUint32(42), 0, 0)); err != nil {
		t.Fatal(err)
	}
	stored, _, _ := characters.Character(guid, 1, 1)
	if stored.Health != 100 {
		t.Fatalf("healed health=%d", stored.Health)
	}
	active.Health = 100
	if _, err := server.castSpellPacket(active, append(encodeUint32(43), 0, 0)); err != nil {
		t.Fatal(err)
	}
	stored, _, _ = characters.Character(guid, 1, 1)
	if stored.Health != 0 {
		t.Fatalf("killed health=%d", stored.Health)
	}
}

func TestSpellBindAndStuckUseDeathbind(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Caster", Map: 1, PositionX: 10, PositionY: 20, PositionZ: 30, Health: 10})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Caster", Map: 1, PositionX: 10, PositionY: 20, PositionZ: 30, Health: 10}
	server.bindSpellTarget(&spellCast{caster: active}, active)
	bind, found, err := characters.Deathbind(guid)
	if err != nil || !found || bind.Map != 1 || bind.X != 10 || bind.Y != 20 || bind.Z != 30 {
		t.Fatalf("bind=%#v found=%v err=%v", bind, found, err)
	}
	active.Map, active.PositionX, active.PositionY, active.PositionZ = 2, 1, 2, 3
	if err := characters.UpdateLocation(guid, 1, 1, active.Map, active.PositionX, active.PositionY, active.PositionZ, active.Orientation); err != nil {
		t.Fatal(err)
	}
	server.stuckSpellTarget(active)
	stored, found, err := characters.Character(guid, 1, 1)
	if err != nil || !found || stored.Map != 1 || stored.PositionX != 10 || stored.PositionY != 20 || stored.PositionZ != 30 {
		t.Fatalf("stored=%#v found=%v err=%v", stored, found, err)
	}
}
