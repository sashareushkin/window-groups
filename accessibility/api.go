//go:build darwin
// +build darwin

package accessibility

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation

#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>

// AXError enum
typedef enum {
	kAXErrorSuccess = 0,
	kAXErrorFailure = 1,
	kAXErrorIllegalArgument = 2,
	kAXErrorInvalidUIElement = 3,
	kAXErrorInvalidUIElementObserver = 4,
	kAXErrorCannotComplete = 5,
	kAXErrorAttributeUnsupported = 6,
	kAXErrorActionUnsupported = 7,
	kAXErrorNotificationUnsupported = 8,
	kAXErrorNotImplemented = 9,
	kAXErrorNotificationAlreadyRegistered = 10,
	kAXErrorNotificationNotRegistered = 11,
	kAXErrorAPIDisabled = 12,
	kAXErrorNoValue = 13,
	kAXErrorParameterizedAttributeUnsupported = 14,
	kAXErrorNotEnoughPrecision = 15
} AXError;

// AXValueType enum
typedef enum {
	kAXValueTypeUnknown = 0,
	kAXValueTypeNumber = 1,
	kAXValueTypePoint = 2,
	kAXValueTypeSize = 3,
	kAXValueTypeRect = 4,
	kAXValueTypeRange = 5,
	kAXValueTypeString = 6,
	kAXValueTypeAttributedString = 7,
	kAXValueTypeURL = 8,
	kAXValueTypeDate = 9,
	kAXValueTypeData = 10
} AXValueType;

// AX constants
#define kAXFocusedApplicationAttribute "AXFocusedApplication"
#define kAXFocusedWindowAttribute "AXFocusedWindow"
#define kAXMinimizedAttribute "AXMinimized"
#define kAXFullScreenAttribute "AXFullScreen"
#define kAXPositionAttribute "AXPosition"
#define kAXSizeAttribute "AXSize"
#define kAXBundleIdentifierAttribute "AXBundleIdentifier"

