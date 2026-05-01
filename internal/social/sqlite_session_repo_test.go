package social

import (
	"testing"
	"time"

	"stargrazer/internal/db/dbtest"
)

func TestSQLiteSessionRepo_GetAll_DefaultsForUnknownPlatforms(t *testing.T) {
	repo := NewSQLiteSessionRepo(dbtest.NewMemDB(t))
	got := repo.GetAll()
	if len(got) != len(AllPlatforms()) {
		t.Fatalf("expected %d statuses, got %d", len(AllPlatforms()), len(got))
	}
	for _, s := range got {
		if s.LoggedIn {
			t.Errorf("unexpected LoggedIn=true for %s", s.PlatformID)
		}
	}
}

func TestSQLiteSessionRepo_SetLoggedInPersistsAndReads(t *testing.T) {
	repo := NewSQLiteSessionRepo(dbtest.NewMemDB(t))
	repo.SetLoggedIn(Facebook, "tester@example.com")

	got := repo.Get(Facebook)
	if !got.LoggedIn {
		t.Error("expected LoggedIn=true")
	}
	if got.Username != "tester@example.com" {
		t.Errorf("Username: got %q", got.Username)
	}
	if got.LastLogin.IsZero() {
		t.Error("LastLogin not set")
	}
}

func TestSQLiteSessionRepo_SetLoggedOutClearsUsername(t *testing.T) {
	repo := NewSQLiteSessionRepo(dbtest.NewMemDB(t))
	repo.SetLoggedIn(Instagram, "u")
	repo.SetLoggedOut(Instagram)

	got := repo.Get(Instagram)
	if got.LoggedIn {
		t.Error("expected LoggedIn=false")
	}
	if got.Username != "" {
		t.Errorf("Username should be cleared, got %q", got.Username)
	}
	if !got.LastLogin.IsZero() {
		t.Errorf("LastLogin should be cleared on logout, got %v", got.LastLogin)
	}
}

func TestSQLiteSessionRepo_UpdateCheckTimeOnlyTouchesLastCheck(t *testing.T) {
	repo := NewSQLiteSessionRepo(dbtest.NewMemDB(t))
	repo.SetLoggedIn(X, "u")
	first := repo.Get(X)

	time.Sleep(2 * time.Millisecond)
	repo.UpdateCheckTime(X)
	second := repo.Get(X)

	if !second.LastCheck.After(first.LastCheck) {
		t.Errorf("LastCheck not updated: first=%v second=%v", first.LastCheck, second.LastCheck)
	}
	if second.Username != "u" {
		t.Errorf("Username should be unchanged, got %q", second.Username)
	}
}
