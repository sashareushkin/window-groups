package window

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"window-groups/accessibility"
)

// Window represents a single window
type Window struct {
	ID           string  `json:"id"`
	App          string  `json:"app"`
	Title        string  `json:"title"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	Monitor      int     `json:"monitor"`
	BundleID     string  `json:"bundle_id"`
	ProcessID    int     `json:"process_id"`
	WindowID     uint32  `json:"window_id"`
	IsMinimized  bool    `json:"is_minimized"`
	IsFullScreen bool    `json:"is_fullscreen"`
}

// Group represents a window group
type Group struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Windows     []Window  `json:"windows"`
	CreatedAt   time.Time `json:"created_at"`
	BundleIDs   []string  `json:"bundle_ids"`
}

// Manager handles window operations
type Manager struct {
	groupsPath string
	groups     []Group
}

// NewManager creates a new window manager
func NewManager() *Manager {
	// Default to user's Application Support directory
	groupsPath := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "WindowGroups", "groups.json")

	m := &Manager{
		groupsPath: groupsPath,
		groups:     make([]Group, 0),
	}

	// Load existing groups
	m.loadGroups()

	return m
}

// loadGroups loads groups from disk
func (m *Manager) loadGroups() {
	data, err := os.ReadFile(m.groupsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Warning: Could not load groups: %v\n", err)
		}
		return
	}

	if err := json.Unmarshal(data, &m.groups); err != nil {
		fmt.Printf("Warning: Could not parse groups: %v\n", err)
	}
}

// saveGroups saves groups to disk
func (m *Manager) saveGroups() error {
	// Ensure directory exists
	dir := filepath.Dir(m.groupsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create groups directory: %w", err)
	}

	data, err := json.MarshalIndent(m.groups, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal groups: %w", err)
	}

	if err := os.WriteFile(m.groupsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write groups: %w", err)
	}

	return nil
}

// CaptureWindows captures all windows in current workspace
func (m *Manager) CaptureWindows() ([]Window, error) {
	fmt.Println("Capturing windows...")

	windows, err := accessibility.GetWindows()
	if err != nil {
		return nil, fmt.Errorf("failed to get windows: %w", err)
	}

	result := make([]Window, 0, len(windows))
	for _, w := range windows {
		result = append(result, Window{
			ID:           fmt.Sprintf("%d-%d", w.ProcessID, w.WindowID),
			App:          w.AppName,
			Title:        w.Title,
			X:            w.Bounds.X,
			Y:            w.Bounds.Y,
			Width:        w.Bounds.Width,
			Height:       w.Bounds.Height,
			BundleID:     w.BundleID,
			ProcessID:    int(w.ProcessID),
			WindowID:     w.WindowID,
			Monitor:      int(w.DisplayID),
			IsMinimized:  w.IsMinimized,
			IsFullScreen: w.IsFullScreen,
		})
	}

	fmt.Printf("Captured %d windows\n", len(result))
	return result, nil
}

// CreateGroup creates a new window group from bundle IDs with current window positions
func (m *Manager) CreateGroup(name string, bundleIDs []string) (*Group, error) {
	// Capture current windows first
	windows, err := m.CaptureWindows()
	if err != nil {
		return nil, fmt.Errorf("failed to capture windows: %w", err)
	}

	// Filter windows to only include those with matching bundle IDs
	var groupWindows []Window
	for _, w := range windows {
		for _, bid := range bundleIDs {
			if w.BundleID == bid {
				groupWindows = append(groupWindows, w)
				break
			}
		}
	}

	group := Group{
		ID:        generateGroupID(),
		Name:      name,
		BundleIDs: bundleIDs,
		Windows:   groupWindows,
		CreatedAt: time.Now(),
	}

	m.groups = append(m.groups, group)

	if err := m.saveGroups(); err != nil {
		return nil, err
	}

	fmt.Printf("Group created: %s (%s) with %d windows\n", name, group.ID, len(groupWindows))

	return &group, nil
}

// RestoreGroup restores a window group
func (m *Manager) RestoreGroup(groupID string) error {
	group := m.GetGroup(groupID)
	if group == nil {
		return fmt.Errorf("group not found: %s", groupID)
	}

	fmt.Printf("Restoring group: %s (%s)\n", group.Name, groupID)

	// Get current windows for matching
	currentWindows, err := accessibility.GetWindows()
	if err != nil {
		return fmt.Errorf("failed to get current windows: %w", err)
	}

	// Build map of bundle ID -> current windows
	windowMap := make(map[string][]accessibility.WindowInfo)
	for _, w := range currentWindows {
		if w.BundleID != "" {
			windowMap[w.BundleID] = append(windowMap[w.BundleID], w)
		}
	}

	// Process each bundle ID in the group
	for _, bundleID := range group.BundleIDs {
		windows, exists := windowMap[bundleID]

		if !exists || len(windows) == 0 {
			// По текущему продукт-решению: не запускать закрытые приложения автоматически.
			fmt.Printf("Skip %s: app is not running\n", bundleID)
			continue
		}

		// Find matching window in saved group
		var savedWindow *Window
		for i := range group.Windows {
			if group.Windows[i].BundleID == bundleID {
				savedWindow = &group.Windows[i]
				break
			}
		}

		if savedWindow == nil {
			continue
		}

		// Apply saved position and size to first available window
		window := windows[0]
		err = accessibility.SetWindowBounds(
			window.ProcessID,
			window.WindowID,
			accessibility.Rect{
				X:      savedWindow.X,
				Y:      savedWindow.Y,
				Width:  savedWindow.Width,
				Height: savedWindow.Height,
			},
		)
		if err != nil {
			fmt.Printf("Warning: Could not restore window for %s: %v\n", bundleID, err)
			continue
		}

		// Restore minimized state
		if savedWindow.IsMinimized {
			accessibility.MinimizeWindow(window.ProcessID, window.WindowID)
		}

		// Restore fullscreen state
		if savedWindow.IsFullScreen {
			accessibility.SetFullScreen(window.ProcessID, window.WindowID, true)
		}

		fmt.Printf("Restored: %s at (%.0f, %.0f) size (%.0fx%.0f)\n",
			bundleID, savedWindow.X, savedWindow.Y, savedWindow.Width, savedWindow.Height)
	}

	fmt.Printf("Group %s restored successfully\n", group.Name)
	return nil
}

// launchApp launches an application by bundle ID
func (m *Manager) launchApp(bundleID string) error {
	cmd := exec.Command("open", "-b", bundleID)
	return cmd.Run()
}

// RestoreGroupByName restores a group by name
func (m *Manager) RestoreGroupByName(name string) error {
	for _, g := range m.groups {
		if g.Name == name {
			return m.RestoreGroup(g.ID)
		}
	}
	return fmt.Errorf("group not found: %s", name)
}

// DeleteGroup deletes a window group
func (m *Manager) DeleteGroup(groupID string) error {
	for i, g := range m.groups {
		if g.ID == groupID {
			m.groups = append(m.groups[:i], m.groups[i+1:]...)
			return m.saveGroups()
		}
	}
	return fmt.Errorf("group not found: %s", groupID)
}

// GetGroups returns all saved groups
func (m *Manager) GetGroups() []Group {
	return m.groups
}

// GetGroup returns a specific group
func (m *Manager) GetGroup(id string) *Group {
	for _, g := range m.groups {
		if g.ID == id {
			return &g
		}
	}
	return nil
}

// UpdateGroupWindows updates the saved windows for a group
func (m *Manager) UpdateGroupWindows(groupID string) error {
	group := m.GetGroup(groupID)
	if group == nil {
		return fmt.Errorf("group not found: %s", groupID)
	}

	// Capture current windows
	windows, err := m.CaptureWindows()
	if err != nil {
		return fmt.Errorf("failed to capture windows: %w", err)
	}

	// Update windows for matching bundle IDs
	for i := range group.Windows {
		for _, w := range windows {
			if w.BundleID == group.Windows[i].BundleID {
				group.Windows[i].X = w.X
				group.Windows[i].Y = w.Y
				group.Windows[i].Width = w.Width
				group.Windows[i].Height = w.Height
				break
			}
		}
	}

	return m.saveGroups()
}

// generateGroupID generates a unique group ID
func generateGroupID() string {
	return fmt.Sprintf("group-%d", time.Now().UnixNano())
}
