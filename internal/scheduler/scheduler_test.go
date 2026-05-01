package scheduler

import (
	"strings"
	"testing"
	"time"

	"stargrazer/internal/browser"
	"stargrazer/internal/db/dbtest"
	"stargrazer/internal/social"
)

// newTestScheduler creates a fresh Scheduler backed by an in-memory SQLite
// repo, bypassing the singleton.
func newTestScheduler(t *testing.T) *Scheduler {
	t.Helper()

	// Reset browser singleton for test.
	browser.GetInstance()

	s := &Scheduler{
		jobs:     make(map[string]*Job),
		repo:     NewSQLiteRepo(dbtest.NewMemDB(t)),
		browser:  browser.GetInstance(),
		sessions: social.NewSQLiteSessionRepo(dbtest.NewMemDB(t)),
	}
	return s
}

func TestGetInstanceReturnsSingleton(t *testing.T) {
	// Reset singleton for test isolation.
	resetSchedulerForTest()
	defer resetSchedulerForTest()

	b := browser.GetInstance()
	ss := social.NewSQLiteSessionRepo(dbtest.NewMemDB(t))
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))

	s1 := GetInstance(b, ss, repo)
	s2 := GetInstance(b, ss, repo)
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
	mem := dbtest.NewMemDB(t)
	repo := NewSQLiteRepo(mem)

	s1 := &Scheduler{
		jobs:    make(map[string]*Job),
		repo:    repo,
		browser: browser.GetInstance(),
	}

	s1.Create(Job{Name: "Persist Test", Type: JobTypeKeepAlive, CronExpr: "0 */12 * * *"})

	// Verify the row was written by listing through the repo directly.
	got, err := repo.List()
	if err != nil {
		t.Fatalf("repo.List: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 persisted job, got %d", len(got))
	}

	// Create a fresh scheduler bound to the same DB and load through it.
	s2 := &Scheduler{
		jobs:    make(map[string]*Job),
		repo:    NewSQLiteRepo(mem),
		browser: browser.GetInstance(),
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

// --- Start / Stop tests ---

func TestStartAndStop(t *testing.T) {
	s := newTestScheduler(t)

	s.Start()
	s.mu.Lock()
	running := s.running
	s.mu.Unlock()
	if !running {
		t.Error("expected running=true after Start()")
	}

	s.Stop()
	s.mu.Lock()
	running = s.running
	s.mu.Unlock()
	if running {
		t.Error("expected running=false after Stop()")
	}
}

func TestStopWithNilCronRunner(t *testing.T) {
	s := newTestScheduler(t)
	// Stop without Start — cronRunner is nil
	s.Stop()
	if s.running {
		t.Error("expected running=false after Stop()")
	}
}

func TestStartLoadsPersistedJobs(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))

	// Pre-persist a job through the repo directly.
	if err := repo.Save(&Job{
		ID:        "persisted-1",
		Name:      "Persisted Job",
		Type:      JobTypeKeepAlive,
		CronExpr:  "0 */12 * * *",
		Status:    JobStatusActive,
		CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("seed save: %v", err)
	}

	s := &Scheduler{
		jobs:    make(map[string]*Job),
		repo:    repo,
		browser: browser.GetInstance(),
	}

	s.Start()
	defer s.Stop()

	if len(s.jobs) != 1 {
		t.Fatalf("expected 1 loaded job, got %d", len(s.jobs))
	}

	j := s.jobs["persisted-1"]
	if j == nil {
		t.Fatal("expected persisted job to be loaded")
	}
	if j.Name != "Persisted Job" {
		t.Errorf("expected name 'Persisted Job', got %q", j.Name)
	}
}

func TestStartRegistersActiveJobs(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))

	// Pre-persist an active and a paused job through the repo.
	now := time.Now()
	if err := repo.Save(&Job{
		ID: "active-1", Name: "Active", Type: JobTypeKeepAlive,
		CronExpr: "0 */12 * * *", Status: JobStatusActive, CreatedAt: now,
	}); err != nil {
		t.Fatalf("seed active: %v", err)
	}
	if err := repo.Save(&Job{
		ID: "paused-1", Name: "Paused", Type: JobTypeKeepAlive,
		CronExpr: "0 */6 * * *", Status: JobStatusPaused, CreatedAt: now,
	}); err != nil {
		t.Fatalf("seed paused: %v", err)
	}

	s := &Scheduler{
		jobs:    make(map[string]*Job),
		repo:    repo,
		browser: browser.GetInstance(),
	}

	s.Start()
	defer s.Stop()

	// Active job should have a non-zero cronEntryID
	activeJob := s.jobs["active-1"]
	if activeJob.cronEntryID == 0 {
		t.Error("expected active job to be registered with cron runner")
	}

	// Paused job should NOT be registered
	pausedJob := s.jobs["paused-1"]
	if pausedJob.cronEntryID != 0 {
		t.Error("expected paused job to not be registered with cron runner")
	}
}

