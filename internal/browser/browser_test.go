package browser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"stargrazer/internal/config"
)

func TestParseNetscapeCookiesValid(t *testing.T) {
	input := ".example.com\tTRUE\t/\tTRUE\t1700000000\tsession_id\tabc123\n" +
		".example.com\tTRUE\t/path\tFALSE\t1700000001\tother_cookie\txyz789\n"

	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}

	c := cookies[0]
	if c.Domain != ".example.com" {
		t.Errorf("expected domain '.example.com', got %q", c.Domain)
	}
	if c.Path != "/" {
		t.Errorf("expected path '/', got %q", c.Path)
	}
	if !c.Secure {
		t.Error("expected Secure true")
	}
	if c.Expires != 1700000000 {
		t.Errorf("expected expires 1700000000, got %f", c.Expires)
	}
	if c.Name != "session_id" {
		t.Errorf("expected name 'session_id', got %q", c.Name)
	}
	if c.Value != "abc123" {
		t.Errorf("expected value 'abc123', got %q", c.Value)
	}

	c2 := cookies[1]
	if c2.Secure {
		t.Error("expected Secure false for second cookie")
	}
	if c2.Name != "other_cookie" {
		t.Errorf("expected name 'other_cookie', got %q", c2.Name)
	}
}

func TestParseNetscapeCookiesSkipsComments(t *testing.T) {
	input := "# Netscape HTTP Cookie File\n" +
		"# This is a comment\n" +
		".example.com\tTRUE\t/\tTRUE\t1700000000\tname\tvalue\n"

	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie (comments skipped), got %d", len(cookies))
	}
	if cookies[0].Name != "name" {
		t.Errorf("expected name 'name', got %q", cookies[0].Name)
	}
}

func TestParseNetscapeCookiesSkipsEmptyLines(t *testing.T) {
	input := "\n\n.example.com\tTRUE\t/\tTRUE\t1700000000\tname\tvalue\n\n\n"

	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie (empty lines skipped), got %d", len(cookies))
	}
}

func TestParseNetscapeCookiesReturnsEmptyForInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only comments", "# comment\n# another"},
		{"too few fields", "domain\tpath\tvalue\n"},
		{"garbage", "this is not a cookie format"},
		{"only whitespace", "   \n   \n   "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cookies := ParseNetscapeCookies(tc.input)
			if len(cookies) != 0 {
				t.Errorf("expected 0 cookies, got %d", len(cookies))
			}
		})
	}
}

func TestParseNetscapeCookiesHandlesMixedInput(t *testing.T) {
	input := "# Header\n" +
		"\n" +
		".valid.com\tTRUE\t/\tTRUE\t0\tcookie1\tval1\n" +
		"bad line\n" +
		".valid2.com\tFALSE\t/foo\tFALSE\t999\tcookie2\tval2\n"

	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}
}

func TestParseNetscapeCookiesZeroExpiry(t *testing.T) {
	input := ".example.com\tTRUE\t/\tTRUE\t0\tsession\tval\n"

	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Expires != 0 {
		t.Errorf("expected expires 0, got %f", cookies[0].Expires)
	}
}

func TestParseNetscapeCookiesSecureCaseInsensitive(t *testing.T) {
	input := ".example.com\tTRUE\t/\ttrue\t0\tname\tvalue\n"
	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if !cookies[0].Secure {
		t.Error("expected Secure true (case insensitive)")
	}
}

func TestParseNetscapeCookiesWindowsLineEndings(t *testing.T) {
	input := ".example.com\tTRUE\t/\tTRUE\t0\tname1\tval1\r\n" +
		".example.com\tTRUE\t/\tFALSE\t0\tname2\tval2\r\n"

	cookies := ParseNetscapeCookies(input)
	// The split is on \n, so \r may remain in value. The function trims lines.
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}
	// Values may have trailing \r due to TrimSpace on the line.
	if strings.TrimSpace(cookies[0].Name) != "name1" {
		t.Errorf("expected name 'name1', got %q", cookies[0].Name)
	}
}

func TestGetInstanceReturnsSingleton(t *testing.T) {
	// Reset the singleton for test isolation.
	once = sync.Once{}
	instance = nil
	defer func() {
		once = sync.Once{}
		instance = nil
	}()

	m1 := GetInstance()
	m2 := GetInstance()
	if m1 != m2 {
		t.Error("GetInstance() returned different pointers")
	}
}

