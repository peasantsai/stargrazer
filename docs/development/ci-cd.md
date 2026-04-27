# CI/CD

## Workflows

| Workflow | File | Trigger | Purpose |
|----------|------|---------|---------|
| CI | `ci.yml` | Push to `main`, PRs | Lint, type-check, Go tests, frontend tests, build verification |
| Build | `build.yml` | Called by `release.yml` | Reusable per-platform Wails build |
| Release | `release.yml` | Tag push (`v*`) or manual dispatch | Full 6-platform release pipeline |
| Docs | `docs.yml` | Push to `main` | Deploy MkDocs to GitHub Pages |
| SBOM | `sbom.yml` | Push to `main`, PRs | Generate and publish Software Bill of Materials |

## Release Process

### Via Tag Push

```bash
git tag v1.0.0
git push origin v1.0.0
```

### Via GitHub UI

1. Go to **Actions → Release → Run workflow**
2. Enter version (e.g., `v1.0.0`)
3. Uncheck **Dry run** for a real release
4. Select target platforms via checkboxes

### What Happens

1. **Prepare**: Resolves version, finds previous tag, creates `release/v<version>` branch
2. **Build** (parallel): Builds for each selected platform via reusable `build.yml`
   - Linux x64 / ARM64 (with UPX compression)
   - Windows x64 / ARM64 (NSIS installer)
   - macOS Intel / Apple Silicon (codesigned + `productbuild`)
3. **Publish**: Collects artifacts, generates changelog + release notes, publishes GitHub Release

## Target Platforms

| Platform | Binary | Packaging |
|----------|--------|-----------|
| Linux x64 | ELF | `.tar.gz` (UPX compressed) |
| Linux ARM64 | ELF | `.tar.gz` (UPX compressed) |
| Windows x64 | PE | NSIS `.exe` installer |
| Windows ARM64 | PE | NSIS `.exe` installer |
| macOS Intel | Mach-O | `.pkg` via `productbuild` |
| macOS Apple Silicon | Mach-O | `.pkg` via `productbuild` |

## CI Scripts

All build logic lives in `scripts/ci/`:

| Script | Purpose |
|--------|---------|
| `install-deps.sh` | Platform-specific system packages |
| `download-chromium.sh` | Download + extract Ungoogled Chromium |
| `clean-chromium.sh` | Strip non-essential Chromium files |
| `build.sh` | Wails build with platform-specific flags |
| `bundle-assets.sh` | Copy Chromium into build output |
| `package.sh` | Create distributable archives |
| `sign-macos.sh` | macOS codesign + productbuild |
| `generate-changelog.sh` | Git log between tags |
| `generate-release-notes.sh` | Release notes with download table |

## Chromium Caching

Chromium binaries are cached per-platform in GitHub Actions using `actions/cache` keyed on a hash of `download-chromium.sh`. This avoids re-downloading (~200 MB) on every run.

## Running Tests Locally

```bash
# Go backend tests
cd stargrazer
go test ./...

# Frontend tests
cd stargrazer/frontend
npm test

# Frontend coverage
npm run coverage
```
