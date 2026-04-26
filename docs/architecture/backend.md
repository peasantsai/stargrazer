# Backend Packages

## `internal/config`

Viper-based configuration with three-tier loading:

1. Built-in defaults (stealth flags, window size, scheduler)
2. `config.yaml` file
3. CLI flags and environment variables (`STARGRAZER_` prefix)
4. Runtime updates from the frontend UI

Structs: `WindowConfig`, `BrowserConfig`, `SchedulerConfig`, `AppConfig`.

## `internal/browser`

Singleton Chromium lifecycle manager. Responsibilities:

- Start/stop browser process with `--user-data-dir` and CDP flags
- Load cookies extension and pin it in the toolbar
- CDP helpers: evaluate JS, get/set cookies, navigate, open tabs
- Parse Netscape cookie format and inject via `Network.setCookie`
- Export cookies to JSON files on disk

## `internal/social`

Platform definitions and session persistence:

- 6 platforms with login URLs, session domains, and cookie indicators
- `SessionStore` — Thread-safe JSON-backed account status
- `FindPlatform()` helper used across the app
- Shared browser profile directory for all platforms

## `internal/scheduler`

Cron-based job scheduler using `robfig/cron/v3`:

- Job types: `session_keepalive` and `upload`
- CRUD operations with JSON persistence to `schedules.json`
- Auto keep-alive creation from cookie expiry analysis
- Execution: opens tabs for keep-alive, loads workflows for uploads

## `internal/logger`

Thread-safe ring buffer (1000 entries):

- Levels: info, warn, error, debug
- Dual output: stdout + in-memory
- JSON export for the logs modal

## `internal/workflow`

Upload workflow engine:

- Step types: navigate, click, type, upload_file, wait, evaluate
- Template substitution: `{{file}}`, `{{caption}}`, `{{hashtags}}`
- JSON persistence in `workflows/` directory
- Default workflows for all 6 platforms
