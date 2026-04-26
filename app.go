package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
	sessions    *social.SessionStore
	scheduler   *scheduler.Scheduler
	automations *automation.Store
}

func NewApp() *App {
	b := browser.GetInstance()
	s := social.NewSessionStore()
	return &App{
		getCtx:      func() context.Context { return context.Background() },
		browser:     b,
		sessions:    s,
		scheduler:   scheduler.GetInstance(b, s),
		automations: automation.NewStore(social.SharedSessionDir()),
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
	Action string `json:"action"`
	Target string `json:"target"`
	Value  string `json:"value"`
	Label  string `json:"label"`
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
			Value: s.Value, Label: s.Label,
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

// RunAutomation executes a saved automation step-by-step via CDP.
func (a *App) RunAutomation(platformID, id string) RunAutomationResponse {
	if !safePlatformIDPattern.MatchString(platformID) {
		return RunAutomationResponse{Success: false, Message: "invalid platform ID"}
	}
	if !a.browser.IsRunning() {
		return RunAutomationResponse{Success: false, Message: "browser is not running — start it first"}
	}
	configs, err := a.automations.List(platformID)
	if err != nil {
		return RunAutomationResponse{Success: false, Message: err.Error()}
	}
	var cfg *automation.Config
	for i := range configs {
		if configs[i].ID == id {
			cfg = &configs[i]
			break
		}
	}
	if cfg == nil {
		return RunAutomationResponse{Success: false, Message: "automation not found"}
	}
	logger.Info("automation", fmt.Sprintf("Running %q (%d steps) for %s", cfg.Name, len(cfg.Steps), platformID))
	for i, step := range cfg.Steps {
		if err := a.executeStep(step); err != nil {
			msg := fmt.Sprintf("step %d (%s %q): %v", i+1, step.Action, step.Label, err)
			logger.Error("automation", msg)
			return RunAutomationResponse{Success: false, Message: msg}
		}
	}
	if err := a.automations.RecordRun(platformID, id); err != nil {
		logger.Warn("automation", fmt.Sprintf("RecordRun: %v", err))
	}
	return RunAutomationResponse{Success: true, Message: fmt.Sprintf("'%s' completed (%d steps)", cfg.Name, len(cfg.Steps))}
}

// executeStep dispatches a single automation step to the appropriate browser method.
func (a *App) executeStep(step automation.Step) error {
	switch step.Action {
	case automation.ActionNavigate:
		return a.browser.NavigateToURL(step.Target)
	case automation.ActionClick:
		return a.browser.ClickElement(step.Target)
	case automation.ActionType:
		return a.browser.TypeText(step.Target, step.Value)
	case automation.ActionWait:
		ms := 1000
		if n, err := strconv.Atoi(step.Value); err == nil && n > 0 {
			ms = n
		}
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return nil
	case automation.ActionEvaluate:
		_, err := a.browser.EvaluateExpression(step.Value)
		return err
	case automation.ActionScroll:
		return a.browser.ScrollToElement(step.Target)
	default:
		return fmt.Errorf("unknown action: %q", step.Action)
	}
}
