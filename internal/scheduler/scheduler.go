package scheduler

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"stargrazer/internal/browser"
	"stargrazer/internal/config"
	"stargrazer/internal/logger"
	"stargrazer/internal/social"
)

// JobType identifies the kind of scheduled work.
type JobType string

const (
	JobTypeKeepAlive JobType = "session_keepalive"
	JobTypeUpload    JobType = "upload"
)

// JobStatus represents the current state of a job.
type JobStatus string

const (
	JobStatusActive JobStatus = "active"
	JobStatusPaused JobStatus = "paused"
	JobStatusFailed JobStatus = "failed"
)

// UploadConfig holds parameters for a scheduled upload job.
type UploadConfig struct {
	FilePath string   `json:"filePath"`
	Caption  string   `json:"caption"`
	Hashtags []string `json:"hashtags"`
}

// Job represents a single scheduled task.
type Job struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Type         JobType       `json:"type"`
	Platforms    []string      `json:"platforms"`
	CronExpr     string        `json:"cronExpr"`
	NextRun      time.Time     `json:"nextRun"`
	LastRun      time.Time     `json:"lastRun"`
	Status       JobStatus     `json:"status"`
	CreatedAt    time.Time     `json:"createdAt"`
	RunCount     int           `json:"runCount"`
	LastResult   string        `json:"lastResult"`
	Auto         bool          `json:"auto"`
	UploadConfig *UploadConfig `json:"uploadConfig,omitempty"`
	cronEntryID  cron.EntryID  `json:"-"`
}

// Scheduler manages cron-based jobs for session keep-alive and uploads.
type Scheduler struct {
	mu         sync.Mutex
	cronRunner *cron.Cron
	jobs       map[string]*Job
	repo       ScheduleRepo
	browser    *browser.Manager
	sessions   social.SessionRepo
	running    bool
}

var (
	instance *Scheduler
	once     sync.Once
)

// GetInstance returns the singleton scheduler. The first call initialises it
// with the provided dependencies.
func GetInstance(b *browser.Manager, s social.SessionRepo, repo ScheduleRepo) *Scheduler {
	once.Do(func() {
		instance = &Scheduler{
			jobs:     make(map[string]*Job),
			repo:     repo,
			browser:  b,
			sessions: s,
		}
	})
	return instance
}

// resetSchedulerForTest re-zeroes the singleton so tests can build a fresh
// Scheduler with their own dependencies. Lowercase: callable only from
// in-package tests.
func resetSchedulerForTest() {
	instance = nil
	once = sync.Once{}
}

// Start creates the cron runner, loads persisted jobs and begins scheduling.
func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cronRunner = cron.New()
	s.load()

	for _, j := range s.jobs {
		if j.Status == JobStatusActive {
			s.registerJob(j)
		}
	}

	s.cronRunner.Start()
	s.running = true
	logger.Info("scheduler", fmt.Sprintf("started with %d jobs loaded", len(s.jobs)))
}

// Stop halts the cron runner and marks the scheduler stopped. Job state is
// already persisted on every mutation, so no flush is required here.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cronRunner != nil {
		s.cronRunner.Stop()
	}
	s.running = false
	logger.Info("scheduler", "stopped")
}

// Create adds a new job, assigns it an ID, registers its cron entry and persists.
func (s *Scheduler) Create(j Job) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	j.ID = uuid.New().String()
	j.CreatedAt = time.Now()
	j.Status = JobStatusActive

	s.jobs[j.ID] = &j
	s.registerJob(&j)
	s.persistJob(&j)

	logger.Info("scheduler", fmt.Sprintf("created job %s (%s)", j.Name, j.ID))
	return &j
}

// Get returns a single job by ID, or nil if not found.
func (s *Scheduler) Get(id string) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.jobs[id]
}

// List returns all jobs sorted by CreatedAt descending.
func (s *Scheduler) List() []*Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]*Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		result = append(result, j)
	}
	sort.Slice(result, func(i, k int) bool {
		return result[i].CreatedAt.After(result[k].CreatedAt)
	})
	return result
}

// Update applies fn to the job identified by id. If the CronExpr changed
// the job is re-registered with the cron runner.
func (s *Scheduler) Update(id string, fn func(*Job)) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	j, ok := s.jobs[id]
	if !ok {
		return nil
	}

	oldCron := j.CronExpr
	fn(j)

	if j.CronExpr != oldCron {
		s.unregisterJob(j)
		if j.Status == JobStatusActive {
			s.registerJob(j)
		}
	}

	s.persistJob(j)
	return j
}

// Delete removes a job entirely.
func (s *Scheduler) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	j, ok := s.jobs[id]
	if !ok {
		return false
	}

	s.unregisterJob(j)
	delete(s.jobs, id)
	s.persistDelete(id)

	logger.Info("scheduler", fmt.Sprintf("deleted job %s", id))
	return true
}

// Pause stops a job's cron entry without deleting it.
func (s *Scheduler) Pause(id string) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	j, ok := s.jobs[id]
	if !ok {
		return nil
	}

	s.unregisterJob(j)
	j.Status = JobStatusPaused
	s.persistJob(j)

	logger.Info("scheduler", fmt.Sprintf("paused job %s", j.Name))
	return j
}

// Resume re-registers a paused job and marks it active.
func (s *Scheduler) Resume(id string) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	j, ok := s.jobs[id]
	if !ok {
		return nil
	}

	s.registerJob(j)
	j.Status = JobStatusActive
	s.persistJob(j)

	logger.Info("scheduler", fmt.Sprintf("resumed job %s", j.Name))
	return j
}

