# Settings

The Settings view contains all app configuration, searchable from the top search bar.

## Social Media Connections

Platform cards showing login status. Click to connect, click (i) for session info, use Purge to disconnect.

## Browser Configuration

| Setting | Description | Default |
|---------|-------------|---------|
| CDP Port | Chrome DevTools Protocol port | 9222 |
| Chromium Path | Path to browser binary | Auto-detected |
| User Data Dir | Persistent profile directory | `%APPDATA%/stargrazer/sessions/browser_profile` |
| Headless | Run without visible window | Off |
| Window Size | Browser dimensions | 1280x900 |

## Chromium Flags

Organized in 5 collapsible categories with checkboxes:

- **Stealth & Anti-Detection** — Hide automation signals
- **Privacy & Telemetry** — Disable sync, translate, extensions
- **Automation & CDP** — Optimize for programmatic control
- **Display & UI** — Dark mode, notifications, window size
- **Network** — Certificate handling, CORS, GPU

Custom flags can be added as comma-separated values.

## Saving Settings

Click **Save Settings**. If the browser is running, it automatically restarts with the new configuration.

## Logs

Click the log icon (top right) to view application logs. Features:

- Real-time auto-refresh
- Filter by level, source, or message text
- Export as JSON
- Clear log buffer
