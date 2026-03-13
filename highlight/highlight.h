#ifndef HIGHLIGHT_MGR_H
#define HIGHLIGHT_MGR_H

#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import <Quartz/Quartz.h>

// Highlight manager interface
@interface HighlightManager : NSObject
@property (nonatomic, strong) NSMutableDictionary *overlays;
- (instancetype)init;
- (void)highlightWindow:(CGWindowID)windowID;
- (void)removeHighlightForWindow:(CGWindowID)windowID;
- (void)clearAllHighlights;
@end

// C interface functions
void* create_highlight_manager();
void highlight_window(void *mgr, uint32_t windowID);
void remove_highlight(void *mgr, uint32_t windowID);
void clear_highlights(void *mgr);
void destroy_highlight_manager(void *mgr);

#endif // HIGHLIGHT_MGR_H
