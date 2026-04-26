// Package automation manages user-defined browser automation workflows per platform.
package automation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Action is the type of a single automation step.
type Action string

const (
	ActionNavigate   Action = "navigate"
	ActionClick      Action = "click"
	ActionType       Action = "type"
	ActionWait       Action = "wait"
	ActionEvaluate   Action = "evaluate"
	ActionScroll     Action = "scroll"
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

// Store persists automation configs to JSON files, one per platform.
// It is safe for concurrent use.
type Store struct {
	mu      sync.RWMutex
	dataDir string
}

// NewStore creates a Store that saves files under dataDir/automations/.
func NewStore(dataDir string) *Store {
	return &Store{dataDir: dataDir}
}

// List returns all configs saved for the given platform.
func (s *Store) List(platformID string) ([]Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.load(platformID)
}

// Save creates or replaces a config. An empty ID causes a new UUID to be assigned.
func (s *Store) Save(cfg Config) (Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	configs, err := s.load(cfg.PlatformID)
	if err != nil {
		return Config{}, err
	}

	if cfg.ID == "" {
		cfg.ID = uuid.NewString()
		cfg.CreatedAt = time.Now()
	}
	if cfg.Steps == nil {
		cfg.Steps = []Step{}
	}

	replaced := false
	for i, c := range configs {
		if c.ID == cfg.ID {
			configs[i] = cfg
			replaced = true
			break
		}
	}
	if !replaced {
		configs = append(configs, cfg)
	}

	return cfg, s.persist(cfg.PlatformID, configs)
}

// Delete removes a config by ID. Returns false when the ID was not found.
func (s *Store) Delete(platformID, id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	configs, err := s.load(platformID)
	if err != nil {
		return false, err
	}

	next := configs[:0]
	found := false
	for _, c := range configs {
		if c.ID == id {
			found = true
		} else {
			next = append(next, c)
		}
	}
	if !found {
		return false, nil
	}
	return true, s.persist(platformID, next)
}

// RecordRun increments RunCount and sets LastRun to now for the given automation.
func (s *Store) RecordRun(platformID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	configs, err := s.load(platformID)
	if err != nil {
		return err
	}
	for i, c := range configs {
		if c.ID == id {
			configs[i].LastRun = time.Now()
			configs[i].RunCount++
			return s.persist(platformID, configs)
		}
	}
	return fmt.Errorf("automation %q not found for platform %q", id, platformID)
}

// --- private helpers ---

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

func (s *Store) persist(platformID string, configs []Config) error {
	fp := s.filePath(platformID)
	if err := os.MkdirAll(filepath.Dir(fp), 0700); err != nil {
		return fmt.Errorf("creating automations dir: %w", err)
	}
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding automations: %w", err)
	}
	if err := os.WriteFile(fp, data, 0600); err != nil {
		return fmt.Errorf("writing automations file: %w", err)
	}
	return nil
}
