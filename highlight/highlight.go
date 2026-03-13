//go:build darwin
// +build darwin

package highlight

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit -framework Quartz

#import <AppKit/AppKit.h>
#import <Quartz/Quartz.h>

void* create_highlight_manager();
void highlight_window(void *mgr, uint32_t windowID);
void remove_highlight(void *mgr, uint32_t windowID);
void clear_highlights(void *mgr);
void destroy_highlight_manager(void *mgr);
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Manager handles window highlighting
type Manager struct {
	ptr unsafe.Pointer
}

// NewManager creates a new highlight manager
func NewManager() *Manager {
	ptr := C.create_highlight_manager()
	if ptr == nil {
		return nil
	}
	return &Manager{ptr: ptr}
}

// Destroy destroys the highlight manager
func (m *Manager) Destroy() {
	if m.ptr != nil {
		C.destroy_highlight_manager(m.ptr)
		m.ptr = nil
	}
}

// HighlightWindow adds a highlight to a window
func (m *Manager) HighlightWindow(windowID uint32) {
	if m.ptr == nil {
		return
	}
	C.highlight_window(m.ptr, C.uint(windowID))
	fmt.Printf("Highlighted window: %d\n", windowID)
}

// RemoveHighlight removes highlight from a window
func (m *Manager) RemoveHighlight(windowID uint32) {
	if m.ptr == nil {
		return
	}
	C.remove_highlight(m.ptr, C.uint(windowID))
}

// ClearAllHighlights removes all highlights
func (m *Manager) ClearAllHighlights() {
	if m.ptr == nil {
		return
	}
	C.clear_highlights(m.ptr)
	fmt.Println("All highlights cleared")
}