// EnsureKeepAlive creates an automatic keep-alive job for the given platform
// if one does not already exist. The cron interval is derived from cookie
// expiry times (75 % of the shortest non-zero expiry, clamped to 12 h – 7 d).
func (s *Scheduler) EnsureKeepAlive(platformID string, platformName string, cookies []browser.CDPCookie) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check whether a keep-alive already exists for this platform.
	for _, j := range s.jobs {
		if j.Type != JobTypeKeepAlive {
			continue
		}
		for _, pid := range j.Platforms {
			if pid == platformID {
				return // already covered
			}
		}
	}

	cronExpr := s.computeKeepAliveCron(cookies)

	j := &Job{
		ID:        uuid.New().String(),
		Name:      fmt.Sprintf("%s Session Keep-Alive", platformName),
		Type:      JobTypeKeepAlive,
		Platforms: []string{platformID},
		CronExpr:  cronExpr,
		Status:    JobStatusActive,
		CreatedAt: time.Now(),
		Auto:      true,
	}

	s.jobs[j.ID] = j
	s.registerJob(j)
	s.persistJob(j)

	logger.Info("scheduler", fmt.Sprintf("auto-created keep-alive for %s (cron: %s)", platformName, cronExpr))
}

// computeKeepAliveCron derives a cron expression from cookie expiry times.
func (s *Scheduler) computeKeepAliveCron(cookies []browser.CDPCookie) string {
	now := time.Now()
	var shortest time.Duration

	for _, c := range cookies {
		if c.Expires <= 0 {
			continue
		}
		exp := time.Unix(int64(c.Expires), 0)
		dur := exp.Sub(now)
		if dur <= 0 {
			continue
		}
		if shortest == 0 || dur < shortest {
			shortest = dur
		}
	}

	if shortest > 0 {
		interval := time.Duration(float64(shortest) * 0.75)

		// Clamp to [12h, 7d].
		const minInterval = 12 * time.Hour
		const maxInterval = 7 * 24 * time.Hour
		if interval < minInterval {
			interval = minInterval
		}
		if interval > maxInterval {
			interval = maxInterval
		}

		hours := int(interval.Hours())
		if hours < 1 {
			hours = 12
		}
		return fmt.Sprintf("0 */%d * * *", hours)
	}

	// No valid expiry found — fall back to config default.
	return config.GetScheduler().DefaultKeepAliveInterval
}

// --- private helpers --------------------------------------------------------

// registerJob adds a cron entry for the job and updates its NextRun.
func (s *Scheduler) registerJob(j *Job) {
	if s.cronRunner == nil {
		return
	}

	entryID, err := s.cronRunner.AddFunc(j.CronExpr, func() {
		s.execute(j)
	})
	if err != nil {
		logger.Error("scheduler", fmt.Sprintf("failed to register job %s: %v", j.Name, err))
		j.Status = JobStatusFailed
		j.LastResult = fmt.Sprintf("invalid cron: %v", err)
		return
	}

	j.cronEntryID = entryID
	entry := s.cronRunner.Entry(entryID)
	j.NextRun = entry.Next
}

// unregisterJob removes a job's cron entry.
func (s *Scheduler) unregisterJob(j *Job) {
	if s.cronRunner == nil {
		return
	}
	s.cronRunner.Remove(j.cronEntryID)
	j.cronEntryID = 0
}

// execute dispatches the job based on its type. It is called by the cron runner.
func (s *Scheduler) execute(j *Job) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("scheduler", fmt.Sprintf("panic in job %s: %v", j.Name, r))
			s.mu.Lock()
			j.LastResult = fmt.Sprintf("panic: %v", r)
			j.Status = JobStatusFailed
			s.persistJob(j)
			s.mu.Unlock()
		}
	}()

	logger.Info("scheduler", fmt.Sprintf("executing job %s (%s)", j.Name, j.Type))

	switch j.Type {
	case JobTypeKeepAlive:
		s.executeKeepAlive(j)
	case JobTypeUpload:
		s.executeUpload(j)
	default:
		logger.Warn("scheduler", fmt.Sprintf("unknown job type: %s", j.Type))
	}

	s.mu.Lock()
	j.LastRun = time.Now()
	j.RunCount++
	// Refresh NextRun from the cron entry.
	if s.cronRunner != nil {
		entry := s.cronRunner.Entry(j.cronEntryID)
		j.NextRun = entry.Next
	}
	s.persistJob(j)
	s.mu.Unlock()
}

// load reads persisted jobs from the repo into the in-memory map.
func (s *Scheduler) load() {
	jobs, err := s.repo.List()
	if err != nil {
		logger.Warn("scheduler", fmt.Sprintf("load schedules: %v", err))
		return
	}
	for _, j := range jobs {
		s.jobs[j.ID] = j
	}
}

// persistJob writes one job through the repo.
func (s *Scheduler) persistJob(j *Job) {
	if err := s.repo.Save(j); err != nil {
		logger.Error("scheduler", fmt.Sprintf("persist job %s: %v", j.ID, err))
	}
}

// persistDelete removes a job through the repo.
func (s *Scheduler) persistDelete(id string) {
	if err := s.repo.Delete(id); err != nil {
		logger.Error("scheduler", fmt.Sprintf("delete job %s: %v", id, err))
	}
}
