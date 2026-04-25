package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"stargrazer/internal/browser"
	"stargrazer/internal/config"
	"stargrazer/internal/logger"
	"stargrazer/internal/scheduler"
	"stargrazer/internal/social"
	"stargrazer/internal/workflow"
)

// App exposes backend methods to the frontend via Wails bindings.
type App struct {
	ctx       context.Context
	browser   *browser.Manager
	sessions  *social.SessionStore
	scheduler *scheduler.Scheduler
}

func NewApp() *App {
	b := browser.GetInstance()
	s := social.NewSessionStore()
	return &App{
		browser:   b,
		sessions:  s,
		scheduler: scheduler.GetInstance(b, s),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logger.Info("app", "Stargrazer started")
	if config.GetScheduler().Enabled {
		a.scheduler.Start()
	}
}

func (a *App) shutdown(ctx context.Context) {
	logger.Info("app", "Shutting down...")
	a.scheduler.Stop()
	a.browser.Stop()
}

// --- Browser Controls ---

type BrowserStatusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func (a *App) StartBrowser() BrowserStatusResponse {
	logger.Info("browser", "Starting browser...")
	if err := a.browser.Start(); err != nil {
		logger.Error("browser", fmt.Sprintf("Start failed: %v", err))
		return BrowserStatusResponse{Status: string(browser.StatusError), Error: err.Error()}
	}
	logger.Info("browser", "Browser started, CDP active")
	return BrowserStatusResponse{Status: string(browser.StatusRunning)}
}

func (a *App) StopBrowser() BrowserStatusResponse {
	logger.Info("browser", "Stopping browser...")
	if err := a.browser.Stop(); err != nil {
		logger.Error("browser", fmt.Sprintf("Stop failed: %v", err))
		return BrowserStatusResponse{Status: string(browser.StatusError), Error: err.Error()}
	}
	logger.Info("browser", "Browser stopped")
	return BrowserStatusResponse{Status: string(browser.StatusStopped)}
}

func (a *App) GetBrowserStatus() BrowserStatusResponse {
	status, errMsg := a.browser.GetStatus()
	return BrowserStatusResponse{Status: string(status), Error: errMsg}
}

func (a *App) RestartBrowser() BrowserStatusResponse {
	logger.Info("browser", "Restarting browser with new settings...")
	if a.browser.IsRunning() {
		a.browser.Stop()
	}
	sessionDir, _ := social.EnsureSessionDir("")
	if err := a.browser.StartWithOptions(sessionDir, ""); err != nil {
		logger.Error("browser", fmt.Sprintf("Restart failed: %v", err))
		return BrowserStatusResponse{Status: string(browser.StatusError), Error: err.Error()}
	}
	logger.Info("browser", "Browser restarted")
	return BrowserStatusResponse{Status: string(browser.StatusRunning)}
}

// --- Config ---

type BrowserConfigResponse struct {
	ChromiumPath string   `json:"chromiumPath"`
	CDPPort      int      `json:"cdpPort"`
	Headless     bool     `json:"headless"`
	UserDataDir  string   `json:"userDataDir"`
	WindowWidth  int      `json:"windowWidth"`
	WindowHeight int      `json:"windowHeight"`
	ExtraFlags   []string `json:"extraFlags"`
}

func (a *App) GetBrowserConfig() BrowserConfigResponse {
	cfg := config.GetBrowser()
	resp := toBrowserConfigResponse(cfg)
	if resp.ChromiumPath == "" {
		resp.ChromiumPath = a.browser.ResolveChromiumPath()
	}
	return resp
}

func (a *App) ResetBrowserConfig() BrowserConfigResponse {
	updated := config.Reset()
	resp := toBrowserConfigResponse(updated.Browser)
	if resp.ChromiumPath == "" {
		resp.ChromiumPath = a.browser.ResolveChromiumPath()
	}
	return resp
}

func (a *App) UpdateBrowserConfig(update BrowserConfigResponse) BrowserConfigResponse {
	updated := config.Update(func(c *config.AppConfig) {
		if update.CDPPort > 0 {
			c.Browser.CDPPort = update.CDPPort
		}
		if update.ChromiumPath != "" {
			c.Browser.ChromiumPath = update.ChromiumPath
		}
		if update.UserDataDir != "" {
			c.Browser.UserDataDir = update.UserDataDir
		}
		if update.WindowWidth > 0 {
			c.Browser.WindowWidth = update.WindowWidth
		}
		if update.WindowHeight > 0 {
			c.Browser.WindowHeight = update.WindowHeight
		}
		c.Browser.Headless = update.Headless
		if update.ExtraFlags != nil {
			c.Browser.ExtraFlags = update.ExtraFlags
		}
	})
	return toBrowserConfigResponse(updated.Browser)
}

func toBrowserConfigResponse(cfg config.BrowserConfig) BrowserConfigResponse {
	flags := cfg.ExtraFlags
	if flags == nil {
		flags = []string{}
	}
	return BrowserConfigResponse{
		ChromiumPath: cfg.ChromiumPath, CDPPort: cfg.CDPPort, Headless: cfg.Headless,
		UserDataDir: cfg.UserDataDir, WindowWidth: cfg.WindowWidth, WindowHeight: cfg.WindowHeight,
		ExtraFlags: flags,
	}
}

// --- Social Media ---

type PlatformResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	LoggedIn   bool   `json:"loggedIn"`
	Username   string `json:"username"`
	LastLogin  string `json:"lastLogin"`
	LastCheck  string `json:"lastCheck"`
	SessionDir string `json:"sessionDir"`
}

