//go:build darwin
// +build darwin

package menu

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit -framework Foundation

#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>
#include <stdint.h>

extern int go_menu_toggle_create(uintptr_t handle);
extern void go_menu_restore_group(uintptr_t handle, const char *groupName);
extern void go_menu_delete_group(uintptr_t handle, const char *groupName);
extern void go_menu_quit(uintptr_t handle);

void init_appkit() {
    [NSApplication sharedApplication];
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    [NSApp finishLaunching];
}

void run_appkit() {
    [NSApp run];
}

// Menu bar controller
@interface MenuController : NSObject
@property (nonatomic, assign) uintptr_t goHandle;
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
    if (groups) {
        [_groupNames addObjectsFromArray:groups];
    }
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
    NSString *groupName = ((NSMenuItem *)sender).title;
    if (_goHandle != 0) {
        go_menu_restore_group(_goHandle, [groupName UTF8String]);
    }
}

- (void)deleteGroup:(id)sender {
    NSString *groupName = ((NSMenuItem *)sender).title;
    if (_goHandle != 0) {
        go_menu_delete_group(_goHandle, [groupName UTF8String]);
    }
}

- (void)toggleCreateMode:(id)sender {
    if (_goHandle != 0) {
        int selecting = go_menu_toggle_create(_goHandle);
        _isSelectingWindows = (selecting != 0);
    } else {
        _isSelectingWindows = !_isSelectingWindows;
    }
    [self rebuildMenu];
}

- (void)cancelSelection:(id)sender {
    _isSelectingWindows = NO;
    [_selectedWindows removeAllObjects];
    [self rebuildMenu];
}

- (void)quit:(id)sender {
    if (_goHandle != 0) {
        go_menu_quit(_goHandle);
    }
    [[NSApplication sharedApplication] terminate:nil];
}

@end

// C interface
void* create_menu_controller() {
    MenuController *controller = [[MenuController alloc] init];
    return (__bridge_retained void *)controller;
}

void set_menu_go_handle(void *controller, uintptr_t handle) {
    if (controller) {
        MenuController *c = (__bridge MenuController *)controller;
        c.goHandle = handle;
    }
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
        MenuController *c = (__bridge_transfer MenuController *)controller;
        [[NSStatusBar systemStatusBar] removeStatusItem:c.statusItem];
    }
}
*/
import "C"

import (
	"fmt"
	"os"
	"runtime/cgo"
	"unsafe"

	"window-groups/accessibility"
	"window-groups/highlight"
	"window-groups/window"
)

// MenuController wraps the Objective-C menu controller
type MenuController struct {
	ptr unsafe.Pointer
}

// NewMenuController creates a new menu controller
func NewMenuController() *MenuController {
	C.init_appkit()
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
	handle          cgo.Handle
	isSelecting     bool
	selectedWindows map[uint32]string // windowID -> bundleID
}

// NewMenuBar creates a new menu bar
func NewMenuBar(wm *window.Manager) *MenuBar {
	mb := &MenuBar{
		wm:              wm,
		controller:      NewMenuController(),
		highlighter:     highlight.NewManager(),
		selectedWindows: make(map[uint32]string),
	}
	if mb.controller == nil {
		return nil
	}
	mb.handle = cgo.NewHandle(mb)
	C.set_menu_go_handle(mb.controller.ptr, C.uintptr_t(mb.handle))
	return mb
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
	C.run_appkit()
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

	// Fallback: if interactive selection is empty, capture all currently visible windows.
	if len(m.selectedWindows) == 0 {
		if windows, err := accessibility.GetWindows(); err == nil {
			for _, w := range windows {
				if w.BundleID != "" {
					m.selectedWindows[w.WindowID] = w.BundleID
				}
			}
		}
	}

	if len(m.selectedWindows) == 0 {
		return "", fmt.Errorf("no windows selected")
	}

	// Extract unique bundle IDs
	seen := make(map[string]bool)
	bundleIDs := make([]string, 0, len(m.selectedWindows))
	for _, bid := range m.selectedWindows {
		if bid == "" || seen[bid] {
			continue
		}
		seen[bid] = true
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
	var groupID string
	for _, g := range m.wm.GetGroups() {
		if g.Name == groupName {
			groupID = g.ID
			break
		}
	}
	if groupID == "" {
		return fmt.Errorf("group not found: %s", groupName)
	}

	err := m.wm.DeleteGroup(groupID)
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
	if m.handle != 0 {
		m.handle.Delete()
		m.handle = 0
	}
}

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

// generateGroupName generates a city-based name for the group
func generateGroupName(windowCount int) string {
	cities := []string{
		"Village",    // 1-2 windows
		"Town",       // 3-4 windows
		"City",       // 5-6 windows
		"Metropolis", // 7-9 windows
		"Megacity",   // 10+ windows
	}

	if windowCount <= 2 {
		return cities[0]
	}
	if windowCount <= 4 {
		return cities[1]
	}
	if windowCount <= 6 {
		return cities[2]
	}
	if windowCount <= 9 {
		return cities[3]
	}
	return cities[4]
}
