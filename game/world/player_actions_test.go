package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

func TestPlayerActionState(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	guid, err := characters.Create(realm.Character{AccountID: 1, RealmID: 1, Name: "Actioner", Race: 1, Class: 1, Level: 4, Totaltime: 20, Leveltime: 5})
	if err != nil {
		t.Fatal(err)
	}
	server := &WorldServer{Characters: characters}
	active := realm.Character{GUID: guid, Totaltime: 20, Leveltime: 5}
	actionData := []byte{3, 0xfe, 0xff, 0xff, 0xff}
	if err := server.setActionButton(active, actionData); err != nil {
		t.Fatal(err)
	}
	buttons, err := characters.Buttons(guid)
	if err != nil || buttons[3] != -2 {
		t.Fatalf("buttons=%v err=%v", buttons, err)
	}
	if err := server.setActionButton(active, []byte{3, 0, 0, 0, 0}); err != nil {
		t.Fatal(err)
	}
	buttons, err = characters.Buttons(guid)
	if err != nil || len(buttons) != 0 {
		t.Fatalf("deleted buttons=%v err=%v", buttons, err)
	}
	spellData := make([]byte, 8)
	binary.LittleEndian.PutUint32(spellData, 42)
	binary.LittleEndian.PutUint32(spellData[4:], ^uint32(0))
	if err := server.newSpellSlot(active, spellData); err != nil {
		t.Fatal(err)
	}
	var index int64
	if err := databases.DB(database.Realm).QueryRow(`SELECT "index" FROM character_spell_book WHERE owner = ? AND spell = ?`, guid, 42).Scan(&index); err != nil || index != -1 {
		t.Fatalf("spell index=%d err=%v", index, err)
	}
	response, err := server.playedTime(active)
	parsed, parseErr := packet.Parse(response)
	if err != nil || parseErr != nil || parsed.Opcode != packet.SMSGPlayedTime || binary.LittleEndian.Uint32(parsed.Data) != 20 || binary.LittleEndian.Uint32(parsed.Data[4:]) != 5 {
		t.Fatalf("played=%#v err=%v parse=%v", parsed, err, parseErr)
	}
	response, err = server.lookingForGroup(active)
	parsed, parseErr = packet.Parse(response)
	if err != nil || parseErr != nil || binary.LittleEndian.Uint32(parsed.Data) != 0 {
		t.Fatalf("lfg initial=%#v err=%v parse=%v", parsed, err, parseErr)
	}
	server.setLookingForGroup(active, []byte{1, 0, 0, 0})
	if server.getGroupStatus(guid) != 2 {
		t.Fatalf("lfg status=%d", server.getGroupStatus(guid))
	}
	server.setGroupStatus(guid, 1)
	server.setLookingForGroup(active, []byte{0, 0, 0, 0})
	if server.getGroupStatus(guid) != 1 {
		t.Fatalf("party status=%d", server.getGroupStatus(guid))
	}
	rollData := make([]byte, 8)
	binary.LittleEndian.PutUint32(rollData, 4)
	binary.LittleEndian.PutUint32(rollData[4:], 4)
	response, err = server.randomRoll(active, rollData)
	parsed, parseErr = packet.Parse(response)
	if err != nil || parseErr != nil || parsed.Opcode != packet.MSGRandomRoll || binary.LittleEndian.Uint32(parsed.Data) != 4 || binary.LittleEndian.Uint32(parsed.Data[4:]) != 4 || binary.LittleEndian.Uint32(parsed.Data[8:]) != 4 || int64(binary.LittleEndian.Uint64(parsed.Data[12:])) != guid {
		t.Fatalf("roll=%#v err=%v parse=%v", parsed, err, parseErr)
	}
	server.registerPlayer(active)
	selection := make([]byte, 8)
	binary.LittleEndian.PutUint64(selection, 7)
	server.setSelection(active, selection)
	server.setTarget(active, selection)
	server.players.mu.RLock()
	if server.players.selection[guid] != 7 || server.players.target[guid] != 7 {
		t.Fatalf("selection=%v target=%v", server.players.selection, server.players.target)
	}
	server.players.mu.RUnlock()
}
