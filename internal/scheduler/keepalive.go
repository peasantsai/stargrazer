package scheduler

import (
	"fmt"
	"strings"
	"time"

	"stargrazer/internal/browser"
	"stargrazer/internal/logger"
	"stargrazer/internal/social"
)

// executeKeepAlive refreshes sessions for each platform in the job.
// If the browser is not running it is started automatically, then stopped
// again when the job finishes so that scheduled jobs never silently skip.
func (s *Scheduler) executeKeepAlive(job *Job) {
	var results []string

	// Auto-start the browser if it is not running.
	browserWasStarted := false
	if !s.browser.IsRunning() {
		logger.Info("scheduler", "browser not running — auto-starting for keep-alive job")
		if err := s.browser.Start(); err != nil {
			msg := fmt.Sprintf("auto-start browser failed: %v", err)
			logger.Error("scheduler", msg)
			job.LastResult = msg
			return
		}
		browserWasStarted = true
		// Give the browser a moment to initialise CDP.
		time.Sleep(3 * time.Second)
		logger.Info("scheduler", "browser auto-started successfully")
	}

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

		logger.Info("scheduler", fmt.Sprintf("%s: opening keep-alive tab", platform.Name))

		_, err = s.browser.OpenNewTab(platform.URL)
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
	}

	// Auto-stop the browser if we started it.
	if browserWasStarted {
		logger.Info("scheduler", "stopping auto-started browser")
		if err := s.browser.Stop(); err != nil {
			logger.Error("scheduler", fmt.Sprintf("auto-stop browser failed: %v", err))
		} else {
			logger.Info("scheduler", "browser auto-stopped successfully")
		}
	}

	// Store a summary on the job so the UI can display it.
	job.LastResult = strings.Join(results, "; ")
}

// executeUpload is a stub that logs the upload intent. Full upload
// orchestration is handled by TriggerUpload in app.go; this placeholder
// ensures the scheduler can reference upload jobs without crashing.
// Like keep-alive, it auto-starts/stops the browser when needed.
func (s *Scheduler) executeUpload(job *Job) {
	if job.UploadConfig == nil {
		logger.Warn("scheduler", fmt.Sprintf("upload job %s has no upload config", job.Name))
		job.LastResult = "no upload config"
		return
	}

	// Auto-start the browser if it is not running.
	browserWasStarted := false
	if !s.browser.IsRunning() {
		logger.Info("scheduler", fmt.Sprintf("upload job %s: auto-starting browser", job.Name))
		if err := s.browser.Start(); err != nil {
			msg := fmt.Sprintf("auto-start browser failed: %v", err)
			logger.Error("scheduler", msg)
			job.LastResult = msg
			return
		}
		browserWasStarted = true
		time.Sleep(3 * time.Second)
		logger.Info("scheduler", "browser auto-started for upload job")
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

	// Auto-stop the browser if we started it.
	if browserWasStarted {
		logger.Info("scheduler", "stopping auto-started browser after upload job")
		if err := s.browser.Stop(); err != nil {
			logger.Error("scheduler", fmt.Sprintf("auto-stop browser failed: %v", err))
		} else {
			logger.Info("scheduler", "browser auto-stopped after upload job")
		}
	}
}
