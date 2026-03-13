//go:build darwin && !fallback
// +build darwin,!fallback

package shortcuts

/*
#cgo LDFLAGS: -framework AppKit -framework Foundation -framework Carbon

#import <AppKit/AppKit.h>
#import <Carbon/Carbon.h>

// Global hotkey callback
static OSStatus hotkeyHandler(EventHandlerCallRef nextHandler, EventRef theEvent, void *userData) {
    UInt32 hotkeyID = 0;
    GetEventParameter(theEvent, kEventParamDirectObject, typeUInt32, NULL, sizeof(hotkeyID), NULL, &hotkeyID);
    
    if (userData) {
        void (*callback)(UInt32) = (void (*)(UInt32))userData;
        callback(hotkeyID);
    }
    
    return noErr;
}

// Register global hotkey
int register_hotkey(void *callback, UInt32 modifiers, UInt32 keyCode) {
    EventHotKeyRef hotkeyRef;
    EventHotKeyID hotkeyID;
    hotkeyID.signature = 'WGRP';
    hotkeyID.id = 1;
    
    OSStatus status = RegisterEventHotKey(keyCode, modifiers, hotkeyID, 
                                          GetApplicationEventTarget(), 0, &hotkeyRef);
    return status == noErr ? 0 : -1;
}

// Unregister global hotkey
void unregister_hotkey(int refCon) {
    // Would need to store and release EventHotKeyRef
}

// Carbon modifier constants
int mod_cmd = cmdKey;
int mod_shift = shiftKey;
int mod_option = optionKey;
int mod_control = controlKey;
*/
import "C"

import (
	"fmt"
	"unsafe"

	"window-groups/window"
)

// Shortcut represents a keyboard shortcut
type Shortcut struct {
	KeyCode   int
	Modifiers int
	GroupName string
}

// HotkeyManager manages global hotkeys
type HotkeyManager struct {
	wm          *window.Manager
	shortcuts   map[int]Shortcut // hotkeyID -> shortcut
	callback    func(string)     // groupName -> callback
	nextHotkeyID int
}

// NewHotkeyManager creates a new hotkey manager
func NewHotkeyManager(wm *window.Manager) *HotkeyManager {
	return &HotkeyManager{
		wm:          wm,
		shortcuts:   make(map[int]Shortcut),
		nextHotkeyID: 1,
	}
}

// RegisterShortcut registers a global hotkey for a group
func (m *HotkeyManager) RegisterShortcut(groupName string, keyCode int, modifiers int) error {
	if m.wm == nil {
		return fmt.Errorf("window manager not set")
	}

	// Validate group exists
	group := m.wm.GetGroup(groupName)
	if group == nil {
		return fmt.Errorf("group not found: %s", groupName)
	}

	// Convert modifiers
	cModifiers := m.convertModifiers(modifiers)
	cKeyCode := C.UInt32(keyCode)

	// Register hotkey (would use Carbon API in real implementation)
	result := C.register_hotkey(nil, C.UInt32(cModifiers), cKeyCode)
	if result != 0 {
		return fmt.Errorf("failed to register hotkey")
	}

	hotkeyID := m.nextHotkeyID
	m.shortcuts[hotkeyID] = Shortcut{
		KeyCode:   keyCode,
		Modifiers: modifiers,
		GroupName: groupName,
	}
	m.nextHotkeyID++

	fmt.Printf("Registered hotkey for group %s: modifiers=%d, keyCode=%d\n", groupName, modifiers, keyCode)
	return nil
}

// UnregisterShortcut unregisters a hotkey
func (m *HotkeyManager) UnregisterShortcut(groupName string) error {
	for id, sc := range m.shortcuts {
		if sc.GroupName == groupName {
			C.unregister_hotkey(C.int(id))
			delete(m.shortcuts, id)
			fmt.Printf("Unregistered hotkey for group %s\n", groupName)
			return nil
		}
	}
	return fmt.Errorf("no hotkey found for group: %s", groupName)
}

// GetShortcuts returns all registered shortcuts
func (m *HotkeyManager) GetShortcuts() []Shortcut {
	result := make([]Shortcut, 0, len(m.shortcuts))
	for _, sc := range m.shortcuts {
		result = append(result, sc)
	}
	return result
}

// GetShortcutForGroup returns shortcut for a specific group
func (m *HotkeyManager) GetShortcutForGroup(groupName string) *Shortcut {
	for _, sc := range m.shortcuts {
		if sc.GroupName == groupName {
			return &sc
		}
	}
	return nil
}

