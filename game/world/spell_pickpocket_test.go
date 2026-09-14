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

func TestPickpocketSpellLootLifecycle(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO item_template (entry, name, display_id, stackable) VALUES (200, 'Pocket Loot', 20, 20); INSERT INTO creature_template (entry, display_id1, name, faction, unit_class, level_min, pickpocket_loot_id, health_multiplier, mana_multiplier) VALUES (100, 10, 'Pickpocket Target', 14, 1, 1, 500, 1, 1); INSERT INTO pickpocketing_loot_template (entry, item, ChanceOrQuestChance, mincountOrRef, maxcount) VALUES (500, 200, 100, 1, 1)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Looter", Health: 20})
	if err != nil {
		t.Fatal(err)
	}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Looter", Health: 20, Map: 0}
	template, found, err := worlddb.NewStore(databases).CreatureTemplate(100)
	if err != nil || !found || template.PickpocketLootID != 500 {
		t.Fatalf("template=%#v found=%v err=%v", template, found, err)
	}
	target := &creatureState{GUID: 0xf130000000000001, Template: template, Health: 40, MaxHealth: 40}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	client, connection := net.Pipe()
	defer client.Close()
	defer connection.Close()
	server.attachPlayer(active.GUID, connection)
	cast := &spellCast{caster: active, targetCreature: target, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectPickpocket)}}}}
	done := make(chan struct{})
	go func() {
		server.applySpellEffects(cast)
		close(done)
	}()
	query, err := sockets.ReadPacket(client)
	if err != nil || (query.Opcode != packet.SMSGItemQuerySingleResponse && query.Opcode != packet.SMSGItemQueryMultipleResponse) {
		t.Fatalf("query=%#v err=%v", query, err)
	}
	loot, err := sockets.ReadPacket(client)
	if err != nil || loot.Opcode != packet.SMSGLootResponse || binary.LittleEndian.Uint32(loot.Data[8:12]) != lootTypePicklock || loot.Data[16] != 1 || binary.LittleEndian.Uint32(loot.Data[18:]) != 200 {
		t.Fatalf("loot=%#v err=%v", loot, err)
	}
	<-done
	server.lootMu.Lock()
	state := server.loots[lootKey{guid: target.GUID, sourceType: lootPickpocketSource}]
	if state == nil || !state.generated || len(state.items) != 1 || state.items[0].entry != 200 {
		server.lootMu.Unlock()
		t.Fatalf("pickpocket state=%#v", state)
	}
	server.lootMu.Unlock()
	done = make(chan struct{})
	go func() {
		server.applySpellEffects(cast)
		close(done)
	}()
	if _, err := sockets.ReadPacket(client); err != nil {
		t.Fatal(err)
	}
	if loot, err = sockets.ReadPacket(client); err != nil || loot.Opcode != packet.SMSGLootResponse || loot.Data[16] != 1 || binary.LittleEndian.Uint32(loot.Data[18:]) != 200 {
		t.Fatalf("repeat loot=%#v err=%v", loot, err)
	}
	<-done
	if responses, err := server.lootItem(active, []byte{0}); err != nil || len(responses) != 4 {
		t.Fatalf("loot item responses=%d err=%v", len(responses), err)
	}
	items, err := characters.WorldInventory(guid)
	if err != nil || len(items) != 1 || items[0].ItemTemplate != 200 {
		t.Fatalf("looted items=%#v err=%v", items, err)
	}
	if _, err := server.sendLoot(active, target.GUID, lootCreatureSource, target.Template.Entry, worlddb.GameObjectTemplate{}, 0); err != nil {
		t.Fatal(err)
	}
	server.lootMu.Lock()
	_, normalFound := server.loots[lootKey{guid: target.GUID, sourceType: lootCreatureSource}]
	_, pickpocketFound := server.loots[lootKey{guid: target.GUID, sourceType: lootPickpocketSource}]
	server.lootMu.Unlock()
	if !normalFound || !pickpocketFound {
		t.Fatalf("normal loot=%v pickpocket loot=%v", normalFound, pickpocketFound)
	}
}
