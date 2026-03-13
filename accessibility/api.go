//go:build darwin && !fallback
// +build darwin,!fallback

package accessibility

/*
#cgo LDFLAGS: -framework ApplicationServices

#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>

// Check if we have accessibility permissions
bool has_permissions() {
    return AXIsProcessTrusted();
}

// Get windows using CGWindowList
CFDictionaryRef* get_window_list(uint32_t options, uint32_t windowID, int *count) {
    CFArrayRef list = CGWindowListCopyWindowInfo(options, windowID);
    if (list == NULL) {
        *count = 0;
        return NULL;
    }
    
    *count = CFArrayGetCount(list);
    CFDictionaryRef **result = malloc(*count * sizeof(CFDictionaryRef*));
    
    for (int i = 0; i < *count; i++) {
        result[i] = (CFDictionaryRef*)CFArrayGetValueAtIndex(list, i);
    }
    
    CFRelease(list);
    return result;
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

// Get string value from dictionary
char* get_string(CFDictionaryRef dict, CFStringRef key) {
    CFStringRef value;
    if (!CFDictionaryGetValueIfPresent(dict, key, (const void**)&value)) {
        return "";
    }
    
    char *result = malloc(1024);
    if (CFStringGetCString(value, result, 1024, kCFStringEncodingUTF8)) {
        return result;
    }
    free(result);
    return "";
}

// Get sub-dictionary
void* get_dict(CFDictionaryRef dict, CFStringRef key) {
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
	defer C.free(unsafe.Pointer(windowList))

	windows := make([]WindowInfo, 0, int(count))

	// Get keys
	kCGWindowNumber := C.CFStringCreateWithUTF8String(C.CFSTR("kCGWindowNumber"))
	kCGWindowOwnerName := C.CFStringCreateWithUTF8String(C.CFSTR("kCGWindowOwnerName"))
	kCGWindowOwnerPID := C.CFStringCreateWithUTF8String(C.CFSTR("kCGWindowOwnerPID"))
	kCGWindowBounds := C.CFStringCreateWithUTF8String(C.CFSTR("kCGWindowBounds"))
	kCGWindowLayer := C.CFStringCreateWithUTF8String(C.CFSTR("kCGWindowLayer"))
	kCGWindowName := C.CFStringCreateWithUTF8String(C.CFSTR("kCGWindowName"))
	kCGBoundsX := C.CFStringCreateWithUTF8String(C.CFSTR("X"))
	kCGBoundsY := C.CFStringCreateWithUTF8String(C.CFSTR("Y"))
	kCGBoundsWidth := C.CFStringCreateWithUTF8String(C.CFSTR("Width"))
	kCGBoundsHeight := C.CFStringCreateWithUTF8String(C.CFSTR("Height"))

	// Get active displays for monitor determination
	displays, displayCount := GetDisplays()

	for i := 0; i < int(count); i++ {
		dict := windowList[i]

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
		bounds.X = float64(C.get_double(boundsDict, kCGBoundsX))
		bounds.Y = float64(C.get_double(boundsDict, kCGBoundsY))
		bounds.Width = float64(C.get_double(boundsDict, kCGBoundsWidth))
		bounds.Height = float64(C.get_double(boundsDict, kCGBoundsHeight))

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
			Title:      windowName,
			Bounds:     bounds,
			BundleID:   bundleID,
			AppName:    ownerName,
			ProcessID:  int32(pid),
			WindowID:   uint32(windowNumber),
			DisplayID:  displayID,
			IsMinimized: isMinimized,
			IsFullScreen: isFullScreen,
		})
	}

	return windows, nil
}

// getBundleIDFromPID gets bundle ID from process ID
func getBundleIDFromPID(pid C.int32_t) string {
	app := C.AXUIElementCreateApplication(pid)
	if app == nil {
		return ""
	}
	defer C.release(unsafe.Pointer(app))

	var appElement C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(app, C.kAXFocusedApplicationAttribute, (*C.AXValue)(unsafe.Pointer(&appElement)))
	if result != C.AXError(0) {
		return ""
	}
	defer C.release(unsafe.Pointer(appElement))

	var bundleIDRef C.CFTypeRef
	result = C.AXUIElementCopyAttributeValue(appElement, C.kAXAXBundleIdentifierAttribute, &bundleIDRef)
	if result != C.AXError(0) {
		return ""
	}
	defer C.CFRelease(bundleIDRef)

	cfStr := C.CFStringRef(bundleIDRef)
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
	if app == nil {
		return false, false
	}
	defer C.release(unsafe.Pointer(app))

	// Get the focused window
	var window C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(app, C.kAXFocusedWindowAttribute, (*C.AXValue)(unsafe.Pointer(&window)))
	if result != C.AXError(0) {
		return false, false
	}
	defer C.release(unsafe.Pointer(window))

	// Check minimized state
	var minimizedRef C.CFTypeRef
	result = C.AXUIElementCopyAttributeValue(window, C.kAXMinimizedAttribute, &minimizedRef)
	if result == C.AXError(0) && minimizedRef != nil {
		if cfBool, ok := minimizedRef.(C.CFBooleanRef); ok {
			isMinimized = C.CFBooleanGetValue(cfBool) != 0
		}
		C.CFRelease(minimizedRef)
	}

	// Check full screen state
	var fullScreenRef C.CFTypeRef
	result = C.AXUIElementCopyAttributeValue(window, C.kAXFullScreenAttribute, &fullScreenRef)
	if result == C.AXError(0) && fullScreenRef != nil {
		if cfBool, ok := fullScreenRef.(C.CFBooleanRef); ok {
			isFullScreen = C.CFBooleanGetValue(cfBool) != 0
		}
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
			ID:       displayID,
			Bounds:   bounds,
			IsMain:   isMain,
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
	if app == nil {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.release(unsafe.Pointer(app))

	var window C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(app, C.kAXFocusedWindowAttribute, (*C.AXValue)(unsafe.Pointer(&window)))
	if result != C.AXError(0) {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	defer C.release(unsafe.Pointer(window))

	position := C.CGPoint{X: C.CGFloat(x), Y: C.CGFloat(y)}
	positionValue := C.CAXValueCreate(C.kAXValueTypeCGPoint, unsafe.Pointer(&position))
	if positionValue == nil {
		return errors.New("failed to create position value")
	}
	defer C.CFRelease(unsafe.Pointer(positionValue))

	setResult := C.AXUIElementSetAttributeValue(window, C.kAXPositionAttribute, C.CFTypeRef(positionValue))
	if setResult != C.AXError(0) {
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
	if app == nil {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.release(unsafe.Pointer(app))

	var window C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(app, C.kAXFocusedWindowAttribute, (*C.AXValue)(unsafe.Pointer(&window)))
	if result != C.AXError(0) {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	defer C.release(unsafe.Pointer(window))

	size := C.CGSize{Width: C.CGFloat(width), Height: C.CGFloat(height)}
	sizeValue := C.CAXValueCreate(C.kAXValueTypeCGSize, unsafe.Pointer(&size))
	if sizeValue == nil {
		return errors.New("failed to create size value")
	}
	defer C.CFRelease(unsafe.Pointer(sizeValue))

	setResult := C.AXUIElementSetAttributeValue(window, C.kAXSizeAttribute, C.CFTypeRef(sizeValue))
	if setResult != C.AXError(0) {
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
	if app == nil {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.release(unsafe.Pointer(app))

	var window C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(app, C.kAXFocusedWindowAttribute, (*C.AXValue)(unsafe.Pointer(&window)))
	if result != C.AXError(0) {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	defer C.release(unsafe.Pointer(window))

	minimized := C.CFBooleanCreate(C.kCFBooleanTrue)
	defer C.CFRelease(unsafe.Pointer(minimized))

	setResult := C.AXUIElementSetAttributeValue(window, C.kAXMinimizedAttribute, C.CFTypeRef(minimized))
	if setResult != C.AXError(0) {
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
	if app == nil {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.release(unsafe.Pointer(app))

	var window C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(app, C.kAXFocusedWindowAttribute, (*C.AXValue)(unsafe.Pointer(&window)))
	if result != C.AXError(0) {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	defer C.release(unsafe.Pointer(window))

	notMinimized := C.CFBooleanCreate(C.kCFBooleanFalse)
	defer C.CFRelease(unsafe.Pointer(notMinimized))

	setResult := C.AXUIElementSetAttributeValue(window, C.kAXMinimizedAttribute, C.CFTypeRef(notMinimized))
	if setResult != C.AXError(0) {
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
	if app == nil {
		return fmt.Errorf("failed to create AXUIElement for pid %d", pid)
	}
	defer C.release(unsafe.Pointer(app))

	var window C.AXUIElement
	result := C.AXUIElementCopyAttributeValue(app, C.kAXFocusedWindowAttribute, (*C.AXValue)(unsafe.Pointer(&window)))
	if result != C.AXError(0) {
		return fmt.Errorf("failed to get focused window: %v", result)
	}
	defer C.release(unsafe.Pointer(window))

	var fsValue C.CFBooleanRef
	if fullscreen {
		fsValue = C.CFBooleanCreate(C.kCFBooleanTrue)
	} else {
		fsValue = C.CFBooleanCreate(C.kCFBooleanFalse)
	}
	defer C.CFRelease(unsafe.Pointer(fsValue))

	setResult := C.AXUIElementSetAttributeValue(window, C.kAXFullScreenAttribute, C.CFTypeRef(fsValue))
	if setResult != C.AXError(0) {
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
