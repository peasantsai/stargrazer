# Development Setup

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.23+ | [golang.org](https://go.dev/dl/) |
| Node.js | 24+ | [nodejs.org](https://nodejs.org/) |
| Wails CLI | latest | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |

## Clone and Run

```bash
git clone https://github.com/peasantsai/stargrazer.git
cd stargrazer
wails dev
```

`wails dev` starts:

- Go backend with hot reload
- Vite dev server for frontend
- Opens the app window

## Bundled Chromium

The `assets/uc-*` directories are gitignored. For development, download Ungoogled Chromium for your platform and extract to `assets/`. The app auto-detects the binary.

## Checking Your Environment

```bash
wails doctor
```

## Building

```bash
# Standard
wails build

# Windows + NSIS installer
wails build -nsis

# Linux (Ubuntu 24.04)
wails build -tags webkit2_41

# Obfuscated
wails build -obfuscated
```
