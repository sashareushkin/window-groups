#import "highlight.h"

// Highlight manager implementation
@implementation HighlightManager

- (instancetype)init {
    self = [super init];
    if (self) {
        _overlays = [[NSMutableDictionary alloc] init];
    }
    return self;
}

- (void)highlightWindow:(CGWindowID)windowID {
    // Remove existing overlay if any
    [self.overlays removeObjectForKey:@(windowID)];
    
    // Get window info
    CFArrayRef windowList = CGWindowListCopyWindowInfo(
        kCGWindowListOptionIncludingWindow,
        windowID
    );
    
    if (windowList == NULL) {
        return;
    }
    
    // Get the first window from the list (toll-free bridged to NSArray)
    NSArray *windows = (__bridge_transfer NSArray *)windowList;
    
    if (windows.count == 0) {
        return;
    }
    
    NSDictionary *windowInfo = windows[0];
    NSDictionary *boundsDict = windowInfo[(NSString *)kCGWindowBounds];
    
    if (!boundsDict) {
        return;
    }
    
    // Extract bounds - CGRectValue returns a CGRect from NSValue
    CGRect bounds;
    CGRectMakeWithDictionaryRepresentation((__bridge CFDictionaryRef)boundsDict, &bounds);
    
    // Create overlay window
    NSWindow *overlay = [[NSWindow alloc] initWithContentRect:bounds
                                                   styleMask:NSWindowStyleMaskBorderless
                                                     backing:NSBackingStoreBuffered
                                                       defer:NO];
    
    overlay.level = NSScreenSaverWindowLevel + 1;
    overlay.backgroundColor = [NSColor clearColor];
    overlay.opaque = NO;
    overlay.ignoresMouseEvents = YES;
    overlay.hasShadow = NO;
    
    // Add border layer
    CALayer *borderLayer = [CALayer layer];
    borderLayer.borderWidth = 3.0;
    borderLayer.borderColor = [[NSColor systemBlueColor] CGColor];
    borderLayer.frame = bounds;
    
    overlay.contentView.layer = borderLayer;
    overlay.contentView.wantsLayer = YES;
    
    [overlay orderFront:nil];
    
    // Store overlay
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

// C interface implementation
void* create_highlight_manager() {
    HighlightManager *mgr = [[HighlightManager alloc] init];
    return (__bridge_retained void *)mgr;
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
        HighlightManager *m = (__bridge_transfer HighlightManager *)mgr;
        [m clearAllHighlights];
    }
}
