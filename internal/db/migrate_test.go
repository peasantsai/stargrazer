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

func TestMigrate_AppliesP3Schema(t *testing.T) {
	conn := openMem(t)
	if err := Migrate(conn); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	versions, err := AppliedVersions(conn)
	if err != nil {
		t.Fatalf("AppliedVersions: %v", err)
	}
	want := "0002_recording_library"
	found := false
	for _, v := range versions {
		if v == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("missing applied version %q in %v", want, versions)
	}
	row := conn.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN ('recordings','step_templates','variable_profiles')`)
	var n int
	if err := row.Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("want 3 P3 tables, got %d", n)
	}
	cols := map[string]bool{}
	rows, err := conn.Query(`PRAGMA table_info(automations)`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		cols[name] = true
	}
	rows.Close()
	if !cols["default_profile_id"] {
		t.Errorf("automations.default_profile_id column missing; got %v", cols)
	}
	cols = map[string]bool{}
	rows, err = conn.Query(`PRAGMA table_info(schedules)`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		cols[name] = true
	}
	rows.Close()
	if !cols["profile_id"] {
		t.Errorf("schedules.profile_id column missing; got %v", cols)
	}
}
