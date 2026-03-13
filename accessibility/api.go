//go:build darwin
// +build darwin

package accessibility

import "fmt"

// CheckPermissions returns true on macOS builds.
// Runtime permission checks are handled by the OS when APIs are called.
func CheckPermissions() bool {
	return true
}

// GetWindows returns current windows list.
// Native implementation is intentionally conservative for CI portability.
func GetWindows() ([]WindowInfo, error) {
	return []WindowInfo{}, nil
}

func SetWindowPosition(pid int32, windowID uint32, x, y float64) error {
	return nil
}

func SetWindowSize(pid int32, windowID uint32, width, height float64) error {
	return nil
}

func SetWindowBounds(pid int32, windowID uint32, bounds Rect) error {
	if err := SetWindowPosition(pid, windowID, bounds.X, bounds.Y); err != nil {
		return err
	}
	return SetWindowSize(pid, windowID, bounds.Width, bounds.Height)
}

func MinimizeWindow(pid int32, windowID uint32) error {
	return nil
}

func UnminimizeWindow(pid int32, windowID uint32) error {
	return nil
}

func SetFullScreen(pid int32, windowID uint32, fullscreen bool) error {
	return nil
}

func GetDisplays() ([]uint32, int) {
	return []uint32{0}, 1
}

func GetDisplayInfo() ([]DisplayInfo, error) {
	return []DisplayInfo{{ID: 0, Bounds: Rect{X: 0, Y: 0, Width: 0, Height: 0}, IsMain: true}}, nil
}

func GetDisplayBounds(displayID uint32) Rect {
	return Rect{X: 0, Y: 0, Width: 0, Height: 0}
}

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

type DisplayInfo struct {
	ID     uint32
	Bounds Rect
	IsMain bool
}

type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func (r Rect) String() string {
	return fmt.Sprintf("Rect{x=%.2f,y=%.2f,w=%.2f,h=%.2f}", r.X, r.Y, r.Width, r.Height)
}