func (a *App) GetPlatforms() []PlatformResponse {
	platforms := social.AllPlatforms()
	statuses := a.sessions.GetAll()
	statusMap := make(map[social.Platform]social.AccountStatus)
	for _, s := range statuses {
		statusMap[s.PlatformID] = s
	}
	result := make([]PlatformResponse, 0, len(platforms))
	for _, p := range platforms {
		s := statusMap[p.ID]
		result = append(result, toPlatformResponse(&p, s))
	}
	return result
}

func (a *App) OpenPlatform(platformID string) BrowserStatusResponse {
	platform := social.FindPlatform(social.Platform(platformID))
	if platform == nil {
		return BrowserStatusResponse{Status: "error", Error: fmt.Sprintf("unknown platform: %s", platformID)}
	}
	sessionDir, err := social.EnsureSessionDir(social.Platform(platformID))
	if err != nil {
		return BrowserStatusResponse{Status: "error", Error: err.Error()}
	}
	if a.browser.IsRunning() {
		logger.Info("social", fmt.Sprintf("Opening %s in new tab", platform.Name))
		if _, err := a.browser.OpenNewTab(platform.URL); err != nil {
			return BrowserStatusResponse{Status: "error", Error: fmt.Sprintf("new tab failed: %v", err)}
		}
		return BrowserStatusResponse{Status: "running"}
	}
	logger.Info("social", fmt.Sprintf("Starting browser for %s", platform.Name))
	if err := a.browser.StartWithOptions(sessionDir, platform.URL); err != nil {
		return BrowserStatusResponse{Status: string(browser.StatusError), Error: err.Error()}
	}
	return BrowserStatusResponse{Status: string(browser.StatusRunning)}
}

func (a *App) CheckLoginStatus(platformID string) PlatformResponse {
	pid := social.Platform(platformID)
	platform := social.FindPlatform(pid)
	if platform == nil {
		return PlatformResponse{ID: platformID}
	}
	if !a.browser.IsRunning() {
		return toPlatformResponse(platform, a.sessions.Get(pid))
	}
	loggedIn, username, err := a.browser.HasCookies(platform.SessionDomains, platform.LoginCookies)
	if err != nil {
		return toPlatformResponse(platform, a.sessions.Get(pid))
	}
	if loggedIn {
		a.sessions.SetLoggedIn(pid, username)
		dataDir := social.SharedSessionDir()
		a.browser.ExportCookiesToDisk(platform.SessionDomains, string(pid), dataDir)
	} else {
		a.sessions.SetLoggedOut(pid)
	}
	return toPlatformResponse(platform, a.sessions.Get(pid))
}

func (a *App) CheckAllLoginStatus() []PlatformResponse {
	result := make([]PlatformResponse, 0)
	for _, p := range social.AllPlatforms() {
		result = append(result, a.CheckLoginStatus(string(p.ID)))
	}
	return result
}

// PurgeSession removes all stored session data for a platform so the user can reconnect.
func (a *App) PurgeSession(platformID string) PlatformResponse {
	pid := social.Platform(platformID)
	platform := social.FindPlatform(pid)
	if platform == nil {
		return PlatformResponse{ID: platformID}
	}

	// Mark as logged out
	a.sessions.SetLoggedOut(pid)

	// Delete cookies file from disk
	cookieFile := filepath.Join(social.SharedSessionDir(), "cookies", platformID+".json")
	os.Remove(cookieFile)

	logger.Info("social", fmt.Sprintf("%s session purged", platform.Name))
	return toPlatformResponse(platform, a.sessions.Get(pid))
}

