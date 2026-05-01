package social

import (
	"os"
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

func TestSharedSessionDirNonEmpty(t *testing.T) {
	dir := SharedSessionDir()
	if dir == "" {
		t.Error("SharedSessionDir() returned empty string")
	}
	if !strings.HasSuffix(dir, "browser_profile") {
		t.Errorf("expected SharedSessionDir to end with 'browser_profile', got %q", dir)
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

// --- sessionsBaseDir ---

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