func TestNewManagerReturnsSameAsGetInstance(t *testing.T) {
	once = sync.Once{}
	instance = nil
	defer func() {
		once = sync.Once{}
		instance = nil
	}()

	m1 := GetInstance()
	m2 := NewManager()
	if m1 != m2 {
		t.Error("NewManager() returned different pointer than GetInstance()")
	}
}

func TestManagerStartsInStatusStopped(t *testing.T) {
	once = sync.Once{}
	instance = nil
	defer func() {
		once = sync.Once{}
		instance = nil
	}()

	m := GetInstance()
	status, errMsg := m.GetStatus()
	if status != StatusStopped {
		t.Errorf("expected StatusStopped, got %s", status)
	}
	if errMsg != "" {
		t.Errorf("expected empty error message, got %q", errMsg)
	}
}

func TestManagerIsRunningReturnsFalseInitially(t *testing.T) {
	once = sync.Once{}
	instance = nil
	defer func() {
		once = sync.Once{}
		instance = nil
	}()

	m := GetInstance()
	if m.IsRunning() {
		t.Error("expected IsRunning() false for new manager")
	}
}

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusStopped, "stopped"},
		{StatusStarting, "starting"},
		{StatusRunning, "running"},
		{StatusError, "error"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.status) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.status))
			}
		})
	}
}

func TestCDPCookieStruct(t *testing.T) {
	c := CDPCookie{
		Name:     "session",
		Value:    "abc",
		Domain:   ".example.com",
		Path:     "/",
		Expires:  1700000000,
		HTTPOnly: true,
		Secure:   true,
	}

	if c.Name != "session" {
		t.Errorf("expected name 'session', got %q", c.Name)
	}
	if !c.HTTPOnly {
		t.Error("expected HTTPOnly true")
	}
	if !c.Secure {
		t.Error("expected Secure true")
	}
}

// --- buildArgs tests ---

func resetSingleton(t *testing.T) *Manager {
	t.Helper()
	once = sync.Once{}
	instance = nil
	t.Cleanup(func() {
		once = sync.Once{}
		instance = nil
	})
	return GetInstance()
}

func TestBuildArgsDefault(t *testing.T) {
	m := resetSingleton(t)

	cfg := config.BrowserConfig{CDPPort: 9222}
	args := m.buildArgs(cfg)

	found := false
	for _, a := range args {
		if a == "--remote-debugging-port=9222" {
			found = true
		}
		if a == "--headless=new" {
			t.Error("--headless=new should not be present when Headless is false")
		}
		if strings.HasPrefix(a, "--user-data-dir=") {
			t.Error("--user-data-dir should not be present when UserDataDir is empty")
		}
		if strings.HasPrefix(a, "--window-size=") {
			t.Error("--window-size should not be present when dimensions are 0")
		}
	}
	if !found {
		t.Error("expected --remote-debugging-port=9222 in args")
	}

	// Should always contain --no-first-run
	hasNoFirstRun := false
	for _, a := range args {
		if a == "--no-first-run" {
			hasNoFirstRun = true
		}
	}
	if !hasNoFirstRun {
		t.Error("expected --no-first-run in args")
	}

	// Should always contain --no-default-browser-check
	hasNoCheck := false
	for _, a := range args {
		if a == "--no-default-browser-check" {
			hasNoCheck = true
		}
	}
	if !hasNoCheck {
		t.Error("expected --no-default-browser-check in args")
	}
}

func TestBuildArgsHeadless(t *testing.T) {
	m := resetSingleton(t)

	cfg := config.BrowserConfig{CDPPort: 9222, Headless: true}
	args := m.buildArgs(cfg)

	found := false
	for _, a := range args {
		if a == "--headless=new" {
			found = true
		}
	}
	if !found {
		t.Error("expected --headless=new when Headless is true")
	}
}

func TestBuildArgsUserDataDir(t *testing.T) {
	m := resetSingleton(t)

	cfg := config.BrowserConfig{CDPPort: 9222, UserDataDir: "/tmp/profile"}
	args := m.buildArgs(cfg)

	found := false
	for _, a := range args {
		if a == "--user-data-dir=/tmp/profile" {
			found = true
		}
	}
	if !found {
		t.Error("expected --user-data-dir=/tmp/profile in args")
	}
}

func TestBuildArgsWindowSize(t *testing.T) {
	m := resetSingleton(t)

	cfg := config.BrowserConfig{CDPPort: 9222, WindowWidth: 1920, WindowHeight: 1080}
	args := m.buildArgs(cfg)

	found := false
	for _, a := range args {
		if a == "--window-size=1920,1080" {
			found = true
		}
	}
	if !found {
		t.Error("expected --window-size=1920,1080 in args")
	}
}

