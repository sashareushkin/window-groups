package main

import (
	"fmt"
	"os"
	"runtime"

	"window-groups/accessibility"
	"window-groups/menu"
	"window-groups/shortcuts"
	"window-groups/window"
)

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	fmt.Println("Window Groups starting...")

	// Check Accessibility permissions
	if !accessibility.CheckPermissions() {
		fmt.Println("Warning: Accessibility permissions not granted")
		fmt.Println("Please enable in System Preferences > Security & Privacy > Privacy > Accessibility")
	}

	// Initialize window manager
	wm := window.NewManager()

	// Initialize hotkey manager
	hm := shortcuts.NewHotkeyManager(wm)

	// Initialize menu bar
	mb := menu.NewMenuBar(wm)
	if mb == nil {
		fmt.Println("Failed to initialize menu bar")
		os.Exit(1)
	}
	defer mb.Destroy()

	// Register default hotkeys for existing groups
	if err := registerDefaultHotkeys(hm, wm); err != nil {
		fmt.Printf("Warning: Could not register hotkeys: %v\n", err)
	}

	fmt.Println("Window Groups ready")
	fmt.Println("Use menu bar icon to create and restore window groups")
	fmt.Println("Assign hotkeys to groups for quick restoration")

	// Demo mode
	if len(os.Args) > 1 && os.Args[1] == "--demo" {
		runDemo(wm, hm)
	}

	// Start menu bar + AppKit run loop (blocking)
	mb.Run()
}

func registerDefaultHotkeys(hm *shortcuts.HotkeyManager, wm *window.Manager) error {
	groups := wm.GetGroups()
	
	// Default hotkeys: ⌃⌥⌘1..9,0 to reduce conflicts with app shortcuts.
	defaultKeys := []int{shortcuts.Key1, shortcuts.Key2, shortcuts.Key3, shortcuts.Key4, shortcuts.Key5, shortcuts.Key6, shortcuts.Key7, shortcuts.Key8, shortcuts.Key9, shortcuts.Key0}
	for i, group := range groups {
		if i >= len(defaultKeys) {
			break
		}

		keyCode := defaultKeys[i]
		modifiers := shortcuts.ModCommand | shortcuts.ModOption | shortcuts.ModControl

		if err := hm.RegisterShortcut(group.Name, keyCode, modifiers); err != nil {
			fmt.Printf("Warning: Could not register hotkey for %s: %v\n", group.Name, err)
		}
	}
	
	return nil
}

func runDemo(wm *window.Manager, hm *shortcuts.HotkeyManager) {
	fmt.Println("\n=== Demo Mode ===")

	// Create a test group
	group, err := wm.CreateGroup("Demo City", []string{
		"com.apple.Safari",
		"com.apple.Terminal",
	})
	if err != nil {
		fmt.Printf("Error creating group: %v\n", err)
		return
	}

	fmt.Printf("Created group: %s\n", group.Name)

	// Register a hotkey for this group
	err = hm.RegisterShortcut(group.Name, shortcuts.Key1, shortcuts.ModCommand)
	if err != nil {
		fmt.Printf("Error registering hotkey: %v\n", err)
	} else {
		fmt.Printf("Registered hotkey ⌘1 for group %s\n", group.Name)
	}

	// List groups with their shortcuts
	groups := wm.GetGroups()
	fmt.Printf("\nTotal groups: %d\n", len(groups))
	for _, g := range groups {
		shortcut := hm.GetShortcutForGroup(g.Name)
		if shortcut != nil {
			fmt.Printf("  - %s: %s\n", g.Name, shortcut.String())
		} else {
			fmt.Printf("  - %s: (no hotkey)\n", g.Name)
		}
	}

	// Try to get windows (will fail without permissions)
	windows, err := accessibility.GetWindows()
	if err != nil {
		fmt.Printf("Note: %v\n", err)
		fmt.Println("Run on macOS with Accessibility permissions to capture windows")
	} else {
		fmt.Printf("Found %d windows\n", len(windows))
	}
}
