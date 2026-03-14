//go:build darwin
// +build darwin

package shortcuts

/*
#include <stdint.h>
*/
import "C"

//export go_hotkey_pressed
func go_hotkey_pressed(hotkeyID C.uint) {
	m := getActiveManager()
	if m == nil {
		return
	}
	m.HandleHotkey(int(hotkeyID))
}
