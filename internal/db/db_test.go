package db

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestOpen_AppliesPragmas(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	var fk int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Errorf("expected foreign_keys=1, got %d", fk)
	}

	var mode string
	if err := db.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("expected journal_mode=wal, got %q", mode)
	}
}

func TestWithTx_CommitsOnNilError(t *testing.T) {
	db := openMem(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	err := WithTx(db, func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO accounts(id) VALUES (?)`, "facebook")
		return err
	})
	if err != nil {
		t.Fatalf("WithTx: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM accounts WHERE id='facebook'`).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 account, got %d", count)
	}
}

func TestWithTx_RollsBackOnError(t *testing.T) {
	db := openMem(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	want := errors.New("boom")
	got := WithTx(db, func(tx *sql.Tx) error {
		_, _ = tx.Exec(`INSERT INTO accounts(id) VALUES (?)`, "instagram")
		return want
	})
	if !errors.Is(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM accounts WHERE id='instagram'`).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected rollback, found %d rows", count)
	}
}
