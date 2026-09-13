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

func TestTextEmotePackets(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO EmotesText (ID, EmoteID) VALUES (1, 3), (2, 13)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Emoter", Race: 1, Class: 1})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{DBC: dbc.NewStore(databases), Characters: characters, WorldData: worlddb.NewStore(databases)}
	active := realm.Character{GUID: guid, Name: "Emoter", Race: 1, Class: 1}
	server.registerPlayer(active)
	data := make([]byte, 12)
	data[0] = 1
	responses, err := server.textEmote(active, data)
	if err != nil || len(responses) != 2 {
		t.Fatalf("wave responses=%d err=%v", len(responses), err)
	}
	for index, expected := range []packet.Opcode{packet.SMSGTextEmote, packet.SMSGEmote} {
		parsed, parseErr := packet.Parse(responses[index])
		if parseErr != nil || parsed.Opcode != expected {
			t.Fatalf("wave[%d]=%#v err=%v", index, parsed, parseErr)
		}
	}
	data[0] = 2
	responses, err = server.textEmote(active, data)
	if err != nil || len(responses) != 1 {
		t.Fatalf("sit responses=%d err=%v", len(responses), err)
	}
	server.players.mu.RLock()
	state := server.players.standState[guid]
	server.players.mu.RUnlock()
	if state != 1 {
		t.Fatalf("stand state=%d", state)
	}
}
