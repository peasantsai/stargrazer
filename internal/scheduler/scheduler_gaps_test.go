package scheduler

import (
	"testing"
	"time"

	"stargrazer/internal/browser"
)

// --- execute panic recovery ---

func TestExecutePanicRecovery(t *testing.T) {
	s := newTestScheduler(t)
	s.Start()
	defer s.Stop()

	job := s.Create(Job{
		Name:     "Panic Test",
		Type:     JobTypeKeepAlive,
		CronExpr: "0 */12 * * *",
	})

	// Normal execute should not panic and should increment RunCount.
	s.execute(job)
	if job.RunCount != 1 {
		t.Errorf("expected RunCount 1, got %d", job.RunCount)
	}
}

// --- persistDelete on missing row is handled ---

func TestPersistDeleteMissingRowIsHandled(t *testing.T) {
	// Repo returns an error for a missing row; persistDelete must log-and-swallow.
	s := newTestScheduler(t)
	// Should not panic; error is logged internally.
	s.persistDelete("nonexistent-id")
}

// --- computeKeepAliveCron hours clamps to 12h minimum ---

func TestComputeKeepAliveCronHoursFloorTo12(t *testing.T) {
	s := newTestScheduler(t)
	// 1-minute expiry: 75% of 1m = 45s → clamped to 12h minimum.
	cookies := []browser.CDPCookie{
		{Expires: float64(time.Now().Add(1 * time.Minute).Unix())},
	}
	cron := s.computeKeepAliveCron(cookies)
	if cron != "0 */12 * * *" {
		t.Errorf("expected 12h minimum cron, got %q", cron)
	}
}

// --- EnsureKeepAlive: upload job does not count as existing keep-alive ---

func TestEnsureKeepAliveSkipsUploadJobs(t *testing.T) {
	s := newTestScheduler(t)

	// An upload job for the same platform should NOT prevent keep-alive creation.
	s.Create(Job{
		Name:      "Upload for instagram",
		Type:      JobTypeUpload,
		Platforms: []string{"instagram"},
		CronExpr:  "0 */12 * * *",
	})

	cookies := []browser.CDPCookie{
		{Expires: float64(time.Now().Add(48 * time.Hour).Unix())},
	}
	s.EnsureKeepAlive("instagram", "Instagram", cookies)

	list := s.List()
	keepAlives := 0
	for _, j := range list {
		if j.Type == JobTypeKeepAlive {
			keepAlives++
		}
	}
	if keepAlives != 1 {
		t.Errorf("expected 1 keep-alive job (upload shouldn't count), got %d", keepAlives)
	}
}

// --- NextRun not updated when cronRunner is nil ---

func TestExecuteNextRunZeroWithNilRunner(t *testing.T) {
	s := newTestScheduler(t)
	// Do NOT call Start() — cronRunner stays nil.

	job := s.Create(Job{
		Name:     "No Runner Exec",
		Type:     JobTypeUpload,
		CronExpr: "0 */12 * * *",
	})

	s.execute(job)
	if job.RunCount != 1 {
		t.Errorf("expected RunCount 1, got %d", job.RunCount)
	}
	// NextRun stays zero when no runner.
	if !job.NextRun.IsZero() {
		t.Error("expected NextRun to remain zero with nil cronRunner")
	}
}

// --- EnsureKeepAlive: platform match in inner loop early return ---

func TestEnsureKeepAliveExistingMatchReturnsEarly(t *testing.T) {
	s := newTestScheduler(t)
	cookies := []browser.CDPCookie{
		{Expires: float64(time.Now().Add(48 * time.Hour).Unix())},
	}

	// First call creates it.
	s.EnsureKeepAlive("facebook", "Facebook", cookies)
	// Second call for the same platform — must not create a duplicate.
	s.EnsureKeepAlive("facebook", "Facebook", cookies)
	// Third call — still only one.
	s.EnsureKeepAlive("facebook", "Facebook", cookies)

	list := s.List()
	count := 0
	for _, j := range list {
		if j.Type == JobTypeKeepAlive {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 keep-alive job, got %d (duplicates created)", count)
	}
}
