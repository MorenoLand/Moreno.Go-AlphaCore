package database

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
)

func TestImportIntoSQLiteSchema(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate importer test")
	}
	databases, err := OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	root := filepath.Join(filepath.Dir(file), "..", "etc", "databases")
	if err := Import(context.Background(), databases, root); err != nil {
		t.Fatal(err)
	}
	var spells int
	if err := databases.DB(DBC).QueryRow("SELECT count(*) FROM Spell").Scan(&spells); err != nil {
		t.Fatal(err)
	}
	if spells == 0 {
		t.Fatal("imported DBC contains no spells")
	}
}
