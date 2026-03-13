//go:build !darwin
// +build !darwin

package shortcuts

import (
	"fmt"

	"window-groups/window"
)

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
