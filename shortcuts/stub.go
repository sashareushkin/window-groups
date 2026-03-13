//go:build !darwin || fallback
// +build !darwin fallback

package shortcuts

import (
	"fmt"

	"window-groups/window"
)

// Shortcut represents a keyboard shortcut
type Shortcut struct {
	GroupName string
	KeyCode   int
	Modifiers int
}

// Key constants (stub)
const (
	Key0 = 29
	Key1 = 18
	KeySpace = 49
)

// Modifier constants (stub)
const (
	ModCommand = 1
	ModShift   = 2
	ModOption  = 4
	ModControl = 8
)

// String returns a human-readable string for the shortcut
func (s Shortcut) String() string {
	return "N/A"
}

// HotkeyManager is a stub for non-darwin platforms
type HotkeyManager struct {
	wm *window.Manager
}

// NewHotkeyManager creates a stub hotkey manager
func NewHotkeyManager(wm *window.Manager) *HotkeyManager {
	fmt.Println("Note: Global hotkeys are only available on macOS")
	return &HotkeyManager{wm: wm}
}

// RegisterShortcut is a stub
func (m *HotkeyManager) RegisterShortcut(groupName string, keyCode int, modifiers int) error {
	return fmt.Errorf("hotkeys not available on this platform")
}

// UnregisterShortcut is a stub
func (m *HotkeyManager) UnregisterShortcut(groupName string) error {
	return fmt.Errorf("hotkeys not available on this platform")
}

// GetShortcuts returns empty slice
func (m *HotkeyManager) GetShortcuts() []Shortcut {
	return []Shortcut{}
}

// GetShortcutForGroup returns nil
func (m *HotkeyManager) GetShortcutForGroup(groupName string) *Shortcut {
	return nil
}

// RestoreGroup is a stub
func (m *HotkeyManager) RestoreGroup(groupName string) {}
