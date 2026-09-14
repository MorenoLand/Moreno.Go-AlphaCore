package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestGMCheats(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "GM", Class: 1, Level: 1, Money: 10, Health: 100, Power1: 50})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO player_classlevelstats ("class", level, basehp, basemana) VALUES (1, 2, 150, 200)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, stackable, bonding) VALUES (700, 'Test Item', 20, 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO Spell (ID) VALUES (9001)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "GM", Class: 1, Level: 1, Money: 10, Health: 100, Power1: 50}
	server.registerPlayer(active)
	if responses, err := server.gmCheat(&active, packet.CMSGCheatSetMoney, cheatUint32(100), 0); err != nil || len(responses) != 0 || active.Money != 10 {
		t.Fatalf("non-gm money responses=%d err=%v money=%d", len(responses), err, active.Money)
	}
	if responses, err := server.gmCheat(&active, packet.CMSGCheatSetMoney, cheatUint32(100), 1); err != nil || len(responses) != 1 || active.Money != 110 {
		t.Fatalf("money responses=%d err=%v money=%d", len(responses), err, active.Money)
	}
	if responses, err := server.gmCheat(&active, packet.CMSGLevelCheat, cheatUint32(2), 1); err != nil || len(responses) == 0 || active.Level != 2 || active.Talentpoints != 10 || active.Skillpoints != 1 || active.Health != 150 || active.Power1 != 200 {
		t.Fatalf("level responses=%d err=%v character=%#v", len(responses), err, active)
	}
	if responses, err := server.gmCheat(&active, packet.CMSGLevelUpCheat, nil, 1); err != nil || len(responses) == 0 || active.Level != 3 || active.Talentpoints != 20 || active.Skillpoints != 2 {
		t.Fatalf("levelup responses=%d err=%v character=%#v", len(responses), err, active)
	}
	if responses, err := server.gmCheat(&active, packet.CMSGLearnSpell, cheatUint32(9001), 1); err != nil || len(responses) != 1 {
		t.Fatalf("learn responses=%d err=%v", len(responses), err)
	} else if learned, parseErr := packet.Parse(responses[0]); parseErr != nil || learned.Opcode != packet.SMSGLearnedSpell || binary.LittleEndian.Uint16(learned.Data) != 9001 {
		t.Fatalf("learn packet=%#v err=%v", learned, parseErr)
	}
	if responses, err := server.gmCheat(&active, packet.CMSGCreateItem, cheatUint32(700), 1); err != nil || len(responses) != 2 {
		t.Fatalf("item responses=%d err=%v", len(responses), err)
	}
	if responses, err := server.gmCheat(&active, packet.CMSGCreateItem, cheatUint32(700), 1); err != nil || len(responses) != 2 {
		t.Fatalf("stack responses=%d err=%v", len(responses), err)
	}
	items, err := characters.InventoryItems(guid, 23, 23, 39)
	if err != nil || len(items) != 1 || items[0].StackCount != 2 || items[0].Flags != itemDynBound {
		t.Fatalf("items=%#v err=%v", items, err)
	}
	active.Power1 = 10
	server.updatePlayer(active)
	if responses, err := server.gmCheat(&active, packet.CMSGRecharge, nil, 1); err != nil || len(responses) != 1 || active.Power1 != 1000 {
		t.Fatalf("recharge responses=%d err=%v power=%d", len(responses), err, active.Power1)
	}
	server.setSpellCooldown(active, dbc.Spell{ID: 9001, RecoveryTime: 1000})
	if !server.spellOnCooldown(guid, 9001) {
		t.Fatal("cooldown was not set")
	}
	if responses, err := server.gmCheat(&active, packet.CMSGCooldownCheat, nil, 1); err != nil || len(responses) != 1 {
		t.Fatalf("cooldown responses=%d err=%v", len(responses), err)
	}
	if server.spellOnCooldown(guid, 9001) {
		t.Fatal("cooldown was not cleared")
	}
	if responses, err := server.gmCheat(&active, packet.CMSGGodMode, []byte{1}, 1); err != nil || len(responses) != 2 || !server.isGodMode(guid) {
		t.Fatalf("godmode responses=%d err=%v enabled=%v", len(responses), err, server.isGodMode(guid))
	}
	health := active.Health
	if err := server.changePlayerHealth(&active, -10); err != nil || active.Health != health {
		t.Fatalf("godmode health=%d before=%d err=%v", active.Health, health, err)
	}
	if _, err := server.gmCheat(&active, packet.CMSGGodMode, []byte{0}, 1); err != nil || server.isGodMode(guid) {
		t.Fatalf("disable godmode err=%v enabled=%v", err, server.isGodMode(guid))
	}
	if err := server.changePlayerHealth(&active, -10); err != nil || active.Health != health-10 {
		t.Fatalf("normal health=%d want=%d err=%v", active.Health, health-10, err)
	}
	if _, err := server.gmCheat(&active, packet.CMSGEnableDebugCombatLogging, cheatUint32(1), 1); err != nil || server.unitFlags(guid)&unitFlagDebugCombatLog == 0 {
		t.Fatalf("debug logging err=%v flags=%x", err, server.unitFlags(guid))
	}
	stored, found, err := characters.CharacterByGUID(guid)
	if err != nil || !found || stored.Level != 3 || stored.Money != 110 || stored.Power1 != 1000 {
		t.Fatalf("stored=%#v found=%v err=%v", stored, found, err)
	}
}

func cheatUint32(value uint32) []byte {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, value)
	return data
}
