package world

import (
	"context"
	"encoding/binary"
	"testing"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/database/realm"
)

func TestBugReportStoresCleanTicket(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	characters := realm.NewStore(databases)
	server := &WorldServer{Characters: characters}
	account := &auth.Account{ID: 4, Name: "PLAYER"}
	active := realm.Character{GUID: 1, AccountID: 4, RealmID: 1, Name: "Tester"}
	report, category := []byte("broken Username: hidden\x00"), []byte("UI\x00")
	data := make([]byte, 8+len(report)+4+len(category))
	binary.LittleEndian.PutUint32(data, 0)
	binary.LittleEndian.PutUint32(data[4:], uint32(len(report)))
	copy(data[8:], report)
	offset := 8 + len(report)
	binary.LittleEndian.PutUint32(data[offset:], uint32(len(category)))
	copy(data[offset+4:], category)
	if err := server.bugReport(account, active, data); err != nil {
		t.Fatal(err)
	}
	var isBug int
	var text string
	if err := databases.DB(database.Realm).QueryRow(`SELECT is_bug, text_body FROM tickets LIMIT 1`).Scan(&isBug, &text); err != nil {
		t.Fatal(err)
	}
	if isBug != 1 || text != "[UI] broken" {
		t.Fatalf("ticket=%d %q", isBug, text)
	}
}
