package browser

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"stargrazer/internal/config"

	"github.com/gorilla/websocket"
)

// Status represents the browser lifecycle state.
type Status string

const (
	StatusStopped  Status = "stopped"
	StatusStarting Status = "starting"
	StatusRunning  Status = "running"
	StatusError    Status = "error"
)

const errCDPWebsocket = "CDP websocket: %w"

// Manager controls the bundled Chromium browser lifecycle.
// It is a singleton — only one instance exists app-wide.
type Manager struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	done    chan struct{}
	status  Status
	lastErr string
}

var (
	instance *Manager
	once     sync.Once
)

// GetInstance returns the singleton browser manager.
func GetInstance() *Manager {
	once.Do(func() {
		instance = &Manager{status: StatusStopped}
	})
	return instance
}

// NewManager returns the singleton browser manager.
// Kept for backward compatibility — same as GetInstance().
func NewManager() *Manager {
	return GetInstance()
}

// Start launches the bundled Chromium with CDP enabled.
func (m *Manager) Start() error {
	return m.StartWithOptions("", "")
}

// StartWithOptions launches Chromium with an optional session dir override and initial URL.
func (m *Manager) StartWithOptions(sessionDir, initialURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status == StatusRunning || m.status == StatusStarting {
		return fmt.Errorf("browser already %s", m.status)
	}

	m.status = StatusStarting
	m.lastErr = ""

	cfg := config.GetBrowser()
	if sessionDir != "" {
		cfg.UserDataDir = sessionDir
	}
	chromiumPath := m.resolveChromiumPath(cfg.ChromiumPath)

	if _, err := os.Stat(chromiumPath); err != nil {
		m.status = StatusError
		m.lastErr = fmt.Sprintf("chromium not found at: %s", chromiumPath)
		return fmt.Errorf("%s", m.lastErr)
	}

	args := m.buildArgs(cfg)
	if initialURL != "" {
		args = append(args, initialURL)
	}

	// Pin the cookies extension in the toolbar
	if cfg.UserDataDir != "" {
		m.pinExtension(cfg.UserDataDir)
	}

	m.cmd = exec.Command(chromiumPath, args...)
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	if err := m.cmd.Start(); err != nil {
		m.status = StatusError
		m.lastErr = err.Error()
		return fmt.Errorf("starting chromium: %w", err)
	}

	m.done = make(chan struct{})
	go m.waitForExit()

	if err := m.waitForCDP(cfg.CDPPort, 10*time.Second); err != nil {
		m.status = StatusError
		m.lastErr = err.Error()
		return err
	}

	m.status = StatusRunning
	return nil
}

// Stop terminates the running browser and all its child processes.
func (m *Manager) Stop() error {
	m.mu.Lock()

	if m.cmd == nil || m.cmd.Process == nil {
		m.status = StatusStopped
		m.mu.Unlock()
		return nil
	}

	pid := m.cmd.Process.Pid
	done := m.done

	// On Windows, kill the entire process tree. On other OS, kill the process group.
	if runtime.GOOS == "windows" {
		// taskkill /T kills the process tree, /F forces it
		kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid))
		kill.Run()
	} else {
		// Negative PID sends signal to the process group
		m.cmd.Process.Kill()
	}

	m.mu.Unlock()

	// Wait for the waitForExit goroutine to finish (it calls cmd.Wait)
	if done != nil {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}

	m.mu.Lock()
	m.cmd = nil
	m.status = StatusStopped
	m.mu.Unlock()

	// Small delay to let the OS free the port before a potential restart
	time.Sleep(500 * time.Millisecond)

	return nil
}

// GetStatus returns the current browser status and any error message.
func (m *Manager) GetStatus() (Status, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status, m.lastErr
}

// IsRunning returns whether the browser process is alive.
func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status == StatusRunning
}

// ResolveChromiumPath returns the auto-detected path for the bundled Chromium.
func (m *Manager) ResolveChromiumPath() string {
	return m.resolveChromiumPath("")
}

