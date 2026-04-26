# Contributing to Stargrazer

Thank you for your interest in contributing to Stargrazer!

## Development Setup

### Prerequisites

- Go 1.23+
- Node.js 24+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Bundled Chromium in `assets/` (see below)

### Getting Started

```bash
git clone https://github.com/peasantsai/stargrazer.git
cd stargrazer
wails dev
```

The app will open with hot-reload enabled for frontend changes.

### Bundled Chromium

The Chromium browser in `assets/` is gitignored due to its size. For development, download the appropriate Ungoogled Chromium build and extract it to `assets/uc-<version>/`. The app auto-detects it on startup.

## Project Structure

| Path | Purpose |
|------|---------|
| `main.go` | Wails entry point |
| `app.go` | All backend-to-frontend bindings |
| `internal/config/` | Viper-based configuration |
| `internal/browser/` | Chromium lifecycle + CDP |
| `internal/social/` | Platform definitions + sessions |
| `internal/scheduler/` | Cron job scheduler |
| `internal/logger/` | Ring-buffer logging |
| `internal/workflow/` | Upload workflow engine |
| `frontend/src/App.tsx` | All React views and components |
| `frontend/src/styles/` | CSS theme and global styles |
| `scripts/ci/` | CI/CD shell scripts |

## Coding Standards

- **Go**: Follow SOLID principles. New domain logic goes in `internal/<package>/`, not `app.go`.
- **React**: Components live in `App.tsx`. Style with CSS variables from `theme.css`.
- **CSS**: Use vars from `[data-theme]` — never hardcode colors.
- **Commits**: Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`). Single-line, no body.

## Workflow

1. Fork and clone the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes
4. Run linting: `go vet ./...` and `cd frontend && npx tsc --noEmit`
5. Commit with conventional commit message
6. Push and open a Pull Request

## CI Pipeline

Every PR runs:
- Frontend type checking (`tsc --noEmit`)
- Go vet and build verification
- SonarCloud quality gate

## Reporting Issues

Use [GitHub Issues](https://github.com/peasantsai/stargrazer/issues) with:
- Clear title and description
- Steps to reproduce
- Expected vs actual behavior
- OS and app version
