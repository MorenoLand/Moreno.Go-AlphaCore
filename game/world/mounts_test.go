package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
)

func TestSummonMountTogglesState(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO CreatureDisplayInfo (ID, ModelID, CreatureModelScale) VALUES (1000, 1, 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO creature_template (entry, display_id1, name, faction, unit_class) VALUES (123, 1000, 'Horse', 1, 1)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Rider", Health: 100})
	if err != nil {
		t.Fatal(err)
	}
	owner := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Rider", Health: 100}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	server.registerPlayer(owner)
	server.summonMount(owner, 123)
	if server.mountDisplayID(guid) != 1000 || server.unitFlags(guid)&0x3000 != 0x3000 {
		t.Fatalf("mounted display=%d flags=%x", server.mountDisplayID(guid), server.unitFlags(guid))
	}
	server.summonMount(owner, 123)
	if server.mountDisplayID(guid) != 0 || server.unitFlags(guid)&0x3000 != 0 {
		t.Fatalf("dismounted display=%d flags=%x", server.mountDisplayID(guid), server.unitFlags(guid))
	}
	aura := &auraState{effect: dbc.SpellEffect{Aura: int64(packet.AuraModMounted), MiscValue: 123}}
	server.auraEffectChange(owner, aura, false)
	if server.mountDisplayID(guid) != 1000 {
		t.Fatalf("aura mounted display=%d", server.mountDisplayID(guid))
	}
	server.auraEffectChange(owner, aura, true)
	if server.mountDisplayID(guid) != 0 {
		t.Fatalf("aura dismounted display=%d", server.mountDisplayID(guid))
	}
}
