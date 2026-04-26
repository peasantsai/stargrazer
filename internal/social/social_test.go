package social

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllPlatformsReturns6(t *testing.T) {
	platforms := AllPlatforms()
	if len(platforms) != 6 {
		t.Errorf("expected 6 platforms, got %d", len(platforms))
	}
}

func TestAllPlatformsContainsExpectedIDs(t *testing.T) {
	platforms := AllPlatforms()
	expected := map[Platform]bool{
		Facebook: false, Instagram: false, TikTok: false,
		YouTube: false, LinkedIn: false, X: false,
	}

	for _, p := range platforms {
		if _, ok := expected[p.ID]; ok {
			expected[p.ID] = true
		}
	}

	for id, found := range expected {
		if !found {
			t.Errorf("platform %s not found in AllPlatforms()", id)
		}
	}
}

func TestAllPlatformsHaveRequiredFields(t *testing.T) {
	for _, p := range AllPlatforms() {
		t.Run(string(p.ID), func(t *testing.T) {
			if p.Name == "" {
				t.Error("Name is empty")
			}
			if p.URL == "" {
				t.Error("URL is empty")
			}
			if p.LoginURL == "" {
				t.Error("LoginURL is empty")
			}
			if len(p.SessionDomains) == 0 {
				t.Error("SessionDomains is empty")
			}
			if len(p.LoginCookies) == 0 {
				t.Error("LoginCookies is empty")
			}
		})
	}
}

func TestFindPlatformExisting(t *testing.T) {
	tests := []struct {
		id   Platform
		name string
	}{
		{Facebook, "Facebook"},
		{Instagram, "Instagram"},
		{TikTok, "TikTok"},
		{YouTube, "YouTube"},
		{LinkedIn, "LinkedIn"},
		{X, "X"},
	}

	for _, tc := range tests {
		t.Run(string(tc.id), func(t *testing.T) {
			p := FindPlatform(tc.id)
			if p == nil {
				t.Fatalf("FindPlatform(%q) returned nil", tc.id)
			}
			if p.Name != tc.name {
				t.Errorf("expected name %q, got %q", tc.name, p.Name)
			}
		})
	}
}

func TestFindPlatformUnknown(t *testing.T) {
	p := FindPlatform("nonexistent")
	if p != nil {
		t.Errorf("expected nil for unknown platform, got %+v", p)
	}
}

func TestNewSessionStoreCreatesEmptyStore(t *testing.T) {
	// Use a temp dir to avoid touching real session files.
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")

	s := &SessionStore{
		accounts: make(map[Platform]*AccountStatus),
		filePath: fp,
	}

	all := s.GetAll()
	if len(all) != 6 {
		t.Errorf("expected 6 entries from GetAll, got %d", len(all))
	}
	for _, a := range all {
		if a.LoggedIn {
			t.Errorf("platform %s should not be logged in", a.PlatformID)
		}
	}
}

func TestSetLoggedInAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")

	s := &SessionStore{
		accounts: make(map[Platform]*AccountStatus),
		filePath: fp,
	}

	s.SetLoggedIn(Instagram, "testuser")

	status := s.Get(Instagram)
	if !status.LoggedIn {
		t.Error("expected LoggedIn true")
	}
	if status.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %q", status.Username)
	}
	if status.PlatformID != Instagram {
		t.Errorf("expected platform Instagram, got %s", status.PlatformID)
	}
	if status.LastLogin.IsZero() {
		t.Error("expected non-zero LastLogin")
	}
}

func TestSetLoggedOut(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")

	s := &SessionStore{
		accounts: make(map[Platform]*AccountStatus),
		filePath: fp,
	}

	s.SetLoggedIn(Facebook, "fbuser")
	s.SetLoggedOut(Facebook)

	status := s.Get(Facebook)
	if status.LoggedIn {
		t.Error("expected LoggedIn false after SetLoggedOut")
	}
	if status.Username != "" {
		t.Errorf("expected empty username after logout, got %q", status.Username)
	}
}

func TestGetReturnsDefaultForUnknownPlatform(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")

	s := &SessionStore{
		accounts: make(map[Platform]*AccountStatus),
		filePath: fp,
	}

	status := s.Get(TikTok)
	if status.LoggedIn {
		t.Error("expected LoggedIn false for unset platform")
	}
	if status.PlatformID != TikTok {
		t.Errorf("expected PlatformID TikTok, got %s", status.PlatformID)
	}
}

func TestGetAllReturnsAll6Entries(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")

	s := &SessionStore{
		accounts: make(map[Platform]*AccountStatus),
		filePath: fp,
	}

	s.SetLoggedIn(Instagram, "ig_user")
	s.SetLoggedIn(X, "x_user")

	all := s.GetAll()
	if len(all) != 6 {
		t.Fatalf("expected 6 entries, got %d", len(all))
	}

	loggedIn := 0
	for _, a := range all {
		if a.LoggedIn {
			loggedIn++
		}
	}
	if loggedIn != 2 {
		t.Errorf("expected 2 logged-in platforms, got %d", loggedIn)
	}
}

func TestSharedSessionDirNonEmpty(t *testing.T) {
	dir := SharedSessionDir()
	if dir == "" {
		t.Error("SharedSessionDir() returned empty string")
	}
	if !strings.HasSuffix(dir, "browser_profile") {
		t.Errorf("expected SharedSessionDir to end with 'browser_profile', got %q", dir)
	}
}

func TestSchedulesFilePath(t *testing.T) {
	fp := SchedulesFilePath()
	if fp == "" {
		t.Error("SchedulesFilePath() returned empty string")
	}
	if !strings.HasSuffix(fp, "schedules.json") {
		t.Errorf("expected SchedulesFilePath to end with 'schedules.json', got %q", fp)
	}
}

func TestEnsureSessionDirCreatesDirectory(t *testing.T) {
	// Override the base dir by using a temp environment variable.
	tmpDir := t.TempDir()
	origAppData := os.Getenv("APPDATA")
	origHome := os.Getenv("HOME")

	// Set environment so sessionsBaseDir resolves under tmpDir.
	t.Setenv("APPDATA", tmpDir)
	t.Setenv("HOME", tmpDir)

	dir, err := EnsureSessionDir(Instagram)
	if err != nil {
		t.Fatalf("EnsureSessionDir returned error: %v", err)
	}
	if dir == "" {
		t.Error("EnsureSessionDir returned empty path")
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("path is not a directory")
	}

	// Restore (t.Setenv handles cleanup, but be explicit for clarity).
	_ = origAppData
	_ = origHome
}

func TestSessionDirReturnsSameAsShared(t *testing.T) {
	// SessionDir should delegate to SharedSessionDir regardless of platform.
	dir := SessionDir(Instagram)
	shared := SharedSessionDir()
	if dir != shared {
		t.Errorf("SessionDir(%q) = %q, SharedSessionDir() = %q; expected same", Instagram, dir, shared)
	}
}

func TestSessionStorePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")

	// Create and populate a store.
	s1 := &SessionStore{
		accounts: make(map[Platform]*AccountStatus),
		filePath: fp,
	}
	s1.SetLoggedIn(YouTube, "yt_user")

	// Create a new store from the same file to test loading.
	s2 := &SessionStore{
		accounts: make(map[Platform]*AccountStatus),
		filePath: fp,
	}
	s2.load()

	status := s2.Get(YouTube)
	if !status.LoggedIn {
		t.Error("expected YouTube to be logged in after reload")
	}
	if status.Username != "yt_user" {
		t.Errorf("expected username 'yt_user', got %q", status.Username)
	}
}