// --- computeKeepAliveCron tests ---

func TestComputeKeepAliveCronDefaultFallback(t *testing.T) {
	s := newTestScheduler(t)
	// No cookies — should fall back to config default
	cronExpr := s.computeKeepAliveCron(nil)
	if cronExpr == "" {
		t.Error("expected non-empty cron expression")
	}
}

func TestComputeKeepAliveCronEmptyCookies(t *testing.T) {
	s := newTestScheduler(t)
	cronExpr := s.computeKeepAliveCron([]browser.CDPCookie{})
	if cronExpr == "" {
		t.Error("expected non-empty cron expression for empty cookies")
	}
}

func TestComputeKeepAliveCronFromCookies(t *testing.T) {
	s := newTestScheduler(t)
	// Use a large enough future time that rounding doesn't affect the hour count.
	cookies := []browser.CDPCookie{
		{Expires: float64(time.Now().Add(49 * time.Hour).Unix())},
	}
	cronExpr := s.computeKeepAliveCron(cookies)
	// 75% of ~49h ≈ 36h, clamped to [12h, 7d]
	// Accept 36 or 37 due to time.Now() drift
	if cronExpr != "0 */36 * * *" && cronExpr != "0 */37 * * *" {
		t.Errorf("expected '0 */36 * * *' or '0 */37 * * *', got %q", cronExpr)
	}
}

func TestComputeKeepAliveCronClampsMinimum(t *testing.T) {
	s := newTestScheduler(t)
	cookies := []browser.CDPCookie{
		{Expires: float64(time.Now().Add(4 * time.Hour).Unix())},
	}
	cronExpr := s.computeKeepAliveCron(cookies)
	// 75% of 4h = 3h, clamped to min 12h
	if cronExpr != "0 */12 * * *" {
		t.Errorf("expected '0 */12 * * *', got %q", cronExpr)
	}
}

func TestComputeKeepAliveCronClampsMaximum(t *testing.T) {
	s := newTestScheduler(t)
	cookies := []browser.CDPCookie{
		{Expires: float64(time.Now().Add(30 * 24 * time.Hour).Unix())},
	}
	cronExpr := s.computeKeepAliveCron(cookies)
	// 75% of 30d = 22.5d, clamped to max 7d = 168h
	if cronExpr != "0 */168 * * *" {
		t.Errorf("expected '0 */168 * * *', got %q", cronExpr)
	}
}

func TestComputeKeepAliveCronSkipsExpired(t *testing.T) {
	s := newTestScheduler(t)
	cookies := []browser.CDPCookie{
		{Expires: float64(time.Now().Add(-1 * time.Hour).Unix())},
	}
	// All expired — should fall back to config default
	cronExpr := s.computeKeepAliveCron(cookies)
	if cronExpr == "" {
		t.Error("expected non-empty cron expression for expired cookies")
	}
}

func TestComputeKeepAliveCronSkipsZeroExpiry(t *testing.T) {
	s := newTestScheduler(t)
	cookies := []browser.CDPCookie{
		{Expires: 0},
		{Expires: float64(time.Now().Add(49 * time.Hour).Unix())},
	}
	cronExpr := s.computeKeepAliveCron(cookies)
	// Should use the valid ~49h cookie, ignoring the 0-expiry one
	if cronExpr != "0 */36 * * *" && cronExpr != "0 */37 * * *" {
		t.Errorf("expected '0 */36 * * *' or '0 */37 * * *', got %q", cronExpr)
	}
}

func TestComputeKeepAliveCronUsesShortestExpiry(t *testing.T) {
	s := newTestScheduler(t)
	cookies := []browser.CDPCookie{
		{Expires: float64(time.Now().Add(96 * time.Hour).Unix())},
		{Expires: float64(time.Now().Add(49 * time.Hour).Unix())},
		{Expires: float64(time.Now().Add(72 * time.Hour).Unix())},
	}
	cronExpr := s.computeKeepAliveCron(cookies)
	// Shortest is ~49h, 75% ≈ 36-37h
	if cronExpr != "0 */36 * * *" && cronExpr != "0 */37 * * *" {
		t.Errorf("expected '0 */36 * * *' or '0 */37 * * *', got %q", cronExpr)
	}
}

