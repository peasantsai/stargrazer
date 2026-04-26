package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	filePath   string
	browser    *browser.Manager
	sessions   *social.SessionStore
	running    bool
}

var (
	instance *Scheduler
	once     sync.Once
)

// GetInstance returns the singleton scheduler. The first call initialises
// it with the given browser manager and session store references.
func GetInstance(b *browser.Manager, s *social.SessionStore) *Scheduler {
	once.Do(func() {
		instance = &Scheduler{
			jobs:     make(map[string]*Job),
			filePath: social.SchedulesFilePath(),
			browser:  b,
			sessions: s,
		}
	})
	return instance
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

// Stop halts the cron runner, persists state and marks the scheduler stopped.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cronRunner != nil {
		s.cronRunner.Stop()
	}
	s.persist()
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
	s.persist()

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

	s.persist()
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
	s.persist()

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
	s.persist()

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
	s.persist()

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
	s.persist()

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
			s.persist()
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
	s.persist()
	s.mu.Unlock()
}

// persist writes all jobs to disk as indented JSON.
func (s *Scheduler) persist() {
	jobs := make([]*Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		jobs = append(jobs, j)
	}

	data, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		logger.Error("scheduler", fmt.Sprintf("failed to marshal jobs: %v", err))
		return
	}

	os.MkdirAll(filepath.Dir(s.filePath), 0700)
	if err := os.WriteFile(s.filePath, data, 0600); err != nil {
		logger.Error("scheduler", fmt.Sprintf("failed to write jobs: %v", err))
	}
}

// load reads persisted jobs from disk into the jobs map.
func (s *Scheduler) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		// File doesn't exist yet — that's fine on first run.
		return
	}

	var jobs []*Job
	if err := json.Unmarshal(data, &jobs); err != nil {
		logger.Warn("scheduler", fmt.Sprintf("failed to parse jobs file: %v", err))
		return
	}

	for _, j := range jobs {
		s.jobs[j.ID] = j
	}
}