// RestoreGroup restores a group (callback for hotkey)
func (m *HotkeyManager) RestoreGroup(groupName string) {
	if m.wm == nil {
		return
	}

	err := m.wm.RestoreGroupByName(groupName)
	if err != nil {
		fmt.Printf("Failed to restore group %s: %v\n", groupName, err)
		return
	}

	fmt.Printf("Group %s restored via hotkey\n", groupName)
}

// convertModifiers converts Go modifiers to Carbon modifiers
func (m *HotkeyManager) convertModifiers(modifiers int) int {
	result := 0
	if modifiers&1 != 0 { // Command
		result |= int(C.mod_cmd)
	}
	if modifiers&2 != 0 { // Shift
		result |= int(C.mod_shift)
	}
	if modifiers&4 != 0 { // Option
		result |= int(C.mod_option)
	}
	if modifiers&8 != 0 { // Control
		result |= int(C.mod_control)
	}
	return result
}

// Key codes for common keys
const (
	KeyA = 0
	KeyB = 11
	KeyC = 8
	KeyD = 2
	KeyE = 14
	KeyF = 3
	KeyG = 5
	KeyH = 4
	KeyI = 34
	KeyJ = 38
	KeyK = 40
	KeyL = 37
	KeyM = 46
	KeyN = 45
	KeyO = 31
	KeyP = 35
	KeyQ = 12
	KeyR = 15
	KeyS = 1
	KeyT = 17
	KeyU = 32
	KeyV = 9
	KeyW = 13
	KeyX = 7
	KeyY = 16
	KeyZ = 6

	Key0 = 29
	Key1 = 18
	Key2 = 19
	Key3 = 20
	Key4 = 21
	Key5 = 23
	Key6 = 22
	Key7 = 26
	Key8 = 28
	Key9 = 25

	KeySpace = 49
	KeyReturn = 36
	KeyTab = 48
	KeyEscape = 53

	KeyF1 = 122
	KeyF2 = 120
	KeyF3 = 99
	KeyF4 = 118
	KeyF5 = 96
	KeyF6 = 97
	KeyF7 = 98
	KeyF8 = 100
	KeyF9 = 101
	KeyF10 = 109
	KeyF11 = 103
	KeyF12 = 111
)

// Modifier constants
const (
	ModCommand = 1
	ModShift   = 2
	ModOption  = 4
	ModControl = 8
)

// KeyEvent represents a keyboard event
type KeyEvent struct {
	KeyCode   int
	Modifiers int
}

// String returns a human-readable string for the shortcut
func (s Shortcut) String() string {
	result := ""
	
	if s.Modifiers&ModControl != 0 {
		result += "⌃"
	}
	if s.Modifiers&ModOption != 0 {
		result += "⌥"
	}
	if s.Modifiers&ModShift != 0 {
		result += "⇧"
	}
	if s.Modifiers&ModCommand != 0 {
		result += "⌘"
	}
	
	result += keyCodeToString(s.KeyCode)
	
	return result
}

// keyCodeToString converts a key code to a string
func keyCodeToString(code int) string {
	keys := map[int]string{
		KeyA: "A", KeyB: "B", KeyC: "C", KeyD: "D", KeyE: "E",
		KeyF: "F", KeyG: "G", KeyH: "H", KeyI: "I", KeyJ: "J",
		KeyK: "K", KeyL: "L", KeyM: "M", KeyN: "N", KeyO: "O",
		KeyP: "P", KeyQ: "Q", KeyR: "R", KeyS: "S", KeyT: "T",
		KeyU: "U", KeyV: "V", KeyW: "W", KeyX: "X", KeyY: "Y",
		KeyZ: "Z",
		Key0: "0", Key1: "1", Key2: "2", Key3: "3", Key4: "4",
		Key5: "5", Key6: "6", Key7: "7", Key8: "8", Key9: "9",
		KeySpace: "Space", KeyReturn: "Return", KeyTab: "Tab", KeyEscape: "Esc",
		KeyF1: "F1", KeyF2: "F2", KeyF3: "F3", KeyF4: "F4",
		KeyF5: "F5", KeyF6: "F6", KeyF7: "F7", KeyF8: "F8",
		KeyF9: "F9", KeyF10: "F10", KeyF11: "F11", KeyF12: "F12",
	}
	
	if key, ok := keys[code]; ok {
		return key
	}
	
	return fmt.Sprintf("Key%d", code)
}
