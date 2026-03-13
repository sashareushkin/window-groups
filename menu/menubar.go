//go:build darwin
// +build darwin

package menu

/*
#cgo LDFLAGS: -framework AppKit -framework Foundation

#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>

// Menu bar controller
@interface MenuController : NSObject
@property (nonatomic, assign) void *goManager;
@property (nonatomic, strong) NSStatusItem *statusItem;
@property (nonatomic, strong) NSMenu *menu;
@property (nonatomic, assign) BOOL isSelectingWindows;
@property (nonatomic, strong) NSMutableArray *selectedWindows;
@property (nonatomic, strong) NSMutableArray *groupNames;
@end

@implementation MenuController

- (instancetype)init {
    self = [super init];
    if (self) {
        _statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
        _menu = [[NSMenu alloc] init];
        _isSelectingWindows = NO;
        _selectedWindows = [[NSMutableArray alloc] init];
        _groupNames = [[NSMutableArray alloc] init];
        
        // Setup status item
        if (_statusItem.button) {
            _statusItem.button.image = [NSImage imageWithSystemSymbolName:@"rectangle.3.group" accessibilityDescription:@"Window Groups"];
            _statusItem.button.toolTip = @"Window Groups";
        }
        _statusItem.menu = _menu;
        
        [self rebuildMenu];
    }
    return self;
}

- (void)updateGroups:(NSArray *)groups {
    [_groupNames removeAllObjects];
    [_groupNames addObjectsFromArray:groups];
    [self rebuildMenu];
}

- (void)rebuildMenu {
    [_menu removeAllItems];
    
    // Header
    NSMenuItem *headerItem = [[NSMenuItem alloc] initWithTitle:@"Window Groups" action:nil keyEquivalent:@""];
    headerItem.enabled = NO;
    [_menu addItem:headerItem];
    
    [_menu addItem:[NSMenuItem separatorItem]];
    
    // Groups section
    if (_groupNames.count > 0) {
        NSMenuItem *groupsHeader = [[NSMenuItem alloc] initWithTitle:@"Saved Groups" action:nil keyEquivalent:@""];
        groupsHeader.enabled = NO;
        [_menu addItem:groupsHeader];
        
        for (NSString *groupName in _groupNames) {
            NSMenuItem *groupItem = [[NSMenuItem alloc] initWithTitle:groupName 
                                                              action:@selector(restoreGroup:) 
                                                       keyEquivalent:@""];
            groupItem.target = self;
            [_menu addItem:groupItem];
        }
        
        [_menu addItem:[NSMenuItem separatorItem]];
    }
    
    // Delete groups submenu
    if (_groupNames.count > 0) {
        NSMenuItem *deleteItem = [[NSMenuItem alloc] initWithTitle:@"Delete Group" action:nil keyEquivalent:@""];
        NSMenu *deleteMenu = [[NSMenu alloc] init];
        
        for (NSString *groupName in _groupNames) {
            NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:groupName 
                                                         action:@selector(deleteGroup:) 
                                                  keyEquivalent:@""];
            item.target = self;
            [deleteMenu addItem:item];
        }
        
        deleteItem.submenu = deleteMenu;
        [_menu addItem:deleteItem];
        
        [_menu addItem:[NSMenuItem separatorItem]];
    }
    
    // Create group button
    NSString *createTitle = _isSelectingWindows ? @"Save Group" : @"Create Group";
    NSMenuItem *createItem = [[NSMenuItem alloc] initWithTitle:createTitle 
                                                        action:@selector(toggleCreateMode:) 
                                                 keyEquivalent:@"n"];
    createItem.target = self;
    [_menu addItem:createItem];
    
    // Cancel selection (only in selection mode)
    if (_isSelectingWindows) {
        NSMenuItem *cancelItem = [[NSMenuItem alloc] initWithTitle:@"Cancel" 
                                                            action:@selector(cancelSelection:) 
                                                     keyEquivalent:@""];
        cancelItem.target = self;
        [_menu addItem:cancelItem];
    }
    
    [_menu addItem:[NSMenuItem separatorItem]];
    
    // Quit
    NSMenuItem *quitItem = [[NSMenuItem alloc] initWithTitle:@"Quit" 
                                                        action:@selector(quit:) 
                                                 keyEquivalent:@"q"];
    quitItem.target = self;
    [_menu addItem:quitItem];
}

- (void)restoreGroup:(id)sender {
    NSLog(@"Restore group requested");
}

- (void)deleteGroup:(id)sender {
    NSLog(@"Delete group requested");
}

- (void)toggleCreateMode:(id)sender {
    _isSelectingWindows = !_isSelectingWindows;
    
    if (_isSelectingWindows) {
        NSLog(@"Starting window selection mode");
    } else {
        NSLog(@"Saving group with %lu windows", (unsigned long)_selectedWindows.count);
        [_selectedWindows removeAllObjects];
    }
    
    [self rebuildMenu];
}

- (void)cancelSelection:(id)sender {
    _isSelectingWindows = NO;
    [_selectedWindows removeAllObjects];
    [self rebuildMenu];
}

- (void)quit:(id)sender {
    [[NSApplication sharedApplication] terminate:nil];
}

@end

// C interface
void* create_menu_controller() {
    MenuController *controller = [[MenuController alloc] init];
    return (__bridge void *)controller;
}

void update_menu_groups(void *controller, char **groups, int count) {
    if (controller) {
        MenuController *c = (__bridge MenuController *)controller;
        NSMutableArray *arr = [[NSMutableArray alloc] init];
        for (int i = 0; i < count; i++) {
            [arr addObject:[NSString stringWithUTF8String:groups[i]]];
        }
        [c updateGroups:arr];
    }
}

void destroy_menu_controller(void *controller) {
    if (controller) {
        MenuController *c = (__bridge MenuController *)controller;
        [[NSStatusBar systemStatusBar] removeStatusItem:c.statusItem];
    }
}
*/
import "C"

