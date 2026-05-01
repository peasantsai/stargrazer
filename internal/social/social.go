package social

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Platform identifies a social media service.
type Platform string

const (
	Facebook  Platform = "facebook"
	Instagram Platform = "instagram"
	TikTok    Platform = "tiktok"
	YouTube   Platform = "youtube"
	LinkedIn  Platform = "linkedin"
	X         Platform = "x"
)

// PlatformInfo holds static metadata about a platform.
type PlatformInfo struct {
	ID       Platform `json:"id"`
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	LoginURL string   `json:"loginUrl"`
	// Domains whose cookies indicate a valid session.
	SessionDomains []string `json:"sessionDomains"`
	// Cookie names that indicate the user is logged in.
	LoginCookies []string `json:"loginCookies"`
}

// AllPlatforms returns the supported platform definitions.
func AllPlatforms() []PlatformInfo {
	return []PlatformInfo{
		{
			ID: Facebook, Name: "Facebook", URL: "https://www.facebook.com",
			LoginURL:       "https://www.facebook.com/login",
			SessionDomains: []string{".facebook.com"},
			LoginCookies:   []string{"c_user", "xs"},
		},
		{
			ID: Instagram, Name: "Instagram", URL: "https://www.instagram.com",
			LoginURL:       "https://www.instagram.com/accounts/login/",
			SessionDomains: []string{".instagram.com"},
			LoginCookies:   []string{"sessionid", "ds_user_id"},
		},
		{
			ID: TikTok, Name: "TikTok", URL: "https://www.tiktok.com",
			LoginURL:       "https://www.tiktok.com/login",
			SessionDomains: []string{".tiktok.com"},
			LoginCookies:   []string{"sessionid", "sid_tt"},
		},
		{
			ID: YouTube, Name: "YouTube", URL: "https://www.youtube.com",
			LoginURL:       "https://accounts.google.com/ServiceLogin?service=youtube",
			SessionDomains: []string{".youtube.com", ".google.com"},
			LoginCookies:   []string{"SID", "SSID", "LOGIN_INFO"},
		},
		{
			ID: LinkedIn, Name: "LinkedIn", URL: "https://www.linkedin.com",
			LoginURL:       "https://www.linkedin.com/login",
			SessionDomains: []string{".linkedin.com"},
			LoginCookies:   []string{"li_at", "JSESSIONID"},
		},
		{
			ID: X, Name: "X", URL: "https://x.com",
			LoginURL:       "https://x.com/i/flow/login",
			SessionDomains: []string{".x.com", ".twitter.com"},
			LoginCookies:   []string{"auth_token", "ct0"},
		},
	}
}

// AccountStatus represents the session state of a single account.
type AccountStatus struct {
	PlatformID Platform  `json:"platformId"`
	LoggedIn   bool      `json:"loggedIn"`
	Username   string    `json:"username"`
	LastLogin  time.Time `json:"lastLogin"`
	LastCheck  time.Time `json:"lastCheck"`
}

// SessionStore reads per-platform login state from a JSON file. Read-only as
// of DSG-001-P2.
type SessionStore struct {
	mu       sync.RWMutex
	accounts map[Platform]*AccountStatus
	filePath string
}

// NewSessionStore creates or loads the session store from disk.
func NewSessionStore() *SessionStore {
	fp := sessionFilePath()
	s := &SessionStore{
		accounts: make(map[Platform]*AccountStatus),
		filePath: fp,
	}
	s.load()
	return s
}

// GetAll returns the status for every platform.
func (s *SessionStore) GetAll() []AccountStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	platforms := AllPlatforms()
	result := make([]AccountStatus, 0, len(platforms))
	for _, p := range platforms {
		if acct, ok := s.accounts[p.ID]; ok {
			result = append(result, *acct)
		} else {
			result = append(result, AccountStatus{PlatformID: p.ID})
		}
	}
	return result
}

// Get returns the status for a single platform.
func (s *SessionStore) Get(id Platform) AccountStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if acct, ok := s.accounts[id]; ok {
		return *acct
	}
	return AccountStatus{PlatformID: id}
}

// SessionDir returns the persistent user-data-dir for a platform's browser session.
// Kept for backward compat display purposes.
func SessionDir(id Platform) string {
	return SharedSessionDir()
}

// SharedSessionDir returns the single shared user-data-dir for all platforms.
// All social logins share one browser profile so cookies persist across platforms.
func SharedSessionDir() string {
	base := sessionsBaseDir()
	return filepath.Join(base, "browser_profile")
}

// FindPlatform looks up a platform by ID.
func FindPlatform(id Platform) *PlatformInfo {
	for _, p := range AllPlatforms() {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// SchedulesFilePath returns the path to the schedules persistence file.
func SchedulesFilePath() string {
	return filepath.Join(sessionsBaseDir(), "schedules.json")
}

// SharedSessionDirParent is the directory that holds accounts.json,
// schedules.json, and the automations/ subdir. Exported for the backfill
// orchestrator only — production code should use the repos.
func SharedSessionDirParent() string { return sessionsBaseDir() }

func sessionsBaseDir() string {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
		if base == "" {
			base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	case "darwin":
		base = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	default:
		base = os.Getenv("XDG_DATA_HOME")
		if base == "" {
			base = filepath.Join(os.Getenv("HOME"), ".local", "share")
		}
	}
	return filepath.Join(base, "stargrazer", "sessions")
}

func sessionFilePath() string {
	return filepath.Join(sessionsBaseDir(), "accounts.json")
}

func (s *SessionStore) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}
	var accounts []AccountStatus
	if err := json.Unmarshal(data, &accounts); err != nil {
		return
	}
	for i := range accounts {
		a := accounts[i]
		s.accounts[a.PlatformID] = &a
	}
}

// EnsureSessionDir creates the shared session directory if it doesn't exist.
func EnsureSessionDir(id Platform) (string, error) {
	dir := SharedSessionDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("creating session dir: %w", err)
	}
	return dir, nil
}
