package main

import (
	"database/sql"
	"embed"
	"strings"
)

//go:embed schema.sql seed.sql
var migrations embed.FS

func initSchema(db *sql.DB) error {
	schemaSQL, err := migrations.ReadFile("schema.sql")
	if err != nil {
		return err
	}

	seedSQL, err := migrations.ReadFile("seed.sql")
	if err != nil {
		return err
	}

	if _, err := db.Exec(string(schemaSQL)); err != nil {
		return err
	}

	if _, err := db.Exec(`ALTER TABLE tasks ADD COLUMN checklist_state TEXT NOT NULL DEFAULT '[]';`); err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tasks;`).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	_, err = db.Exec(string(seedSQL))
	return err
}

func runSeedSQL(execable interface {
	Exec(string, ...any) (sql.Result, error)
}) error {
	seedSQL, err := migrations.ReadFile("seed.sql")
	if err != nil {
		return err
	}

	_, err = execable.Exec(string(seedSQL))
	return err
}
