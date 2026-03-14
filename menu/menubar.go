//go:build darwin
// +build darwin

package menu

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit -framework Foundation -framework Quartz

#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>
#include <stdint.h>
#include <stdlib.h>

extern int go_menu_toggle_create(uintptr_t handle);
extern void go_menu_restore_group(uintptr_t handle, char *groupName);
extern void go_menu_delete_group(uintptr_t handle, char *groupName);
extern void go_menu_quit(uintptr_t handle);
extern void go_menu_add_frontmost(uintptr_t handle, char *bundleID);
extern int go_menu_toggle_window(uintptr_t handle, uint32_t windowID, char *bundleID);
extern void go_menu_cancel_selection(uintptr_t handle);

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
@property (nonatomic, strong) NSMutableDictionary<NSNumber*, NSString*> *selectedWindows;
@property (nonatomic, strong) NSMutableArray *groupNames;
@property (nonatomic, strong) id globalMonitor;
@end

@implementation MenuController

- (instancetype)init {
    self = [super init];
    if (self) {
        _statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
        _menu = [[NSMenu alloc] init];
        _isSelectingWindows = NO;
        _selectedWindows = [[NSMutableDictionary alloc] init];
        _groupNames = [[NSMutableArray alloc] init];
        _globalMonitor = nil;
        
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

- (NSDictionary *)topWindowUnderMouse {
    NSPoint point = [NSEvent mouseLocation];
    CFArrayRef windowsRef = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly | kCGWindowListExcludeDesktopElements, kCGNullWindowID);
    if (windowsRef == NULL) {
        return nil;
    }

    NSArray *windows = (__bridge_transfer NSArray *)windowsRef;
    for (NSDictionary *info in windows) {
        NSNumber *layer = info[(NSString *)kCGWindowLayer];
        if (layer && layer.intValue != 0) {
            continue;
        }

        NSDictionary *boundsDict = info[(NSString *)kCGWindowBounds];
        if (!boundsDict) {
            continue;
        }

        CGRect bounds = CGRectZero;
        if (!CGRectMakeWithDictionaryRepresentation((__bridge CFDictionaryRef)boundsDict, &bounds)) {
            continue;
        }

        if (!CGRectContainsPoint(bounds, CGPointMake(point.x, point.y))) {
            continue;
        }

        return info;
    }

    return nil;
}

- (void)startSelectionMonitor {
    if (self.globalMonitor != nil) {
        return;
    }

    __weak typeof(self) weakSelf = self;
    self.globalMonitor = [NSEvent addGlobalMonitorForEventsMatchingMask:NSEventMaskLeftMouseDown handler:^(NSEvent *event) {
        __strong typeof(weakSelf) selfStrong = weakSelf;
        if (!selfStrong || !selfStrong.isSelectingWindows || selfStrong.goHandle == 0) {
            return;
        }

        if ((event.modifierFlags & NSEventModifierFlagOption) == 0) {
            return;
        }

        NSDictionary *windowInfo = [selfStrong topWindowUnderMouse];
        if (!windowInfo) {
            return;
        }

        NSNumber *windowNumber = windowInfo[(NSString *)kCGWindowNumber];
        NSNumber *pidNumber = windowInfo[(NSString *)kCGWindowOwnerPID];
        if (!windowNumber || !pidNumber) {
            return;
        }

        NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:(pid_t)pidNumber.intValue];
        NSString *bundle = app.bundleIdentifier;
        if (!bundle || bundle.length == 0) {
            return;
        }

        int selected = go_menu_toggle_window(selfStrong.goHandle, (uint32_t)windowNumber.unsignedIntValue, (char *)bundle.UTF8String);
        NSNumber *key = @((uint32_t)windowNumber.unsignedIntValue);
        if (selected != 0) {
            selfStrong.selectedWindows[key] = bundle;
        } else {
            [selfStrong.selectedWindows removeObjectForKey:key];
        }

        dispatch_async(dispatch_get_main_queue(), ^{
            [selfStrong rebuildMenu];
        });
    }];
}

