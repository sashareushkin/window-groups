//go:build darwin
// +build darwin

package accessibility

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation

#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>

// NOTE: AXError, AXValueType, and constants are already defined in the SDK headers
// No need to redefine them - just use them directly

// CFSTR wrapper function to avoid macro issues in CGO
static CFStringRef goCFSTR(const char *str) {
    return CFStringCreateWithCString(kCFAllocatorDefault, str, kCFStringEncodingUTF8);
}

// Wrapper macros for CGO compatibility
#define CFSTR2(str) goCFSTR(str)
#define AXUIElementRef(x) ((AXUIElement)(x))
#define AXValueRef(x) ((AXValue)(x))
#define AXErrorRef(x) ((AXError)(x))

// AXValueCreate wrapper
static AXValue goAXValueCreate(AXValueType type, const void *value) {
    return AXValueCreate(type, value);
}
*/

// Check if we have accessibility permissions
bool has_permissions() {
    return AXIsProcessTrusted();
}

// Get windows using CGWindowList
CFArrayRef get_window_list(uint32_t options, uint32_t windowID, int *count) {
    CFArrayRef list = CGWindowListCopyWindowInfo(options, windowID);
    if (list == NULL) {
        *count = 0;
        return NULL;
    }
    
    *count = (int)CFArrayGetCount(list);
    CFRetain(list);
    return list;
}

// Get integer value from dictionary
int get_int(CFDictionaryRef dict, CFStringRef key) {
    CFNumberRef value;
    if (!CFDictionaryGetValueIfPresent(dict, key, (const void**)&value)) {
        return 0;
    }
    int result;
    CFNumberGetValue(value, kCFNumberIntType, &result);
    return result;
}

// Get double value from dictionary
double get_double(CFDictionaryRef dict, CFStringRef key) {
    CFNumberRef value;
    if (!CFDictionaryGetValueIfPresent(dict, key, (const void**)&value)) {
        return 0.0;
    }
    double result;
    CFNumberGetValue(value, kCFNumberFloat64Type, &result);
    return result;
}

// Get string value from dictionary - returns allocated string
char* get_string(CFDictionaryRef dict, CFStringRef key) {
    CFStringRef value;
    if (!CFDictionaryGetValueIfPresent(dict, key, (const void**)&value)) {
        char *empty = malloc(1);
        empty[0] = '\0';
        return empty;
    }
    
    CFIndex length = CFStringGetLength(value) + 1;
    char *result = malloc(length);
    if (CFStringGetCString(value, result, length, kCFStringEncodingUTF8)) {
        return result;
    }
    free(result);
    char *empty = malloc(1);
    empty[0] = '\0';
    return empty;
}

