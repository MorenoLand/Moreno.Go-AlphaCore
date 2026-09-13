package database

import (
	"context"
	"database/sql"
	"fmt"

	"Moreno.AlphaCore/utils"
	_ "modernc.org/sqlite"
)

type Name string

const (
	Auth  Name = "auth"
	Realm Name = "realm"
	World Name = "world"
	DBC   Name = "dbc"
)

var names = []Name{Auth, Realm, World, DBC}

type Databases struct {
	work  utils.Workspace
	items map[Name]*sql.DB
}

func Open(ctx context.Context, work utils.Workspace) (*Databases, error) {
	databases := &Databases{work: work, items: make(map[Name]*sql.DB, len(names))}
	for _, name := range names {
		db, err := sql.Open("sqlite", work.Database(string(name)))
		if err != nil {
			databases.Close()
			return nil, fmt.Errorf("open %s database: %w", name, err)
		}
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		databases.items[name] = db
		if err := configure(ctx, db); err != nil {
			databases.Close()
			return nil, fmt.Errorf("configure %s database: %w", name, err)
		}
		if err := initialize(ctx, db, name); err != nil {
			databases.Close()
			return nil, fmt.Errorf("initialize %s database: %w", name, err)
		}
	}
	return databases, nil
}

func configure(ctx context.Context, db *sql.DB) error {
	for _, statement := range []string{"PRAGMA busy_timeout = 5000", "PRAGMA foreign_keys = ON", "PRAGMA journal_mode = WAL"} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func initialize(ctx context.Context, db *sql.DB, name Name) error {
	schema, ok := map[Name]string{Auth: authSchema, Realm: realmSchema, World: worldSchema, DBC: dbcSchema}[name]
	if !ok {
		return fmt.Errorf("unknown database %q", name)
	}
	_, err := db.ExecContext(ctx, schema)
	return err
}

func (d *Databases) DB(name Name) *sql.DB { return d.items[name] }

func (d *Databases) Paths() []string {
	paths := make([]string, 0, len(names))
	for _, name := range names {
		paths = append(paths, d.work.Database(string(name)))
	}
	return paths
}

func (d *Databases) Close() error {
	var firstErr error
	for _, db := range d.items {
		if err := db.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