import (
	"fmt"
	"unsafe"

	"window-groups/highlight"
	"window-groups/window"
)

// MenuController wraps the Objective-C menu controller
type MenuController struct {
	ptr unsafe.Pointer
}

// NewMenuController creates a new menu controller
func NewMenuController() *MenuController {
	ptr := C.create_menu_controller()
	if ptr == nil {
		return nil
	}
	return &MenuController{ptr: ptr}
}

// UpdateGroups updates the groups in the menu
func (m *MenuController) UpdateGroups(groups []string) {
	if m.ptr == nil || len(groups) == 0 {
		return
	}

	cGroups := make([]*C.char, len(groups))
	for i, g := range groups {
		cGroups[i] = C.CString(g)
		defer C.free(unsafe.Pointer(cGroups[i]))
	}

	C.update_menu_groups(m.ptr, &cGroups[0], C.int(len(groups)))
}

// Destroy destroys the menu controller
func (m *MenuController) Destroy() {
	if m.ptr != nil {
		C.destroy_menu_controller(m.ptr)
		m.ptr = nil
	}
}

// MenuBar represents the menu bar UI
type MenuBar struct {
	wm              *window.Manager
	controller      *MenuController
	highlighter     *highlight.Manager
	isSelecting     bool
	selectedWindows map[uint32]string // windowID -> bundleID
}

// NewMenuBar creates a new menu bar
func NewMenuBar(wm *window.Manager) *MenuBar {
	return &MenuBar{
		wm:              wm,
		controller:      NewMenuController(),
		highlighter:     highlight.NewManager(),
		selectedWindows: make(map[uint32]string),
	}
}

// Run starts the menu bar (blocking)
func (m *MenuBar) Run() {
	if m.controller == nil {
		fmt.Println("Warning: Menu controller not initialized")
		return
	}

	// Update menu with existing groups
	m.refreshMenu()

	fmt.Println("Menu bar started")
	// In a real macOS app, this would run the main loop
}

// refreshMenu updates the menu with current groups
func (m *MenuBar) refreshMenu() {
	groups := m.wm.GetGroups()
	names := make([]string, len(groups))
	for i, g := range groups {
		names[i] = g.Name
	}
	m.controller.UpdateGroups(names)
}

// StartWindowSelection starts the window selection mode
func (m *MenuBar) StartWindowSelection() error {
	if m.controller == nil {
		return fmt.Errorf("menu controller not initialized")
	}

	m.isSelecting = true
	m.selectedWindows = make(map[uint32]string)
	m.highlighter.ClearAllHighlights()

	fmt.Println("Window selection mode started")
	fmt.Println("Click on windows to add them to the group")

	return nil
}

// AddSelectedWindow adds a window to the selection with highlight
func (m *MenuBar) AddSelectedWindow(windowID uint32, bundleID string) {
	if !m.isSelecting {
		return
	}

	// Check if already selected
	if _, exists := m.selectedWindows[windowID]; exists {
		return
	}

	m.selectedWindows[windowID] = bundleID
	m.highlighter.HighlightWindow(windowID)

	fmt.Printf("Added window %d: %s (total: %d)\n", windowID, bundleID, len(m.selectedWindows))
}

