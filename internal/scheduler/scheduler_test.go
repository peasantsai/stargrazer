package scheduler

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"stargrazer/internal/browser"
	"stargrazer/internal/social"
)

// newTestScheduler creates a fresh Scheduler with a temp file for persistence,
// bypassing the singleton. It also starts the cron runner.
func newTestScheduler(t *testing.T) *Scheduler {
	t.Helper()
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "schedules.json")

	// Reset browser singleton for test.
	browser.GetInstance()

	s := &Scheduler{
		jobs:     make(map[string]*Job),
		filePath: fp,
		browser:  browser.GetInstance(),
		sessions: &social.SessionStore{},
	}
	return s
}

func TestGetInstanceReturnsSingleton(t *testing.T) {
	// Reset singleton for test isolation.
	once = sync.Once{}
	instance = nil
	defer func() {
		once = sync.Once{}
		instance = nil
	}()

	b := browser.GetInstance()
	ss := social.NewSessionStore()

	s1 := GetInstance(b, ss)
	s2 := GetInstance(b, ss)
	if s1 != s2 {
		t.Error("GetInstance() returned different pointers")
	}
}

func TestCreateAssignsIDAndCreatedAt(t *testing.T) {
	s := newTestScheduler(t)

	job := s.Create(Job{
		Name:     "Test Job",
		Type:     JobTypeKeepAlive,
		CronExpr: "0 */12 * * *",
	})

	if job.ID == "" {
		t.Error("expected non-empty ID")
	}
	if job.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if job.Status != JobStatusActive {
		t.Errorf("expected status active, got %s", job.Status)
	}
	if job.Name != "Test Job" {
		t.Errorf("expected name 'Test Job', got %q", job.Name)
	}
}

func TestGetRetrievesCreatedJob(t *testing.T) {
	s := newTestScheduler(t)

	created := s.Create(Job{
		Name:     "Retrievable Job",
		Type:     JobTypeKeepAlive,
		CronExpr: "0 */6 * * *",
	})

	got := s.Get(created.ID)
	if got == nil {
		t.Fatal("Get() returned nil for existing job")
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, got.ID)
	}
	if got.Name != "Retrievable Job" {
		t.Errorf("expected name 'Retrievable Job', got %q", got.Name)
	}
}

func TestGetReturnsNilForMissing(t *testing.T) {
	s := newTestScheduler(t)

	got := s.Get("nonexistent-id")
	if got != nil {
		t.Error("expected nil for nonexistent job")
	}
}

func TestListReturnsAllJobsSortedByCreatedAt(t *testing.T) {
	s := newTestScheduler(t)

	// Create jobs with slight time gaps to ensure ordering.
	j1 := s.Create(Job{Name: "First", Type: JobTypeKeepAlive, CronExpr: "0 */12 * * *"})
	time.Sleep(10 * time.Millisecond)
	j2 := s.Create(Job{Name: "Second", Type: JobTypeKeepAlive, CronExpr: "0 */12 * * *"})
	time.Sleep(10 * time.Millisecond)
	j3 := s.Create(Job{Name: "Third", Type: JobTypeUpload, CronExpr: "0 */12 * * *"})

	list := s.List()
	if len(list) != 3 {
		t.Fatalf("expected 3 jobs, got %d", len(list))
	}

	// List is sorted by CreatedAt descending (newest first).
	if list[0].ID != j3.ID {
		t.Errorf("expected newest job first, got %q", list[0].Name)
	}
	if list[1].ID != j2.ID {
		t.Errorf("expected second newest next, got %q", list[1].Name)
	}
	if list[2].ID != j1.ID {
		t.Errorf("expected oldest last, got %q", list[2].Name)
	}
}

func TestListEmptyScheduler(t *testing.T) {
	s := newTestScheduler(t)
	list := s.List()
	if len(list) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(list))
	}
}

func TestDeleteRemovesJob(t *testing.T) {
	s := newTestScheduler(t)

	job := s.Create(Job{Name: "Deletable", Type: JobTypeKeepAlive, CronExpr: "0 */12 * * *"})

	ok := s.Delete(job.ID)
	if !ok {
		t.Error("Delete() returned false for existing job")
	}

	got := s.Get(job.ID)
	if got != nil {
		t.Error("expected nil after Delete")
	}

	list := s.List()
	if len(list) != 0 {
		t.Errorf("expected 0 jobs after delete, got %d", len(list))
	}
}

