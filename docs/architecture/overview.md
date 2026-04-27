# Architecture Overview

Stargrazer is a [Wails v2](https://wails.io) desktop application with a Go backend and React frontend.

## Stack

| Layer | Technology |
|-------|-----------|
| Desktop framework | Wails v2.12.0 |
| Backend | Go 1.23+ |
| Frontend | React 18, TypeScript, Vite |
| Configuration | Viper (yaml + CLI flags + env vars + runtime mutations) |
| Scheduler | robfig/cron/v3 |
| Browser | Ungoogled Chromium via Chrome DevTools Protocol (CDP) |
| IPC | WebSocket (CDP), Wails bindings (Go↔JS via generated `App.js`) |

## Data Flow

```
User → React UI → Wails Bindings → Go App (app.go)
                                        ├─ Browser Manager  → CDP WebSocket → Chromium
                                        ├─ Scheduler        → Keep-alive / Upload jobs
                                        ├─ SessionStore     → accounts.json on disk
                                        ├─ Automation Store → automations/<platform>.json
                                        ├─ Workflow Engine  → workflows/<platform>_upload.json
                                        └─ Logger           → Ring buffer (1000 entries) + stdout
```

## Persistence

All data lives under a platform-specific base:

- **Windows**: `%APPDATA%\stargrazer\sessions\`
- **macOS**: `~/Library/Application Support/stargrazer/sessions/`
- **Linux**: `$XDG_DATA_HOME/stargrazer/sessions/` (falls back to `~/.local/share/`)

Directory layout:

```
sessions/
  accounts.json              # Platform login status (SessionStore)
  schedules.json             # Scheduled cron jobs (Scheduler)
  browser_profile/           # Shared Chromium user-data-dir (all platforms share one profile)
    cookies/
      facebook.json          # Exported cookies per platform
      instagram.json
      ...
  data/
    uploads/                 # Upload attempt records (timestamped JSON)
  automations/
    facebook.json            # Saved automation configs per platform
    instagram.json
    ...
```

Workflow files are stored separately, relative to the executable or cwd:

```
workflows/
  facebook_upload.json
  instagram_upload.json
  tiktok_upload.json
  youtube_upload.json
  linkedin_upload.json
  x_upload.json
```

## Singletons

Four packages use the singleton pattern:

- `browser.Manager` — One Chromium process at a time (`sync.Once` + package-level `instance`)
- `config` — One Viper instance; `Get()` falls back to defaults when not initialised
- `scheduler.Scheduler` — One cron runner (`sync.Once` + package-level `instance`)
- `logger.ringBuffer` — One in-memory ring buffer for log entries

## Security

- `safePlatformIDPattern` (`^[a-z0-9_-]+$`) guards all filesystem operations that use `platformID` in paths, preventing path traversal
- Cookie files written with `0600` permissions; directories with `0700`
- Browser launched without `--password-store` or credential sync flags
- CDP port bound only on loopback (`127.0.0.1`)
