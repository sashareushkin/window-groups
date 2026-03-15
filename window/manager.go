package window

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Windows           []Window  `json:"windows"`
	CreatedAt         time.Time `json:"created_at"`
	BundleIDs         []string  `json:"bundle_ids"`
	SelectedWindowIDs []uint32  `json:"selected_window_ids,omitempty"`
}

// Manager handles window operations
type Manager struct {
	groupsPath string
	groups     []Group
}

// NewManager creates a new window manager
func NewManager() *Manager {
	groupsPath := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "WindowGroups", "groups.json")

	m := &Manager{
		groupsPath: groupsPath,
		groups:     make([]Group, 0),
	}
	m.loadGroups()
	return m
}

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

	// Backfill selected ids for legacy groups.
	for gi := range m.groups {
		if len(m.groups[gi].SelectedWindowIDs) == 0 && len(m.groups[gi].Windows) > 0 {
			ids := make([]uint32, 0, len(m.groups[gi].Windows))
			for _, w := range m.groups[gi].Windows {
				if w.WindowID != 0 {
					ids = append(ids, w.WindowID)
				}
			}
			m.groups[gi].SelectedWindowIDs = dedupeWindowIDs(ids)
		}
	}
}

func (m *Manager) saveGroups() error {
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

	return result, nil
}

// CreateGroup creates a new group from explicitly selected window IDs.
func (m *Manager) CreateGroup(name string, selectedWindows map[uint32]string) (*Group, error) {
	if len(m.groups) >= 10 {
		return nil, fmt.Errorf("group limit reached: maximum 10 groups")
	}
	if len(selectedWindows) == 0 {
		return nil, fmt.Errorf("no windows selected")
	}

	captured, err := m.CaptureWindows()
	if err != nil {
		return nil, fmt.Errorf("failed to capture windows: %w", err)
	}

	selectedIDs := make([]uint32, 0, len(selectedWindows))
	for id := range selectedWindows {
		if id == 0 {
			continue
		}
		selectedIDs = append(selectedIDs, id)
	}
	selectedIDs = dedupeWindowIDs(selectedIDs)

	groupWindows := make([]Window, 0, len(selectedIDs))
	capturedByID := make(map[uint32]Window, len(captured))
	for _, w := range captured {
		capturedByID[w.WindowID] = w
	}
	for _, id := range selectedIDs {
		if w, ok := capturedByID[id]; ok {
			groupWindows = append(groupWindows, w)
		}
	}

	if len(groupWindows) == 0 {
		return nil, fmt.Errorf("no capturable windows found for selected windows")
	}

	bundleSeen := make(map[string]bool)
	bundleIDs := make([]string, 0, len(groupWindows))
	for _, w := range groupWindows {
		if w.BundleID == "" || bundleSeen[w.BundleID] {
			continue
		}
		bundleSeen[w.BundleID] = true
		bundleIDs = append(bundleIDs, w.BundleID)
	}

	group := Group{
		ID:                generateGroupID(),
		Name:              name,
		BundleIDs:         bundleIDs,
		Windows:           groupWindows,
		SelectedWindowIDs: selectedIDs,
		CreatedAt:         time.Now(),
	}

	m.groups = append(m.groups, group)
	if err := m.saveGroups(); err != nil {
		return nil, err
	}

	fmt.Printf("Group created: %s (%s) with %d selected windows\n", name, group.ID, len(groupWindows))
	return &group, nil
}

// RestoreGroup restores a window group
func (m *Manager) RestoreGroup(groupID string) error {
	group := m.GetGroup(groupID)
	if group == nil {
		return fmt.Errorf("group not found: %s", groupID)
	}

	currentWindows, err := accessibility.GetWindows()
	if err != nil {
		return fmt.Errorf("failed to get current windows: %w", err)
	}

	if len(group.Windows) == 0 {
		return fmt.Errorf("group has no saved windows")
	}

	byBundle := make(map[string][]accessibility.WindowInfo)
	for _, cw := range currentWindows {
		if cw.BundleID == "" {
			continue
		}
		byBundle[cw.BundleID] = append(byBundle[cw.BundleID], cw)
	}

	restoredCount := 0
	activated := make(map[string]bool)
	usedCurrent := make(map[uint32]bool)

	for _, saved := range group.Windows {
		candidates := byBundle[saved.BundleID]
		if len(candidates) == 0 {
			fmt.Printf("Skip window %d (%s): app is not running\n", saved.WindowID, saved.BundleID)
			continue
		}

		target, ok := matchWindow(saved, candidates, usedCurrent)
		if !ok {
			fmt.Printf("Skip window %d (%s): no matching current window\n", saved.WindowID, saved.BundleID)
			continue
		}
		usedCurrent[target.WindowID] = true

		if !activated[saved.BundleID] {
			if err := accessibility.ActivateApp(saved.BundleID); err != nil {
				fmt.Printf("Warning: Could not activate app %s: %v\n", saved.BundleID, err)
			}
			activated[saved.BundleID] = true
		}

		err = accessibility.SetWindowBounds(
			target.ProcessID,
			target.WindowID,
			accessibility.Rect{X: saved.X, Y: saved.Y, Width: saved.Width, Height: saved.Height},
		)
		if err != nil {
			fmt.Printf("Warning: Could not restore window %d (%s): %v\n", target.WindowID, saved.BundleID, err)
			continue
		}

		if saved.IsMinimized {
			_ = accessibility.MinimizeWindow(target.ProcessID, target.WindowID)
		} else {
			_ = accessibility.UnminimizeWindow(target.ProcessID, target.WindowID)
		}
		if saved.IsFullScreen {
			_ = accessibility.SetFullScreen(target.ProcessID, target.WindowID, true)
		}

		restoredCount++
		fmt.Printf("Restored window %d -> %d (%s)\n", saved.WindowID, target.WindowID, saved.BundleID)
	}

	if restoredCount == 0 {
		return fmt.Errorf("no windows were restored (apps not running or no matched windows)")
	}
	fmt.Printf("Group %s restored successfully (%d windows)\n", group.Name, restoredCount)
	return nil
}

