package database

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestImportSQLNormalizesMariaDBDump(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	data := `/*!40101 SET NAMES utf8mb4 */;
CREATE TABLE "things" (
  "id" int(11) unsigned NOT NULL AUTO_INCREMENT,
  "name" varchar(255) NOT NULL COMMENT 'thing',
  "value" int NOT NULL DEFAULT 0,
  PRIMARY KEY ("id"),
		KEY "idx_name" ("name")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
LOCK TABLES "things" WRITE;
INSERT INTO "things" VALUES (1,'A\'B',4);
	UNLOCK TABLES;`
	if err := importSQL(context.Background(), db, data, "fixture"); err != nil {
		t.Fatal(err)
	}
	var name string
	var value int
	if err := db.QueryRow("SELECT name, value FROM things WHERE id = 1").Scan(&name, &value); err != nil {
		t.Fatal(err)
	}
	if name != "A'B" || value != 4 {
		t.Fatalf("name=%q value=%d", name, value)
	}
}