func TestBuildArgsWindowSizePartialZero(t *testing.T) {
	m := resetSingleton(t)

	// Only width set, height is 0 — should NOT add --window-size
	cfg := config.BrowserConfig{CDPPort: 9222, WindowWidth: 1920, WindowHeight: 0}
	args := m.buildArgs(cfg)

	for _, a := range args {
		if strings.HasPrefix(a, "--window-size=") {
			t.Error("--window-size should not be present when one dimension is 0")
		}
	}
}

func TestBuildArgsExtraFlags(t *testing.T) {
	m := resetSingleton(t)

	cfg := config.BrowserConfig{
		CDPPort:    9222,
		ExtraFlags: []string{"--disable-gpu", "--mute-audio"},
	}
	args := m.buildArgs(cfg)

	hasGPU := false
	hasMute := false
	for _, a := range args {
		if a == "--disable-gpu" {
			hasGPU = true
		}
		if a == "--mute-audio" {
			hasMute = true
		}
	}
	if !hasGPU {
		t.Error("expected --disable-gpu in args")
	}
	if !hasMute {
		t.Error("expected --mute-audio in args")
	}
}

func TestBuildArgsCombined(t *testing.T) {
	m := resetSingleton(t)

	cfg := config.BrowserConfig{
		CDPPort:      9333,
		Headless:     true,
		UserDataDir:  "/data/session",
		WindowWidth:  800,
		WindowHeight: 600,
		ExtraFlags:   []string{"--custom-flag"},
	}
	args := m.buildArgs(cfg)

	checks := map[string]bool{
		"--remote-debugging-port=9333": false,
		"--headless=new":               false,
		"--user-data-dir=/data/session": false,
		"--window-size=800,600":         false,
		"--custom-flag":                 false,
		"--no-first-run":                false,
	}
	for _, a := range args {
		if _, ok := checks[a]; ok {
			checks[a] = true
		}
	}
	for expected, found := range checks {
		if !found {
			t.Errorf("expected %q in args", expected)
		}
	}
}

// --- resolveChromiumPath tests ---

func TestResolveChromiumPathOverride(t *testing.T) {
	m := resetSingleton(t)
	result := m.resolveChromiumPath("/custom/path/to/chrome")
	if result != "/custom/path/to/chrome" {
		t.Errorf("expected override path, got %q", result)
	}
}

func TestResolveChromiumPathFallback(t *testing.T) {
	m := resetSingleton(t)
	result := m.resolveChromiumPath("")
	// Should return something non-empty (at minimum the binary name)
	if result == "" {
		t.Error("expected non-empty fallback path")
	}
}

func TestResolveChromiumPathPublicMethod(t *testing.T) {
	m := resetSingleton(t)
	result := m.ResolveChromiumPath()
	if result == "" {
		t.Error("ResolveChromiumPath() returned empty string")
	}
}

// --- resolveExtensionPath tests ---

func TestResolveExtensionPathInTestEnv(t *testing.T) {
	m := resetSingleton(t)
	// In test environment, the assets/cookies-extension directory likely does not exist
	// so this should return empty or a valid path. Either is acceptable.
	result := m.resolveExtensionPath()
	// Just verify it doesn't panic; the result depends on the test environment
	_ = result
}

// --- Stop when not running ---

func TestStopWhenNotRunning(t *testing.T) {
	m := resetSingleton(t)
	err := m.Stop()
	if err != nil {
		t.Errorf("Stop() on non-running manager returned error: %v", err)
	}
	status, _ := m.GetStatus()
	if status != StatusStopped {
		t.Errorf("expected StatusStopped after Stop(), got %s", status)
	}
}

// --- StartWithOptions error paths ---

func TestStartWithOptionsAlreadyRunning(t *testing.T) {
	m := resetSingleton(t)
	m.mu.Lock()
	m.status = StatusRunning
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		m.status = StatusStopped
		m.mu.Unlock()
	}()

	err := m.StartWithOptions("", "")
	if err == nil {
		t.Fatal("expected error when browser already running")
	}
	if !strings.Contains(err.Error(), "already") {
		t.Errorf("expected 'already' in error, got %q", err.Error())
	}
}

func TestStartWithOptionsAlreadyStarting(t *testing.T) {
	m := resetSingleton(t)
	m.mu.Lock()
	m.status = StatusStarting
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		m.status = StatusStopped
		m.mu.Unlock()
	}()

	err := m.StartWithOptions("", "")
	if err == nil {
		t.Fatal("expected error when browser already starting")
	}
	if !strings.Contains(err.Error(), "already") {
		t.Errorf("expected 'already' in error, got %q", err.Error())
	}
}