func matchWindow(saved Window, candidates []accessibility.WindowInfo, used map[uint32]bool) (accessibility.WindowInfo, bool) {
	for _, c := range candidates {
		if used[c.WindowID] {
			continue
		}
		if c.WindowID == saved.WindowID {
			return c, true
		}
	}

	bestScore := math.MaxFloat64
	var best accessibility.WindowInfo
	found := false
	for _, c := range candidates {
		if used[c.WindowID] {
			continue
		}
		score := math.Abs(c.Bounds.X-saved.X) + math.Abs(c.Bounds.Y-saved.Y) + math.Abs(c.Bounds.Width-saved.Width) + math.Abs(c.Bounds.Height-saved.Height)
		if saved.Title != "" && c.Title != "" && saved.Title != c.Title {
			score += 10000
		}
		if score < bestScore {
			bestScore = score
			best = c
			found = true
		}
	}
	return best, found
}

// launchApp launches an application by bundle ID
func (m *Manager) launchApp(bundleID string) error {
	cmd := exec.Command("open", "-b", bundleID)
	return cmd.Run()
}

func (m *Manager) RestoreGroupByName(name string) error {
	for _, g := range m.groups {
		if g.Name == name {
			return m.RestoreGroup(g.ID)
		}
	}
	return fmt.Errorf("group not found: %s", name)
}

func (m *Manager) DeleteGroup(groupID string) error {
	for i, g := range m.groups {
		if g.ID == groupID {
			m.groups = append(m.groups[:i], m.groups[i+1:]...)
			return m.saveGroups()
		}
	}
	return fmt.Errorf("group not found: %s", groupID)
}

func (m *Manager) GetGroups() []Group { return m.groups }

func (m *Manager) GetGroup(id string) *Group {
	for i := range m.groups {
		if m.groups[i].ID == id {
			return &m.groups[i]
		}
	}
	return nil
}

func (m *Manager) GetGroupByName(name string) *Group {
	for i := range m.groups {
		if m.groups[i].Name == name {
			return &m.groups[i]
		}
	}
	return nil
}

func (m *Manager) UpdateGroupWindows(groupID string) error {
	group := m.GetGroup(groupID)
	if group == nil {
		return fmt.Errorf("group not found: %s", groupID)
	}

	captured, err := m.CaptureWindows()
	if err != nil {
		return fmt.Errorf("failed to capture windows: %w", err)
	}

	capturedByID := make(map[uint32]Window, len(captured))
	for _, w := range captured {
		capturedByID[w.WindowID] = w
	}

	for i := range group.Windows {
		if w, ok := capturedByID[group.Windows[i].WindowID]; ok {
			group.Windows[i] = w
			continue
		}
		for _, c := range captured {
			if c.BundleID == group.Windows[i].BundleID {
				group.Windows[i].X = c.X
				group.Windows[i].Y = c.Y
				group.Windows[i].Width = c.Width
				group.Windows[i].Height = c.Height
				break
			}
		}
	}

	ids := make([]uint32, 0, len(group.Windows))
	for _, w := range group.Windows {
		if w.WindowID != 0 {
			ids = append(ids, w.WindowID)
		}
	}
	group.SelectedWindowIDs = dedupeWindowIDs(ids)

	return m.saveGroups()
}

func generateGroupID() string { return fmt.Sprintf("group-%d", time.Now().UnixNano()) }

func dedupeWindowIDs(ids []uint32) []uint32 {
	seen := make(map[uint32]bool, len(ids))
	out := make([]uint32, 0, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