func TestDeleteReturnsFalseForMissing(t *testing.T) {
	s := newTestScheduler(t)

	ok := s.Delete("nonexistent-id")
	if ok {
		t.Error("Delete() returned true for nonexistent job")
	}
}

func TestPauseSetsStatusToPaused(t *testing.T) {
	s := newTestScheduler(t)

	job := s.Create(Job{Name: "Pausable", Type: JobTypeKeepAlive, CronExpr: "0 */12 * * *"})

	paused := s.Pause(job.ID)
	if paused == nil {
		t.Fatal("Pause() returned nil")
	}
	if paused.Status != JobStatusPaused {
		t.Errorf("expected status paused, got %s", paused.Status)
	}

	// Verify via Get too.
	got := s.Get(job.ID)
	if got.Status != JobStatusPaused {
		t.Errorf("Get after Pause: expected paused, got %s", got.Status)
	}
}

func TestPauseReturnsNilForMissing(t *testing.T) {
	s := newTestScheduler(t)

	paused := s.Pause("nonexistent")
	if paused != nil {
		t.Error("expected nil for nonexistent job")
	}
}

func TestResumeSetsStatusToActive(t *testing.T) {
	s := newTestScheduler(t)

	job := s.Create(Job{Name: "Resumable", Type: JobTypeKeepAlive, CronExpr: "0 */12 * * *"})
	s.Pause(job.ID)

	resumed := s.Resume(job.ID)
	if resumed == nil {
		t.Fatal("Resume() returned nil")
	}
	if resumed.Status != JobStatusActive {
		t.Errorf("expected status active, got %s", resumed.Status)
	}
}

func TestResumeReturnsNilForMissing(t *testing.T) {
	s := newTestScheduler(t)

	resumed := s.Resume("nonexistent")
	if resumed != nil {
		t.Error("expected nil for nonexistent job")
	}
}

func TestUpdateModifiesJobFields(t *testing.T) {
	s := newTestScheduler(t)

	job := s.Create(Job{
		Name:      "Updatable",
		Type:      JobTypeKeepAlive,
		CronExpr:  "0 */12 * * *",
		Platforms: []string{"instagram"},
	})

	updated := s.Update(job.ID, func(j *Job) {
		j.Name = "Updated Name"
		j.Platforms = []string{"instagram", "facebook"}
	})

	if updated == nil {
		t.Fatal("Update() returned nil")
	}
	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %q", updated.Name)
	}
	if len(updated.Platforms) != 2 {
		t.Errorf("expected 2 platforms, got %d", len(updated.Platforms))
	}
}

func TestUpdateCronExprReRegisters(t *testing.T) {
	s := newTestScheduler(t)

	job := s.Create(Job{
		Name:     "Cron Change",
		Type:     JobTypeKeepAlive,
		CronExpr: "0 */12 * * *",
	})

	updated := s.Update(job.ID, func(j *Job) {
		j.CronExpr = "0 */6 * * *"
	})

	if updated == nil {
		t.Fatal("Update() returned nil")
	}
	if updated.CronExpr != "0 */6 * * *" {
		t.Errorf("expected cron '0 */6 * * *', got %q", updated.CronExpr)
	}
}

func TestUpdateReturnsNilForMissing(t *testing.T) {
	s := newTestScheduler(t)

	updated := s.Update("nonexistent", func(j *Job) {
		j.Name = "nope"
	})
	if updated != nil {
		t.Error("expected nil for nonexistent job")
	}
}

func TestEnsureKeepAliveCreatesAutoJob(t *testing.T) {
	s := newTestScheduler(t)

	cookies := []browser.CDPCookie{
		{Name: "session", Expires: float64(time.Now().Add(48 * time.Hour).Unix())},
	}

	s.EnsureKeepAlive("instagram", "Instagram", cookies)

	list := s.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 job, got %d", len(list))
	}

	job := list[0]
	if job.Type != JobTypeKeepAlive {
		t.Errorf("expected type keepalive, got %s", job.Type)
	}
	if !job.Auto {
		t.Error("expected Auto true")
	}
	if len(job.Platforms) != 1 || job.Platforms[0] != "instagram" {
		t.Errorf("expected platforms [instagram], got %v", job.Platforms)
	}
	if job.Status != JobStatusActive {
		t.Errorf("expected status active, got %s", job.Status)
	}
	if job.Name == "" {
		t.Error("expected non-empty name")
	}
}