func TestStartWithOptionsChromiumNotFound(t *testing.T) {
	m := resetSingleton(t)

	// Set a chromium path override to a nonexistent binary
	config.Update(func(c *config.AppConfig) {
		c.Browser.ChromiumPath = "/nonexistent/path/to/chrome_binary_xyz"
	})
	defer config.Reset()

	err := m.StartWithOptions("", "")
	if err == nil {
		t.Fatal("expected error when chromium not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got %q", err.Error())
	}

	status, lastErr := m.GetStatus()
	if status != StatusError {
		t.Errorf("expected StatusError, got %s", status)
	}
	if lastErr == "" {
		t.Error("expected non-empty lastErr")
	}
}

func TestStartCallsStartWithOptions(t *testing.T) {
	m := resetSingleton(t)

	// Start without any chromium installed should fail
	config.Update(func(c *config.AppConfig) {
		c.Browser.ChromiumPath = "/nonexistent/chrome_for_start_test"
	})
	defer config.Reset()

	err := m.Start()
	if err == nil {
		t.Fatal("expected error from Start() when chromium not found")
	}
}

// --- LoadCookiesFromDisk tests ---

func TestLoadCookiesFromDiskValid(t *testing.T) {
	tmpDir := t.TempDir()
	cookiesDir := filepath.Join(tmpDir, "cookies")
	os.MkdirAll(cookiesDir, 0700)

	data := `[{"name":"session","value":"abc","domain":".example.com","path":"/","expires":1700000000}]`
	os.WriteFile(filepath.Join(cookiesDir, "test.json"), []byte(data), 0600)

	cookies, err := LoadCookiesFromDisk("test", tmpDir)
	if err != nil {
		t.Fatalf("LoadCookiesFromDisk returned error: %v", err)
	}
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "session" {
		t.Errorf("expected name 'session', got %q", cookies[0].Name)
	}
	if cookies[0].Value != "abc" {
		t.Errorf("expected value 'abc', got %q", cookies[0].Value)
	}
	if cookies[0].Domain != ".example.com" {
		t.Errorf("expected domain '.example.com', got %q", cookies[0].Domain)
	}
}

func TestLoadCookiesFromDiskMultipleCookies(t *testing.T) {
	tmpDir := t.TempDir()
	cookiesDir := filepath.Join(tmpDir, "cookies")
	os.MkdirAll(cookiesDir, 0700)

	data := `[
		{"name":"a","value":"1","domain":".ex.com"},
		{"name":"b","value":"2","domain":".ex.com"},
		{"name":"c","value":"3","domain":".other.com"}
	]`
	os.WriteFile(filepath.Join(cookiesDir, "multi.json"), []byte(data), 0600)

	cookies, err := LoadCookiesFromDisk("multi", tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cookies) != 3 {
		t.Fatalf("expected 3 cookies, got %d", len(cookies))
	}
}

func TestLoadCookiesFromDiskMissing(t *testing.T) {
	_, err := LoadCookiesFromDisk("nonexistent", t.TempDir())
	if err == nil {
		t.Error("expected error for missing cookies file")
	}
}

func TestLoadCookiesFromDiskInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cookiesDir := filepath.Join(tmpDir, "cookies")
	os.MkdirAll(cookiesDir, 0700)
	os.WriteFile(filepath.Join(cookiesDir, "bad.json"), []byte("not valid json {["), 0600)

	_, err := LoadCookiesFromDisk("bad", tmpDir)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadCookiesFromDiskEmptyArray(t *testing.T) {
	tmpDir := t.TempDir()
	cookiesDir := filepath.Join(tmpDir, "cookies")
	os.MkdirAll(cookiesDir, 0700)
	os.WriteFile(filepath.Join(cookiesDir, "empty.json"), []byte("[]"), 0600)

	cookies, err := LoadCookiesFromDisk("empty", tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cookies) != 0 {
		t.Errorf("expected 0 cookies, got %d", len(cookies))
	}
}

// --- pinExtension tests ---

