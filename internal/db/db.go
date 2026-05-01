// Package db owns the SQLite handle for Stargrazer: opening, configuration,
// migrations, and a small WithTx helper. Domain repositories live in their
// own packages and consume the Executor interface so they work against
// either a *sql.DB (production) or a *sql.Tx (migrator).
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Executor is the subset of database/sql that SQLite repositories need.
// Both *sql.DB and *sql.Tx satisfy it implicitly, so repos accept either.
// Used to let the backfill orchestrator drive cross-domain transactions
// without changing the repo call sites.
type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Open opens (or creates) a SQLite database at path and applies pragmas.
// The returned *sql.DB is safe for concurrent use.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite at %s: %w", path, err)
	}
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA synchronous = NORMAL",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("apply %q: %w", p, err)
		}
	}
	return db, nil
}

// WithTx runs fn inside an IMMEDIATE transaction. Commits on nil error,
// rolls back otherwise. SQLITE_BUSY is retried once with a 50ms backoff.
func WithTx(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		if isBusy(err) {
			time.Sleep(50 * time.Millisecond)
			tx, err = db.Begin()
		}
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func isBusy(err error) bool {
	return err != nil && strings.Contains(err.Error(), "database is locked")
}

// Sentinel error for the migrator (unused by Open / WithTx but exported here
// alongside other db-package errors so callers have one import).
var ErrAlreadyApplied = errors.New("migration already applied")
