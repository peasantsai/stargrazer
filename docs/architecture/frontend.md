# Frontend Architecture

## Stack

- **React 18** with TypeScript
- **Vite** for bundling and hot reload
- **Plain CSS** with custom properties (no CSS-in-JS)
- **Wails bindings** for Go backend calls

## Directory Structure

```
frontend/src/
├── App.tsx                # Root: view routing, browser state, platform list
├── components/
│   ├── Sidebar.tsx        # Nav links, browser status dot, theme/account toggles
│   ├── ChatPanel.tsx      # Upload form, hashtag bubbles, file picker, message log
│   ├── SchedulesPanel.tsx # Cron job cards
│   ├── ConfigPanel.tsx    # Browser settings, chromium flags, social connections, logs
│   ├── PlatformPage.tsx   # Per-platform: session info, automation builder, cookie import
│   ├── HamburgerBtn.tsx   # Sidebar toggle button
│   ├── modals/
│   │   ├── AccountModal.tsx         # Session details overlay
│   │   ├── CookiePasteModal.tsx     # Paste Netscape cookie text for import
│   │   ├── CreateScheduleModal.tsx  # New scheduled job form
│   │   ├── LogsModal.tsx            # Log viewer with filter and JSON export
│   │   └── ScheduleDetailModal.tsx  # Job stats, pause/resume, delete
│   └── settings/
│       └── SocialMediaSection.tsx   # Platform cards in the Settings view
├── hooks/
│   ├── useTheme.ts        # Toggle dark/light; persists to localStorage
│   └── useAccount.ts      # Display name, email, avatar from localStorage
├── types/index.ts         # All wire types mirroring Go structs in app.go
├── constants/
│   ├── platforms.tsx      # Platform metadata (icons, labels, IDs)
│   └── chromiumFlags.ts   # Grouped Chromium flag definitions
├── styles/
│   ├── theme.css          # CSS custom properties for dark/light themes
│   └── global.css         # All component styles
└── test/
    ├── setup.ts            # Vitest global setup (jsdom)
    ├── wailsMock.ts        # Wails runtime + binding mocks
    └── modelsMock.ts       # Fixture data for tests
```

## Views

The `View` discriminated union drives navigation in `App.tsx`:

| View value | Component | Description |
|------------|-----------|-------------|
| `'chat'` | `ChatPanel` | Upload form, message log, browser start/stop |
| `'schedules'` | `SchedulesPanel` | Cron job list; create/detail modals |
| `'config'` | `ConfigPanel` | Social connections, browser config, flags, logs |
| `'platform:<id>'` | `PlatformPage` | Session info, automation steps builder, cookie import |

## Automation Builder (`PlatformPage`)

`PlatformPage` provides per-platform functionality:

- **Session info** — login status, last login/check time, username
- **Cookie import** — paste Netscape cookie text → `ImportCookies()` → keep-alive job created automatically
- **Automation builder** — Create, edit, delete, and run named automation sequences (navigate, click, type, wait, evaluate, scroll steps)
- **Actions menu** — Open platform, check login, purge session

## Modals

All modals use the overlay pattern with click-outside dismissal:

```tsx
<div className="modal-overlay" onClick={onClose}>
  <div className="modal-content" onClick={e => e.stopPropagation()}>
    <div className="modal-header">...</div>
    <div className="modal-body">...</div>
    <div className="modal-footer">...</div>
  </div>
</div>
```

## Hooks

| Hook | Purpose |
|------|---------|
| `useTheme` | Toggle dark/light theme; persists selection to `localStorage` |
| `useAccount` | Read/write display name, email, avatar URL in `localStorage` |

## Types

All wire-format types live in `src/types/index.ts` and mirror Go struct definitions in `app.go`:

| Type | Description |
|------|-------------|
| `View` | Discriminated union of navigable views |
| `PlatformResponse` | Platform session status from `GetPlatforms()` |
| `ScheduleResponse` | Scheduled job from `GetSchedules()` |
| `BrowserConfigResponse` | Browser settings from `GetBrowserConfig()` |
| `AutomationData` / `AutomationStepData` | Automation builder model |
| `LogEntryResponse` | Log entry from `GetLogs()` |

## Styling

Two CSS files:

- `styles/theme.css` — CSS variables for `[data-theme="dark"]` and `[data-theme="light"]`
- `styles/global.css` — All component styles consuming theme variables; no inline styles or CSS-in-JS

## Wails Bindings

Auto-generated in `frontend/wailsjs/go/main/App.js`. Import and call:

```tsx
import { StartBrowser, GetPlatforms, GetAutomations, RunAutomation } from '../wailsjs/go/main/App';
const platforms = await GetPlatforms();
```

Do **not** hand-edit generated binding files.

## Testing

Tests use **Vitest** + **@testing-library/react** with a jsdom environment. Wails runtime is mocked in `test/wailsMock.ts`. Run:

```bash
cd frontend
npm test          # watch mode
npm run coverage  # coverage report
```