func TestPinExtensionCreatesPrefsFile(t *testing.T) {
	m := resetSingleton(t)
	tmpDir := t.TempDir()

	m.pinExtension(tmpDir)

	prefsPath := filepath.Join(tmpDir, "Default", "Preferences")
	data, err := os.ReadFile(prefsPath)
	if err != nil {
		t.Fatalf("Preferences file not created: %v", err)
	}

	var prefs map[string]interface{}
	if err := json.Unmarshal(data, &prefs); err != nil {
		t.Fatalf("Preferences is not valid JSON: %v", err)
	}

	// Verify the extension is pinned
	extensions, ok := prefs["extensions"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'extensions' key in prefs")
	}
	toolbar, ok := extensions["toolbar"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'toolbar' key in extensions")
	}
	pinned, ok := toolbar["pinned_extensions"].([]interface{})
	if !ok {
		t.Fatal("expected 'pinned_extensions' key in toolbar")
	}
	if len(pinned) != 1 {
		t.Fatalf("expected 1 pinned extension, got %d", len(pinned))
	}
	if pinned[0] != "edoacekkjanmingkbkgjndndibhkegno" {
		t.Errorf("unexpected extension ID: %v", pinned[0])
	}
}

func TestPinExtensionMergesExistingPrefs(t *testing.T) {
	m := resetSingleton(t)
	tmpDir := t.TempDir()

	// Write existing prefs
	defaultDir := filepath.Join(tmpDir, "Default")
	os.MkdirAll(defaultDir, 0700)
	existing := `{"browser":{"theme":"dark"}}`
	os.WriteFile(filepath.Join(defaultDir, "Preferences"), []byte(existing), 0600)

	m.pinExtension(tmpDir)

	data, err := os.ReadFile(filepath.Join(defaultDir, "Preferences"))
	if err != nil {
		t.Fatalf("failed to read prefs: %v", err)
	}

	var prefs map[string]interface{}
	json.Unmarshal(data, &prefs)

	// Should have both "browser" and "extensions" keys
	if _, ok := prefs["browser"]; !ok {
		t.Error("expected 'browser' key preserved in merged prefs")
	}
	if _, ok := prefs["extensions"]; !ok {
		t.Error("expected 'extensions' key added in merged prefs")
	}

	// Verify browser.theme is preserved
	browserMap, _ := prefs["browser"].(map[string]interface{})
	if browserMap["theme"] != "dark" {
		t.Error("expected browser.theme 'dark' to be preserved")
	}
}

func TestPinExtensionIdempotent(t *testing.T) {
	m := resetSingleton(t)
	tmpDir := t.TempDir()

	// Pin twice — should not duplicate the extension ID
	m.pinExtension(tmpDir)
	m.pinExtension(tmpDir)

	prefsPath := filepath.Join(tmpDir, "Default", "Preferences")
	data, _ := os.ReadFile(prefsPath)
	var prefs map[string]interface{}
	json.Unmarshal(data, &prefs)

	extensions := prefs["extensions"].(map[string]interface{})
	toolbar := extensions["toolbar"].(map[string]interface{})
	pinned := toolbar["pinned_extensions"].([]interface{})

	if len(pinned) != 1 {
		t.Errorf("expected 1 pinned extension after double pin, got %d", len(pinned))
	}
}

func TestPinExtensionWithExistingExtensions(t *testing.T) {
	m := resetSingleton(t)
	tmpDir := t.TempDir()

	// Write prefs with existing extensions/toolbar structure
	defaultDir := filepath.Join(tmpDir, "Default")
	os.MkdirAll(defaultDir, 0700)
	existing := `{"extensions":{"toolbar":{"pinned_extensions":["some_other_ext"]}}}`
	os.WriteFile(filepath.Join(defaultDir, "Preferences"), []byte(existing), 0600)

	m.pinExtension(tmpDir)

	data, _ := os.ReadFile(filepath.Join(defaultDir, "Preferences"))
	var prefs map[string]interface{}
	json.Unmarshal(data, &prefs)

	extensions := prefs["extensions"].(map[string]interface{})
	toolbar := extensions["toolbar"].(map[string]interface{})
	pinned := toolbar["pinned_extensions"].([]interface{})

	if len(pinned) != 2 {
		t.Errorf("expected 2 pinned extensions, got %d", len(pinned))
	}
}

// --- findChromiumInAssets ---

func TestFindChromiumInAssetsFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake Chromium directory structure: assets/uc-123/chrome.exe
	assetDir := filepath.Join(tmpDir, "uc-123")
	os.MkdirAll(assetDir, 0700)
	os.WriteFile(filepath.Join(assetDir, "chrome.exe"), []byte("fake"), 0700)

	result := findChromiumInAssets(tmpDir, "chrome.exe")
	if result == "" {
		t.Error("expected to find chrome.exe in assets subdirectory")
	}
	if !strings.HasSuffix(result, "chrome.exe") {
		t.Errorf("expected result to end with 'chrome.exe', got %q", result)
	}
}

func TestFindChromiumInAssetsNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	// Empty directory — nothing to find.
	result := findChromiumInAssets(tmpDir, "chrome.exe")
	if result != "" {
		t.Errorf("expected empty result for missing binary, got %q", result)
	}
}

func TestFindChromiumInAssetsMissingDirectory(t *testing.T) {
	// Non-existent directory should return "".
	result := findChromiumInAssets("/nonexistent/path/xyz", "chrome.exe")
	if result != "" {
		t.Errorf("expected empty result for missing directory, got %q", result)
	}
}

func TestFindChromiumInAssetsSkipsFiles(t *testing.T) {
	tmpDir := t.TempDir()
	// A regular file (not a dir) should be ignored in the search.
	os.WriteFile(filepath.Join(tmpDir, "chrome.exe"), []byte("fake"), 0700)

	result := findChromiumInAssets(tmpDir, "chrome.exe")
	// findChromiumInAssets only looks inside subdirectories, so this should not be found.
	if result != "" {
		t.Errorf("expected empty result (file at root, not in subdir), got %q", result)
	}
}

// --- CDPCookie JSON serialization ---

func TestCDPCookieJSONRoundTrip(t *testing.T) {
	cookie := CDPCookie{
		Name:     "session_id",
		Value:    "abc123",
		Domain:   ".example.com",
		Path:     "/path",
		Expires:  1700000000.5,
		HTTPOnly: true,
		Secure:   false,
	}

	data, err := json.Marshal(cookie)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var decoded CDPCookie
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if decoded.Name != cookie.Name {
		t.Errorf("Name mismatch: %q vs %q", decoded.Name, cookie.Name)
	}
	if decoded.Value != cookie.Value {
		t.Errorf("Value mismatch: %q vs %q", decoded.Value, cookie.Value)
	}
	if decoded.Domain != cookie.Domain {
		t.Errorf("Domain mismatch: %q vs %q", decoded.Domain, cookie.Domain)
	}
	if decoded.Expires != cookie.Expires {
		t.Errorf("Expires mismatch: %f vs %f", decoded.Expires, cookie.Expires)
	}
	if decoded.HTTPOnly != cookie.HTTPOnly {
		t.Errorf("HTTPOnly mismatch: %v vs %v", decoded.HTTPOnly, cookie.HTTPOnly)
	}
}

// --- ParseNetscapeCookies value with tabs ---

func TestParseNetscapeCookiesValueWithSpaces(t *testing.T) {
	// Value field may contain spaces (not split further)
	input := ".example.com\tTRUE\t/\tFALSE\t0\ttoken\tmy token value\n"
	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Value != "my token value" {
		t.Errorf("expected value 'my token value', got %q", cookies[0].Value)
	}
}

func TestParseNetscapeCookiesExactlySevenFields(t *testing.T) {
	// Exactly 7 tab-separated fields — should be parsed.
	input := ".ex.com\tTRUE\t/\tFALSE\t0\tn\tv\n"
	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie for exactly 7 fields, got %d", len(cookies))
	}
}

func TestParseNetscapeCookiesSixFields(t *testing.T) {
	// Only 6 fields — should be skipped (len < 7).
	input := ".ex.com\tTRUE\t/\tFALSE\t0\tname\n"
	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 0 {
		t.Errorf("expected 0 cookies for 6-field line, got %d", len(cookies))
	}
}

// --- ExportCookiesToDisk ---

func TestExportCookiesToDiskEmptyCookies(t *testing.T) {
	m := resetSingleton(t)
	tmpDir := t.TempDir()

	// GetCookiesForDomains will fail because no browser is running.
	// But we can test the "empty cookies" path by writing an empty file first.
	cookiesDir := filepath.Join(tmpDir, "cookies")
	os.MkdirAll(cookiesDir, 0700)
	os.WriteFile(filepath.Join(cookiesDir, "testplatform.json"), []byte("[]"), 0600)

	// LoadCookiesFromDisk with empty array — not ExportCookiesToDisk directly
	// (Export needs a running browser), but we verify LoadCookiesFromDisk handles []
	cookies, err := LoadCookiesFromDisk("testplatform", tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cookies) != 0 {
		t.Errorf("expected 0 cookies from empty file, got %d", len(cookies))
	}
	_ = m // used via resetSingleton
}
