//go:build darwin
// +build darwin

package shortcuts

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit -framework Foundation -framework Carbon

#import <AppKit/AppKit.h>
#import <Carbon/Carbon.h>

extern void go_hotkey_pressed(unsigned int hotkeyID);

static EventHotKeyRef gHotkeyRefs[256];
static int gHotkeyHandlerInstalled = 0;

static OSStatus hotkeyHandler(EventHandlerCallRef nextHandler, EventRef theEvent, void *userData) {
    EventHotKeyID hkCom;
    OSStatus res = GetEventParameter(theEvent, kEventParamDirectObject, typeEventHotKeyID,
                                     NULL, sizeof(hkCom), NULL, &hkCom);
    if (res == noErr) {
        go_hotkey_pressed((unsigned int)hkCom.id);
    }
    return noErr;
}

static int ensure_hotkey_handler() {
    if (gHotkeyHandlerInstalled) {
        return 0;
    }

    EventTypeSpec eventType;
    eventType.eventClass = kEventClassKeyboard;
    eventType.eventKind = kEventHotKeyPressed;

    OSStatus status = InstallApplicationEventHandler(&hotkeyHandler, 1, &eventType, NULL, NULL);
    if (status != noErr) {
        return -1;
    }

    gHotkeyHandlerInstalled = 1;
    return 0;
}

int register_hotkey(UInt32 hotkeyID, UInt32 modifiers, UInt32 keyCode) {
    if (hotkeyID >= 256) {
        return -1;
    }

    if (ensure_hotkey_handler() != 0) {
        return -1;
    }

    if (gHotkeyRefs[hotkeyID] != NULL) {
        UnregisterEventHotKey(gHotkeyRefs[hotkeyID]);
        gHotkeyRefs[hotkeyID] = NULL;
    }

    EventHotKeyID hkID;
    hkID.signature = 'WGRP';
    hkID.id = hotkeyID;

    EventHotKeyRef hotkeyRef = NULL;
    OSStatus status = RegisterEventHotKey(keyCode, modifiers, hkID,
                                          GetApplicationEventTarget(), 0, &hotkeyRef);
    if (status != noErr) {
        return -1;
    }

    gHotkeyRefs[hotkeyID] = hotkeyRef;
    return 0;
}

void unregister_hotkey(UInt32 hotkeyID) {
    if (hotkeyID >= 256) {
        return;
    }

    if (gHotkeyRefs[hotkeyID] != NULL) {
        UnregisterEventHotKey(gHotkeyRefs[hotkeyID]);
        gHotkeyRefs[hotkeyID] = NULL;
    }
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
	"sync"

	"window-groups/window"
)

var (
	activeHotkeyManager *HotkeyManager
	activeMu            sync.RWMutex
)

// Shortcut represents a keyboard shortcut
type Shortcut struct {
	KeyCode   int
	Modifiers int
	GroupName string
}

// HotkeyManager manages global hotkeys
type HotkeyManager struct {
	wm           *window.Manager
	shortcuts    map[int]Shortcut // hotkeyID -> shortcut
	nextHotkeyID int
}

// NewHotkeyManager creates a new hotkey manager
func NewHotkeyManager(wm *window.Manager) *HotkeyManager {
	hm := &HotkeyManager{
		wm:           wm,
		shortcuts:    make(map[int]Shortcut),
		nextHotkeyID: 1,
	}

	activeMu.Lock()
	activeHotkeyManager = hm
	activeMu.Unlock()

	return hm
}

// RegisterShortcut registers a global hotkey for a group
func (m *HotkeyManager) RegisterShortcut(groupName string, keyCode int, modifiers int) error {
	if m.wm == nil {
		return fmt.Errorf("window manager not set")
	}

	group := m.wm.GetGroup(groupName)
	if group == nil {
		return fmt.Errorf("group not found: %s", groupName)
	}

	hotkeyID := m.nextHotkeyID
	if hotkeyID >= 256 {
		return fmt.Errorf("hotkey limit reached")
	}

	cModifiers := m.convertModifiers(modifiers)
	result := C.register_hotkey(C.UInt32(hotkeyID), C.UInt32(cModifiers), C.UInt32(keyCode))
	if result != 0 {
		return fmt.Errorf("failed to register hotkey")
	}

	m.shortcuts[hotkeyID] = Shortcut{
		KeyCode:   keyCode,
		Modifiers: modifiers,
		GroupName: groupName,
	}
	m.nextHotkeyID++

	fmt.Printf("Registered hotkey for group %s: %s\n", groupName, m.shortcuts[hotkeyID].String())
	return nil
}

// UnregisterShortcut unregisters a hotkey
func (m *HotkeyManager) UnregisterShortcut(groupName string) error {
	for id, sc := range m.shortcuts {
		if sc.GroupName == groupName {
			C.unregister_hotkey(C.UInt32(id))
			delete(m.shortcuts, id)
			fmt.Printf("Unregistered hotkey for group %s\n", groupName)
			return nil
		}
	}
	return fmt.Errorf("no hotkey found for group: %s", groupName)
}

// HandleHotkey triggers restore by hotkey id.
func (m *HotkeyManager) HandleHotkey(hotkeyID int) {
	sc, ok := m.shortcuts[hotkeyID]
	if !ok {
		return
	}
	m.RestoreGroup(sc.GroupName)
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

// getActiveManager returns current active hotkey manager
func getActiveManager() *HotkeyManager {
	activeMu.RLock()
	defer activeMu.RUnlock()
	return activeHotkeyManager
}

// convertModifiers converts Go modifiers to Carbon modifiers
func (m *HotkeyManager) convertModifiers(modifiers int) int {
	result := 0
	if modifiers&1 != 0 {
		result |= int(C.mod_cmd)
	}
	if modifiers&2 != 0 {
		result |= int(C.mod_shift)
	}
	if modifiers&4 != 0 {
		result |= int(C.mod_option)
	}
	if modifiers&8 != 0 {
		result |= int(C.mod_control)
	}
	return result
}

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

	KeySpace  = 49
	KeyReturn = 36
	KeyTab    = 48
	KeyEscape = 53

	KeyF1  = 122
	KeyF2  = 120
	KeyF3  = 99
	KeyF4  = 118
	KeyF5  = 96
	KeyF6  = 97
	KeyF7  = 98
	KeyF8  = 100
	KeyF9  = 101
	KeyF10 = 109
	KeyF11 = 103
	KeyF12 = 111
)

const (
	ModCommand = 1
	ModShift   = 2
	ModOption  = 4
	ModControl = 8
)

type KeyEvent struct {
	KeyCode   int
	Modifiers int
}

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
