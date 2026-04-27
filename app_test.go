package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"stargrazer/internal/browser"
	"stargrazer/internal/config"
)

// --- extractUsernameFromCookies ---

func TestExtractUsernameFromCookiesFacebook(t *testing.T) {
	cookies := []browser.CDPCookie{
		{Name: "c_user", Value: "100001234567"},
		{Name: "xs", Value: "somevalue"},
	}
	got := extractUsernameFromCookies(cookies)
	if got != "100001234567" {
		t.Errorf("expected '100001234567', got %q", got)
	}
}

func TestExtractUsernameFromCookiesInstagram(t *testing.T) {
	cookies := []browser.CDPCookie{
		{Name: "ds_user_id", Value: "999888777"},
	}
	got := extractUsernameFromCookies(cookies)
	if got != "999888777" {
		t.Errorf("expected '999888777', got %q", got)
	}
}

func TestExtractUsernameFromCookiesTikTok(t *testing.T) {
	cookies := []browser.CDPCookie{
		{Name: "uid_tt", Value: "tiktok_uid_123"},
	}
	got := extractUsernameFromCookies(cookies)
	if got != "tiktok_uid_123" {
		t.Errorf("expected 'tiktok_uid_123', got %q", got)
	}
}

func TestExtractUsernameFromCookiesXTwitter(t *testing.T) {
	// twid value is URL-encoded: u%3D<user_id>
	cookies := []browser.CDPCookie{
		{Name: "twid", Value: "u%3D12345678"},
	}
	got := extractUsernameFromCookies(cookies)
	if got != "12345678" {
		t.Errorf("expected '12345678', got %q", got)
	}
}

func TestExtractUsernameFromCookiesXTwitterInvalidEncoding(t *testing.T) {
	// If URL-decode fails, should not panic and return ""
	cookies := []browser.CDPCookie{
		{Name: "twid", Value: "%ZZinvalid"},
	}
	got := extractUsernameFromCookies(cookies)
	// Returns "" because url.QueryUnescape fails
	if got != "" {
		t.Errorf("expected empty string for invalid encoding, got %q", got)
	}
}

func TestExtractUsernameFromCookiesUnknownCookies(t *testing.T) {
	cookies := []browser.CDPCookie{
		{Name: "some_other_cookie", Value: "value"},
		{Name: "another_cookie", Value: "value2"},
	}
	got := extractUsernameFromCookies(cookies)
	if got != "" {
		t.Errorf("expected empty string for unrecognised cookies, got %q", got)
	}
}

func TestExtractUsernameFromCookiesEmpty(t *testing.T) {
	got := extractUsernameFromCookies([]browser.CDPCookie{})
	if got != "" {
		t.Errorf("expected empty string for empty cookies, got %q", got)
	}
}

func TestExtractUsernameFromCookiesFirstMatch(t *testing.T) {
	// Only the first matching cookie should be returned
	cookies := []browser.CDPCookie{
		{Name: "c_user", Value: "first"},
		{Name: "ds_user_id", Value: "second"},
	}
	got := extractUsernameFromCookies(cookies)
	if got != "first" {
		t.Errorf("expected 'first' (first match wins), got %q", got)
	}
}

// --- persistCookiesToDisk ---

func TestPersistCookiesToDiskCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	cookies := []browser.CDPCookie{
		{Name: "session", Value: "abc123", Domain: ".example.com", Path: "/"},
	}

	persistCookiesToDisk("testplatform", "Test Platform", cookies, tmpDir)

	fp := filepath.Join(tmpDir, "cookies", "testplatform.json")
	if _, err := os.Stat(fp); err != nil {
		t.Fatalf("expected cookie file to be created, got: %v", err)
	}
}

func TestPersistCookiesToDiskCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	cookies := []browser.CDPCookie{
		{Name: "session", Value: "abc", Domain: ".example.com"},
	}

	persistCookiesToDisk("facebook", "Facebook", cookies, tmpDir)

	cookiesDir := filepath.Join(tmpDir, "cookies")
	info, err := os.Stat(cookiesDir)
	if err != nil {
		t.Fatalf("cookies directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected cookies path to be a directory")
	}
}

func TestPersistCookiesToDiskWritesValidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cookies := []browser.CDPCookie{
		{Name: "c_user", Value: "12345", Domain: ".facebook.com", Path: "/", Secure: true},
	}

	persistCookiesToDisk("facebook", "Facebook", cookies, tmpDir)

	data, err := os.ReadFile(filepath.Join(tmpDir, "cookies", "facebook.json"))
	if err != nil {
		t.Fatalf("failed to read cookie file: %v", err)
	}

	// Verify it's valid JSON by checking basic structure
	if len(data) == 0 {
		t.Error("expected non-empty cookie file")
	}
	// Should start with '['
	if data[0] != '[' {
		t.Errorf("expected JSON array, got: %c", data[0])
	}
}

// --- formatTime ---

func TestFormatTimeZero(t *testing.T) {
	got := formatTime(time.Time{})
	if got != "" {
		t.Errorf("expected empty string for zero time, got %q", got)
	}
}

func TestFormatTimeNonZero(t *testing.T) {
	ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	got := formatTime(ts)
	if got == "" {
		t.Error("expected non-empty string for non-zero time")
	}
	// Should be RFC3339 format
	if got != "2025-01-15T10:30:00Z" {
		t.Errorf("expected RFC3339 format, got %q", got)
	}
}

func TestFormatTimeRFC3339(t *testing.T) {
	ts := time.Now()
	got := formatTime(ts)
	// Should parse back to a valid time
	parsed, err := time.Parse(time.RFC3339, got)
	if err != nil {
		t.Errorf("formatTime returned invalid RFC3339: %v", err)
	}
	// Truncate sub-second precision for comparison
	if parsed.Unix() != ts.Unix() {
		t.Errorf("time round-trip failed: got %v, want %v", parsed, ts)
	}
}

// --- toBrowserConfigResponse ---

func TestToBrowserConfigResponseBasic(t *testing.T) {
	cfg := config.BrowserConfig{
		ChromiumPath: "/path/to/chrome",
		CDPPort:      9222,
		Headless:     true,
		UserDataDir:  "/tmp/profile",
		WindowWidth:  1280,
		WindowHeight: 900,
		ExtraFlags:   []string{"--disable-gpu"},
	}

	resp := toBrowserConfigResponse(cfg)

	if resp.ChromiumPath != "/path/to/chrome" {
		t.Errorf("expected ChromiumPath '/path/to/chrome', got %q", resp.ChromiumPath)
	}
	if resp.CDPPort != 9222 {
		t.Errorf("expected CDPPort 9222, got %d", resp.CDPPort)
	}
	if !resp.Headless {
		t.Error("expected Headless true")
	}
	if resp.UserDataDir != "/tmp/profile" {
		t.Errorf("expected UserDataDir '/tmp/profile', got %q", resp.UserDataDir)
	}
	if resp.WindowWidth != 1280 {
		t.Errorf("expected WindowWidth 1280, got %d", resp.WindowWidth)
	}
	if resp.WindowHeight != 900 {
		t.Errorf("expected WindowHeight 900, got %d", resp.WindowHeight)
	}
	if len(resp.ExtraFlags) != 1 || resp.ExtraFlags[0] != "--disable-gpu" {
		t.Errorf("expected ExtraFlags [--disable-gpu], got %v", resp.ExtraFlags)
	}
}

func TestToBrowserConfigResponseNilFlagsBecomesEmptySlice(t *testing.T) {
	cfg := config.BrowserConfig{
		CDPPort:    9222,
		ExtraFlags: nil,
	}

	resp := toBrowserConfigResponse(cfg)

	if resp.ExtraFlags == nil {
		t.Error("expected non-nil ExtraFlags for nil input")
	}
	if len(resp.ExtraFlags) != 0 {
		t.Errorf("expected empty ExtraFlags slice, got %v", resp.ExtraFlags)
	}
}

func TestToBrowserConfigResponseEmptyFlags(t *testing.T) {
	cfg := config.BrowserConfig{
		CDPPort:    9222,
		ExtraFlags: []string{},
	}

	resp := toBrowserConfigResponse(cfg)

	if resp.ExtraFlags == nil {
		t.Error("expected non-nil ExtraFlags")
	}
}
