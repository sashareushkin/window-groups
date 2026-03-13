//go:build !darwin
// +build !darwin

package highlight

import (
	"fmt"
)

// Manager is a stub for non-darwin platforms
type Manager struct{}

// NewManager creates a stub manager (no-op on non-darwin)
func NewManager() *Manager {
	fmt.Println("Note: Window highlighting is only available on macOS")
	return &Manager{}
}

// Destroy is a stub
func (m *Manager) Destroy() {}

// HighlightWindow is a stub
func (m *Manager) HighlightWindow(windowID uint32) {
	fmt.Printf("Highlight not available on this platform: window %d\n", windowID)
}

// RemoveHighlight is a stub
func (m *Manager) RemoveHighlight(windowID uint32) {}

// ClearAllHighlights is a stub
func (m *Manager) ClearAllHighlights() {
	fmt.Println("Highlights cleared (no-op)")
}
