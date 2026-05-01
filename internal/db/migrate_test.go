package db

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func openMem(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open mem: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestMigrate_AppliesInitialMigration(t *testing.T) {
	db := openMem(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	versions, err := AppliedVersions(db)
	if err != nil {
		t.Fatalf("AppliedVersions: %v", err)
	}
	if len(versions) == 0 || versions[0] != "0001_init" {
		t.Errorf("expected first applied version 0001_init, got %v", versions)
	}

	for _, table := range []string{"schema_migrations", "accounts", "schedules", "automations"} {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Errorf("table %q missing: %v", table, err)
		}
	}
}

func TestMigrate_IsIdempotent(t *testing.T) {
	db := openMem(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
	versions, _ := AppliedVersions(db)
	count := 0
	for _, v := range versions {
		if v == "0001_init" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 0001_init applied once, got %d", count)
	}
}
