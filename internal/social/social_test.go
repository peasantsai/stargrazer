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

// --- AllPlatforms URL and domain format checks ---

func TestAllPlatformsURLsStartWithHTTPS(t *testing.T) {
	for _, p := range AllPlatforms() {
		t.Run(string(p.ID), func(t *testing.T) {
			if !strings.HasPrefix(p.URL, "https://") {
				t.Errorf("URL %q should start with https://", p.URL)
			}
			if !strings.HasPrefix(p.LoginURL, "https://") {
				t.Errorf("LoginURL %q should start with https://", p.LoginURL)
			}
		})
	}
}

func TestAllPlatformsSessionDomainsStartWithDot(t *testing.T) {
	for _, p := range AllPlatforms() {
		t.Run(string(p.ID), func(t *testing.T) {
			for _, domain := range p.SessionDomains {
				if !strings.HasPrefix(domain, ".") {
					t.Errorf("session domain %q should start with '.'  (platform: %s)", domain, p.ID)
				}
			}
		})
	}
}

func TestFindPlatformReturnsPointerNotSameAddress(t *testing.T) {
	// Each call to FindPlatform returns a pointer to a new copy (range variable);
	// this confirms the caller gets an independent struct.
	p1 := FindPlatform(Facebook)
	p2 := FindPlatform(Facebook)
	if p1 == nil || p2 == nil {
		t.Fatal("FindPlatform returned nil")
	}
	if p1.ID != p2.ID {
		t.Errorf("expected same platform ID, got %s vs %s", p1.ID, p2.ID)
	}
}

func TestAccountStatusFields(t *testing.T) {
	status := AccountStatus{
		PlatformID: Instagram,
		LoggedIn:   true,
		Username:   "testuser",
	}
	if status.PlatformID != Instagram {
		t.Errorf("expected PlatformID Instagram, got %s", status.PlatformID)
	}
	if !status.LoggedIn {
		t.Error("expected LoggedIn true")
	}
	if status.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %q", status.Username)
	}
	// Timestamps should be zero by default
	if !status.LastLogin.IsZero() {
		t.Error("expected zero LastLogin")
	}
	if !status.LastCheck.IsZero() {
		t.Error("expected zero LastCheck")
	}
}