func TestEnsureKeepAliveDoesNotDuplicate(t *testing.T) {
	s := newTestScheduler(t)

	cookies := []browser.CDPCookie{
		{Name: "session", Expires: float64(time.Now().Add(48 * time.Hour).Unix())},
	}

	s.EnsureKeepAlive("instagram", "Instagram", cookies)
	s.EnsureKeepAlive("instagram", "Instagram", cookies)
	s.EnsureKeepAlive("instagram", "Instagram", cookies)

	list := s.List()
	if len(list) != 1 {
		t.Errorf("expected 1 job (no duplicates), got %d", len(list))
	}
}

func TestEnsureKeepAliveCreatesSeparateForDifferentPlatforms(t *testing.T) {
	s := newTestScheduler(t)

	cookies := []browser.CDPCookie{
		{Name: "session", Expires: float64(time.Now().Add(48 * time.Hour).Unix())},
	}

	s.EnsureKeepAlive("instagram", "Instagram", cookies)
	s.EnsureKeepAlive("facebook", "Facebook", cookies)

	list := s.List()
	if len(list) != 2 {
		t.Errorf("expected 2 jobs for different platforms, got %d", len(list))
	}
}

func TestJobTypeConstants(t *testing.T) {
	if string(JobTypeKeepAlive) != "session_keepalive" {
		t.Errorf("expected 'session_keepalive', got %q", JobTypeKeepAlive)
	}
	if string(JobTypeUpload) != "upload" {
		t.Errorf("expected 'upload', got %q", JobTypeUpload)
	}
}

func TestJobStatusConstants(t *testing.T) {
	if string(JobStatusActive) != "active" {
		t.Errorf("expected 'active', got %q", JobStatusActive)
	}
	if string(JobStatusPaused) != "paused" {
		t.Errorf("expected 'paused', got %q", JobStatusPaused)
	}
	if string(JobStatusFailed) != "failed" {
		t.Errorf("expected 'failed', got %q", JobStatusFailed)
	}
}

func TestPersistAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "schedules.json")

	s1 := &Scheduler{
		jobs:     make(map[string]*Job),
		filePath: fp,
		browser:  browser.GetInstance(),
	}

	s1.Create(Job{Name: "Persist Test", Type: JobTypeKeepAlive, CronExpr: "0 */12 * * *"})

	// Verify file was written.
	if _, err := os.Stat(fp); err != nil {
		t.Fatalf("schedules file not created: %v", err)
	}

	// Create a new scheduler pointing to the same file and load.
	s2 := &Scheduler{
		jobs:     make(map[string]*Job),
		filePath: fp,
		browser:  browser.GetInstance(),
	}
	s2.load()

	if len(s2.jobs) != 1 {
		t.Fatalf("expected 1 job after load, got %d", len(s2.jobs))
	}

	for _, j := range s2.jobs {
		if j.Name != "Persist Test" {
			t.Errorf("expected name 'Persist Test', got %q", j.Name)
		}
	}
}

func TestCreateWithUploadConfig(t *testing.T) {
	s := newTestScheduler(t)

	job := s.Create(Job{
		Name:      "Upload Job",
		Type:      JobTypeUpload,
		Platforms: []string{"instagram", "tiktok"},
		CronExpr:  "0 9 * * 1",
		UploadConfig: &UploadConfig{
			FilePath: "/tmp/video.mp4",
			Caption:  "My upload",
			Hashtags: []string{"#test"},
		},
	})

	if job.UploadConfig == nil {
		t.Fatal("expected non-nil UploadConfig")
	}
	if job.UploadConfig.FilePath != "/tmp/video.mp4" {
		t.Errorf("expected filePath '/tmp/video.mp4', got %q", job.UploadConfig.FilePath)
	}
	if job.UploadConfig.Caption != "My upload" {
		t.Errorf("expected caption 'My upload', got %q", job.UploadConfig.Caption)
	}
}
