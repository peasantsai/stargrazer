# Stargrazer

[![CI](https://github.com/peasantsai/stargrazer/actions/workflows/ci.yml/badge.svg)](https://github.com/peasantsai/stargrazer/actions/workflows/ci.yml)
[![Release](https://github.com/peasantsai/stargrazer/actions/workflows/release.yml/badge.svg)](https://github.com/peasantsai/stargrazer/actions/workflows/release.yml)
[![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=peasantsai_stargrazer&metric=alert_status)](https://sonarcloud.io/dashboard?id=peasantsai_stargrazer)

A desktop application for social media automation, built with Go, React, and a bundled Ungoogled Chromium browser controlled via the Chrome DevTools Protocol (CDP).

## Features

- **Bundled stealth browser** — Ungoogled Chromium with anti-detection flags, no telemetry
- **6 platforms** — Facebook, Instagram, TikTok, YouTube, LinkedIn, X
- **Cookie-based auth** — Paste Netscape cookies from the pinned extension, sessions persist across restarts
- **Auto keep-alive** — Scheduled jobs refresh sessions before cookies expire
- **Content uploads** — File + caption + hashtags to multiple platforms at once
- **Upload workflows** — JSON-defined CDP step sequences, customizable per platform
- **Job scheduler** — Cron-based scheduling for keep-alive and uploads
- **Dark/light themes** — System-aware with manual toggle
- **Cross-platform** — Windows, Linux, macOS (x64 and ARM64)

## Quick Start

```bash
# Prerequisites: Go 1.23+, Node.js 24+, Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone and run
git clone https://github.com/peasantsai/stargrazer.git
cd stargrazer
wails dev
```

## Documentation

Full docs at [peasantsai.github.io/stargrazer](https://peasantsai.github.io/stargrazer).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).

## License

MIT