// findChromiumInAssets searches for the chromium binary in an assets directory.
func findChromiumInAssets(assetsDir, binary string) string {
	entries, err := os.ReadDir(assetsDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			candidate := filepath.Join(assetsDir, e.Name(), binary)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}
	return ""
}

func (m *Manager) resolveChromiumPath(override string) string {
	if override != "" {
		return override
	}

	var binary string
	switch runtime.GOOS {
	case "windows":
		binary = "chrome.exe"
	case "darwin":
		binary = "Chromium.app/Contents/MacOS/Chromium"
	default:
		binary = "chrome"
	}

	exe, _ := os.Executable()
	if found := findChromiumInAssets(filepath.Join(filepath.Dir(exe), "assets"), binary); found != "" {
		return found
	}

	cwd, _ := os.Getwd()
	if found := findChromiumInAssets(filepath.Join(cwd, "assets"), binary); found != "" {
		return found
	}

	return binary
}

// pinExtension writes Chrome Preferences to pin the cookies extension in the toolbar.
func (m *Manager) pinExtension(userDataDir string) {
	defaultDir := filepath.Join(userDataDir, "Default")
	os.MkdirAll(defaultDir, 0700)
	prefsPath := filepath.Join(defaultDir, "Preferences")

	// Read existing preferences or start fresh
	var prefs map[string]interface{}
	if data, err := os.ReadFile(prefsPath); err == nil {
		json.Unmarshal(data, &prefs)
	}
	if prefs == nil {
		prefs = make(map[string]interface{})
	}

	// Get the extension ID from the loaded extension path
	extID := "edoacekkjanmingkbkgjndndibhkegno" // default cookies extension ID

	// Set pinned extensions in toolbar
	extensions, ok := prefs["extensions"].(map[string]interface{})
	if !ok {
		extensions = make(map[string]interface{})
	}
	toolbar, ok := extensions["toolbar"].(map[string]interface{})
	if !ok {
		toolbar = make(map[string]interface{})
	}

	// Ensure extension is in the pinned list
	pinnedList, _ := toolbar["pinned_extensions"].([]interface{})
	found := false
	for _, id := range pinnedList {
		if id == extID {
			found = true
			break
		}
	}
	if !found {
		pinnedList = append(pinnedList, extID)
	}
	toolbar["pinned_extensions"] = pinnedList
	extensions["toolbar"] = toolbar
	prefs["extensions"] = extensions

	data, _ := json.MarshalIndent(prefs, "", "  ")
	os.WriteFile(prefsPath, data, 0600)
}

// resolveExtensionPath finds the cookies extension in the assets folder.
func (m *Manager) resolveExtensionPath() string {
	// Check relative to executable
	exe, _ := os.Executable()
	candidate := filepath.Join(filepath.Dir(exe), "assets", "cookies-extension", "1.0_0")
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate
	}
	// Fallback: relative to cwd (dev mode)
	cwd, _ := os.Getwd()
	candidate = filepath.Join(cwd, "assets", "cookies-extension", "1.0_0")
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate
	}
	return ""
}

func (m *Manager) buildArgs(cfg config.BrowserConfig) []string {
	args := []string{
		fmt.Sprintf("--remote-debugging-port=%d", cfg.CDPPort),
		"--no-first-run",
		"--no-default-browser-check",
	}

	if cfg.Headless {
		args = append(args, "--headless=new")
	}

	if cfg.UserDataDir != "" {
		args = append(args, fmt.Sprintf("--user-data-dir=%s", cfg.UserDataDir))
	}

	if cfg.WindowWidth > 0 && cfg.WindowHeight > 0 {
		args = append(args, fmt.Sprintf("--window-size=%d,%d", cfg.WindowWidth, cfg.WindowHeight))
	}

	// Load the cookies extension for reliable cookie access
	extPath := m.resolveExtensionPath()
	if extPath != "" {
		args = append(args, fmt.Sprintf("--load-extension=%s", extPath))
		args = append(args, fmt.Sprintf("--allowlisted-extension-id=edoacekkjanmingkbkgjndndibhkegno"))
	}

	args = append(args, cfg.ExtraFlags...)

	return args
}