- (void)stopSelectionMonitor {
    if (self.globalMonitor != nil) {
        [NSEvent removeMonitor:self.globalMonitor];
        self.globalMonitor = nil;
    }
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
    
    // Selection controls
    if (_isSelectingWindows) {
        NSMenuItem *addFrontmostItem = [[NSMenuItem alloc] initWithTitle:@"Add Frontmost App" 
                                                                   action:@selector(addFrontmostWindow:) 
                                                            keyEquivalent:@"a"];
        addFrontmostItem.target = self;
        [_menu addItem:addFrontmostItem];

        if (_selectedWindows.count > 0) {
            [_menu addItem:[NSMenuItem separatorItem]];
            NSMenuItem *selectedHeader = [[NSMenuItem alloc] initWithTitle:@"Selected windows (Option+Click)" action:nil keyEquivalent:@""];
            selectedHeader.enabled = NO;
            [_menu addItem:selectedHeader];
            for (NSNumber *windowID in _selectedWindows) {
                NSString *bundle = _selectedWindows[windowID];
                NSString *title = [NSString stringWithFormat:@"%@ (#%@)", bundle, windowID];
                NSMenuItem *b = [[NSMenuItem alloc] initWithTitle:title action:nil keyEquivalent:@""];
                b.enabled = NO;
                [_menu addItem:b];
            }
        }

        [_menu addItem:[NSMenuItem separatorItem]];
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

    if (_isSelectingWindows) {
        [self startSelectionMonitor];
    } else {
        [self stopSelectionMonitor];
        [_selectedWindows removeAllObjects];
    }

    [self rebuildMenu];
}

- (void)addFrontmostWindow:(id)sender {
    NSRunningApplication *app = [[NSWorkspace sharedWorkspace] frontmostApplication];
    NSString *bundle = app.bundleIdentifier;
    if (!bundle || bundle.length == 0) {
        return;
    }

    NSNumber *fakeID = @((uint32_t)(100000 + _selectedWindows.count + 1));
    _selectedWindows[fakeID] = bundle;
    if (_goHandle != 0) {
        go_menu_add_frontmost(_goHandle, (char *)[bundle UTF8String]);
    }

    [self rebuildMenu];
}

- (void)cancelSelection:(id)sender {
    _isSelectingWindows = NO;
    [self stopSelectionMonitor];
    [_selectedWindows removeAllObjects];
    if (_goHandle != 0) {
        go_menu_cancel_selection(_goHandle);
    }
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
        [c stopSelectionMonitor];
        [[NSStatusBar systemStatusBar] removeStatusItem:c.statusItem];
    }
}
*/
import "C"

