//go:build darwin
// +build darwin

package accessibility

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// CheckPermissions returns true on macOS builds.
// Runtime permission checks are handled by the OS when APIs are called.
func CheckPermissions() bool {
	return true
}

// GetWindows returns current windows list via System Events.
func GetWindows() ([]WindowInfo, error) {
	script := `tell application "System Events"
set rows to {}
repeat with p in (every application process whose background only is false)
	set pidValue to unix id of p
	set bundleValue to bundle identifier of p
	set appNameValue to name of p
	set ws to windows of p
	repeat with w in ws
		try
			set pos to position of w
			set sz to size of w
			set t to name of w
			set end of rows to ((pidValue as text) & "|" & (bundleValue as text) & "|" & (appNameValue as text) & "|" & (t as text) & "|" & (item 1 of pos as text) & "|" & (item 2 of pos as text) & "|" & (item 1 of sz as text) & "|" & (item 2 of sz as text))
		end try
	end repeat
end repeat
return rows as text
end tell`

	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("osascript GetWindows failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	text := strings.TrimSpace(string(out))
	if text == "" {
		return []WindowInfo{}, nil
	}

	// AppleScript list as text uses comma+space delimiters.
	items := strings.Split(text, ", ")
	result := make([]WindowInfo, 0, len(items))
	for _, item := range items {
		parts := strings.Split(item, "|")
		if len(parts) < 8 {
			continue
		}
		pid, _ := strconv.Atoi(parts[0])
		x, _ := strconv.ParseFloat(parts[4], 64)
		y, _ := strconv.ParseFloat(parts[5], 64)
		w, _ := strconv.ParseFloat(parts[6], 64)
		h, _ := strconv.ParseFloat(parts[7], 64)
		result = append(result, WindowInfo{
			Title:     parts[3],
			Bounds:    Rect{X: x, Y: y, Width: w, Height: h},
			BundleID:  parts[1],
			AppName:   parts[2],
			ProcessID: int32(pid),
			WindowID:  0,
			DisplayID: 0,
		})
	}

	return result, nil
}

func SetWindowPosition(pid int32, windowID uint32, x, y float64) error {
	script := fmt.Sprintf(`tell application "System Events"
set p to first application process whose unix id is %d
if (count of windows of p) is 0 then error "no windows"
set position of front window of p to {%d, %d}
end tell`, pid, int(x), int(y))
	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("set position failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func SetWindowSize(pid int32, windowID uint32, width, height float64) error {
	script := fmt.Sprintf(`tell application "System Events"
set p to first application process whose unix id is %d
if (count of windows of p) is 0 then error "no windows"
set size of front window of p to {%d, %d}
end tell`, pid, int(width), int(height))
	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("set size failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func SetWindowBounds(pid int32, windowID uint32, bounds Rect) error {
	if err := SetWindowPosition(pid, windowID, bounds.X, bounds.Y); err != nil {
		return err
	}
	return SetWindowSize(pid, windowID, bounds.Width, bounds.Height)
}

func MinimizeWindow(pid int32, windowID uint32) error {
	script := fmt.Sprintf(`tell application "System Events"
set p to first application process whose unix id is %d
if (count of windows of p) is 0 then return
set value of attribute "AXMinimized" of front window of p to true
end tell`, pid)
	_, err := exec.Command("osascript", "-e", script).CombinedOutput()
	return err
}

func UnminimizeWindow(pid int32, windowID uint32) error {
	script := fmt.Sprintf(`tell application "System Events"
set p to first application process whose unix id is %d
if (count of windows of p) is 0 then return
set value of attribute "AXMinimized" of front window of p to false
end tell`, pid)
	_, err := exec.Command("osascript", "-e", script).CombinedOutput()
	return err
}

func SetFullScreen(pid int32, windowID uint32, fullscreen bool) error {
	value := "false"
	if fullscreen {
		value = "true"
	}
	script := fmt.Sprintf(`tell application "System Events"
set p to first application process whose unix id is %d
if (count of windows of p) is 0 then return
set value of attribute "AXFullScreen" of front window of p to %s
end tell`, pid, value)
	_, err := exec.Command("osascript", "-e", script).CombinedOutput()
	return err
}

func ActivateApp(bundleID string) error {
	if strings.TrimSpace(bundleID) == "" {
		return fmt.Errorf("empty bundleID")
	}
	out, err := exec.Command("open", "-b", bundleID).CombinedOutput()
	if err != nil {
		return fmt.Errorf("open -b failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}
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