func (m *Manager) waitForCDP(port int, timeout time.Duration) error {
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}

	return fmt.Errorf("CDP not reachable on port %d after %s", port, timeout)
}

func (m *Manager) waitForExit() {
	if m.cmd != nil {
		m.cmd.Wait()
	}

	// Signal that the process has exited
	close(m.done)

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.status == StatusRunning {
		m.status = StatusStopped
	}
}

// --- CDP helpers ---

// cdpTarget represents a Chrome DevTools target.
type cdpTarget struct {
	ID                string `json:"id"`
	Type              string `json:"type"`
	Title             string `json:"title"`
	URL               string `json:"url"`
	WebSocketDebugURL string `json:"webSocketDebuggerUrl"`
}

// GetCDPTargets lists all debuggable targets from the CDP endpoint.
func (m *Manager) GetCDPTargets() ([]cdpTarget, error) {
	port := config.GetBrowser().CDPPort
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/json", port))
	if err != nil {
		return nil, fmt.Errorf("CDP targets: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var targets []cdpTarget
	if err := json.Unmarshal(body, &targets); err != nil {
		return nil, fmt.Errorf("parsing CDP targets: %w", err)
	}
	return targets, nil
}

// cdpMessage is a generic CDP protocol message.
type cdpMessage struct {
	ID     int             `json:"id"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

// findServiceWorkerTarget finds the cookies extension's service worker CDP target.
func (m *Manager) findServiceWorkerTarget() (string, error) {
	targets, err := m.GetCDPTargets()
	if err != nil {
		return "", err
	}
	// Look for the extension's service worker
	for _, t := range targets {
		if t.Type == "service_worker" && strings.Contains(t.URL, "cookies") {
			return t.WebSocketDebugURL, nil
		}
	}
	// Fallback: any service worker
	for _, t := range targets {
		if t.Type == "service_worker" && t.WebSocketDebugURL != "" {
			return t.WebSocketDebugURL, nil
		}
	}
	return "", fmt.Errorf("no extension service worker target found")
}

// findAnyPageTarget finds any page target's websocket URL.
func (m *Manager) findAnyPageTarget() (string, error) {
	targets, err := m.GetCDPTargets()
	if err != nil {
		return "", err
	}
	for _, t := range targets {
		if t.Type == "page" && t.WebSocketDebugURL != "" {
			return t.WebSocketDebugURL, nil
		}
	}
	return "", fmt.Errorf("no page target found")
}

// cdpEval connects to a target and evaluates a JS expression, returning raw JSON result.
func (m *Manager) cdpEval(wsURL, expression string) (json.RawMessage, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf(errCDPWebsocket, err)
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	params, _ := json.Marshal(map[string]interface{}{
		"expression":    expression,
		"awaitPromise":  true,
		"returnByValue": true,
	})
	msg := cdpMessage{ID: 1, Method: "Runtime.evaluate", Params: params}
	if err := conn.WriteJSON(msg); err != nil {
		return nil, fmt.Errorf("sending evaluate: %w", err)
	}

	for {
		var resp cdpMessage
		if err := conn.ReadJSON(&resp); err != nil {
			return nil, fmt.Errorf("reading evaluate response: %w", err)
		}
		if resp.ID == 1 && resp.Result != nil {
			return resp.Result, nil
		}
	}
}

// GetCookiesForDomains uses the cookies extension via CDP Runtime.evaluate
// to call chrome.cookies.getAll for each domain. Falls back to page-level JS.
func (m *Manager) GetCookiesForDomains(domains []string) ([]CDPCookie, error) {
	// Try extension service worker first (has chrome.cookies API)
	wsURL, err := m.findServiceWorkerTarget()
	if err == nil {
		return m.getCookiesViaExtension(wsURL, domains)
	}

	// Fallback: use page target with document.cookie (less reliable but works)
	pageWS, err := m.findAnyPageTarget()
	if err != nil {
		return nil, fmt.Errorf("no CDP target available: %w", err)
	}
	return m.getCookiesViaPage(pageWS, domains)
}

// getCookiesViaExtension uses chrome.cookies.getAll in the extension context.
func (m *Manager) getCookiesViaExtension(wsURL string, domains []string) ([]CDPCookie, error) {
	var allCookies []CDPCookie

	for _, domain := range domains {
		// Build JS to call chrome.cookies.getAll and return JSON
		js := fmt.Sprintf(
			`(async () => {
				const cookies = await chrome.cookies.getAll({domain: %q});
				return JSON.stringify(cookies);
			})()`,
			domain,
		)

		raw, err := m.cdpEval(wsURL, js)
		if err != nil {
			continue
		}

		// Parse the Runtime.evaluate result structure
		var evalResult struct {
			Result struct {
				Value string `json:"value"`
			} `json:"result"`
		}
		if err := json.Unmarshal(raw, &evalResult); err != nil {
			continue
		}

		var cookies []CDPCookie
		if err := json.Unmarshal([]byte(evalResult.Result.Value), &cookies); err != nil {
			continue
		}
		allCookies = append(allCookies, cookies...)
	}

	return allCookies, nil
}

// getCookiesViaPage uses document.cookie on a page target (fallback).
func (m *Manager) getCookiesViaPage(wsURL string, domains []string) ([]CDPCookie, error) {
	raw, err := m.cdpEval(wsURL, `document.cookie`)
	if err != nil {
		return nil, err
	}

	var evalResult struct {
		Result struct {
			Value string `json:"value"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &evalResult); err != nil {
		return nil, err
	}

	// Parse "name=value; name2=value2" format
	var cookies []CDPCookie
	for _, pair := range strings.Split(evalResult.Result.Value, ";") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			cookies = append(cookies, CDPCookie{Name: parts[0], Value: parts[1]})
		}
	}
	return cookies, nil
}

