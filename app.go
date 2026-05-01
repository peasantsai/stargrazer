package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"stargrazer/internal/automation"
	"stargrazer/internal/browser"
	"stargrazer/internal/config"
	"stargrazer/internal/logger"
	"stargrazer/internal/scheduler"
	"stargrazer/internal/social"
	"stargrazer/internal/workflow"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// safePlatformIDPattern ensures platform IDs contain only safe characters,
// preventing path traversal attacks since platformID is used in file paths.
var safePlatformIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// App exposes backend methods to the frontend via Wails bindings.
type App struct {
	getCtx      func() context.Context
	browser     *browser.Manager
	sessions    social.SessionRepo
	scheduler   *scheduler.Scheduler
	automations automation.Repository

	currentRunMu          sync.Mutex
	currentRunCancel      context.CancelFunc
	currentAutomationName string
	currentAutomationID   string
}

// NewApp constructs an App with explicit dependencies. Callers (main.go) own
// the lifecycle of the SQLite handle and the repos that wrap it.
func NewApp(automations automation.Repository, sessions social.SessionRepo, sched *scheduler.Scheduler, b *browser.Manager) *App {
	return &App{
		getCtx:      func() context.Context { return context.Background() },
		browser:     b,
		sessions:    sessions,
		scheduler:   sched,
		automations: automations,
	}
}

