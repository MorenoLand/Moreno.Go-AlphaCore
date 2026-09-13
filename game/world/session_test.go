package world

import (
	"context"
	"encoding/binary"
	"math"
	"net"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

func TestWorldSessionLifecycle(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	accounts := auth.NewStore(databases)
	if err := accounts.CreateAccount("PLAYER", "PASSWORD", "", 0); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO ChrRaces (ID, FactionID, MaleDisplayId, FemaleDisplayId, CreatureType) VALUES (1, 1, 49, 50, 7)`); err != nil {
		t.Fatal(err)
	}
	if _, err := databases.DB(database.DBC).Exec(`INSERT INTO AreaTable (ID, AreaNumber, ContinentID, ParentAreaNum) VALUES (12, 393216, 0, 0)`); err != nil {
		t.Fatal(err)
	}
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Testone", Race: 1, Class: 1, Level: 1, Map: 0, Zone: 12, PositionX: 1, PositionY: 2, PositionZ: 3, Orientation: 4, Health: 20})
	if err != nil {
		t.Fatal(err)
	}
	if err := characters.AddSpell(guid, 42); err != nil {
		t.Fatal(err)
	}
	if err := characters.SetSpellButton(guid, 42, -1); err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Accounts: accounts, Characters: characters, DBC: dbc.NewStore(databases), WorldData: worlddb.NewStore(databases), SupportedClient: 3368, ServerSeed: []byte{1, 2, 3, 4}}
	client, connection := net.Pipe()
	done := make(chan struct{})
	go func() {
		server.handle(connection)
		close(done)
	}()
	defer client.Close()
	stream := client
	challenge, err := sockets.ReadPacket(stream)
	if err != nil || challenge.Opcode != packet.SMSGAuthChallenge {
		t.Fatalf("challenge=%#v err=%v", challenge, err)
	}
	authData := make([]byte, 8)
	binary.LittleEndian.PutUint32(authData, 3368)
	authData = append(authData, []byte("PLAYER\x00PASSWORD\x00")...)
	message, _ := packet.Encode(packet.CMSGAuthSession, authData)
	if _, err := client.Write(message); err != nil {
		t.Fatal(err)
	}
	response, err := sockets.ReadPacket(stream)
	if err != nil || response.Data[0] != byte(packet.AuthOK) {
		t.Fatalf("auth=%#v err=%v", response, err)
	}
	loginData := make([]byte, 8)
	binary.LittleEndian.PutUint64(loginData, uint64(guid))
	message, _ = packet.Encode(packet.CMSGPlayerLogin, loginData)
	client.Write(message)
	for index, expected := range []packet.Opcode{packet.SMSGLoginSetTimeSpeed, packet.SMSGNewWorld} {
		response, err = sockets.ReadPacket(stream)
		if err != nil || response.Opcode != expected {
			t.Fatalf("login[%d]=%#v err=%v", index, response, err)
		}
	}
	message, _ = packet.Encode(packet.MSGMoveWorldportAck, nil)
	client.Write(message)
	for index, expected := range []packet.Opcode{packet.SMSGInitializeFactions, packet.SMSGInitialSpells, packet.SMSGActionButtons, packet.SMSGCompressedUpdateObject} {
		response, err = sockets.ReadPacket(stream)
		if err != nil || response.Opcode != expected {
			t.Fatalf("initial[%d]=%#v err=%v", index, response, err)
		}
		if expected == packet.SMSGInitialSpells && (len(response.Data) != 9 || binary.LittleEndian.Uint16(response.Data[1:]) != 1 || binary.LittleEndian.Uint16(response.Data[3:]) != 42 || int16(binary.LittleEndian.Uint16(response.Data[5:])) != -1) {
			t.Fatalf("initial spells=%#v", response.Data)
		}
	}
	message, _ = packet.Encode(packet.Opcode(0x7fff), nil)
	client.Write(message)
	pingData := []byte{1, 2, 3, 4}
	message, _ = packet.Encode(packet.CMSGPing, pingData)
	client.Write(message)
	response, err = sockets.ReadPacket(stream)
	if err != nil || response.Opcode != packet.SMSGPong {
		t.Fatalf("pong=%#v err=%v", response, err)
	}
	message, _ = packet.Encode(packet.CMSGQueryTime, nil)
	client.Write(message)
	response, err = sockets.ReadPacket(stream)
	if err != nil || response.Opcode != packet.SMSGQueryTimeResponse || len(response.Data) != 4 {
		t.Fatalf("time=%#v err=%v", response, err)
	}
	queryData := make([]byte, 8)
	binary.LittleEndian.PutUint64(queryData, uint64(guid))
	message, _ = packet.Encode(packet.CMSGNameQuery, queryData)
	client.Write(message)
	response, err = sockets.ReadPacket(stream)
	if err != nil || response.Opcode != packet.SMSGNameQueryResponse {
		t.Fatalf("name=%#v err=%v", response, err)
	}
	whoData := make([]byte, 26)
	binary.LittleEndian.PutUint32(whoData, 1)
	binary.LittleEndian.PutUint32(whoData[4:], 60)
	whoData[8] = 0
	whoData[9] = 0
	binary.LittleEndian.PutUint32(whoData[10:], 0xffffffff)
	binary.LittleEndian.PutUint32(whoData[14:], 0xffffffff)
	binary.LittleEndian.PutUint32(whoData[18:], 0)
	binary.LittleEndian.PutUint32(whoData[22:], 0)
	message, _ = packet.Encode(packet.CMSGWho, whoData)
	client.Write(message)
	response, err = sockets.ReadPacket(stream)
	if err != nil || response.Opcode != packet.SMSGWho || len(response.Data) < 8 || binary.LittleEndian.Uint32(response.Data) != 1 || binary.LittleEndian.Uint32(response.Data[4:]) != 1 {
		t.Fatalf("who=%#v err=%v", response, err)
	}
	name, err := packet.ReadString(response.Data, 8, 0)
	if err != nil || name != "Testone" {
		t.Fatalf("who name=%q err=%v", name, err)
	}
	chatData := []byte{0, 0, 0, 0, 7, 0, 0, 0}
	chatData = append(chatData, []byte("Hello\x00")...)
	message, _ = packet.Encode(packet.CMSGMessageChat, chatData)
	client.Write(message)
	response, err = sockets.ReadPacket(stream)
	if err != nil || response.Opcode != packet.SMSGMessageChat {
		t.Fatalf("chat=%#v err=%v", response, err)
	}
	zoneData := make([]byte, 4)
	binary.LittleEndian.PutUint32(zoneData, 42)
	message, _ = packet.Encode(packet.CMSGZoneUpdate, zoneData)
	client.Write(message)
	moveData := make([]byte, 48)
	binary.LittleEndian.PutUint32(moveData[24:], math.Float32bits(10))
	binary.LittleEndian.PutUint32(moveData[28:], math.Float32bits(20))
	binary.LittleEndian.PutUint32(moveData[32:], math.Float32bits(30))
	binary.LittleEndian.PutUint32(moveData[36:], math.Float32bits(40))
	message, _ = packet.Encode(packet.Opcode(0x00b5), moveData)
	client.Write(message)
	message, _ = packet.Encode(packet.CMSGLogoutRequest, nil)
	client.Write(message)
	response, err = sockets.ReadPacket(stream)
	if err != nil || response.Opcode != packet.SMSGLogoutResponse || response.Data[0] != 1 {
		t.Fatalf("logout request=%#v err=%v", response, err)
	}
	server.players.mu.RLock()
	zone := server.players.players[guid].Zone
	server.players.mu.RUnlock()
	if zone != 42 {
		t.Fatalf("registry zone=%d", zone)
	}
	message, _ = packet.Encode(packet.CMSGLogoutCancel, nil)
	client.Write(message)
	response, err = sockets.ReadPacket(stream)
	if err != nil || response.Opcode != packet.SMSGLogoutCancelAck {
		t.Fatalf("logout cancel=%#v err=%v", response, err)
	}
	message, _ = packet.Encode(packet.CMSGPlayerLogout, nil)
	client.Write(message)
	response, err = sockets.ReadPacket(stream)
	if err != nil || response.Opcode != packet.SMSGLogoutComplete {
		t.Fatalf("logout=%#v err=%v", response, err)
	}
	client.Close()
	<-done
	stored, err := server.Characters.Characters(1, 1)
	if err != nil || len(stored) != 1 || stored[0].Online != 0 || stored[0].PositionX != 10 || stored[0].PositionY != 20 || stored[0].PositionZ != 30 || stored[0].Orientation != 40 || stored[0].Zone != 42 {
		t.Fatalf("stored character=%#v err=%v", stored, err)
	}
}
