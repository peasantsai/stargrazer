package backfill

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"stargrazer/internal/automation"
	"stargrazer/internal/db/dbtest"
)

func TestRunIfNeeded_NoFiles_NoOp(t *testing.T) {
	db := dbtest.NewMemDB(t)
	if err := RunIfNeeded(db, t.TempDir()); err != nil {
		t.Fatalf("RunIfNeeded: %v", err)
	}

	// Sentinel must still be written so subsequent runs short-circuit.
	if !sentinelPresent(t, db) {
		t.Error("sentinel not written on empty backfill")
	}
}

func TestRunIfNeeded_BackfillsAutomations(t *testing.T) {
	dir := t.TempDir()
	autoDir := filepath.Join(dir, "automations")
	if err := os.MkdirAll(autoDir, 0o700); err != nil {
		t.Fatal(err)
	}
	cfgs := []automation.Config{{
		ID:         "a1",
		PlatformID: "facebook",
		Name:       "Sample",
		Steps:      []automation.Step{{Action: automation.ActionNavigate, Target: "https://www.facebook.com"}},
	}}
	body, _ := json.MarshalIndent(cfgs, "", "  ")
	if err := os.WriteFile(filepath.Join(autoDir, "facebook.json"), body, 0o600); err != nil {
		t.Fatal(err)
	}

	db := dbtest.NewMemDB(t)
	if err := RunIfNeeded(db, dir); err != nil {
		t.Fatalf("RunIfNeeded: %v", err)
	}

	repo := automation.NewSQLiteRepo(db)
	got, _ := repo.List("facebook")
	if len(got) != 1 || got[0].Name != "Sample" {
		t.Errorf("automation not migrated: %+v", got)
	}

	// Source archived.
	if _, err := os.Stat(filepath.Join(autoDir, "facebook.json")); !os.IsNotExist(err) {
		t.Error("expected facebook.json renamed away")
	}
	if _, err := os.Stat(filepath.Join(autoDir, "facebook.json.preP2.bak")); err != nil {
		t.Errorf("expected backup file present: %v", err)
	}
}

func TestRunIfNeeded_IsIdempotent(t *testing.T) {
	dir := t.TempDir()
	db := dbtest.NewMemDB(t)
	if err := RunIfNeeded(db, dir); err != nil {
		t.Fatalf("first RunIfNeeded: %v", err)
	}
	if err := RunIfNeeded(db, dir); err != nil {
		t.Fatalf("second RunIfNeeded: %v", err)
	}
}

func sentinelPresent(t *testing.T, db *sql.DB) bool {
	t.Helper()
	var v string
	err := db.QueryRow(`SELECT version FROM schema_migrations WHERE version='backfill_p2'`).Scan(&v)
	return err == nil && v == "backfill_p2"
}