// Get sub-dictionary
CFDictionaryRef get_dict(CFDictionaryRef dict, CFStringRef key) {
    CFDictionaryRef value;
    if (!CFDictionaryGetValueIfPresent(dict, key, (const void**)&value)) {
        return NULL;
    }
    return value;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// CheckPermissions checks if the app has Accessibility permissions
func CheckPermissions() bool {
	return C.has_permissions()
}

// GetWindows returns all windows from all applications with full state info
func GetWindows() ([]WindowInfo, error) {
	if !CheckPermissions() {
		return nil, errors.New("accessibility permissions not granted")
	}

	options := C.uint(C.kCGWindowListOptionOnScreenOnly | C.kCGWindowListExcludeDesktopElements)
	windowID := C.uint(0)

	var count C.int
	windowList := C.get_window_list(options, windowID, &count)
	if windowList == nil || count == 0 {
		return []WindowInfo{}, nil
	}
	defer C.CFRelease(C.CFTypeRef(windowList))

	// Get keys - use CFSTR macro for constant strings
	kCGWindowNumber := C.goCFSTR("kCGWindowNumber")
	kCGWindowOwnerName := C.goCFSTR("kCGWindowOwnerName")
	kCGWindowOwnerPID := C.goCFSTR("kCGWindowOwnerPID")
	kCGWindowBounds := C.goCFSTR("kCGWindowBounds")
	kCGWindowLayer := C.goCFSTR("kCGWindowLayer")
	kCGWindowName := C.goCFSTR("kCGWindowName")

	// Get active displays for monitor determination
	displays, displayCount := GetDisplays()

	for i := 0; i < int(count); i++ {
		var dict C.CFDictionaryRef
		CFArrayGetValueAtIndex(C.CFArrayRef(windowList), C.CFIndex(i), unsafe.Pointer(&dict))

		// Get window layer (should be 0 for normal windows)
		layer := C.get_int(dict, kCGWindowLayer)
		if layer != 0 {
			continue
		}

		// Get window bounds
		boundsDict := C.get_dict(dict, kCGWindowBounds)
		if boundsDict == nil {
			continue
		}

		var bounds Rect
		bounds.X = float64(C.get_double(boundsDict, C.goCFSTR("X")))
		bounds.Y = float64(C.get_double(boundsDict, C.goCFSTR("Y")))
		bounds.Width = float64(C.get_double(boundsDict, C.goCFSTR("Width")))
		bounds.Height = float64(C.get_double(boundsDict, C.goCFSTR("Height")))

		// Skip very small windows
		if bounds.Width < 50 || bounds.Height < 50 {
			continue
		}

		// Get owner PID
		pid := C.get_int(dict, kCGWindowOwnerPID)
		if pid == 0 {
			continue
		}

		// Get owner name
		ownerName := C.GoString(C.get_string(dict, kCGWindowOwnerName))

		// Get window name
		windowName := C.GoString(C.get_string(dict, kCGWindowName))

		// Get window number
		windowNumber := C.get_int(dict, kCGWindowNumber)

		// Get bundle ID
		bundleID := getBundleIDFromPID(C.int32_t(pid))

		// Determine which display the window is on
		displayID := determineDisplay(bounds, displays, displayCount)

		// Get window state (minimized, maximized)
		isMinimized, isFullScreen := getWindowState(C.int32_t(pid), uint32(windowNumber))

		windows = append(windows, WindowInfo{
			Title:         windowName,
			Bounds:        bounds,
			BundleID:      bundleID,
			AppName:       ownerName,
			ProcessID:     int32(pid),
			WindowID:      uint32(windowNumber),
			DisplayID:     displayID,
			IsMinimized:   isMinimized,
			IsFullScreen:  isFullScreen,
		})
	}

	return windows, nil
}

// getBundleIDFromPID gets bundle ID from process ID
func getBundleIDFromPID(pid C.int32_t) string {
	app := C.AXUIElementRefCreateApplication(pid)
	if app == 0 {
		return ""
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var bundleIDRef C.CFTypeRef
	result := C.AXUIElementRefCopyAttributeValue(C.AXUIElementRef(app), C.goCFSTR("AXFocusedApplication"), &bundleIDRef)
	if result != C.kAXErrorSuccess {
		return ""
	}
	if bundleIDRef != 0 {
		defer C.CFRelease(bundleIDRef)
	}

	// Get bundle ID from the application element
	var bundleID C.CFTypeRef
	result = C.AXUIElementRefCopyAttributeValue(C.AXUIElementRef(app), C.goCFSTR("AXBundleIdentifier"), &bundleID)
	if result != C.kAXErrorSuccess {
		return ""
	}
	if bundleID != 0 {
		defer C.CFRelease(bundleID)
	}

	cfStr := C.CFStringRef(bundleID)
	length := C.CFStringGetLength(cfStr)
	if length == 0 {
		return ""
	}

	var buf [1024]C.char
	if C.CFStringGetCString(cfStr, &buf[0], 1024, C.kCFStringEncodingUTF8) {
		return C.GoString(&buf[0])
	}
	return ""
}

// getWindowState gets minimized and fullscreen state for a window
func getWindowState(pid C.int32_t, windowID uint32) (isMinimized, isFullScreen bool) {
	app := C.AXUIElementRefCreateApplication(pid)
	if app == 0 {
		return false, false
	}
	defer C.CFRelease(C.CFTypeRef(app))

	// Get the focused window
	var window C.AXUIElementRef
	result := C.AXUIElementRefCopyAttributeValue(C.AXUIElementRef(app), C.goCFSTR("AXFocusedWindow"), (*C.AXValueRef)(unsafe.Pointer(&window)))
	if result != C.kAXErrorSuccess {
		return false, false
	}
	if window != 0 {
		defer C.CFRelease(C.CFTypeRef(window))
	}

	// Check minimized state
	var minimizedRef C.CFTypeRef
	result = C.AXUIElementRefCopyAttributeValue(C.AXUIElementRef(window), C.goCFSTR("AXMinimized"), &minimizedRef)
	if result == C.kAXErrorSuccess && minimizedRef != 0 {
		isMinimized = C.CFBooleanGetValue(C.CFBooleanRef(minimizedRef)) != 0
		C.CFRelease(minimizedRef)
	}

	// Check full screen state
	var fullScreenRef C.CFTypeRef
	result = C.AXUIElementRefCopyAttributeValue(C.AXUIElementRef(window), C.goCFSTR("AXFullScreen"), &fullScreenRef)
	if result == C.kAXErrorSuccess && fullScreenRef != 0 {
		isFullScreen = C.CFBooleanGetValue(C.CFBooleanRef(fullScreenRef)) != 0
		C.CFRelease(fullScreenRef)
	}

	return isMinimized, isFullScreen
}

// GetDisplays returns list of active display IDs and count
func GetDisplays() ([]uint32, int) {
	var displayCount C.uint32_t
	var displays [16]C.CGDirectDisplayID

	result := C.CGGetActiveDisplayList(16, &displays[0], &displayCount)
	if result != C.CGError(0) {
		return []uint32{uint32(C.CGMainDisplayID())}, 1
	}

	resultDisplays := make([]uint32, 0, int(displayCount))
	for i := 0; i < int(displayCount); i++ {
		resultDisplays = append(resultDisplays, uint32(displays[i]))
	}

	if len(resultDisplays) == 0 {
		return []uint32{uint32(C.CGMainDisplayID())}, 1
	}

	return resultDisplays, int(displayCount)
}

// determineDisplay determines which display contains the window center
func determineDisplay(bounds Rect, displays []uint32, displayCount int) uint32 {
	// Calculate window center
	centerX := bounds.X + bounds.Width/2
	centerY := bounds.Y + bounds.Height/2

	// Check each display
	for _, displayID := range displays {
		displayBounds := GetDisplayBounds(displayID)
		if centerX >= displayBounds.X && centerX < displayBounds.X+displayBounds.Width &&
			centerY >= displayBounds.Y && centerY < displayBounds.Y+displayBounds.Height {
			return displayID
		}
	}

	// Default to main display
	return uint32(C.CGMainDisplayID())
}

// GetDisplayBounds returns the bounds of a specific display
func GetDisplayBounds(displayID uint32) Rect {
	display := C.CGDirectDisplayID(displayID)
	bounds := C.CGDisplayBounds(display)
	return Rect{
		X:      float64(bounds.origin.x),
		Y:      float64(bounds.origin.y),
		Width:  float64(bounds.size.width),
		Height: float64(bounds.size.height),
	}
}

// GetDisplayInfo returns information about all connected displays
func GetDisplayInfo() ([]DisplayInfo, error) {
	displays, count := GetDisplays()
	result := make([]DisplayInfo, 0, count)

	for _, displayID := range displays {
		bounds := GetDisplayBounds(displayID)
		isMain := displayID == uint32(C.CGMainDisplayID())

		result = append(result, DisplayInfo{
			ID:      displayID,
			Bounds:  bounds,
			IsMain:  isMain,
		})
	}

	return result, nil
}

// SetWindowPosition sets window position
func SetWindowPosition(pid int32, windowID uint32, x, y float64) error {
	if !CheckPermissions() {
		return errors.New("accessibility permissions not granted")
	}

	app := C.AXUIElementRefCreateApplication(C.int32_t(pid))
	if app == 0 {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var window C.AXUIElementRef
	result := C.AXUIElementRefCopyAttributeValue(C.AXUIElementRef(app), C.goCFSTR("AXFocusedWindow"), (*C.AXValueRef)(unsafe.Pointer(&window)))
	if result != C.kAXErrorSuccess {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	if window != 0 {
		defer C.CFRelease(C.CFTypeRef(window))
	}

	position := C.CGPoint{X: C.CGFloat(x), Y: C.CGFloat(y)}
	positionValue := C.goAXValueCreate(C.int(kAXValueTypePoint), unsafe.Pointer(&position))
	if positionValue == 0 {
		return errors.New("failed to create position value")
	}
	defer C.CFRelease(C.CFTypeRef(positionValue))

	setResult := C.AXUIElementRefSetAttributeValue(C.AXUIElementRef(window), C.goCFSTR("AXPosition"), C.CFTypeRef(positionValue))
	if setResult != C.kAXErrorSuccess {
		return fmt.Errorf("failed to set position: %v", setResult)
	}

	return nil
}

// SetWindowSize sets window size
func SetWindowSize(pid int32, windowID uint32, width, height float64) error {
	if !CheckPermissions() {
		return errors.New("accessibility permissions not granted")
	}

	app := C.AXUIElementRefCreateApplication(C.int32_t(pid))
	if app == 0 {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var window C.AXUIElementRef
	result := C.AXUIElementRefCopyAttributeValue(C.AXUIElementRef(app), C.goCFSTR("AXFocusedWindow"), (*C.AXValueRef)(unsafe.Pointer(&window)))
	if result != C.kAXErrorSuccess {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	if window != 0 {
		defer C.CFRelease(C.CFTypeRef(window))
	}

	size := C.CGSize{Width: C.CGFloat(width), Height: C.CGFloat(height)}
	sizeValue := C.goAXValueCreate(C.int(kAXValueTypeSize), unsafe.Pointer(&size))
	if sizeValue == 0 {
		return errors.New("failed to create size value")
	}
	defer C.CFRelease(C.CFTypeRef(sizeValue))

	setResult := C.AXUIElementRefSetAttributeValue(C.AXUIElementRef(window), C.goCFSTR("AXSize"), C.CFTypeRef(sizeValue))
	if setResult != C.kAXErrorSuccess {
		return fmt.Errorf("failed to set size: %v", setResult)
	}

	return nil
}

// SetWindowBounds sets both position and size
func SetWindowBounds(pid int32, windowID uint32, bounds Rect) error {
	if err := SetWindowPosition(pid, windowID, bounds.X, bounds.Y); err != nil {
		return err
	}
	return SetWindowSize(pid, windowID, bounds.Width, bounds.Height)
}

// MinimizeWindow minimizes a window
func MinimizeWindow(pid int32, windowID uint32) error {
	if !CheckPermissions() {
		return errors.New("accessibility permissions not granted")
	}

	app := C.AXUIElementRefCreateApplication(C.int32_t(pid))
	if app == 0 {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var window C.AXUIElementRef
	result := C.AXUIElementRefCopyAttributeValue(C.AXUIElementRef(app), C.goCFSTR("AXFocusedWindow"), (*C.AXValueRef)(unsafe.Pointer(&window)))
	if result != C.kAXErrorSuccess {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	if window != 0 {
		defer C.CFRelease(C.CFTypeRef(window))
	}

	setResult := C.AXUIElementRefSetAttributeValue(C.AXUIElementRef(window), C.goCFSTR("AXMinimized"), C.CFTypeRef(C.kCFBooleanTrue))
	if setResult != C.kAXErrorSuccess {
		return fmt.Errorf("failed to minimize window: %v", setResult)
	}

	return nil
}

// UnminimizeWindow unminimizes a window
func UnminimizeWindow(pid int32, windowID uint32) error {
	if !CheckPermissions() {
		return errors.New("accessibility permissions not granted")
	}

	app := C.AXUIElementRefCreateApplication(C.int32_t(pid))
	if app == 0 {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var window C.AXUIElementRef
	result := C.AXUIElementRefCopyAttributeValue(C.AXUIElementRef(app), C.goCFSTR("AXFocusedWindow"), (*C.AXValueRef)(unsafe.Pointer(&window)))
	if result != C.kAXErrorSuccess {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	if window != 0 {
		defer C.CFRelease(C.CFTypeRef(window))
	}

	setResult := C.AXUIElementRefSetAttributeValue(C.AXUIElementRef(window), C.goCFSTR("AXMinimized"), C.CFTypeRef(C.kCFBooleanFalse))
	if setResult != C.kAXErrorSuccess {
		return fmt.Errorf("failed to unminimize window: %v", setResult)
	}

	return nil
}

// SetFullScreen sets window fullscreen state
func SetFullScreen(pid int32, windowID uint32, fullscreen bool) error {
	if !CheckPermissions() {
		return errors.New("accessibility permissions not granted")
	}

	app := C.AXUIElementRefCreateApplication(C.int32_t(pid))
	if app == 0 {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var window C.AXUIElementRef
	result := C.AXUIElementRefCopyAttributeValue(C.AXUIElementRef(app), C.goCFSTR("AXFocusedWindow"), (*C.AXValueRef)(unsafe.Pointer(&window)))
	if result != C.kAXErrorSuccess {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	if window != 0 {
		defer C.CFRelease(C.CFTypeRef(window))
	}

	var fsValue C.CFBooleanRef
	if fullscreen {
		fsValue = C.kCFBooleanTrue
	} else {
		fsValue = C.kCFBooleanFalse
	}

	setResult := C.AXUIElementRefSetAttributeValue(C.AXUIElementRef(window), C.goCFSTR("AXFullScreen"), C.CFTypeRef(fsValue))
	if setResult != C.kAXErrorSuccess {
		return fmt.Errorf("failed to set fullscreen: %v", setResult)
	}

	return nil
}

// WindowInfo contains window information with full state
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

// DisplayInfo contains display information
type DisplayInfo struct {
	ID      uint32
	Bounds  Rect
	IsMain  bool
}

// Rect represents window bounds
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}
