# Backend Packages

## `internal/config`

Viper-based configuration with four-tier loading:

1. Built-in defaults (stealth flags, window size, scheduler)
2. `config.yaml` file (`.` and `$HOME/.stargrazer` search paths)
3. CLI flags (`--cdp-port`, `--headless`, `--chromium-path`) and environment variables (`STARGRAZER_` prefix)
4. Runtime mutations via `Update()` from the frontend UI

Structs: `WindowConfig`, `BrowserConfig`, `SchedulerConfig`, `AppConfig`.

Accessors: `Get()`, `GetBrowser()`, `GetWindow()`, `GetScheduler()`, `Update()`, `Reset()`.

## `internal/browser`

Singleton Chromium lifecycle manager (`Manager`). Responsibilities:

- Start/stop browser process with `--user-data-dir`, CDP flags, and stealth flags
- Load the bundled cookies extension and pin it in the Chrome toolbar via `Preferences`
- CDP helpers over WebSocket:
  - `GetCookiesForDomains()` — reads cookies via extension service worker or page fallback
  - `SetCookiesViaCDP()` — injects cookies via `Network.setCookie`
  - `NavigateToURL()` — `Page.navigate` on the first page target
  - `OpenNewTab()` — creates a tab via `/json/new`
  - `ClickElement()`, `TypeText()`, `EvaluateExpression()`, `ScrollToElement()` — automation primitives
- Parse Netscape cookie format (`ParseNetscapeCookies`)
- Export/load cookies to/from JSON on disk (`ExportCookiesToDisk`, `LoadCookiesFromDisk`)
- Resolve bundled Chromium binary automatically (exe-relative and cwd-relative `assets/` search)

## `internal/social`

Platform definitions and session persistence:

- 6 platforms: Facebook, Instagram, TikTok, YouTube, LinkedIn, X
- Each `PlatformInfo` has login URL, session domains, and login cookie names
- `SessionStore` — thread-safe, JSON-backed account status (`accounts.json`)
- `SharedSessionDir()` — single shared Chromium profile for all platforms
- `EnsureSessionDir()` — creates profile directory on first use
- `FindPlatform()` helper used by `app.go`, `scheduler`, and `keepalive`

## `internal/scheduler`

Cron-based job scheduler using `robfig/cron/v3`:

- Job types: `session_keepalive` and `upload`
- CRUD with JSON persistence to `schedules.json`
- `EnsureKeepAlive()` — auto-creates a keep-alive job on cookie import; derives cron interval from cookie expiry (75% of shortest, clamped to 12 h – 7 d)
- Keep-alive execution (`keepalive.go`): auto-starts browser if needed, opens a tab, waits, re-exports cookies; auto-stops browser when done
- Upload execution (stub in `keepalive.go`): auto-starts browser; logs intent; full orchestration via `TriggerUpload` in `app.go`
- `registerJob` / `unregisterJob` maintain `cron.EntryID` per job

## `internal/logger`

Thread-safe ring buffer (1 000 entries max):

- Levels: `info`, `warn`, `error`, `debug`
- Dual output: stdout + in-memory ring buffer
- `GetAll()` returns entries in chronological order (handles wrap-around)
- `Export()` returns JSON bytes for the logs modal
- `Clear()` flushes the buffer

## `internal/workflow`

Upload workflow engine:

- Step types: `navigate`, `click`, `type`, `upload_file`, `wait`, `wait_navigation`, `evaluate`
- Template substitution in step values: `{{file}}`, `{{caption}}`, `{{hashtags}}`
- `PrepareSteps()` — replaces placeholders without mutating originals
- JSON persistence in `workflows/<platform>_upload.json` (resolved relative to exe or cwd)
- `DefaultWorkflows()` — built-in steps for all 6 platforms

## `internal/automation`

Per-platform, user-defined browser automation sequences:

- Step actions: `navigate`, `click`, `type`, `wait`, `evaluate`, `scroll`
- `Config` — named automation with ordered `[]Step`; tracks `RunCount` and `LastRun`
- `Store` — thread-safe JSON persistence; one file per platform (`automations/<platformID>.json`)
- `Save()` — creates (assigns UUID) or updates (replaces by ID)
- `Delete()` — removes by ID; returns `false` when not found
- `RecordRun()` — increments `RunCount` and sets `LastRun`
- Executed step-by-step via CDP from `app.go`:`executeStep()`

## `app.go` (Wails bindings surface)

All methods callable from the frontend via Wails bindings:

| Group | Methods |
|-------|---------|
| Browser | `StartBrowser`, `StopBrowser`, `GetBrowserStatus`, `RestartBrowser` |
| Config | `GetBrowserConfig`, `UpdateBrowserConfig`, `ResetBrowserConfig` |
| Social | `GetPlatforms`, `OpenPlatform`, `CheckLoginStatus`, `CheckAllLoginStatus`, `PurgeSession`, `ImportCookies` |
| Schedules | `GetSchedules`, `CreateSchedule`, `UpdateSchedule`, `DeleteSchedule`, `PauseSchedule`, `ResumeSchedule`, `GetScheduleStats` |
| Logs | `GetLogs`, `ExportLogs`, `ClearLogs` |
| Upload | `SelectFile`, `TriggerUpload` |
| Automation | `GetAutomations`, `SaveAutomation`, `DeleteAutomation`, `RunAutomation` |