// CDPCookie represents a browser cookie from CDP.
type CDPCookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	HTTPOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
}

// ParseNetscapeCookies parses Netscape/curl cookie.txt format into CDPCookies.
// Format: domain\tincludeSubdomains\tpath\tsecure\texpiry\tname\tvalue
func ParseNetscapeCookies(text string) []CDPCookie {
	var cookies []CDPCookie
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}
		expires, _ := strconv.ParseFloat(fields[4], 64)
		secure := strings.EqualFold(fields[3], "TRUE")
		cookies = append(cookies, CDPCookie{
			Domain:  fields[0],
			Path:    fields[2],
			Secure:  secure,
			Expires: expires,
			Name:    fields[5],
			Value:   fields[6],
		})
	}
	return cookies
}

// waitForCDPResponse reads CDP messages until the response with the given ID arrives.
func waitForCDPResponse(conn *websocket.Conn, id int) error {
	for {
		var resp cdpMessage
		if err := conn.ReadJSON(&resp); err != nil {
			return err
		}
		if resp.ID == id {
			return nil
		}
	}
}

// SetCookiesViaCDP injects cookies into the browser via CDP Network.setCookie.
func (m *Manager) SetCookiesViaCDP(cookies []CDPCookie) error {
	wsURL, err := m.findAnyPageTarget()
	if err != nil {
		return err
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf(errCDPWebsocket, err)
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))

	for i, c := range cookies {
		domain := c.Domain
		// Build a URL for setCookie — required by CDP
		scheme := "https"
		if !c.Secure {
			scheme = "http"
		}
		host := domain
		if strings.HasPrefix(host, ".") {
			host = "www" + host
		}
		url := fmt.Sprintf("%s://%s%s", scheme, host, c.Path)

		params, _ := json.Marshal(map[string]interface{}{
			"name":     c.Name,
			"value":    c.Value,
			"domain":   domain,
			"path":     c.Path,
			"secure":   c.Secure,
			"httpOnly": c.HTTPOnly,
			"url":      url,
			"expires":  c.Expires,
		})
		msg := cdpMessage{ID: i + 1, Method: "Network.setCookie", Params: params}
		if err := conn.WriteJSON(msg); err != nil {
			return fmt.Errorf("setting cookie %s: %w", c.Name, err)
		}

		if err := waitForCDPResponse(conn, i+1); err != nil {
			return fmt.Errorf("reading setCookie response for %s: %w", c.Name, err)
		}
	}
	return nil
}