// ImportCookies parses Netscape cookie text, saves to disk, injects if browser is running,
// marks as logged in, and auto-creates a keep-alive schedule.
func (a *App) ImportCookies(platformID string, cookieText string) PlatformResponse {
	pid := social.Platform(platformID)
	platform := social.FindPlatform(pid)
	if platform == nil {
		return PlatformResponse{ID: platformID}
	}

	cookies := browser.ParseNetscapeCookies(cookieText)
	if len(cookies) == 0 {
		logger.Warn("social", fmt.Sprintf("No cookies parsed for %s", platform.Name))
		return toPlatformResponse(platform, a.sessions.Get(pid))
	}
	logger.Info("social", fmt.Sprintf("Parsed %d cookies for %s", len(cookies), platform.Name))

	// Inject cookies if browser is running (don't start browser just for import)
	if a.browser.IsRunning() {
		if err := a.browser.SetCookiesViaCDP(cookies); err != nil {
			logger.Warn("social", fmt.Sprintf("CDP cookie inject failed for %s: %v", platform.Name, err))
		} else {
			logger.Info("social", fmt.Sprintf("%s cookies injected into running browser", platform.Name))
		}
	}

	// Extract username
	username := ""
	for _, c := range cookies {
		switch c.Name {
		case "c_user", "ds_user_id":
			if username == "" {
				username = c.Value
			}
		}
	}

	a.sessions.SetLoggedIn(pid, username)

	// Persist cookies to disk
	dataDir := social.SharedSessionDir()
	cookieData, _ := json.MarshalIndent(cookies, "", "  ")
	cookiesDir := filepath.Join(dataDir, "cookies")
	os.MkdirAll(cookiesDir, 0755)
	os.WriteFile(filepath.Join(cookiesDir, platformID+".json"), cookieData, 0644)

	logger.Info("social", fmt.Sprintf("%s session saved (user: %s)", platform.Name, username))

	// Auto-create keep-alive schedule
	a.scheduler.EnsureKeepAlive(platformID, platform.Name, cookies)

	return toPlatformResponse(platform, a.sessions.Get(pid))
}

func toPlatformResponse(p *social.PlatformInfo, s social.AccountStatus) PlatformResponse {
	return PlatformResponse{
		ID: string(p.ID), Name: p.Name, URL: p.URL,
		LoggedIn: s.LoggedIn, Username: s.Username,
		LastLogin: formatTime(s.LastLogin), LastCheck: formatTime(s.LastCheck),
		SessionDir: social.SharedSessionDir(),
	}
}

// --- Schedules ---

type ScheduleResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Platforms  []string `json:"platforms"`
	CronExpr   string   `json:"cronExpr"`
	NextRun    string   `json:"nextRun"`
	LastRun    string   `json:"lastRun"`
	Status     string   `json:"status"`
	CreatedAt  string   `json:"createdAt"`
	RunCount   int      `json:"runCount"`
	LastResult string   `json:"lastResult"`
	Auto       bool     `json:"auto"`
	FilePath   string   `json:"filePath,omitempty"`
	Caption    string   `json:"caption,omitempty"`
	Hashtags   []string `json:"hashtags,omitempty"`
}

type CreateScheduleRequest struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Platforms []string `json:"platforms"`
	CronExpr  string   `json:"cronExpr"`
	FilePath  string   `json:"filePath,omitempty"`
	Caption   string   `json:"caption,omitempty"`
	Hashtags  []string `json:"hashtags,omitempty"`
}

func toScheduleResponse(j *scheduler.Job) ScheduleResponse {
	r := ScheduleResponse{
		ID: j.ID, Name: j.Name, Type: string(j.Type),
		Platforms: j.Platforms, CronExpr: j.CronExpr,
		NextRun: formatTime(j.NextRun), LastRun: formatTime(j.LastRun),
		Status: string(j.Status), CreatedAt: formatTime(j.CreatedAt),
		RunCount: j.RunCount, LastResult: j.LastResult, Auto: j.Auto,
	}
	if j.UploadConfig != nil {
		r.FilePath = j.UploadConfig.FilePath
		r.Caption = j.UploadConfig.Caption
		r.Hashtags = j.UploadConfig.Hashtags
	}
	return r
}

func (a *App) GetSchedules() []ScheduleResponse {
	jobs := a.scheduler.List()
	result := make([]ScheduleResponse, len(jobs))
	for i, j := range jobs {
		result[i] = toScheduleResponse(j)
	}
	return result
}

