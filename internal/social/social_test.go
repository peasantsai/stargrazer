package social

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

// --- UpdateCheckTime tests ---

func TestUpdateCheckTime(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")
	s := &SessionStore{accounts: make(map[Platform]*AccountStatus), filePath: fp}

	s.SetLoggedIn(Instagram, "user")
	before := s.Get(Instagram).LastCheck

	time.Sleep(10 * time.Millisecond)
	s.UpdateCheckTime(Instagram)

	after := s.Get(Instagram).LastCheck
	if !after.After(before) {
		t.Errorf("expected LastCheck to be updated; before=%v, after=%v", before, after)
	}
}

func TestUpdateCheckTimeNoOp(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")
	s := &SessionStore{accounts: make(map[Platform]*AccountStatus), filePath: fp}

	// UpdateCheckTime on non-existent platform should be no-op (no crash)
	s.UpdateCheckTime(TikTok)

	status := s.Get(TikTok)
	if status.LoggedIn {
		t.Error("expected platform to remain not logged in")
	}
}

func TestUpdateCheckTimePersists(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")
	s := &SessionStore{accounts: make(map[Platform]*AccountStatus), filePath: fp}

	s.SetLoggedIn(Facebook, "fb_user")
	time.Sleep(10 * time.Millisecond)
	s.UpdateCheckTime(Facebook)

	// Reload and verify
	s2 := &SessionStore{accounts: make(map[Platform]*AccountStatus), filePath: fp}
	s2.load()

	status := s2.Get(Facebook)
	if !status.LoggedIn {
		t.Error("expected Facebook to still be logged in after reload")
	}
	if status.LastCheck.IsZero() {
		t.Error("expected non-zero LastCheck after reload")
	}
}

// --- load edge cases ---

func TestLoadWithInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")
	os.WriteFile(fp, []byte("not valid json {["), 0600)

	s := &SessionStore{accounts: make(map[Platform]*AccountStatus), filePath: fp}
	s.load()
	// Should not crash, accounts should remain empty
	if len(s.accounts) != 0 {
		t.Errorf("expected 0 accounts after invalid JSON load, got %d", len(s.accounts))
	}
}

func TestLoadWithMissingFile(t *testing.T) {
	s := &SessionStore{
		accounts: make(map[Platform]*AccountStatus),
		filePath: "/nonexistent/path/accounts.json",
	}
	s.load()
	// Should not crash
	if len(s.accounts) != 0 {
		t.Errorf("expected 0 accounts for missing file, got %d", len(s.accounts))
	}
}

func TestLoadWithEmptyArray(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")
	os.WriteFile(fp, []byte("[]"), 0600)

	s := &SessionStore{accounts: make(map[Platform]*AccountStatus), filePath: fp}
	s.load()
	if len(s.accounts) != 0 {
		t.Errorf("expected 0 accounts for empty array, got %d", len(s.accounts))
	}
}

// --- Platform constants ---

func TestPlatformConstants(t *testing.T) {
	tests := []struct {
		platform Platform
		expected string
	}{
		{Facebook, "facebook"},
		{Instagram, "instagram"},
		{TikTok, "tiktok"},
		{YouTube, "youtube"},
		{LinkedIn, "linkedin"},
		{X, "x"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.platform) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.platform))
			}
		})
	}
}

// --- sessionsBaseDir / sessionFilePath ---

func TestSessionsBaseDirReturnsNonEmpty(t *testing.T) {
	dir := sessionsBaseDir()
	if dir == "" {
		t.Error("sessionsBaseDir() returned empty string")
	}
	if !strings.Contains(dir, "stargrazer") {
		t.Errorf("expected 'stargrazer' in base dir, got %q", dir)
	}
	if !strings.HasSuffix(dir, "sessions") {
		t.Errorf("expected base dir to end with 'sessions', got %q", dir)
	}
}

func TestSessionFilePathReturnsNonEmpty(t *testing.T) {
	fp := sessionFilePath()
	if fp == "" {
		t.Error("sessionFilePath() returned empty string")
	}
	if !strings.HasSuffix(fp, "accounts.json") {
		t.Errorf("expected sessionFilePath to end with 'accounts.json', got %q", fp)
	}
}

// --- NewSessionStore ---

func TestNewSessionStoreLoadsExistingFile(t *testing.T) {
	// Set env so sessionFilePath resolves under our temp dir
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "stargrazer", "sessions")
	os.MkdirAll(sessionsDir, 0700)

	// Write a pre-existing accounts file
	data := `[{"platformId":"instagram","loggedIn":true,"username":"pre_user","lastLogin":"2025-01-01T00:00:00Z","lastCheck":"2025-01-01T00:00:00Z"}]`
	os.WriteFile(filepath.Join(sessionsDir, "accounts.json"), []byte(data), 0600)

	t.Setenv("APPDATA", tmpDir)
	t.Setenv("HOME", tmpDir)

	store := NewSessionStore()
	status := store.Get(Instagram)
	if !status.LoggedIn {
		t.Error("expected Instagram to be logged in from pre-existing file")
	}
	if status.Username != "pre_user" {
		t.Errorf("expected username 'pre_user', got %q", status.Username)
	}
}

// --- Multiple SetLoggedIn overwrite ---

func TestSetLoggedInOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")
	s := &SessionStore{accounts: make(map[Platform]*AccountStatus), filePath: fp}

	s.SetLoggedIn(Instagram, "user1")
	s.SetLoggedIn(Instagram, "user2")

	status := s.Get(Instagram)
	if status.Username != "user2" {
		t.Errorf("expected username 'user2' after overwrite, got %q", status.Username)
	}
}

// --- SetLoggedOut on non-existent platform ---

func TestSetLoggedOutNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")
	s := &SessionStore{accounts: make(map[Platform]*AccountStatus), filePath: fp}

	// Should not crash
	s.SetLoggedOut(LinkedIn)
	status := s.Get(LinkedIn)
	if status.LoggedIn {
		t.Error("expected LoggedIn false")
	}
	if status.LastCheck.IsZero() {
		t.Error("expected non-zero LastCheck after SetLoggedOut")
	}
}
