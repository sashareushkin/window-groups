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
    NSWindow *existing = self.overlays[@(windowID)];
    if (existing) {
        [existing orderOut:nil];
        [self.overlays removeObjectForKey:@(windowID)];
    }
    
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
    overlay.releasedWhenClosed = NO;
    overlay.collectionBehavior = NSWindowCollectionBehaviorCanJoinAllSpaces | NSWindowCollectionBehaviorFullScreenAuxiliary;
    
    // Add border layer (in overlay-local coordinates)
    overlay.contentView.wantsLayer = YES;
    CALayer *borderLayer = [CALayer layer];
    borderLayer.borderWidth = 3.0;
    borderLayer.borderColor = [[NSColor systemBlueColor] CGColor];
    borderLayer.backgroundColor = [[NSColor clearColor] CGColor];
    borderLayer.frame = overlay.contentView.bounds;
    borderLayer.autoresizingMask = kCALayerWidthSizable | kCALayerHeightSizable;
    overlay.contentView.layer = borderLayer;
    
    [overlay orderFrontRegardless];
    
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
        dispatch_async(dispatch_get_main_queue(), ^{
            [m highlightWindow:windowID];
        });
    }
}

void remove_highlight(void *mgr, uint32_t windowID) {
    if (mgr) {
        HighlightManager *m = (__bridge HighlightManager *)mgr;
        dispatch_async(dispatch_get_main_queue(), ^{
            [m removeHighlightForWindow:windowID];
        });
    }
}

void clear_highlights(void *mgr) {
    if (mgr) {
        HighlightManager *m = (__bridge HighlightManager *)mgr;
        dispatch_async(dispatch_get_main_queue(), ^{
            [m clearAllHighlights];
        });
    }
}

void destroy_highlight_manager(void *mgr) {
    if (mgr) {
        HighlightManager *m = (__bridge_transfer HighlightManager *)mgr;
        [m clearAllHighlights];
    }
}