func TestComputeKeepAliveCronSkipsNegativeExpiry(t *testing.T) {
	s := newTestScheduler(t)
	cookies := []browser.CDPCookie{
		{Expires: -1},
	}
	cronExpr := s.computeKeepAliveCron(cookies)
	// Negative expiry is skipped, falls back to default
	if cronExpr == "" {
		t.Error("expected non-empty cron for negative expiry")
	}
}

// --- execute tests ---

func TestExecuteKeepAlive(t *testing.T) {
	s := newTestScheduler(t)
	s.Start()
	defer s.Stop()

	job := &Job{
		Name:      "KA Test",
		Type:      JobTypeKeepAlive,
		Platforms: []string{"instagram"},
		CronExpr:  "0 */12 * * *",
		Status:    JobStatusActive,
	}

	s.executeKeepAlive(job)
	// Browser is not running. The keepalive auto-starts it, but no Chromium is
	// present in the test environment, so we expect either an auto-start error
	// or (if cookies are on disk from a real app run) a 'no stored cookies' message.
	if !strings.Contains(job.LastResult, "browser not running") &&
		!strings.Contains(job.LastResult, "no stored cookies") &&
		!strings.Contains(job.LastResult, "auto-start browser failed") {
		t.Errorf("unexpected result, got %q", job.LastResult)
	}
}

func TestExecuteKeepAliveMissingPlatform(t *testing.T) {
	s := newTestScheduler(t)

	job := &Job{
		Name:      "KA Missing",
		Type:      JobTypeKeepAlive,
		Platforms: []string{"nonexistent_platform_xyz"},
	}

	s.executeKeepAlive(job)
	if !strings.Contains(job.LastResult, "not found") {
		t.Errorf("expected 'not found' in result, got %q", job.LastResult)
	}
}

func TestExecuteKeepAliveNoCookies(t *testing.T) {
	s := newTestScheduler(t)

	job := &Job{
		Name:      "KA No Cookies",
		Type:      JobTypeKeepAlive,
		Platforms: []string{"instagram"},
	}

	s.executeKeepAlive(job)
	// Depending on whether cookies exist on disk from real app use,
	// we get "no stored cookies" or "browser not running", or if Chromium
	// is missing we get an auto-start failure message.
	if !strings.Contains(job.LastResult, "no stored cookies") &&
		!strings.Contains(job.LastResult, "browser not running") &&
		!strings.Contains(job.LastResult, "auto-start browser failed") {
		t.Errorf("unexpected result, got %q", job.LastResult)
	}
}

func TestExecuteKeepAliveMultiplePlatforms(t *testing.T) {
	s := newTestScheduler(t)

	job := &Job{
		Name:      "KA Multi",
		Type:      JobTypeKeepAlive,
		Platforms: []string{"instagram", "nonexistent_xyz", "facebook"},
	}

	s.executeKeepAlive(job)
	// Should contain results for each platform separated by "; " OR a single
	// auto-start failure message when no Chromium binary is available.
	if !strings.Contains(job.LastResult, "auto-start browser failed") {
		parts := strings.Split(job.LastResult, "; ")
		if len(parts) < 2 {
			t.Errorf("expected multiple results, got %q", job.LastResult)
		}
	}
}

func TestExecuteUploadNoConfig(t *testing.T) {
	s := newTestScheduler(t)

	job := &Job{
		Name: "Upload No Config",
		Type: JobTypeUpload,
	}

	s.executeUpload(job)
	if job.LastResult != "no upload config" {
		t.Errorf("expected 'no upload config', got %q", job.LastResult)
	}
}

func TestExecuteUploadWithConfig(t *testing.T) {
	s := newTestScheduler(t)

	job := &Job{
		Name: "Upload With Config",
		Type: JobTypeUpload,
		UploadConfig: &UploadConfig{
			FilePath: "/tmp/test.mp4",
			Caption:  "test caption",
			Hashtags: []string{"#test"},
		},
		Platforms: []string{"instagram"},
	}

	s.executeUpload(job)
	// Either the upload is scheduled (stub path), or browser auto-start failed
	// because no Chromium binary is present in the CI/test environment.
	if !strings.Contains(job.LastResult, "upload scheduled") &&
		!strings.Contains(job.LastResult, "auto-start browser failed") {
		t.Errorf("unexpected result, got %q", job.LastResult)
	}
	if strings.Contains(job.LastResult, "upload scheduled") {
		if !strings.Contains(job.LastResult, "/tmp/test.mp4") {
			t.Errorf("expected file path in result, got %q", job.LastResult)
		}
	}
}