import (
	"fmt"
	"math/rand"
	"runtime/cgo"
	"time"
	"unsafe"

	"window-groups/highlight"
	"window-groups/shortcuts"
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
	if m.ptr == nil {
		return
	}
	if len(groups) == 0 {
		C.update_menu_groups(m.ptr, nil, 0)
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
	hm              *shortcuts.HotkeyManager
	controller      *MenuController
	highlighter     *highlight.Manager
	handle          cgo.Handle
	isSelecting     bool
	selectedWindows map[uint32]string // windowID -> bundleID
}

// NewMenuBar creates a new menu bar
func NewMenuBar(wm *window.Manager, hm *shortcuts.HotkeyManager) *MenuBar {
	mb := &MenuBar{
		wm:              wm,
		hm:              hm,
		controller:      NewMenuController(),
		highlighter:     highlight.NewManager(),
		selectedWindows: make(map[uint32]string),
	}
	if mb.controller == nil {
		return nil
	}
	mb.handle = cgo.NewHandle(mb)
	C.set_menu_go_handle(mb.controller.ptr, C.uintptr_t(uintptr(mb.handle)))
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
	m.syncSlotHotkeys()
}

func (m *MenuBar) syncSlotHotkeys() {
	if m.hm == nil {
		return
	}
	m.hm.ClearAllShortcuts()
	defaultKeys := []int{shortcuts.Key1, shortcuts.Key2, shortcuts.Key3, shortcuts.Key4, shortcuts.Key5, shortcuts.Key6, shortcuts.Key7, shortcuts.Key8, shortcuts.Key9, shortcuts.Key0}
	for i, group := range m.wm.GetGroups() {
		if i >= len(defaultKeys) {
			break
		}
		if err := m.hm.RegisterShortcut(group.Name, defaultKeys[i], shortcuts.ModCommand|shortcuts.ModOption|shortcuts.ModControl); err != nil {
			fmt.Printf("Warning: Could not register hotkey for slot %d (%s): %v\n", i+1, group.Name, err)
		}
	}
}

// StartWindowSelection starts the window selection mode
func (m *MenuBar) StartWindowSelection() error {
	if m.controller == nil {
		return fmt.Errorf("menu controller not initialized")
	}

	m.isSelecting = true
	m.selectedWindows = make(map[uint32]string)
	if m.highlighter != nil {
		m.highlighter.ClearAllHighlights()
	}

	fmt.Println("Window selection mode started")
	fmt.Println("Use Option + Left Click to toggle windows")

	return nil
}

// AddSelectedBundle adds an app (bundle id) to the selection.
func (m *MenuBar) AddSelectedBundle(bundleID string) {
	if !m.isSelecting || bundleID == "" {
		return
	}
	for _, existing := range m.selectedWindows {
		if existing == bundleID {
			return
		}
	}
	id := uint32(len(m.selectedWindows) + 1)
	m.selectedWindows[id] = bundleID
	fmt.Printf("Selected app: %s (total: %d)\n", bundleID, len(m.selectedWindows))
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
	if m.highlighter != nil {
		m.highlighter.HighlightWindow(windowID)
	}

	fmt.Printf("Added window %d: %s (total: %d)\n", windowID, bundleID, len(m.selectedWindows))
}

// RemoveSelectedWindow removes a window from selection
func (m *MenuBar) RemoveSelectedWindow(windowID uint32) {
	if !m.isSelecting {
		return
	}

	if _, exists := m.selectedWindows[windowID]; exists {
		delete(m.selectedWindows, windowID)
		if m.highlighter != nil {
			m.highlighter.RemoveHighlight(windowID)
		}
		fmt.Printf("Removed window %d (total: %d)\n", windowID, len(m.selectedWindows))
	}
}

// ToggleWindowSelection toggles a window in selection
func (m *MenuBar) ToggleWindowSelection(windowID uint32, bundleID string) bool {
	if _, exists := m.selectedWindows[windowID]; exists {
		m.RemoveSelectedWindow(windowID)
		return false
	}
	m.AddSelectedWindow(windowID, bundleID)
	return true
}

// SaveGroup saves the selected windows as a group
func (m *MenuBar) SaveGroup() (string, error) {
	if !m.isSelecting {
		return "", fmt.Errorf("not in selection mode")
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

	name := m.nextGroupName()

	// Create group via window manager (captures current positions)
	group, err := m.wm.CreateGroup(name, bundleIDs)
	if err != nil {
		return "", err
	}

	// Clear highlights only after successful save.
	if m.highlighter != nil {
		m.highlighter.ClearAllHighlights()
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
	if m.highlighter != nil {
		m.highlighter.ClearAllHighlights()
	}
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

func (m *MenuBar) nextGroupName() string {
	used := make(map[string]bool)
	for _, g := range m.wm.GetGroups() {
		used[g.Name] = true
	}

	pool := append([]string(nil), cityNames...)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	for _, name := range pool {
		if !used[name] {
			return name
		}
	}

	// All base names are occupied: deterministic collision-free fallback.
	base := pool[0]
	if base == "" {
		base = "City"
	}
	for n := 2; ; n++ {
		candidate := fmt.Sprintf("%s-%d", base, n)
		if !used[candidate] {
			return candidate
		}
	}
}

var cityNames = []string{
	"Moscow", "London", "Paris", "Berlin", "Madrid", "Rome", "Amsterdam", "Vienna", "Prague", "Warsaw",
	"New York", "Los Angeles", "Chicago", "Toronto", "Vancouver", "Mexico City", "Sao Paulo", "Buenos Aires", "Santiago", "Lima",
	"Tokyo", "Seoul", "Beijing", "Shanghai", "Hong Kong", "Singapore", "Bangkok", "Dubai", "Istanbul", "Sydney",
}
