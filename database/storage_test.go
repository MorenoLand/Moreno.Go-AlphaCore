package database

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSQLiteSchemas(t *testing.T) {
	for _, name := range names {
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		if err := configure(context.Background(), db); err != nil {
			db.Close()
			t.Fatal(err)
		}
		if err := initialize(context.Background(), db, name); err != nil {
			db.Close()
			t.Fatalf("%s: %v", name, err)
		}
		var count int
		if err := db.QueryRow("SELECT count(*) FROM schema_meta").Scan(&count); err != nil || count != 2 {
			db.Close()
			t.Fatalf("%s metadata count=%d err=%v", name, count, err)
		}
		if name == Auth {
			var realm string
			if err := db.QueryRow("SELECT realm_name FROM realmlist WHERE realm_id = 1").Scan(&realm); err != nil || realm != "Moreno.AlphaCore" {
				db.Close()
				t.Fatalf("realm=%q err=%v", realm, err)
			}
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
