// Package dbtest exposes a NewMemDB helper for repository tests so each test
// gets a fresh in-memory SQLite with all migrations applied.
package dbtest

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"stargrazer/internal/db"
)

// NewMemDB returns a fresh in-memory SQLite database with every embedded
// migration applied. The DB is closed automatically via t.Cleanup.
func NewMemDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open mem: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := db.Migrate(conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return conn
}
