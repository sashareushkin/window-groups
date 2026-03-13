package main

import (
	"fmt"
	"os"

	"window-groups/window"
	"window-groups/menu"
	"window-groups/accessibility"
	"window-groups/shortcuts"
)

func main() {
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
	_ = menu.NewMenuBar(wm)

	// Register default hotkeys for existing groups
	if err := registerDefaultHotkeys(hm, wm); err != nil {
		fmt.Printf("Warning: Could not register hotkeys: %v\n", err)
	}

	// Start menu bar (blocking)
	fmt.Println("Window Groups ready")
	fmt.Println("Use menu bar icon to create and restore window groups")
	fmt.Println("Assign hotkeys to groups for quick restoration")

	// Demo mode
	if len(os.Args) > 1 && os.Args[1] == "--demo" {
		runDemo(wm, hm)
	}

	// Keep running
	select {}
}

func registerDefaultHotkeys(hm *shortcuts.HotkeyManager, wm *window.Manager) error {
	groups := wm.GetGroups()
	
	// Default hotkey assignments based on group index
	// First group: ⌘1, second: ⌘2, etc.
	for i, group := range groups {
		if i > 9 { // Max 10 groups
			break
		}
		
		keyCode := shortcuts.Key0 + i
		modifiers := shortcuts.ModCommand
		
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