func (a *App) CreateSchedule(req CreateScheduleRequest) ScheduleResponse {
	job := scheduler.Job{
		Name:      req.Name,
		Type:      scheduler.JobType(req.Type),
		Platforms: req.Platforms,
		CronExpr:  req.CronExpr,
	}
	if req.Type == string(scheduler.JobTypeUpload) {
		job.UploadConfig = &scheduler.UploadConfig{
			FilePath: req.FilePath,
			Caption:  req.Caption,
			Hashtags: req.Hashtags,
		}
	}
	created := a.scheduler.Create(job)
	logger.Info("scheduler", fmt.Sprintf("Schedule created: %s (%s)", created.Name, created.ID))
	return toScheduleResponse(created)
}

func (a *App) UpdateSchedule(id string, req CreateScheduleRequest) ScheduleResponse {
	updated := a.scheduler.Update(id, func(j *scheduler.Job) {
		if req.Name != "" {
			j.Name = req.Name
		}
		if req.CronExpr != "" {
			j.CronExpr = req.CronExpr
		}
		if req.Platforms != nil {
			j.Platforms = req.Platforms
		}
		if req.Type != "" {
			j.Type = scheduler.JobType(req.Type)
		}
		if req.Type == string(scheduler.JobTypeUpload) {
			j.UploadConfig = &scheduler.UploadConfig{
				FilePath: req.FilePath, Caption: req.Caption, Hashtags: req.Hashtags,
			}
		}
	})
	if updated == nil {
		return ScheduleResponse{}
	}
	return toScheduleResponse(updated)
}

func (a *App) DeleteSchedule(id string) bool {
	return a.scheduler.Delete(id)
}

func (a *App) PauseSchedule(id string) ScheduleResponse {
	j := a.scheduler.Pause(id)
	if j == nil {
		return ScheduleResponse{}
	}
	return toScheduleResponse(j)
}

func (a *App) ResumeSchedule(id string) ScheduleResponse {
	j := a.scheduler.Resume(id)
	if j == nil {
		return ScheduleResponse{}
	}
	return toScheduleResponse(j)
}

func (a *App) GetScheduleStats(id string) ScheduleResponse {
	j := a.scheduler.Get(id)
	if j == nil {
		return ScheduleResponse{}
	}
	return toScheduleResponse(j)
}

// --- Logs ---

type LogEntryResponse struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Message   string `json:"message"`
}

func (a *App) GetLogs() []LogEntryResponse {
	entries := logger.GetAll()
	result := make([]LogEntryResponse, len(entries))
	for i, e := range entries {
		result[i] = LogEntryResponse{
			Timestamp: e.Timestamp.Format(time.RFC3339Nano),
			Level:     string(e.Level), Source: e.Source, Message: e.Message,
		}
	}
	return result
}

func (a *App) ExportLogs() string { return string(logger.Export()) }
func (a *App) ClearLogs()         { logger.Clear() }

// --- Upload / Workflow ---

type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (a *App) TriggerUpload(req workflow.UploadRequest) UploadResponse {
	// Caption-only, hashtags-only, or file-only are all valid
	if len(req.Platforms) == 0 {
		return UploadResponse{Success: false, Message: "No platforms selected"}
	}
	if req.FilePath == "" && req.Caption == "" && len(req.Hashtags) == 0 {
		return UploadResponse{Success: false, Message: "Provide at least a file, caption, or hashtags"}
	}
	if !a.browser.IsRunning() {
		return UploadResponse{Success: false, Message: "Browser is not running"}
	}

	logger.Info("upload", fmt.Sprintf("Upload triggered for %v", req.Platforms))

	// Save upload record to disk
	dataDir := social.SharedSessionDir()
	uploadsDir := filepath.Join(dataDir, "data", "uploads")
	os.MkdirAll(uploadsDir, 0755)
	record := map[string]interface{}{
		"platforms": req.Platforms, "filePath": req.FilePath,
		"caption": req.Caption, "hashtags": req.Hashtags,
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "queued",
	}
	recordData, _ := json.MarshalIndent(record, "", "  ")
	recordFile := filepath.Join(uploadsDir, fmt.Sprintf("upload_%d.json", time.Now().UnixMilli()))
	os.WriteFile(recordFile, recordData, 0644)

	for _, pid := range req.Platforms {
		wf, err := workflow.LoadWorkflow(pid)
		if err != nil {
			logger.Warn("upload", fmt.Sprintf("No workflow for %s: %v", pid, err))
			continue
		}
		steps := workflow.PrepareSteps(wf.Steps, req)
		logger.Info("upload", fmt.Sprintf("Prepared %d steps for %s", len(steps), pid))
	}

	return UploadResponse{Success: true, Message: fmt.Sprintf("Upload queued for %d platform(s)", len(req.Platforms))}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
