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

func TestSpellEffectQuestComplete(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Player", Level: 1, Health: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.World).Exec(`INSERT INTO quest_template (entry, QuestLevel, Title) VALUES (500, 1, 'Quest')`); err != nil {
		t.Fatal(err)
	}
	if err := characters.SaveQuestState(realm.QuestState{GUID: guid, Quest: 500, State: questAccepted}); err != nil {
		t.Fatal(err)
	}
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Player", Level: 1, Health: 1}
	server := &WorldServer{Characters: characters, WorldData: worlddb.NewStore(databases)}
	server.registerPlayer(active)
	client, connection := net.Pipe()
	defer client.Close()
	defer connection.Close()
	server.attachPlayer(guid, connection)
	done := make(chan struct{})
	go func() {
		server.applySpellEffects(&spellCast{caster: active, spell: dbc.Spell{Effects: [3]dbc.SpellEffect{{Type: int64(packet.SpellEffectQuestComplete), MiscValue: 500}}}, effectTargets: map[int][]realm.Character{0: {active}}})
		close(done)
	}()
	response, err := sockets.ReadPacket(client)
	if err != nil || response.Opcode != packet.SMSGQuestUpdateComplete || len(response.Data) != 4 || binary.LittleEndian.Uint32(response.Data) != 500 {
		t.Fatalf("response=%#v err=%v", response, err)
	}
	<-done
	state, found, err := characters.QuestState(guid, 500)
	if err != nil || !found || state.State != questReward || !state.Explored || state.Rewarded {
		t.Fatalf("state=%#v found=%v err=%v", state, found, err)
	}
}
