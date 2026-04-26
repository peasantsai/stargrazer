package automation

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)
	if s == nil {
		t.Fatal("expected non-nil Store")
	}
	if s.dataDir != tmpDir {
		t.Errorf("expected dataDir %q, got %q", tmpDir, s.dataDir)
	}
}

func TestListEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	configs, err := s.List("instagram")
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("expected 0 configs, got %d", len(configs))
	}
}

func TestSaveNewConfigAssignsID(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	cfg := Config{
		PlatformID:  "instagram",
		Name:        "My Automation",
		Description: "Test automation",
		Steps: []Step{
			{Action: ActionNavigate, Target: "https://instagram.com", Label: "Open Instagram"},
		},
	}

	saved, err := s.Save(cfg)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if saved.ID == "" {
		t.Error("expected non-empty ID")
	}
	if saved.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if saved.Name != "My Automation" {
		t.Errorf("expected Name 'My Automation', got %q", saved.Name)
	}
}

func TestSaveNilStepsBecomesEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	cfg := Config{
		PlatformID: "facebook",
		Name:       "Nil Steps",
		Steps:      nil,
	}

	saved, err := s.Save(cfg)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if saved.Steps == nil {
		t.Error("expected Steps to be non-nil empty slice, got nil")
	}
	if len(saved.Steps) != 0 {
		t.Errorf("expected 0 steps, got %d", len(saved.Steps))
	}
}

func TestSaveWithPreassignedIDPreservesIt(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	cfg := Config{
		ID:         "custom-id-123",
		PlatformID: "linkedin",
		Name:       "Custom ID",
		CreatedAt:  time.Now(),
	}

	saved, err := s.Save(cfg)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if saved.ID != "custom-id-123" {
		t.Errorf("expected ID 'custom-id-123', got %q", saved.ID)
	}
}

func TestSaveUpdateReplacesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	cfg, _ := s.Save(Config{PlatformID: "youtube", Name: "Original"})
	cfg.Name = "Updated"
	updated, err := s.Save(cfg)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if updated.ID != cfg.ID {
		t.Errorf("expected same ID %q, got %q", cfg.ID, updated.ID)
	}
	if updated.Name != "Updated" {
		t.Errorf("expected 'Updated', got %q", updated.Name)
	}

	configs, _ := s.List("youtube")
	if len(configs) != 1 {
		t.Errorf("expected 1 config after update (not 2), got %d", len(configs))
	}
	if configs[0].Name != "Updated" {
		t.Errorf("expected 'Updated' in list, got %q", configs[0].Name)
	}
}

func TestSaveMultipleConfigs(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	s.Save(Config{PlatformID: "linkedin", Name: "First"})
	s.Save(Config{PlatformID: "linkedin", Name: "Second"})
	s.Save(Config{PlatformID: "linkedin", Name: "Third"})

	configs, err := s.List("linkedin")
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(configs) != 3 {
		t.Errorf("expected 3 configs, got %d", len(configs))
	}
}

func TestDeleteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	cfg, _ := s.Save(Config{PlatformID: "x", Name: "To Delete"})

	found, err := s.Delete("x", cfg.ID)
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if !found {
		t.Error("expected Delete() to return true for existing config")
	}

	configs, _ := s.List("x")
	if len(configs) != 0 {
		t.Errorf("expected 0 configs after delete, got %d", len(configs))
	}
}

func TestDeleteNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	found, err := s.Delete("instagram", "nonexistent-id")
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if found {
		t.Error("expected Delete() to return false for nonexistent ID")
	}
}

func TestDeleteOneOfMany(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	_, _ = s.Save(Config{PlatformID: "instagram", Name: "Keep 1"})
	c2, _ := s.Save(Config{PlatformID: "instagram", Name: "Delete Me"})
	_, _ = s.Save(Config{PlatformID: "instagram", Name: "Keep 2"})

	s.Delete("instagram", c2.ID)

	configs, _ := s.List("instagram")
	if len(configs) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(configs))
	}
	for _, c := range configs {
		if c.ID == c2.ID {
			t.Error("deleted config still in list")
		}
	}
}

func TestDeletePersistsToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	cfg, _ := s.Save(Config{PlatformID: "x", Name: "Persist Delete"})
	s.Delete("x", cfg.ID)

	// New store from the same directory should see 0 configs.
	s2 := NewStore(tmpDir)
	configs, _ := s2.List("x")
	if len(configs) != 0 {
		t.Errorf("expected 0 configs after delete+reload, got %d", len(configs))
	}
}

func TestRecordRunUpdatesStats(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	cfg, _ := s.Save(Config{PlatformID: "facebook", Name: "Runnable"})

	err := s.RecordRun("facebook", cfg.ID)
	if err != nil {
		t.Fatalf("RecordRun() error: %v", err)
	}

	configs, _ := s.List("facebook")
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].RunCount != 1 {
		t.Errorf("expected RunCount 1, got %d", configs[0].RunCount)
	}
	if configs[0].LastRun.IsZero() {
		t.Error("expected non-zero LastRun")
	}
}

func TestRecordRunIncrementsCount(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	cfg, _ := s.Save(Config{PlatformID: "tiktok", Name: "Multiple Runs"})

	s.RecordRun("tiktok", cfg.ID)
	s.RecordRun("tiktok", cfg.ID)
	s.RecordRun("tiktok", cfg.ID)

	configs, _ := s.List("tiktok")
	if configs[0].RunCount != 3 {
		t.Errorf("expected RunCount 3, got %d", configs[0].RunCount)
	}
}

func TestRecordRunNonExistentReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	err := s.RecordRun("instagram", "nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent automation ID")
	}
}

func TestRecordRunPersistsToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	cfg, _ := s.Save(Config{PlatformID: "facebook", Name: "Persistent Run"})
	s.RecordRun("facebook", cfg.ID)

	s2 := NewStore(tmpDir)
	configs, _ := s2.List("facebook")
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].RunCount != 1 {
		t.Errorf("expected RunCount 1 after reload, got %d", configs[0].RunCount)
	}
	if configs[0].LastRun.IsZero() {
		t.Error("expected non-zero LastRun after reload")
	}
}

func TestListInvalidJSONReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	dir := filepath.Join(tmpDir, "automations")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "bad.json"), []byte("not json {["), 0600)

	_, err := s.List("bad")
	if err == nil {
		t.Error("expected error for invalid JSON automations file")
	}
}

func TestFilePathReturnsCorrectPath(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	fp := s.filePath("instagram")
	expected := filepath.Join(tmpDir, "automations", "instagram.json")
	if fp != expected {
		t.Errorf("expected %q, got %q", expected, fp)
	}
}

func TestListIsolatedByPlatform(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	s.Save(Config{PlatformID: "instagram", Name: "IG Config"})
	s.Save(Config{PlatformID: "facebook", Name: "FB Config"})

	igConfigs, _ := s.List("instagram")
	fbConfigs, _ := s.List("facebook")

	if len(igConfigs) != 1 {
		t.Errorf("expected 1 instagram config, got %d", len(igConfigs))
	}
	if len(fbConfigs) != 1 {
		t.Errorf("expected 1 facebook config, got %d", len(fbConfigs))
	}
	if igConfigs[0].Name != "IG Config" {
		t.Errorf("expected 'IG Config', got %q", igConfigs[0].Name)
	}
}

func TestSaveRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	original := Config{
		PlatformID:  "youtube",
		Name:        "Round Trip",
		Description: "Test description",
		Steps: []Step{
			{Action: ActionNavigate, Target: "https://youtube.com", Label: "Go to YouTube"},
			{Action: ActionClick, Target: "#upload", Label: "Click Upload"},
			{Action: ActionType, Target: "#title", Value: "My Video", Label: "Enter title"},
			{Action: ActionWait, Value: "2000", Label: "Wait"},
			{Action: ActionEvaluate, Value: "document.title", Label: "Get title"},
			{Action: ActionScroll, Target: "body", Label: "Scroll"},
		},
	}

	saved, _ := s.Save(original)

	configs, err := s.List("youtube")
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}

	loaded := configs[0]
	if loaded.ID != saved.ID {
		t.Errorf("ID mismatch: %q vs %q", loaded.ID, saved.ID)
	}
	if loaded.Name != "Round Trip" {
		t.Errorf("Name mismatch: got %q", loaded.Name)
	}
	if len(loaded.Steps) != 6 {
		t.Errorf("expected 6 steps, got %d", len(loaded.Steps))
	}
	if loaded.Steps[0].Action != ActionNavigate {
		t.Errorf("expected ActionNavigate, got %s", loaded.Steps[0].Action)
	}
}

func TestActionConstants(t *testing.T) {
	tests := []struct {
		action   Action
		expected string
	}{
		{ActionNavigate, "navigate"},
		{ActionClick, "click"},
		{ActionType, "type"},
		{ActionWait, "wait"},
		{ActionEvaluate, "evaluate"},
		{ActionScroll, "scroll"},
	}
	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.action) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.action))
			}
		})
	}
}

func TestConcurrentSave(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			cfg := Config{
				PlatformID: "instagram",
				Name:       fmt.Sprintf("Config %d", n),
				Steps:      []Step{{Action: ActionClick, Target: "#btn"}},
			}
			_, err := s.Save(cfg)
			if err != nil {
				t.Errorf("concurrent Save() error: %v", err)
			}
		}(i)
	}
	wg.Wait()

	configs, err := s.List("instagram")
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(configs) != goroutines {
		t.Errorf("expected %d configs after concurrent save, got %d", goroutines, len(configs))
	}
}

func TestConcurrentListAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	// Pre-populate some configs.
	for i := 0; i < 5; i++ {
		s.Save(Config{PlatformID: "tiktok", Name: fmt.Sprintf("Init %d", i)})
	}

	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 10; i++ {
		go func(n int) {
			defer wg.Done()
			s.List("tiktok")
		}(i)
		go func(n int) {
			defer wg.Done()
			s.Save(Config{PlatformID: "tiktok", Name: fmt.Sprintf("Concurrent %d", n)})
		}(i)
	}
	wg.Wait()
}

func TestPersistCreatesNestedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	// automations/ directory does not exist yet; Save should create it.
	_, err := s.Save(Config{PlatformID: "instagram", Name: "Dir Test"})
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	automationsDir := filepath.Join(tmpDir, "automations")
	if _, err := os.Stat(automationsDir); err != nil {
		t.Fatalf("automations directory not created: %v", err)
	}

	fp := filepath.Join(automationsDir, "instagram.json")
	if _, err := os.Stat(fp); err != nil {
		t.Fatalf("automation file not created: %v", err)
	}
}

func TestDeleteOnEmptyPlatformReturnsFalse(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	// No configs for "youtube" at all.
	found, err := s.Delete("youtube", "any-id")
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if found {
		t.Error("expected false for delete on empty platform")
	}
}

func TestRecordRunOnEmptyPlatformReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	err := s.RecordRun("youtube", "any-id")
	if err == nil {
		t.Error("expected error when platform has no configs")
	}
}

func TestSaveLastRunFieldPreservedOnUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewStore(tmpDir)

	cfg, _ := s.Save(Config{PlatformID: "facebook", Name: "Track Runs"})
	s.RecordRun("facebook", cfg.ID)

	configs, _ := s.List("facebook")
	runTime := configs[0].LastRun
	if runTime.IsZero() {
		t.Fatal("expected non-zero LastRun")
	}

	// Update name — LastRun should be preserved.
	configs[0].Name = "Updated Name"
	s.Save(configs[0])

	configs2, _ := s.List("facebook")
	if configs2[0].LastRun != runTime {
		t.Error("LastRun was changed during a name-only update")
	}
}
