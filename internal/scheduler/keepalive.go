package scheduler

import (
	"fmt"
	"strings"
	"time"

	"stargrazer/internal/browser"
	"stargrazer/internal/logger"
	"stargrazer/internal/social"
)

// executeKeepAlive refreshes sessions for each platform in the job by
// navigating to the platform URL (if the browser is running) and
// re-exporting cookies to disk.
func (s *Scheduler) executeKeepAlive(job *Job) {
	var results []string

	for _, pid := range job.Platforms {
		platform := social.FindPlatform(social.Platform(pid))
		if platform == nil {
			msg := fmt.Sprintf("platform %s not found, skipping", pid)
			logger.Warn("scheduler", msg)
			results = append(results, msg)
			continue
		}

		// Load stored cookies from disk.
		cookies, err := browser.LoadCookiesFromDisk(pid, social.SharedSessionDir())
		if err != nil {
			msg := fmt.Sprintf("%s: no stored cookies (%v)", platform.Name, err)
			logger.Warn("scheduler", msg)
			results = append(results, msg)
			continue
		}

		if s.browser.IsRunning() {
			// Browser is up — open a new tab to touch the session.
			logger.Info("scheduler", fmt.Sprintf("%s: browser running, opening keep-alive tab", platform.Name))

			_, err := s.browser.OpenNewTab(platform.URL)
			if err != nil {
				msg := fmt.Sprintf("%s: failed to open tab: %v", platform.Name, err)
				logger.Error("scheduler", msg)
				results = append(results, msg)
				continue
			}

			// Give the page time to load and set cookies.
			time.Sleep(5 * time.Second)

			// Re-export cookies so the on-disk copy stays fresh.
			exportErr := s.browser.ExportCookiesToDisk(
				platform.SessionDomains,
				pid,
				social.SharedSessionDir(),
			)
			if exportErr != nil {
				msg := fmt.Sprintf("%s: cookie export failed: %v", platform.Name, exportErr)
				logger.Warn("scheduler", msg)
				results = append(results, msg)
				continue
			}

			msg := fmt.Sprintf("%s: keep-alive OK (%d cookies refreshed)", platform.Name, len(cookies))
			logger.Info("scheduler", msg)
			results = append(results, msg)
		} else {
			// Browser is not running — defer until next browser start to avoid
			// port conflicts with a headless instance.
			msg := fmt.Sprintf("%s: browser not running, keep-alive will run on next browser start", platform.Name)
			logger.Info("scheduler", msg)
			results = append(results, msg)
		}
	}

	// Store a summary on the job so the UI can display it.
	job.LastResult = strings.Join(results, "; ")
}

// executeUpload is a stub that logs the upload intent. Full upload
// orchestration is handled by TriggerUpload in app.go; this placeholder
// ensures the scheduler can reference upload jobs without crashing.
func (s *Scheduler) executeUpload(job *Job) {
	if job.UploadConfig == nil {
		logger.Warn("scheduler", fmt.Sprintf("upload job %s has no upload config", job.Name))
		job.LastResult = "no upload config"
		return
	}

	logger.Info("scheduler", fmt.Sprintf(
		"upload job %s: would upload %s to %v (caption: %q, hashtags: %v)",
		job.Name,
		job.UploadConfig.FilePath,
		job.Platforms,
		job.UploadConfig.Caption,
		job.UploadConfig.Hashtags,
	))

	job.LastResult = fmt.Sprintf("upload scheduled for %s (stub — use TriggerUpload for now)", job.UploadConfig.FilePath)
}
