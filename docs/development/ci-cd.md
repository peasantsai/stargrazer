# CI/CD

## Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | Push to `main`/`develop`, PRs | Lint + type-check + build verification |
| `build.yml` | Called by `release.yml` | Reusable per-platform build |
| `release.yml` | Tag push (`v*`) or manual dispatch | Full release pipeline |
| `docs.yml` | Push to `main` | Deploy MkDocs to GitHub Pages |
| `sonar.yml` | Push to `main`, PRs | SonarCloud quality analysis |

## Release Process

### Via Tag Push

```bash
git tag v1.0.0
git push origin v1.0.0
```

### Via GitHub UI

1. Go to Actions > Release > Run workflow
2. Enter version (e.g., `v1.0.0`)
3. Uncheck "Dry run" for a real release
4. Select target platforms via checkboxes

### What Happens

1. **Prepare**: Resolves version, finds previous tag, creates release branch
2. **Build**: Parallel builds per selected platform (reusable `build.yml`)
3. **Publish**: Collects artifacts, generates changelog + release notes, publishes GitHub Release

## CI Scripts

All build logic lives in `scripts/ci/`:

| Script | Purpose |
|--------|---------|
| `install-deps.sh` | Platform-specific system packages |
| `download-chromium.sh` | Download + extract Chromium |
| `clean-chromium.sh` | Strip non-essential files |
| `build.sh` | Wails build with platform flags |
| `bundle-assets.sh` | Copy Chromium into build output |
| `package.sh` | Create distributable archives |
| `sign-macos.sh` | macOS codesign + productbuild |
| `generate-changelog.sh` | Git log between tags |
| `generate-release-notes.sh` | Release notes with download table |