// CG constants
#define kCGWindowNumber "kCGWindowNumber"
#define kCGWindowOwnerName "kCGWindowOwnerName"
#define kCGWindowOwnerPID "kCGWindowOwnerPID"
#define kCGWindowBounds "kCGWindowBounds"
#define kCGWindowLayer "kCGWindowLayer"
#define kCGWindowName "kCGWindowName"

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
	kCGWindowNumber := C.CFSTR("kCGWindowNumber")
	kCGWindowOwnerName := C.CFSTR("kCGWindowOwnerName")
	kCGWindowOwnerPID := C.CFSTR("kCGWindowOwnerPID")
	kCGWindowBounds := C.CFSTR("kCGWindowBounds")
	kCGWindowLayer := C.CFSTR("kCGWindowLayer")
	kCGWindowName := C.CFSTR("kCGWindowName")

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
		bounds.X = float64(C.get_double(boundsDict, C.CFSTR("X")))
		bounds.Y = float64(C.get_double(boundsDict, C.CFSTR("Y")))
		bounds.Width = float64(C.get_double(boundsDict, C.CFSTR("Width")))
		bounds.Height = float64(C.get_double(boundsDict, C.CFSTR("Height")))

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
	app := C.AXUIElementCreateApplication(pid)
	if app == 0 {
		return ""
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var bundleIDRef C.CFTypeRef
	result := C.AXUIElementCopyAttributeValue(C.AXUIElement(app), C.CFSTR("AXFocusedApplication"), &bundleIDRef)
	if result != C.AXError(kAXErrorSuccess) {
		return ""
	}
	if bundleIDRef != 0 {
		defer C.CFRelease(bundleIDRef)
	}

	// Get bundle ID from the application element
	var bundleID C.CFTypeRef
	result = C.AXUIElementCopyAttributeValue(C.AXUIElement(app), C.CFSTR("AXBundleIdentifier"), &bundleID)
	if result != C.AXError(kAXErrorSuccess) {
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
	app := C.AXUIElementCreateApplication(pid)
	if app == 0 {
		return false, false
	}
	defer C.CFRelease(C.CFTypeRef(app))

	// Get the focused window
	var window C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(C.AXUIElement(app), C.CFSTR("AXFocusedWindow"), (*C.AXValue)(unsafe.Pointer(&window)))
	if result != C.AXError(kAXErrorSuccess) {
		return false, false
	}
	if window != 0 {
		defer C.CFRelease(C.CFTypeRef(window))
	}

	// Check minimized state
	var minimizedRef C.CFTypeRef
	result = C.AXUIElementCopyAttributeValue(C.AXUIElement(window), C.CFSTR("AXMinimized"), &minimizedRef)
	if result == C.AXError(kAXErrorSuccess) && minimizedRef != 0 {
		isMinimized = C.CFBooleanGetValue(C.CFBooleanRef(minimizedRef)) != 0
		C.CFRelease(minimizedRef)
	}

	// Check full screen state
	var fullScreenRef C.CFTypeRef
	result = C.AXUIElementCopyAttributeValue(C.AXUIElement(window), C.CFSTR("AXFullScreen"), &fullScreenRef)
	if result == C.AXError(kAXErrorSuccess) && fullScreenRef != 0 {
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

	app := C.AXUIElementCreateApplication(C.int32_t(pid))
	if app == 0 {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var window C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(C.AXUIElement(app), C.CFSTR("AXFocusedWindow"), (*C.AXValue)(unsafe.Pointer(&window)))
	if result != C.AXError(kAXErrorSuccess) {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	if window != 0 {
		defer C.CFRelease(C.CFTypeRef(window))
	}

	position := C.CGPoint{X: C.CGFloat(x), Y: C.CGFloat(y)}
	positionValue := C.AXValueCreate(C.enum_AXValueType(kAXValueTypePoint), unsafe.Pointer(&position))
	if positionValue == 0 {
		return errors.New("failed to create position value")
	}
	defer C.CFRelease(C.CFTypeRef(positionValue))

	setResult := C.AXUIElementSetAttributeValue(C.AXUIElement(window), C.CFSTR("AXPosition"), C.CFTypeRef(positionValue))
	if setResult != C.AXError(kAXErrorSuccess) {
		return fmt.Errorf("failed to set position: %v", setResult)
	}

	return nil
}

// SetWindowSize sets window size
func SetWindowSize(pid int32, windowID uint32, width, height float64) error {
	if !CheckPermissions() {
		return errors.New("accessibility permissions not granted")
	}

	app := C.AXUIElementCreateApplication(C.int32_t(pid))
	if app == 0 {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var window C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(C.AXUIElement(app), C.CFSTR("AXFocusedWindow"), (*C.AXValue)(unsafe.Pointer(&window)))
	if result != C.AXError(kAXErrorSuccess) {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	if window != 0 {
		defer C.CFRelease(C.CFTypeRef(window))
	}

	size := C.CGSize{Width: C.CGFloat(width), Height: C.CGFloat(height)}
	sizeValue := C.AXValueCreate(C.enum_AXValueType(kAXValueTypeSize), unsafe.Pointer(&size))
	if sizeValue == 0 {
		return errors.New("failed to create size value")
	}
	defer C.CFRelease(C.CFTypeRef(sizeValue))

	setResult := C.AXUIElementSetAttributeValue(C.AXUIElement(window), C.CFSTR("AXSize"), C.CFTypeRef(sizeValue))
	if setResult != C.AXError(kAXErrorSuccess) {
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

	app := C.AXUIElementCreateApplication(C.int32_t(pid))
	if app == 0 {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var window C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(C.AXUIElement(app), C.CFSTR("AXFocusedWindow"), (*C.AXValue)(unsafe.Pointer(&window)))
	if result != C.AXError(kAXErrorSuccess) {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	if window != 0 {
		defer C.CFRelease(C.CFTypeRef(window))
	}

	setResult := C.AXUIElementSetAttributeValue(C.AXUIElement(window), C.CFSTR("AXMinimized"), C.CFTypeRef(C.kCFBooleanTrue))
	if setResult != C.AXError(kAXErrorSuccess) {
		return fmt.Errorf("failed to minimize window: %v", setResult)
	}

	return nil
}

// UnminimizeWindow unminimizes a window
func UnminimizeWindow(pid int32, windowID uint32) error {
	if !CheckPermissions() {
		return errors.New("accessibility permissions not granted")
	}

	app := C.AXUIElementCreateApplication(C.int32_t(pid))
	if app == 0 {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var window C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(C.AXUIElement(app), C.CFSTR("AXFocusedWindow"), (*C.AXValue)(unsafe.Pointer(&window)))
	if result != C.AXError(kAXErrorSuccess) {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	if window != 0 {
		defer C.CFRelease(C.CFTypeRef(window))
	}

	setResult := C.AXUIElementSetAttributeValue(C.AXUIElement(window), C.CFSTR("AXMinimized"), C.CFTypeRef(C.kCFBooleanFalse))
	if setResult != C.AXError(kAXErrorSuccess) {
		return fmt.Errorf("failed to unminimize window: %v", setResult)
	}

	return nil
}

// SetFullScreen sets window fullscreen state
func SetFullScreen(pid int32, windowID uint32, fullscreen bool) error {
	if !CheckPermissions() {
		return errors.New("accessibility permissions not granted")
	}

	app := C.AXUIElementCreateApplication(C.int32_t(pid))
	if app == 0 {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var window C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(C.AXUIElement(app), C.CFSTR("AXFocusedWindow"), (*C.AXValue)(unsafe.Pointer(&window)))
	if result != C.AXError(kAXErrorSuccess) {
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

	setResult := C.AXUIElementSetAttributeValue(C.AXUIElement(window), C.CFSTR("AXFullScreen"), C.CFTypeRef(fsValue))
	if setResult != C.AXError(kAXErrorSuccess) {
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
