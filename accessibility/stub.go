//go:build !darwin || fallback
// +build !darwin fallback

package accessibility

import (
	"fmt"
)

// CheckPermissions returns false on non-darwin
func CheckPermissions() bool {
	return false
}

// GetWindows returns empty list on non-darwin
func GetWindows() ([]WindowInfo, error) {
	fmt.Println("Note: Accessibility API is only available on macOS")
	return []WindowInfo{}, nil
}

// SetWindowPosition is a stub
func SetWindowPosition(pid int32, windowID uint32, x, y float64) error {
	return fmt.Errorf("accessibility not available on this platform")
}

// SetWindowSize is a stub
func SetWindowSize(pid int32, windowID uint32, width, height float64) error {
	return fmt.Errorf("accessibility not available on this platform")
}

// SetWindowBounds is a stub
func SetWindowBounds(pid int32, windowID uint32, bounds Rect) error {
	return fmt.Errorf("accessibility not available on this platform")
}

// MinimizeWindow is a stub
func MinimizeWindow(pid int32, windowID uint32) error {
	return fmt.Errorf("accessibility not available on this platform")
}

// UnminimizeWindow is a stub
func UnminimizeWindow(pid int32, windowID uint32) error {
	return fmt.Errorf("accessibility not available on this platform")
}

// SetFullScreen is a stub
func SetFullScreen(pid int32, windowID uint32, fullscreen bool) error {
	return fmt.Errorf("accessibility not available on this platform")
}

// GetDisplays returns main display only on non-darwin
func GetDisplays() ([]uint32, int) {
	return []uint32{0}, 1
}

// GetDisplayInfo returns empty on non-darwin
func GetDisplayInfo() ([]DisplayInfo, error) {
	return []DisplayInfo{}, nil
}

// GetDisplayBounds returns zero rect on non-darwin
func GetDisplayBounds(displayID uint32) Rect {
	return Rect{}
}

// WindowInfo is a stub
type WindowInfo struct {
	Title        string
	Bounds       Rect
	BundleID     string
	AppName      string
	ProcessID    int32
	WindowID     uint32
	DisplayID    uint32
	IsMinimized  bool
	IsFullScreen bool
}

// DisplayInfo is a stub
type DisplayInfo struct {
	ID     uint32
	Bounds Rect
	IsMain bool
}

// Rect is a stub
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}
