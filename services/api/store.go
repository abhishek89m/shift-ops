package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type store struct {
	db *sql.DB
}

func newStore(dbPath string) (*store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := initSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &store{db: db}, nil
}

func (s *store) close() {
	_ = s.db.Close()
}
