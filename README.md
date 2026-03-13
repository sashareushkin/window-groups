# Window Groups

Window manager for macOS — remembers and restores window positions.

## Requirements

- macOS 12+
- Go 1.21+
- Xcode Command Line Tools

## Features

- **Window Capture**: Save window positions, sizes, and monitors
- **Group Management**: Create, save, and delete window groups
- **Group Restore**: Restore window groups with one click or hotkey
- **Multi-monitor Support**: Works with multiple displays
- **Global Hotkeys**: Assign keyboard shortcuts to groups
- **Menu Bar UI**: All controls in the menu bar

## Installation

### From Release

1. Download the latest release for your Mac:
   - `window-groups-darwin-amd64` for Intel Macs
   - `window-groups-darwin-arm64` for Apple Silicon (M1/M2/M3)

2. Make executable:
   ```bash
   chmod +x window-groups-darwin-*
   ```

3. Run:
   ```bash
   ./window-groups-darwin-*
   ```

### From Source

```bash
# Clone repository
git clone https://github.com/sashareushkin/window-groups.git
cd window-groups

# Build
go build -o window-groups ./cmd/main.go

# Run
./window-groups
```

## Setup

1. Grant Accessibility permissions on first launch:
   - System Preferences → Security & Privacy → Privacy → Accessibility
   - Add `window-groups` to the allowed apps list

2. Usage:
   - Click the menu bar icon
   - Select "Create Group"
   - Click windows to add to the group
   - Save the group

## Keyboard Shortcuts

- Assign hotkeys to groups for quick restoration
- Default: ⌘1, ⌘2, etc. for first 10 groups

## Project Structure

```
window-groups/
├── cmd/            # Entry point
├── app/            # Configuration and storage
├── window/         # Window management
├── menu/           # Menu bar UI
├── highlight/      # Window highlighting
├── accessibility/  # Accessibility API integration
├── shortcuts/      # Global hotkeys
└── README.md
```

## Development

### Build for release

```bash
# AMD64
GOOS=darwin GOARCH=amd64 go build -o window-groups-darwin-amd64 ./cmd/main.go

# ARM64
GOOS=darwin GOARCH=arm64 go build -o window-groups-darwin-arm64 ./cmd/main.go
```

### Testing

```bash
# Run tests
go test ./...

# Run with verbose output
go test -v ./...
```

## License

MIT
