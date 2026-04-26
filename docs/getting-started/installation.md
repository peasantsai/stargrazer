# Installation

## Download

Download the latest release from [GitHub Releases](https://github.com/peasantsai/stargrazer/releases).

| Platform | Architecture | File |
|----------|-------------|------|
| Windows | x64 | `stargrazer-windows-amd64-setup.exe` (installer) or `.zip` (portable) |
| Windows | ARM64 | `stargrazer-windows-arm64-setup.exe` |
| Linux | x64 | `stargrazer-linux-amd64.tar.gz` |
| Linux | ARM64 | `stargrazer-linux-arm64.tar.gz` |
| macOS | Apple Silicon | `stargrazer-macos-arm64.zip` |
| macOS | Intel | `stargrazer-macos-amd64.zip` |

## Windows

### Installer

Run the `.exe` installer. It installs to `Program Files` and creates desktop/start menu shortcuts.

!!! note "WebView2 Runtime"
    The installer checks for WebView2 and offers to install it if missing (required on Windows 10).

### Portable

Extract the `.zip` to any folder and run `stargrazer.exe`.

## Linux

```bash
tar xzf stargrazer-linux-amd64.tar.gz
cd stargrazer
./stargrazer
```

!!! info "Dependencies"
    You may need GTK3 and WebKitGTK:
    ```bash
    # Ubuntu/Debian
    sudo apt install libgtk-3-0 libwebkit2gtk-4.1-0 gstreamer1.0-plugins-good

    # Fedora
    sudo dnf install gtk3 webkit2gtk4.1 gstreamer1-plugins-good
    ```

## macOS

Extract the `.zip` and move `Stargrazer.app` to `/Applications`.

!!! warning "First launch"
    macOS may block the app on first run. Go to **System Preferences > Privacy & Security** and click **Open Anyway**.

## Build from Source

See [Development Setup](../development/setup.md).
