# Architecture Overview

Stargrazer is a [Wails v2](https://wails.io) desktop application with a Go backend and React frontend.

## Stack

| Layer | Technology |
|-------|-----------|
| Desktop framework | Wails v2.12.0 |
| Backend | Go 1.23+ |
| Frontend | React 18, TypeScript, Vite |
| Configuration | Viper (yaml + CLI + env + runtime) |
| Scheduler | robfig/cron/v3 |
| Browser | Ungoogled Chromium via CDP |
| IPC | WebSocket (CDP), Wails bindings (Go↔JS) |

## Data Flow

```
User → React UI → Wails Bindings → Go App → Browser Manager → CDP → Chromium
                                         → Scheduler → Keep-alive / Uploads
                                         → SessionStore → JSON on disk
                                         → Logger → Ring buffer + stdout
```

## Persistence

All data lives in `%APPDATA%/stargrazer/sessions/`:

```
sessions/
  accounts.json              # Platform login status
  schedules.json             # Scheduled jobs
  browser_profile/           # Shared Chromium profile
    cookies/
      facebook.json          # Exported cookies per platform
      instagram.json
```

## Singletons

Three packages use the singleton pattern:

- `browser.Manager` — One Chromium process at a time
- `config` — One Viper instance
- `scheduler.Scheduler` — One cron runner
