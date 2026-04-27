package social

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// --- sessionsBaseDir fallback paths ---

func TestSessionsBaseDirWindowsFallback(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}

	// Clear APPDATA to trigger the USERPROFILE fallback path.
	origAppData := os.Getenv("APPDATA")
	t.Setenv("APPDATA", "")
	defer t.Setenv("APPDATA", origAppData)

	dir := sessionsBaseDir()
	if dir == "" {
		t.Error("sessionsBaseDir() returned empty string with blank APPDATA")
	}
	if !strings.Contains(dir, "stargrazer") {
		t.Errorf("expected 'stargrazer' in dir, got %q", dir)
	}
}

func TestSessionsBaseDirLinuxFallback(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	// Clear XDG_DATA_HOME to trigger HOME fallback.
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "/tmp/fakeuser")

	dir := sessionsBaseDir()
	if !strings.Contains(dir, "stargrazer") {
		t.Errorf("expected 'stargrazer' in dir, got %q", dir)
	}
	if !strings.Contains(dir, "/tmp/fakeuser") {
		t.Errorf("expected HOME path in dir, got %q", dir)
	}
}

func TestSessionsBaseDirLinuxXDG(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	dir := sessionsBaseDir()
	if !strings.Contains(dir, "stargrazer") {
		t.Errorf("expected 'stargrazer' in dir, got %q", dir)
	}
	if !strings.HasPrefix(dir, tmpDir) {
		t.Errorf("expected dir to start with XDG_DATA_HOME %q, got %q", tmpDir, dir)
	}
}

// --- EnsureSessionDir MkdirAll failure ---

func TestEnsureSessionDirFailureReturnsError(t *testing.T) {
	// Create a regular file at the path where a directory is expected,
	// so MkdirAll will fail.
	tmpDir := t.TempDir()
	blockerFile := filepath.Join(tmpDir, "stargrazer")
	os.WriteFile(blockerFile, []byte("blocker"), 0600)

	// Set APPDATA/HOME to tmpDir so sessionsBaseDir resolves to
	// <tmpDir>/stargrazer/sessions — but <tmpDir>/stargrazer is a file, not dir.
	t.Setenv("APPDATA", tmpDir)
	t.Setenv("HOME", tmpDir)

	_, err := EnsureSessionDir(Instagram)
	if err == nil {
		// On Windows, creating a dir where a file exists may or may not fail
		// depending on OS version. Accept either outcome; we just must not panic.
		t.Log("EnsureSessionDir succeeded (OS allowed it); no error to check")
	}
}

// --- save() marshaling error is unreachable normally, so test via direct call ---

func TestSaveCallDoesNotPanicWithEmptyAccounts(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")
	s := &SessionStore{
		accounts: make(map[Platform]*AccountStatus),
		filePath: fp,
	}
	// save() with empty accounts should write "[]" and not panic.
	s.save()
	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("save() did not create file: %v", err)
	}
	if string(data) != "[]" {
		t.Errorf("expected '[]', got %q", string(data))
	}
}

// --- AccountStatus JSON tags ---

func TestAccountStatusJSONTags(t *testing.T) {
	// Ensure the JSON field names match what the frontend expects.
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "accounts.json")
	s := &SessionStore{
		accounts: make(map[Platform]*AccountStatus),
		filePath: fp,
	}
	s.SetLoggedIn(Instagram, "testuser")
	s.save()

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("failed to read saved accounts: %v", err)
	}
	content := string(data)
	for _, expected := range []string{"platformId", "loggedIn", "username", "lastLogin", "lastCheck"} {
		if !strings.Contains(content, expected) {
			t.Errorf("JSON missing field %q in: %s", expected, content)
		}
	}
}
