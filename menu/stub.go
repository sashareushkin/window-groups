//go:build !darwin
// +build !darwin

package menu

import (
	"fmt"

	"window-groups/window"
)

// MenuBar is a stub for non-darwin platforms
type MenuBar struct {
	wm *window.Manager
}

// NewMenuBar creates a stub menu bar (no-op on non-darwin)
func NewMenuBar(wm *window.Manager) *MenuBar {
	fmt.Println("Note: Menu bar UI is only available on macOS")
	return &MenuBar{wm: wm}
}

// Run is a stub
func (m *MenuBar) Run() {
	fmt.Println("Menu bar not available on this platform")
}

// StartWindowSelection is a stub
func (m *MenuBar) StartWindowSelection() error {
	return fmt.Errorf("menu bar not available on this platform")
}

// AddSelectedWindow is a stub
func (m *MenuBar) AddSelectedWindow(windowID uint32, bundleID string) {}

// RemoveSelectedWindow is a stub
func (m *MenuBar) RemoveSelectedWindow(windowID uint32) {}

// ToggleWindowSelection is a stub
func (m *MenuBar) ToggleWindowSelection(windowID uint32, bundleID string) {}

// SaveGroup is a stub
func (m *MenuBar) SaveGroup() (string, error) {
	return "", fmt.Errorf("menu bar not available on this platform")
}

// CancelSelection is a stub
func (m *MenuBar) CancelSelection() {}

// RestoreGroup is a stub
func (m *MenuBar) RestoreGroup(groupName string) error {
	return fmt.Errorf("menu bar not available on this platform")
}

// DeleteGroup is a stub
func (m *MenuBar) DeleteGroup(groupName string) error {
	return fmt.Errorf("menu bar not available on this platform")
}

// IsSelecting returns false
func (m *MenuBar) IsSelecting() bool {
	return false
}

// GetSelectedWindows returns empty map
func (m *MenuBar) GetSelectedWindows() map[uint32]string {
	return make(map[uint32]string)
}

// GetSelectedBundleIDs returns empty slice
func (m *MenuBar) GetSelectedBundleIDs() []string {
	return []string{}
}

// ShowGroups returns empty slice
func (m *MenuBar) ShowGroups() []string {
	return []string{}
}

// Destroy is a stub
func (m *MenuBar) Destroy() {}
