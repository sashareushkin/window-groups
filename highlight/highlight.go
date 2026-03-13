//go:build darwin
// +build darwin

package highlight

/*
#cgo LDFLAGS: -framework AppKit -framework Quartz

#import <AppKit/AppKit.h>
#import <Quartz/Quartz.h>

// Window highlighting overlay
@interface WindowHighlightLayer : CALayer
@property (nonatomic, assign) CGWindowID windowID;
@end

@implementation WindowHighlightLayer

- (instancetype)initWithWindowID:(CGWindowID)wid {
    self = [super init];
    if (self) {
        _windowID = wid;
        self.borderWidth = 3.0;
        self.borderColor = [[NSColor systemBlueColor] CGColor];
        self.backgroundColor = [[NSColor clearColor] CGColor];
    }
    return self;
}

@end

// Highlight manager
@interface HighlightManager : NSObject
@property (nonatomic, strong) NSMutableDictionary *overlays;
@end

@implementation HighlightManager

- (instancetype)init {
    self = [super init];
    if (self) {
        _overlays = [[NSMutableDictionary alloc] init];
    }
    return self;
}

- (void)highlightWindow:(CGWindowID)windowID {
    [self.overlays removeObjectForKey:@(windowID)];
    
    CFArrayRef windowList = CGWindowListCopyWindowInfo(kCGWindowListOptionIncludingWindow, windowID);
    if (windowList == nil) return;
    
    NSDictionary *windowInfo = (__bridge NSDictionary *)CFArrayGetValueAtIndex(windowList, 0);
    NSDictionary *boundsDict = windowInfo[(NSString *)kCGWindowBounds];
    
    CGRect bounds = [boundsDict CGRectValue];
    
    NSWindow *overlay = [[NSWindow alloc] initWithContentRect:bounds
                                                   styleMask:NSWindowStyleMaskBorderless
                                                     backing:NSBackingStoreBuffered
                                                       defer:NO];
    overlay.level = NSScreenSaverWindowLevel + 1;
    overlay.backgroundColor = [NSColor clearColor];
    overlay.opaque = NO;
    overlay.ignoresMouseEvents = YES;
    overlay.hasShadow = NO;
    
    CALayer *borderLayer = [CALayer layer];
    borderLayer.borderWidth = 3.0;
    borderLayer.borderColor = [[NSColor systemBlueColor] CGColor];
    borderLayer.frame = bounds;
    overlay.contentView.layer = borderLayer;
    overlay.contentView.wantsLayer = YES;
    
    [overlay orderFront:nil];
    
    self.overlays[@(windowID)] = overlay;
}

- (void)removeHighlightForWindow:(CGWindowID)windowID {
    NSWindow *overlay = self.overlays[@(windowID)];
    if (overlay) {
        [overlay orderOut:nil];
        [self.overlays removeObjectForKey:@(windowID)];
    }
}

- (void)clearAllHighlights {
    for (NSWindow *overlay in self.overlays.allValues) {
        [overlay orderOut:nil];
    }
    [self.overlays removeAllObjects];
}

@end

// C interface
void* create_highlight_manager() {
    HighlightManager *mgr = [[HighlightManager alloc] init];
    return (__bridge void *)mgr;
}

void highlight_window(void *mgr, uint32_t windowID) {
    if (mgr) {
        HighlightManager *m = (__bridge HighlightManager *)mgr;
        [m highlightWindow:windowID];
    }
}

void remove_highlight(void *mgr, uint32_t windowID) {
    if (mgr) {
        HighlightManager *m = (__bridge HighlightManager *)mgr;
        [m removeHighlightForWindow:windowID];
    }
}

void clear_highlights(void *mgr) {
    if (mgr) {
        HighlightManager *m = (__bridge HighlightManager *)mgr;
        [m clearAllHighlights];
    }
}

void destroy_highlight_manager(void *mgr) {
    if (mgr) {
        HighlightManager *m = (__bridge HighlightManager *)mgr;
        [m clearAllHighlights];
    }
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Manager handles window highlighting
type Manager struct {
	ptr unsafe.Pointer
}

// NewManager creates a new highlight manager
func NewManager() *Manager {
	ptr := C.create_highlight_manager()
	if ptr == nil {
		return nil
	}
	return &Manager{ptr: ptr}
}

// Destroy destroys the highlight manager
func (m *Manager) Destroy() {
	if m.ptr != nil {
		C.destroy_highlight_manager(m.ptr)
		m.ptr = nil
	}
}

// HighlightWindow adds a highlight to a window
func (m *Manager) HighlightWindow(windowID uint32) {
	if m.ptr == nil {
		return
	}
	C.highlight_window(m.ptr, C.uint(windowID))
	fmt.Printf("Highlighted window: %d\n", windowID)
}

// RemoveHighlight removes highlight from a window
func (m *Manager) RemoveHighlight(windowID uint32) {
	if m.ptr == nil {
		return
	}
	C.remove_highlight(m.ptr, C.uint(windowID))
}

// ClearAllHighlights removes all highlights
func (m *Manager) ClearAllHighlights() {
	if m.ptr == nil {
		return
	}
	C.clear_highlights(m.ptr)
	fmt.Println("All highlights cleared")
}
