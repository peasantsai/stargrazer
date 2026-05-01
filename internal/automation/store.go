// Package automation persists user-defined automation workflows.
//
// As of DSG-001-P2 the production read/write path lives in SQLiteRepo
// (sqlite_repo.go). The Store type below is retained read-only as the source
// for the one-shot backfill orchestrator and is removed in a follow-up
// release.
package automation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Step is one command in an automation sequence.
type Step struct {
	Action Action `json:"action"`
	// Target is a CSS selector for click/type/scroll, or a URL for navigate.
	Target string `json:"target"`
	// Value is text to type, JS to evaluate, or milliseconds to wait.
	Value string `json:"value"`
	// Label is a human-readable description shown in the UI.
	Label string `json:"label"`
	// Selectors holds Chrome Recorder-style fallback selector groups.
	// Each outer element is an alternative strategy (CSS, XPath, ARIA, text, pierce).
	// The execution engine tries each until one succeeds.
	// When empty, Target is used as the sole selector.
	Selectors [][]string `json:"selectors,omitempty"`
}

// Config is a named, platform-scoped automation workflow.
type Config struct {
	ID          string    `json:"id"`
	PlatformID  string    `json:"platformId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Steps       []Step    `json:"steps"`
	CreatedAt   time.Time `json:"createdAt"`
	LastRun     time.Time `json:"lastRun,omitempty"`
	RunCount    int       `json:"runCount"`
}

// Store reads automations from per-platform JSON files. Read-only as of
// DSG-001-P2.
type Store struct {
	mu      sync.RWMutex
	dataDir string
}

// NewStore returns a Store that reads from dataDir/automations/<platformID>.json.
func NewStore(dataDir string) *Store { return &Store{dataDir: dataDir} }

// List returns all configs saved for the given platform.
func (s *Store) List(platformID string) ([]Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.load(platformID)
}

func (s *Store) filePath(platformID string) string {
	return filepath.Join(s.dataDir, "automations", platformID+".json")
}

func (s *Store) load(platformID string) ([]Config, error) {
	data, err := os.ReadFile(s.filePath(platformID))
	if os.IsNotExist(err) {
		return []Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading automations for %q: %w", platformID, err)
	}
	var configs []Config
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("parsing automations for %q: %w", platformID, err)
	}
	return configs, nil
}
