package world

import (
	"context"
	"net"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

func TestDynamicCreatureCheats(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, level_min, level_max, faction, scale, unit_class, unit_flags, npc_flags, type, beast_family, health_multiplier, mana_multiplier, armor_multiplier, damage_multiplier, damage_variance, base_attack_time, ranged_attack_time) VALUES (100, 123, 'Wolf', 2, 2, 14, 1, 1, 0, 0, 1, 1, 1, 1, 1, 1, 0.14, 2000, 2000)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_classlevelstats ("class", level, melee_damage, ranged_damage, attack_power, ranged_attack_power, health, base_health, mana, base_mana, strength, agility, stamina, intellect, spirit, armor) VALUES (1, 2, 2.3, 2.3, 4, 4, 55, 29, 0, 0, 20, 21, 23, 20, 20, 10)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_model_info (modelid, bounding_radius, combat_reach, gender) VALUES (123, 0.3, 1.5, 0)`); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: 1, Name: "GM", Health: 100, Map: 0, PositionX: 1, PositionY: 2, PositionZ: 3, Orientation: 4}
	server.registerPlayer(active)
	client, connection := net.Pipe()
	defer client.Close()
	server.attachPlayer(active.GUID, connection)
	done := make(chan error, 1)
	go func() {
		_, createErr := server.gmCheat(&active, packet.CMSGCreateMonster, cheatUint32(100), 1)
		done <- createErr
	}()
	created, err := sockets.ReadPacket(client)
	if err != nil {
		t.Fatal(err)
	}
	if created.Opcode != packet.SMSGUpdateObject && created.Opcode != packet.SMSGCompressedUpdateObject {
		t.Fatalf("create packet=%#v", created)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	server.creatures.mu.Lock()
	if len(server.creatures.active) != 1 {
		server.creatures.mu.Unlock()
		t.Fatalf("active creatures=%d", len(server.creatures.active))
	}
	var guid uint64
	for value := range server.creatures.active {
		guid = value
	}
	server.creatures.mu.Unlock()
	if guid&0xffff000000000000 != 0xf130000000000000 {
		t.Fatalf("dynamic guid=%x", guid)
	}
	if _, err := server.gmCheat(&active, packet.CMSGDestroyMonster, encodeUint64(guid), 1); err != nil {
		t.Fatal(err)
	}
	server.creatures.mu.Lock()
	_, found := server.creatures.active[guid]
	server.creatures.mu.Unlock()
	if !found {
		t.Fatal("non-dev destroyed dynamic creature")
	}
	done = make(chan error, 1)
	go func() {
		_, destroyErr := server.gmCheat(&active, packet.CMSGDestroyMonster, encodeUint64(guid), 2)
		done <- destroyErr
	}()
	destroyed, err := sockets.ReadPacket(client)
	if err != nil || destroyed.Opcode != packet.SMSGDestroyObject {
		t.Fatalf("destroy packet=%#v err=%v", destroyed, err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	server.creatures.mu.Lock()
	_, found = server.creatures.active[guid]
	server.creatures.mu.Unlock()
	if found {
		t.Fatal("dynamic creature remains")
	}
}
