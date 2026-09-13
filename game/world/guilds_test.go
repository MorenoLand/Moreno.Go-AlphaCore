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

func TestGuildLifecycle(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO ChrRaces (ID, FactionID, MaleDisplayId, FemaleDisplayId, BaseLanguage, CreatureType) VALUES (1, 1, 49, 50, 1, 7)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	ownerGUID, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Owner", Race: 1, Class: 1})
	if err != nil {
		t.Fatal(err)
	}
	targetGUID, err := characters.Create(realm.Character{AccountID: 2, RealmID: 1, Name: "Target", Race: 1, Class: 1})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases)}
	owner := realm.Character{GUID: ownerGUID, Name: "Owner", Race: 1, Class: 1}
	target := realm.Character{GUID: targetGUID, Name: "Target", Race: 1, Class: 1}
	server.registerPlayer(owner)
	server.registerPlayer(target)
	created, err := server.guildCreate(owner, []byte("Test Guild\x00"), 1)
	if err != nil || len(created) != 1 {
		t.Fatalf("guild create=%d err=%v", len(created), err)
	}
	createdPacket, err := packet.Parse(created[0])
	if err != nil || createdPacket.Opcode != packet.SMSGGuildEvent {
		t.Fatalf("guild create packet=%#v err=%v", createdPacket, err)
	}
	client, connection := net.Pipe()
	defer client.Close()
	defer connection.Close()
	server.attachPlayer(targetGUID, connection)
	packets := make(chan packet.Packet, 8)
	go func() {
		for {
			value, err := sockets.ReadPacket(client)
			if err != nil {
				return
			}
			packets <- value
		}
	}()
	responses, err := server.guildInvite(owner, []byte("Target\x00"))
	if err != nil || len(responses) != 1 {
		t.Fatalf("guild invite=%d err=%v", len(responses), err)
	}
	command, err := packet.Parse(responses[0])
	if err != nil || command.Opcode != packet.SMSGGuildCommandResult || binary.LittleEndian.Uint32(command.Data) != guildInviteCommand || binary.LittleEndian.Uint32(command.Data[len(command.Data)-4:]) != guildInvited {
		t.Fatalf("guild command=%#v err=%v", command, err)
	}
	invite := <-packets
	if invite.Opcode != packet.SMSGGuildInvite {
		t.Fatalf("guild invite packet=%#v", invite)
	}
	if err := server.guildAccept(target); err != nil {
		t.Fatal(err)
	}
	if joined := <-packets; joined.Opcode != packet.SMSGGuildEvent {
		t.Fatalf("guild joined=%#v", joined)
	}
	state := server.guilds.forPlayer(ownerGUID)
	if state == nil || len(state.Members) != 2 || state.Members[ownerGUID] != 0 || state.Members[targetGUID] != 4 {
		t.Fatalf("guild state=%#v", state)
	}
	if _, err := server.guildPromote(owner, []byte("Target\x00"), false); err != nil {
		t.Fatal(err)
	}
	if promoted := <-packets; promoted.Opcode != packet.SMSGGuildEvent {
		t.Fatalf("promoted=%#v", promoted)
	}
	if _, err := server.guildPromote(owner, []byte("Target\x00"), true); err != nil {
		t.Fatal(err)
	}
	if demoted := <-packets; demoted.Opcode != packet.SMSGGuildEvent {
		t.Fatalf("demoted=%#v", demoted)
	}
	queryData := make([]byte, 4)
	binary.LittleEndian.PutUint32(queryData, uint32(state.Guild.ID))
	query, err := server.guildQuery(queryData)
	if err != nil {
		t.Fatal(err)
	}
	if parsed, err := packet.Parse(query); err != nil || parsed.Opcode != packet.SMSGGuildQueryResponse {
		t.Fatalf("guild query=%#v err=%v", parsed, err)
	}
	if roster, err := server.guildRoster(owner); err != nil {
		t.Fatal(err)
	} else if parsed, err := packet.Parse(roster); err != nil || parsed.Opcode != packet.SMSGGuildRoster {
		t.Fatalf("guild roster=%#v err=%v", parsed, err)
	}
	if info, err := server.guildInfo(owner); err != nil {
		t.Fatal(err)
	} else if parsed, err := packet.Parse(info); err != nil || parsed.Opcode != packet.SMSGGuildInfo {
		t.Fatalf("guild info=%#v err=%v", parsed, err)
	}
	if motd, err := server.guildMOTD(owner, []byte("Hello\x00")); err != nil {
		t.Fatal(err)
	} else if parsed, err := packet.Parse(motd); err != nil || parsed.Opcode != packet.SMSGGuildEvent {
		t.Fatalf("guild motd=%#v err=%v", parsed, err)
	}
	if _, err := server.guildLeave(target); err != nil {
		t.Fatal(err)
	}
	if _, found, err := characters.GuildByPlayer(targetGUID); err != nil || found {
		t.Fatalf("target guild found=%v err=%v", found, err)
	}
	if err := server.guildDisband(state); err != nil {
		t.Fatal(err)
	}
	if _, found, err := characters.GuildByPlayer(ownerGUID); err != nil || found {
		t.Fatalf("owner guild found=%v err=%v", found, err)
	}
}
