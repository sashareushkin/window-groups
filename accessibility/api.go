//go:build darwin
// +build darwin

package accessibility

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit -framework Foundation -framework Quartz

#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>
#import <CoreGraphics/CoreGraphics.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

static char* wg_copy_windows_dump() {
    @autoreleasepool {
        CFArrayRef windowsRef = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly | kCGWindowListExcludeDesktopElements, kCGNullWindowID);
        if (windowsRef == NULL) {
            return strdup("");
        }

        NSArray *windows = (__bridge_transfer NSArray *)windowsRef;
        NSMutableString *out = [NSMutableString string];

        for (NSDictionary *info in windows) {
            NSNumber *layer = info[(NSString *)kCGWindowLayer];
            if (layer && layer.intValue != 0) {
                continue;
            }

            NSNumber *windowNumber = info[(NSString *)kCGWindowNumber];
            NSNumber *pidNumber = info[(NSString *)kCGWindowOwnerPID];
            NSDictionary *boundsDict = info[(NSString *)kCGWindowBounds];
            if (!windowNumber || !pidNumber || !boundsDict) {
                continue;
            }

            CGRect bounds = CGRectZero;
            if (!CGRectMakeWithDictionaryRepresentation((__bridge CFDictionaryRef)boundsDict, &bounds)) {
                continue;
            }
            if (bounds.size.width < 2 || bounds.size.height < 2) {
                continue;
            }

            NSNumber *alpha = info[(NSString *)kCGWindowAlpha];
            if (alpha && alpha.doubleValue <= 0.01) {
                continue;
            }

            NSString *ownerName = info[(NSString *)kCGWindowOwnerName];
            if (ownerName && ([ownerName isEqualToString:@"Window Server"] || [ownerName isEqualToString:@"Dock"])) {
                continue;
            }

            NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:(pid_t)pidNumber.intValue];
            NSString *bundleID = app.bundleIdentifier ?: @"";
            NSString *appName = app.localizedName ?: (ownerName ?: @"");
            NSString *title = info[(NSString *)kCGWindowName] ?: @"";

            NSString *line = [NSString stringWithFormat:@"%d\t%@\t%@\t%@\t%.2f\t%.2f\t%.2f\t%.2f\t%u\n",
                              pidNumber.intValue,
                              bundleID,
                              appName,
                              title,
                              bounds.origin.x,
                              bounds.origin.y,
                              bounds.size.width,
                              bounds.size.height,
                              windowNumber.unsignedIntValue];
            [out appendString:line];
        }

        return strdup(out.UTF8String ?: "");
    }
}
*/
import "C"

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"
)

// CheckPermissions returns true on macOS builds.
// Runtime permission checks are handled by the OS when APIs are called.
func CheckPermissions() bool {
	return true
}

// GetWindows returns current visible app windows via CGWindowListCopyWindowInfo.
func GetWindows() ([]WindowInfo, error) {
	raw := C.wg_copy_windows_dump()
	if raw == nil {
		return nil, fmt.Errorf("CGWindowListCopyWindowInfo failed")
	}
	defer C.free(unsafe.Pointer(raw))

	text := strings.TrimSpace(C.GoString(raw))
	if text == "" {
		return []WindowInfo{}, nil
	}

	lines := strings.Split(text, "\n")
	result := make([]WindowInfo, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 9 {
			continue
		}

		pid, _ := strconv.Atoi(parts[0])
		x, _ := strconv.ParseFloat(parts[4], 64)
		y, _ := strconv.ParseFloat(parts[5], 64)
		w, _ := strconv.ParseFloat(parts[6], 64)
		h, _ := strconv.ParseFloat(parts[7], 64)
		wid, _ := strconv.ParseUint(parts[8], 10, 32)

		result = append(result, WindowInfo{
			Title:     parts[3],
			Bounds:    Rect{X: x, Y: y, Width: w, Height: h},
			BundleID:  parts[1],
			AppName:   parts[2],
			ProcessID: int32(pid),
			WindowID:  uint32(wid),
			DisplayID: 0,
		})
	}

	return result, nil
}

func SetWindowPosition(pid int32, windowID uint32, x, y float64) error {
	script := fmt.Sprintf(`tell application "System Events"
set p to first application process whose unix id is %d
if (count of windows of p) is 0 then error "no windows"
set position of front window of p to {%d, %d}
end tell`, pid, int(x), int(y))
	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("set position failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func SetWindowSize(pid int32, windowID uint32, width, height float64) error {
	script := fmt.Sprintf(`tell application "System Events"
set p to first application process whose unix id is %d
if (count of windows of p) is 0 then error "no windows"
set size of front window of p to {%d, %d}
end tell`, pid, int(width), int(height))
	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("set size failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func SetWindowBounds(pid int32, windowID uint32, bounds Rect) error {
	if err := SetWindowPosition(pid, windowID, bounds.X, bounds.Y); err != nil {
		return err
	}
	return SetWindowSize(pid, windowID, bounds.Width, bounds.Height)
}

func MinimizeWindow(pid int32, windowID uint32) error {
	script := fmt.Sprintf(`tell application "System Events"
set p to first application process whose unix id is %d
if (count of windows of p) is 0 then return
set value of attribute "AXMinimized" of front window of p to true
end tell`, pid)
	_, err := exec.Command("osascript", "-e", script).CombinedOutput()
	return err
}

func UnminimizeWindow(pid int32, windowID uint32) error {
	script := fmt.Sprintf(`tell application "System Events"
set p to first application process whose unix id is %d
if (count of windows of p) is 0 then return
set value of attribute "AXMinimized" of front window of p to false
end tell`, pid)
	_, err := exec.Command("osascript", "-e", script).CombinedOutput()
	return err
}

func SetFullScreen(pid int32, windowID uint32, fullscreen bool) error {
	value := "false"
	if fullscreen {
		value = "true"
	}
	script := fmt.Sprintf(`tell application "System Events"
set p to first application process whose unix id is %d
if (count of windows of p) is 0 then return
set value of attribute "AXFullScreen" of front window of p to %s
end tell`, pid, value)
	_, err := exec.Command("osascript", "-e", script).CombinedOutput()
	return err
}

func ActivateApp(bundleID string) error {
	if strings.TrimSpace(bundleID) == "" {
		return fmt.Errorf("empty bundleID")
	}
	out, err := exec.Command("open", "-b", bundleID).CombinedOutput()
	if err != nil {
		return fmt.Errorf("open -b failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func GetDisplays() ([]uint32, int) {
	return []uint32{0}, 1
}

func GetDisplayInfo() ([]DisplayInfo, error) {
	return []DisplayInfo{{ID: 0, Bounds: Rect{X: 0, Y: 0, Width: 0, Height: 0}, IsMain: true}}, nil
}

func GetDisplayBounds(displayID uint32) Rect {
	return Rect{X: 0, Y: 0, Width: 0, Height: 0}
}

type WindowInfo struct {
	Title        string
	Bounds       Rect
	BundleID     string
	AppName      string
	ProcessID    int32
	WindowID     uint32
	DisplayID    uint32
	IsMinimized  bool
	IsFullScreen bool
}

type DisplayInfo struct {
	ID     uint32
	Bounds Rect
	IsMain bool
}

type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func (r Rect) String() string {
	return fmt.Sprintf("Rect{x=%.2f,y=%.2f,w=%.2f,h=%.2f}", r.X, r.Y, r.Width, r.Height)
}
