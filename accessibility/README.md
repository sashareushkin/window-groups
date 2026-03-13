# Accessibility API Integration

This module provides integration with macOS Accessibility API (AXUIElement) for window management.

## Features

- **Permission Check**: Verify if the app has Accessibility permissions
- **Window Enumeration**: Get list of all visible windows with their properties
- **Window Manipulation**: Set window position and size

## Usage

```go
import "window-groups/accessibility"

// Check permissions
if !accessibility.CheckPermissions() {
    fmt.Println("Please enable Accessibility permissions in System Preferences")
}

// Get all windows
windows, err := accessibility.GetWindows()
if err != nil {
    log.Fatal(err)
}

for _, w := range windows {
    fmt.Printf("%s: %s at (%.0f, %.0f) size (%.0f x %.0f)\n",
        w.AppName, w.BundleID, w.Bounds.X, w.Bounds.Y, w.Bounds.Width, w.Bounds.Height)
}

// Set window position
err = accessibility.SetWindowPosition(pid, windowID, 100, 100)

// Set window size  
err = accessibility.SetWindowSize(pid, windowID, 800, 600)

// Set both position and size
err = accessibility.SetWindowBounds(pid, windowID, accessibility.Rect{
    X: 100, Y: 100, Width: 800, Height: 600,
})
```

## Permissions

To use this API, the app must be granted Accessibility permissions:

1. Open **System Preferences** → **Security & Privacy** → **Privacy** → **Accessibility**
2. Add your application to the allowed list
3. The app will request permissions on first launch

## Data Structures

### WindowInfo

```go
type WindowInfo struct {
    Title     string  // Window title (may be empty)
    Bounds    Rect    // Window position and size
    BundleID  string  // Application bundle identifier
    AppName   string  // Application name
    ProcessID int32   // Process ID
    WindowID  uint32  // Window ID
}
```

### Rect

```go
type Rect struct {
    X      float64 // X position
    Y      float64 // Y position  
    Width  float64 // Window width
    Height float64 // Window height
}
```

## Implementation Details

- Uses `CGWindowListCopyWindowInfo` for efficient window enumeration
- Uses `AXUIElement` API for window manipulation
- Requires macOS 10.6+ (tested on macOS 12+)
- Build with CGO enabled for macOS
