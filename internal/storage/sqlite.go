// Package storage owns Mica's local SQLite bootstrap and migrations.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const schemaVersion = 1

// Open initializes a local SQLite database. Mica keeps a single aggregate
// payload during the early MVP while migrations establish a safe path toward
// normalized incident, evidence, audit, and delivery tables.
func Open(path string) (*sql.DB, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(1)
	if err := database.Ping(); err != nil {
		database.Close()
		return nil, err
	}
	if _, err := database.Exec(`PRAGMA journal_mode = WAL; PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000;`); err != nil {
		database.Close()
		return nil, err
	}
	if err := migrate(database); err != nil {
		database.Close()
		return nil, err
	}
	return database, nil
}

func migrate(database *sql.DB) error {
	var version int
	if err := database.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return err
	}
	if version > schemaVersion {
		return fmt.Errorf("database schema %d is newer than supported schema %d", version, schemaVersion)
	}
	if version < 1 {
		transaction, err := database.Begin()
		if err != nil {
			return err
		}
		defer transaction.Rollback()
		if _, err := transaction.Exec(`CREATE TABLE IF NOT EXISTS mica_state (id INTEGER PRIMARY KEY CHECK (id = 1), payload JSON NOT NULL, updated_at TEXT NOT NULL)`); err != nil {
			return err
		}
		if _, err := transaction.Exec(`PRAGMA user_version = 1`); err != nil {
			return err
		}
		return transaction.Commit()
	}
	return nil
}

func Load(database *sql.DB) ([]byte, bool, error) {
	var payload []byte
	err := database.QueryRow(`SELECT payload FROM mica_state WHERE id = 1`).Scan(&payload)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	return payload, err == nil, err
}

func Save(database *sql.DB, payload []byte) error {
	_, err := database.ExecContext(context.Background(), `INSERT INTO mica_state (id, payload, updated_at) VALUES (1, ?, ?) ON CONFLICT(id) DO UPDATE SET payload = excluded.payload, updated_at = excluded.updated_at`, payload, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}
