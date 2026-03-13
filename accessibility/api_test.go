package accessibility

import (
	"testing"
)

func TestCheckPermissions(t *testing.T) {
	// This will return false on Linux, true on macOS with permissions
	result := CheckPermissions()
	t.Logf("Accessibility permissions: %v", result)
}

func TestGetWindows(t *testing.T) {
	if !CheckPermissions() {
		t.Skip("Accessibility permissions not granted")
	}

	windows, err := GetWindows()
	if err != nil {
		t.Fatalf("GetWindows failed: %v", err)
	}

	t.Logf("Found %d windows", len(windows))
	
	for _, w := range windows {
		t.Logf("Window: %s (PID: %d, BID: %s) at (%.0f, %.0f) size (%.0f x %.0f)",
			w.AppName, w.ProcessID, w.BundleID, w.Bounds.X, w.Bounds.Y, w.Bounds.Width, w.Bounds.Height)
	}
}

func TestWindowInfoFields(t *testing.T) {
	// Test that WindowInfo has all required fields
	w := WindowInfo{
		Title:     "Test Window",
		Bounds:    Rect{X: 100, Y: 100, Width: 800, Height: 600},
		BundleID:  "com.apple.Safari",
		AppName:   "Safari",
		ProcessID: 1234,
		WindowID:  5678,
	}

	if w.Title != "Test Window" {
		t.Errorf("Expected title 'Test Window', got '%s'", w.Title)
	}
	if w.BundleID != "com.apple.Safari" {
		t.Errorf("Expected bundle ID 'com.apple.Safari', got '%s'", w.BundleID)
	}
	if w.Bounds.Width != 800 {
		t.Errorf("Expected width 800, got %f", w.Bounds.Width)
	}
}
