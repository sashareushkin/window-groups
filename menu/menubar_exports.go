//go:build darwin
// +build darwin

package menu

/*
#include <stdint.h>
*/
import "C"

import (
	"fmt"
	"os"
	"runtime/cgo"
)

//export go_menu_toggle_create
func go_menu_toggle_create(handle C.uintptr_t) C.int {
	h := cgo.Handle(handle)
	mb, ok := h.Value().(*MenuBar)
	if !ok || mb == nil {
		return 0
	}

	if mb.IsSelecting() {
		if name, err := mb.SaveGroup(); err != nil {
			fmt.Printf("Save group failed: %v\n", err)
			if err.Error() == "group limit reached: maximum 10 groups" {
				fmt.Println("Cannot create more than 10 groups")
			}
			return 1
		} else {
			fmt.Printf("Group saved: %s\n", name)
		}
		return 0
	}

	if err := mb.StartWindowSelection(); err != nil {
		fmt.Printf("Start selection failed: %v\n", err)
		return 0
	}
	return 1
}

//export go_menu_restore_group
func go_menu_restore_group(handle C.uintptr_t, groupName *C.char) {
	h := cgo.Handle(handle)
	mb, ok := h.Value().(*MenuBar)
	if !ok || mb == nil || groupName == nil {
		return
	}
	name := C.GoString(groupName)
	if err := mb.RestoreGroup(name); err != nil {
		fmt.Printf("Restore group failed: %v\n", err)
	}
}

//export go_menu_delete_group
func go_menu_delete_group(handle C.uintptr_t, groupName *C.char) {
	h := cgo.Handle(handle)
	mb, ok := h.Value().(*MenuBar)
	if !ok || mb == nil || groupName == nil {
		return
	}
	name := C.GoString(groupName)
	if err := mb.DeleteGroup(name); err != nil {
		fmt.Printf("Delete group failed: %v\n", err)
	}
}

//export go_menu_add_frontmost
func go_menu_add_frontmost(handle C.uintptr_t, bundleID *C.char) {
	h := cgo.Handle(handle)
	mb, ok := h.Value().(*MenuBar)
	if !ok || mb == nil || bundleID == nil {
		return
	}
	mb.AddSelectedBundle(C.GoString(bundleID))
}

//export go_menu_toggle_window
func go_menu_toggle_window(handle C.uintptr_t, windowID C.uint, bundleID *C.char) C.int {
	h := cgo.Handle(handle)
	mb, ok := h.Value().(*MenuBar)
	if !ok || mb == nil || bundleID == nil {
		return 0
	}
	selected := mb.ToggleWindowSelection(uint32(windowID), C.GoString(bundleID))
	if selected {
		return 1
	}
	return 0
}

//export go_menu_cancel_selection
func go_menu_cancel_selection(handle C.uintptr_t) {
	h := cgo.Handle(handle)
	mb, ok := h.Value().(*MenuBar)
	if !ok || mb == nil {
		return
	}
	mb.CancelSelection()
}

//export go_menu_quit
func go_menu_quit(handle C.uintptr_t) {
	h := cgo.Handle(handle)
	mb, ok := h.Value().(*MenuBar)
	if !ok || mb == nil {
		return
	}
	mb.Destroy()
	os.Exit(0)
}