func TestExecuteDispatchesKeepAlive(t *testing.T) {
	s := newTestScheduler(t)
	s.Start()
	defer s.Stop()

	job := s.Create(Job{
		Name:      "Exec KA",
		Type:      JobTypeKeepAlive,
		Platforms: []string{"instagram"},
		CronExpr:  "0 */12 * * *",
	})

	oldRunCount := job.RunCount
	s.execute(job)

	if job.RunCount != oldRunCount+1 {
		t.Errorf("expected RunCount %d, got %d", oldRunCount+1, job.RunCount)
	}
	if job.LastRun.IsZero() {
		t.Error("expected non-zero LastRun after execute")
	}
}

func TestExecuteDispatchesUpload(t *testing.T) {
	s := newTestScheduler(t)
	s.Start()
	defer s.Stop()

	job := s.Create(Job{
		Name:     "Exec Upload",
		Type:     JobTypeUpload,
		CronExpr: "0 */12 * * *",
	})

	s.execute(job)

	if job.RunCount != 1 {
		t.Errorf("expected RunCount 1, got %d", job.RunCount)
	}
	// No upload config, should have "no upload config"
	if job.LastResult != "no upload config" {
		t.Errorf("expected 'no upload config', got %q", job.LastResult)
	}
}

func TestExecuteUnknownJobType(t *testing.T) {
	s := newTestScheduler(t)
	s.Start()
	defer s.Stop()

	job := s.Create(Job{
		Name:     "Unknown Type",
		Type:     JobType("unknown_type"),
		CronExpr: "0 */12 * * *",
	})

	// Should not panic
	s.execute(job)
	if job.RunCount != 1 {
		t.Errorf("expected RunCount 1, got %d", job.RunCount)
	}
}

// --- Load with empty repo ---

func TestLoadEmptyRepo(t *testing.T) {
	s := &Scheduler{
		jobs:    make(map[string]*Job),
		repo:    NewSQLiteRepo(dbtest.NewMemDB(t)),
		browser: browser.GetInstance(),
	}
	s.load()
	if len(s.jobs) != 0 {
		t.Errorf("expected 0 jobs from empty repo, got %d", len(s.jobs))
	}
}

// --- registerJob with invalid cron ---

func TestRegisterJobWithInvalidCron(t *testing.T) {
	s := newTestScheduler(t)
	s.Start()
	defer s.Stop()

	job := s.Create(Job{
		Name:     "Bad Cron",
		Type:     JobTypeKeepAlive,
		CronExpr: "invalid cron expression !!",
	})

	if job.Status != JobStatusFailed {
		t.Errorf("expected status failed for invalid cron, got %s", job.Status)
	}
	if !strings.Contains(job.LastResult, "invalid cron") {
		t.Errorf("expected 'invalid cron' in LastResult, got %q", job.LastResult)
	}
}

// --- registerJob with nil cronRunner ---

func TestRegisterJobNilCronRunner(t *testing.T) {
	s := newTestScheduler(t)
	// Don't call Start(), so cronRunner is nil

	job := &Job{
		Name:     "No Runner",
		Type:     JobTypeKeepAlive,
		CronExpr: "0 */12 * * *",
		Status:   JobStatusActive,
	}

	// Should not panic
	s.registerJob(job)
	if job.cronEntryID != 0 {
		t.Error("expected cronEntryID 0 when cronRunner is nil")
	}
}

func TestUnregisterJobNilCronRunner(t *testing.T) {
	s := newTestScheduler(t)

	job := &Job{
		Name:     "No Runner",
		Type:     JobTypeKeepAlive,
		CronExpr: "0 */12 * * *",
	}

	// Should not panic
	s.unregisterJob(job)
}

// --- Persist tests ---

func TestPersistJobWritesThroughRepo(t *testing.T) {
	repo := NewSQLiteRepo(dbtest.NewMemDB(t))
	s := &Scheduler{
		jobs:    make(map[string]*Job),
		repo:    repo,
		browser: browser.GetInstance(),
	}

	j := &Job{ID: "test", Name: "Persist Job Test", Type: JobTypeKeepAlive,
		CronExpr: "0 */12 * * *", Status: JobStatusActive, CreatedAt: time.Now()}
	s.jobs[j.ID] = j
	s.persistJob(j)

	got, err := repo.Get("test")
	if err != nil {
		t.Fatalf("repo.Get: %v", err)
	}
	if got.Name != "Persist Job Test" {
		t.Errorf("expected name 'Persist Job Test', got %q", got.Name)
	}
}