// NavigateToURL opens a URL in the first available page target via CDP.
func (m *Manager) NavigateToURL(targetURL string) error {
	targets, err := m.GetCDPTargets()
	if err != nil {
		return err
	}

	var wsURL string
	for _, t := range targets {
		if t.Type == "page" && t.WebSocketDebugURL != "" {
			wsURL = t.WebSocketDebugURL
			break
		}
	}
	if wsURL == "" {
		return fmt.Errorf("no page target found for CDP")
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf(errCDPWebsocket, err)
	}
	defer conn.Close()

	params, _ := json.Marshal(map[string]string{"url": targetURL})
	msg := cdpMessage{ID: 1, Method: "Page.navigate", Params: params}
	if err := conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("sending navigate: %w", err)
	}

	var resp cdpMessage
	return conn.ReadJSON(&resp)
}

// OpenNewTab creates a new tab via CDP and navigates it to the given URL.
// Returns the new target ID.
func (m *Manager) OpenNewTab(targetURL string) (string, error) {
	port := config.GetBrowser().CDPPort
	// Use the /json/new endpoint to create a new tab
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/json/new?%s", port, url.QueryEscape(targetURL)))
	if err != nil {
		return "", fmt.Errorf("creating new tab: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var target cdpTarget
	if err := json.Unmarshal(body, &target); err != nil {
		return "", fmt.Errorf("parsing new tab response: %w", err)
	}
	return target.ID, nil
}

// HasCookies checks if specific cookie names exist for the given domains.
func (m *Manager) HasCookies(domains, cookieNames []string) (bool, string, error) {
	cookies, err := m.GetCookiesForDomains(domains)
	if err != nil {
		return false, "", err
	}

	nameSet := make(map[string]bool)
	for _, n := range cookieNames {
		nameSet[n] = true
	}

	found := 0
	for _, c := range cookies {
		if nameSet[c.Name] {
			found++
		}
	}

	// If we found any of the expected login cookies, consider logged in
	loggedIn := found > 0

	// Try to extract a username from common cookie patterns
	username := ""
	for _, c := range cookies {
		switch c.Name {
		case "c_user", "ds_user_id", "login_email":
			if username == "" {
				username = c.Value
			}
		}
	}

	return loggedIn, username, nil
}

// ExportCookiesToDisk saves cookies for the given domains to a JSON file
// in the user data directory for persistence across sessions.
func (m *Manager) ExportCookiesToDisk(domains []string, platform string, dataDir string) error {
	cookies, err := m.GetCookiesForDomains(domains)
	if err != nil {
		return fmt.Errorf("getting cookies: %w", err)
	}
	if len(cookies) == 0 {
		return nil
	}

	cookiesDir := filepath.Join(dataDir, "cookies")
	os.MkdirAll(cookiesDir, 0700)

	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cookies: %w", err)
	}

	filePath := filepath.Join(cookiesDir, platform+".json")
	return os.WriteFile(filePath, data, 0600)
}

// LoadCookiesFromDisk reads previously exported cookies for a platform.
func LoadCookiesFromDisk(platform, dataDir string) ([]CDPCookie, error) {
	filePath := filepath.Join(dataDir, "cookies", platform+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var cookies []CDPCookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return nil, err
	}
	return cookies, nil
}