func (a *App) startup(ctx context.Context) {
	a.getCtx = func() context.Context { return ctx }
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
	sessionDir := social.SharedSessionDir()
	if err := a.browser.StartWithOptions(sessionDir, ""); err != nil {
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
	if resp.UserDataDir == "" {
		resp.UserDataDir = social.SharedSessionDir()
	}
	return resp
}

func (a *App) ResetBrowserConfig() BrowserConfigResponse {
	updated := config.Reset()
	resp := toBrowserConfigResponse(updated.Browser)
	if resp.ChromiumPath == "" {
		resp.ChromiumPath = a.browser.ResolveChromiumPath()
	}
	if resp.UserDataDir == "" {
		resp.UserDataDir = social.SharedSessionDir()
	}
	return resp
}

func (a *App) UpdateBrowserConfig(update BrowserConfigResponse) BrowserConfigResponse {
	updated := config.Update(func(c *config.AppConfig) {
		if update.CDPPort > 0 && update.CDPPort <= 65535 {
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
	if !safePlatformIDPattern.MatchString(platformID) {
		return PlatformResponse{ID: platformID}
	}

	// Mark as logged out
	a.sessions.SetLoggedOut(pid)

	// Delete cookies file from disk (best-effort; log on failure)
	cookieFile := filepath.Join(social.SharedSessionDir(), "cookies", platformID+".json")
	if err := os.Remove(cookieFile); err != nil && !os.IsNotExist(err) {
		logger.Warn("social", fmt.Sprintf("removing cookie file for %s: %v", platform.Name, err))
	}

	logger.Info("social", fmt.Sprintf("%s session purged", platform.Name))
	return toPlatformResponse(platform, a.sessions.Get(pid))
}

// ImportCookies parses Netscape cookie text, saves to disk, injects if browser is running,
// marks as logged in, and auto-creates a keep-alive schedule.
func (a *App) ImportCookies(platformID, cookieText string) PlatformResponse {
	pid := social.Platform(platformID)
	platform := social.FindPlatform(pid)
	if platform == nil {
		return PlatformResponse{ID: platformID}
	}
	if !safePlatformIDPattern.MatchString(platformID) {
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

	// Extract username / user ID from platform-specific cookies.
	// Facebook: c_user, Instagram: ds_user_id, TikTok: uid_tt, X: twid (URL-encoded)
	// LinkedIn/YouTube: no standard cookie exposes the username; shown as "Connected"
	username := extractUsernameFromCookies(cookies)
	a.sessions.SetLoggedIn(pid, username)

	// Persist cookies to disk
	dataDir := social.SharedSessionDir()
	persistCookiesToDisk(platformID, platform.Name, cookies, dataDir)

	logger.Info("social", fmt.Sprintf("%s session saved (user: %s)", platform.Name, username))

	// Auto-create keep-alive schedule
	a.scheduler.EnsureKeepAlive(platformID, platform.Name, cookies)

	return toPlatformResponse(platform, a.sessions.Get(pid))
}

// extractUsernameFromCookies scans cookies for platform-specific user identity fields.
func extractUsernameFromCookies(cookies []browser.CDPCookie) string {
	for _, c := range cookies {
		switch c.Name {
		case "c_user", "ds_user_id", "uid_tt":
			return c.Value
		case "twid":
			// twid value is URL-encoded: u%3D<user_id>
			if decoded, err := url.QueryUnescape(c.Value); err == nil {
				return strings.TrimPrefix(decoded, "u=")
			}
		}
	}
	return ""
}

// persistCookiesToDisk writes cookies for a platform to the shared session directory.
func persistCookiesToDisk(platformID, platformName string, cookies []browser.CDPCookie, dataDir string) {
	cookieData, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		logger.Warn("social", fmt.Sprintf("encoding cookies for %s: %v", platformName, err))
		return
	}
	cookiesDir := filepath.Join(dataDir, "cookies")
	if mkErr := os.MkdirAll(cookiesDir, 0700); mkErr != nil {
		logger.Warn("social", fmt.Sprintf("creating cookies dir: %v", mkErr))
		return
	}
	if wErr := os.WriteFile(filepath.Join(cookiesDir, platformID+".json"), cookieData, 0600); wErr != nil {
		logger.Warn("social", fmt.Sprintf("writing cookies for %s: %v", platformName, wErr))
	}
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

// LogFromFrontend allows the React frontend to send log entries to the
// shared application log so they appear alongside backend logs.
func (a *App) LogFromFrontend(level, source, message string) {
	switch level {
	case "warn":
		logger.Warn(source, message)
	case "error":
		logger.Error(source, message)
	case "debug":
		logger.Debug(source, message)
	default:
		logger.Info(source, message)
	}
}

// --- Upload / Workflow ---

// SelectFile opens a native file dialog and returns the selected file path.
func (a *App) SelectFile() string {
	selection, err := wailsRuntime.OpenFileDialog(a.getCtx(), wailsRuntime.OpenDialogOptions{
		Title: "Select file to upload",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Images & Videos", Pattern: "*.jpg;*.jpeg;*.png;*.gif;*.webp;*.mp4;*.mov;*.avi;*.webm"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	if err != nil {
		logger.Error("upload", fmt.Sprintf("File dialog error: %v", err))
		return ""
	}
	return selection
}

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
	os.MkdirAll(uploadsDir, 0700)
	record := map[string]interface{}{
		"platforms": req.Platforms, "filePath": req.FilePath,
		"caption": req.Caption, "hashtags": req.Hashtags,
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "queued",
	}
	recordData, _ := json.MarshalIndent(record, "", "  ")
	recordFile := filepath.Join(uploadsDir, fmt.Sprintf("upload_%d.json", time.Now().UnixMilli()))
	if err := os.WriteFile(recordFile, recordData, 0600); err != nil {
		logger.Warn("upload", fmt.Sprintf("writing upload record: %v", err))
	}

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

// --- Automations ---

// AutomationStepPayload is the wire-format for a single automation step.
type AutomationStepPayload struct {
	Action    string     `json:"action"`
	Target    string     `json:"target"`
	Value     string     `json:"value"`
	Label     string     `json:"label"`
	Selectors [][]string `json:"selectors,omitempty"`
}

// AutomationPayload is the wire-format for an automation config.
type AutomationPayload struct {
	ID          string                  `json:"id"`
	PlatformID  string                  `json:"platformId"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Steps       []AutomationStepPayload `json:"steps"`
	CreatedAt   string                  `json:"createdAt"`
	LastRun     string                  `json:"lastRun"`
	RunCount    int                     `json:"runCount"`
}

// RunAutomationResponse reports the result of executing an automation.
type RunAutomationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func toAutomationPayload(c automation.Config) AutomationPayload {
	steps := make([]AutomationStepPayload, len(c.Steps))
	for i, s := range c.Steps {
		steps[i] = AutomationStepPayload{
			Action: string(s.Action), Target: s.Target,
			Value: s.Value, Label: s.Label, Selectors: s.Selectors,
		}
	}
	return AutomationPayload{
		ID: c.ID, PlatformID: c.PlatformID,
		Name: c.Name, Description: c.Description,
		Steps:     steps,
		CreatedAt: formatTime(c.CreatedAt),
		LastRun:   formatTime(c.LastRun),
		RunCount:  c.RunCount,
	}
}

// GetAutomations returns all saved automations for a platform.
func (a *App) GetAutomations(platformID string) []AutomationPayload {
	if !safePlatformIDPattern.MatchString(platformID) {
		return []AutomationPayload{}
	}
	configs, err := a.automations.List(platformID)
	if err != nil {
		logger.Error("automation", fmt.Sprintf("List: %v", err))
		return []AutomationPayload{}
	}
	result := make([]AutomationPayload, len(configs))
	for i, c := range configs {
		result[i] = toAutomationPayload(c)
	}
	return result
}

// SaveAutomation creates or updates an automation for a platform.
func (a *App) SaveAutomation(platformID string, req AutomationPayload) AutomationPayload {
	if !safePlatformIDPattern.MatchString(platformID) {
		return AutomationPayload{}
	}
	steps := make([]automation.Step, len(req.Steps))
	for i, s := range req.Steps {
		steps[i] = automation.Step{
			Action: automation.Action(s.Action),
			Target: s.Target, Value: s.Value, Label: s.Label,
			Selectors: s.Selectors,
		}
	}
	cfg := automation.Config{
		ID: req.ID, PlatformID: platformID,
		Name: req.Name, Description: req.Description,
		Steps: steps,
	}
	saved, err := a.automations.Save(cfg)
	if err != nil {
		logger.Error("automation", fmt.Sprintf("Save: %v", err))
		return AutomationPayload{}
	}
	logger.Info("automation", fmt.Sprintf("Saved %q for %s", saved.Name, platformID))
	return toAutomationPayload(saved)
}

// DeleteAutomation removes an automation by ID.
func (a *App) DeleteAutomation(platformID, id string) bool {
	if !safePlatformIDPattern.MatchString(platformID) {
		return false
	}
	ok, err := a.automations.Delete(platformID, id)
	if err != nil {
		logger.Error("automation", fmt.Sprintf("Delete: %v", err))
		return false
	}
	return ok
}

// findAutomation loads and returns an automation config by platform and ID.
func (a *App) findAutomation(platformID, id string) (*automation.Config, error) {
	configs, err := a.automations.List(platformID)
	if err != nil {
		return nil, err
	}
	for i := range configs {
		if configs[i].ID == id {
			return &configs[i], nil
		}
	}
	return nil, fmt.Errorf("automation not found")
}

// RunAutomation executes a saved automation step-by-step via chromedp.
func (a *App) RunAutomation(platformID, id string) RunAutomationResponse {
	if !safePlatformIDPattern.MatchString(platformID) {
		return RunAutomationResponse{Success: false, Message: "invalid platform ID"}
	}
	if !a.browser.IsRunning() {
		return RunAutomationResponse{Success: false, Message: "browser is not running — start it first"}
	}
	cfg, err := a.findAutomation(platformID, id)
	if err != nil {
		return RunAutomationResponse{Success: false, Message: err.Error()}
	}

	ctx, cancel, err := a.browser.ConnectChromedp(context.Background())
	if err != nil {
		return RunAutomationResponse{Success: false, Message: fmt.Sprintf("CDP connect: %v", err)}
	}
	defer cancel()

	a.beginRun(cfg.ID, cfg.Name, cancel)
	defer a.endRun()

	logger.Info("automation", fmt.Sprintf("Running %q (%d steps) for %s via chromedp", cfg.Name, len(cfg.Steps), platformID))
	for i, step := range cfg.Steps {
		if err := a.executeStep(ctx, i, len(cfg.Steps), step); err != nil {
			msg := fmt.Sprintf("step %d/%d failed (%s): %v", i+1, len(cfg.Steps), step.Action, err)
			logger.Error("automation", msg)
			return RunAutomationResponse{Success: false, Message: msg}
		}
	}
	a.automations.RecordRun(platformID, id)
	msg := fmt.Sprintf("'%s' completed successfully (%d steps)", cfg.Name, len(cfg.Steps))
	logger.Info("automation", msg)
	return RunAutomationResponse{Success: true, Message: msg}
}

// TestAutomation starts the browser if needed, navigates to the platform,
// executes the automation via chromedp, waits 5 seconds, then cleans up.
func (a *App) TestAutomation(platformID, id string) RunAutomationResponse {
	if !safePlatformIDPattern.MatchString(platformID) {
		return RunAutomationResponse{Success: false, Message: "invalid platform ID"}
	}
	platform := social.FindPlatform(social.Platform(platformID))
	if platform == nil {
		return RunAutomationResponse{Success: false, Message: fmt.Sprintf("unknown platform: %s", platformID)}
	}
	cfg, err := a.findAutomation(platformID, id)
	if err != nil {
		return RunAutomationResponse{Success: false, Message: err.Error()}
	}

	// Auto-start browser if not running.
	browserWasStarted := false
	if !a.browser.IsRunning() {
		logger.Info("automation", "Browser not running — auto-starting for test")
		sessionDir, _ := social.EnsureSessionDir(social.Platform(platformID))
		if err := a.browser.StartWithOptions(sessionDir, ""); err != nil {
			return RunAutomationResponse{Success: false, Message: fmt.Sprintf("auto-start failed: %v", err)}
		}
		browserWasStarted = true
		time.Sleep(2 * time.Second)
	}

	// Connect chromedp to the running browser.
	ctx, cancel, err := a.browser.ConnectChromedp(context.Background())
	if err != nil {
		if browserWasStarted {
			a.browser.Stop()
		}
		return RunAutomationResponse{Success: false, Message: fmt.Sprintf("CDP connect: %v", err)}
	}
	defer cancel()

	a.beginRun(cfg.ID, cfg.Name, cancel)
	defer a.endRun()

	// Navigate to the platform URL first, then wait for page ready.
	logger.Info("automation", fmt.Sprintf("Navigating to %s: %s", platform.Name, platform.URL))
	if err := a.browser.ExecNavigate(ctx, platform.URL); err != nil {
		if browserWasStarted {
			a.browser.Stop()
		}
		return RunAutomationResponse{Success: false, Message: fmt.Sprintf("navigate: %v", err)}
	}
	// Extra wait for dynamic content to render.
	time.Sleep(3 * time.Second)

	// Execute all steps.
	logger.Info("automation", fmt.Sprintf("Testing %q (%d steps) for %s", cfg.Name, len(cfg.Steps), platformID))
	var stepErr error
	for i, step := range cfg.Steps {
		if err := a.executeStep(ctx, i, len(cfg.Steps), step); err != nil {
			stepErr = err
			logger.Error("automation", fmt.Sprintf("step %d/%d failed (%s): %v", i+1, len(cfg.Steps), step.Action, err))
			break
		}
	}

	// Wait 5 seconds so user can observe the result.
	logger.Info("automation", "Waiting 5 seconds before cleanup...")
	time.Sleep(5 * time.Second)

	// Cleanup.
	if browserWasStarted {
		logger.Info("automation", "Stopping auto-started browser")
		a.browser.Stop()
	}

	a.automations.RecordRun(platformID, id)
	if stepErr != nil {
		return RunAutomationResponse{Success: false, Message: stepErr.Error()}
	}
	msg := fmt.Sprintf("Test '%s' completed successfully (%d steps)", cfg.Name, len(cfg.Steps))
	logger.Info("automation", msg)
	return RunAutomationResponse{Success: true, Message: msg}
}

// executeStep dispatches a single automation step via the registered StepHandler
// and emits run.step events for the visibility strip.
func (a *App) executeStep(ctx context.Context, index int, total int, step automation.Step) error {
	a.emitRunStep(runStepEvent{
		AutomationID:   a.currentAutomationID,
		AutomationName: a.currentAutomationName,
		StepIndex:      index,
		Total:          total,
		Action:         string(step.Action),
		Target:         truncTarget(step.Target),
		Status:         "running",
		StartedAt:      time.Now().UTC().Format(time.RFC3339Nano),
	})

	logger.Info("automation", fmt.Sprintf("  step %d/%d: %s → %s", index+1, total, step.Action, truncSel(step)))
	err := browser.RunStep(ctx, a.browser, step)

	status := "success"
	if err != nil {
		status = "failed"
	}
	a.emitRunStep(runStepEvent{
		AutomationID:   a.currentAutomationID,
		AutomationName: a.currentAutomationName,
		StepIndex:      index,
		Total:          total,
		Action:         string(step.Action),
		Status:         status,
		FinishedAt:     time.Now().UTC().Format(time.RFC3339Nano),
		Error:          errString(err),
	})

	if err != nil {
		return fmt.Errorf("%s: %w", step.Action, err)
	}
	return nil
}

// runStepEvent is the payload of the "run.step" Wails event topic. The shape is
// additive — Phase 5 adds RunID, ScreenshotPath, DurationMs.
type runStepEvent struct {
	RunID          string `json:"runId"`
	AutomationID   string `json:"automationId"`
	AutomationName string `json:"automationName"`
	StepIndex      int    `json:"stepIndex"`
	Total          int    `json:"total"`
	Action         string `json:"action"`
	Target         string `json:"target,omitempty"`
	Status         string `json:"status"`
	StartedAt      string `json:"startedAt,omitempty"`
	FinishedAt     string `json:"finishedAt,omitempty"`
	Error          string `json:"error,omitempty"`
}

func (a *App) emitRunStep(ev runStepEvent) {
	wailsRuntime.EventsEmit(a.getCtx(), "run.step", ev)
}

// beginRun stashes the cancel func and automation identity so the visibility
// strip can render and CancelRun can interrupt.
func (a *App) beginRun(automationID, automationName string, cancel context.CancelFunc) {
	a.currentRunMu.Lock()
	defer a.currentRunMu.Unlock()
	a.currentAutomationID = automationID
	a.currentAutomationName = automationName
	a.currentRunCancel = cancel
}

func (a *App) endRun() {
	a.currentRunMu.Lock()
	defer a.currentRunMu.Unlock()
	a.currentAutomationID = ""
	a.currentAutomationName = ""
	a.currentRunCancel = nil
}

// CancelRun cancels the in-flight automation, if any. Idempotent — calling it
// when nothing is running is a no-op and returns false.
func (a *App) CancelRun() bool {
	a.currentRunMu.Lock()
	cancel := a.currentRunCancel
	a.currentRunMu.Unlock()
	if cancel == nil {
		return false
	}
	cancel()
	return true
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func truncTarget(t string) string {
	if len(t) <= 120 {
		return t
	}
	return t[:120] + "…"
}

// ImportResult is what ImportRecording returns to the frontend.
type ImportResult struct {
	Success  bool              `json:"success"`
	Error    string            `json:"error,omitempty"`
	Draft    AutomationPayload `json:"draft"`
	Warnings []string          `json:"warnings,omitempty"`
}

// ImportRecording parses a recorder JSON export (currently Chrome DevTools
// Recorder format) and returns an unpersisted AutomationPayload draft for the
// frontend to populate the editor with.
func (a *App) ImportRecording(rawJSON, platformID string) ImportResult {
	if !safePlatformIDPattern.MatchString(platformID) {
		return ImportResult{Success: false, Error: "invalid platform ID"}
	}
	imp := automation.PickImporter([]byte(rawJSON))
	if imp == nil {
		return ImportResult{Success: false, Error: "unrecognised recorder format"}
	}
	draft, err := imp.Import([]byte(rawJSON), platformID)
	if err != nil {
		return ImportResult{Success: false, Error: err.Error()}
	}
	steps := make([]AutomationStepPayload, len(draft.Steps))
	for i, s := range draft.Steps {
		steps[i] = AutomationStepPayload{
			Action: string(s.Action), Target: s.Target,
			Value: s.Value, Label: s.Label, Selectors: s.Selectors,
		}
	}
	return ImportResult{
		Success: true,
		Draft: AutomationPayload{
			PlatformID:  platformID,
			Name:        draft.Title,
			Description: draft.Description,
			Steps:       steps,
		},
		Warnings: draft.Warnings,
	}
}

func truncSel(s automation.Step) string {
	t := s.Target
	if len(t) > 60 {
		t = t[:60] + "..."
	}
	return t
}

func countSelectors(s automation.Step) int {
	n := len(s.Selectors)
	if n == 0 && s.Target != "" {
		return 1
	}
	return n
}
