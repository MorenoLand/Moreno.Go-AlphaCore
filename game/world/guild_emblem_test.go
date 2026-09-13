package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestGuildSaveEmblem(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Founder", Money: 200000})
	if err != nil {
		t.Fatal(err)
	}
	guild, err := characters.CreateGuild("Emblem", "", guid)
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddGuildMember(guild.ID, guid, 0); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters}
	state := server.guilds.loaded(guild, []realm.GuildMember{{GuildID: guild.ID, GUID: guid, Rank: 0}})
	active := realm.Character{GUID: guid, AccountID: 1, RealmID: 1, Name: "Founder", Money: 200000}
	data := make([]byte, 20)
	for index, value := range []uint32{1, 2, 3, 4, 5} {
		binary.LittleEndian.PutUint32(data[index*4:], value)
	}
	response, err := server.guildSaveEmblem(&active, data)
	if err != nil {
		t.Fatal(err)
	}
	message, err := packet.Parse(response)
	if err != nil || message.Opcode != packet.MSGSaveGuildEmblem || len(message.Data) != 4 || binary.LittleEndian.Uint32(message.Data) != 0 {
		t.Fatalf("response=%#v err=%v", message, err)
	}
	if active.Money != 100000 || state.Guild.EmblemStyle != 1 || state.Guild.BackgroundColor != 5 {
		t.Fatalf("active=%#v guild=%#v", active, state.Guild)
	}
}