// RemoveSelectedWindow removes a window from selection
func (m *MenuBar) RemoveSelectedWindow(windowID uint32) {
	if !m.isSelecting {
		return
	}

	if _, exists := m.selectedWindows[windowID]; exists {
		delete(m.selectedWindows, windowID)
		m.highlighter.RemoveHighlight(windowID)
		fmt.Printf("Removed window %d (total: %d)\n", windowID, len(m.selectedWindows))
	}
}

// ToggleWindowSelection toggles a window in selection
func (m *MenuBar) ToggleWindowSelection(windowID uint32, bundleID string) {
	if _, exists := m.selectedWindows[windowID]; exists {
		m.RemoveSelectedWindow(windowID)
	} else {
		m.AddSelectedWindow(windowID, bundleID)
	}
}

// SaveGroup saves the selected windows as a group
func (m *MenuBar) SaveGroup() (string, error) {
	if !m.isSelecting {
		return "", fmt.Errorf("not in selection mode")
	}

	if len(m.selectedWindows) == 0 {
		return "", fmt.Errorf("no windows selected")
	}

	// Extract bundle IDs
	bundleIDs := make([]string, 0, len(m.selectedWindows))
	for _, bid := range m.selectedWindows {
		bundleIDs = append(bundleIDs, bid)
	}

	// Generate auto-name based on group size
	name := generateGroupName(len(bundleIDs))

	// Clear highlights
	m.highlighter.ClearAllHighlights()

	// Create group via window manager (captures current positions)
	group, err := m.wm.CreateGroup(name, bundleIDs)
	if err != nil {
		return "", err
	}

	fmt.Printf("Group saved: %s with %d windows\n", name, len(bundleIDs))

	// Reset selection
	m.isSelecting = false
	m.selectedWindows = make(map[uint32]string)

	// Update menu
	m.refreshMenu()

	return group.Name, nil
}

// CancelSelection cancels the current selection
func (m *MenuBar) CancelSelection() {
	m.highlighter.ClearAllHighlights()
	m.isSelecting = false
	m.selectedWindows = make(map[uint32]string)
	fmt.Println("Selection cancelled")
	m.refreshMenu()
}

// RestoreGroup restores a saved group
func (m *MenuBar) RestoreGroup(groupName string) error {
	err := m.wm.RestoreGroupByName(groupName)
	if err != nil {
		return fmt.Errorf("failed to restore group: %w", err)
	}
	fmt.Printf("Group restored: %s\n", groupName)
	return nil
}

// DeleteGroup deletes a saved group
func (m *MenuBar) DeleteGroup(groupName string) error {
	group := m.wm.GetGroup(groupName)
	if group == nil {
		return fmt.Errorf("group not found: %s", groupName)
	}

	err := m.wm.DeleteGroup(group.ID)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	fmt.Printf("Group deleted: %s\n", groupName)
	m.refreshMenu()
	return nil
}

// IsSelecting returns true if in selection mode
func (m *MenuBar) IsSelecting() bool {
	return m.isSelecting
}

// GetSelectedWindows returns currently selected windows (windowID -> bundleID)
func (m *MenuBar) GetSelectedWindows() map[uint32]string {
	return m.selectedWindows
}

// GetSelectedBundleIDs returns bundle IDs of selected windows
func (m *MenuBar) GetSelectedBundleIDs() []string {
	bundleIDs := make([]string, 0, len(m.selectedWindows))
	for _, bid := range m.selectedWindows {
		bundleIDs = append(bundleIDs, bid)
	}
	return bundleIDs
}

// ShowGroups displays saved groups in menu
func (m *MenuBar) ShowGroups() []string {
	groups := m.wm.GetGroups()
	names := make([]string, len(groups))
	for i, g := range groups {
		names[i] = g.Name
	}
	return names
}

// Destroy cleans up resources
func (m *MenuBar) Destroy() {
	if m.highlighter != nil {
		m.highlighter.Destroy()
	}
	if m.controller != nil {
		m.controller.Destroy()
	}
}

// generateGroupName generates a city-based name for the group
func generateGroupName(windowCount int) string {
	cities := []string{
		"Village",    // 1-2 windows
		"Town",       // 3-4 windows
		"City",       // 5-6 windows
		"Metropolis", // 7-9 windows
		"Megacity",   // 10+ windows
	}

	if windowCount <= 0 {
		return cities[0]
	}
	if windowCount > len(cities)-1 {
		return cities[len(cities)-1]
	}

	return cities[windowCount]
}
