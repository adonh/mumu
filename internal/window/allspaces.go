package window

/*
#cgo CFLAGS: -x objective-c
#include "../native/mimi.h"
*/
import "C"

import (
	"unsafe"

	derrors "github.com/y3owk1n/mimi/internal/errors"
)

// AcrossSpacesEntry represents one window discovered across all Mission
// Control spaces (not just the active one), for use by the layout
// save/restore capability.
type AcrossSpacesEntry struct {
	// WindowID is the window's CGWindowID, sufficient on its own (no
	// AXUIElementRef needed) to move the window via MoveToSpace.
	WindowID   uint32
	BundleID   string
	Title      string
	SpaceID    uint64
	Fullscreen bool
}

// MoveToSpace moves this window to the Mission Control space with the
// given macOS space ID, without altering its position or size. Unlike the
// frontmost-window move used by "mimi action move_window_to_space", this
// deliberately does not activate the target display, since layout restore
// may move many windows across many displays in one pass.
func (e AcrossSpacesEntry) MoveToSpace(spaceID uint64) error {
	return MoveWindowIDToSpace(e.WindowID, spaceID)
}

// MoveWindowIDToSpace moves the window with the given CGWindowID to the
// Mission Control space with the given macOS space ID. See
// AcrossSpacesEntry.MoveToSpace for behavior notes.
func MoveWindowIDToSpace(windowID uint32, spaceID uint64) error {
	result := C.MimiMoveWindowIDToSpace(C.uint32_t(windowID), C.uint64_t(spaceID))
	if result == 0 {
		return derrors.New(derrors.CodeActionFailed, "failed to move window to space")
	}

	return nil
}

// AllAcrossSpaces enumerates windows across all Mission Control spaces for
// every running regular application, resolving each window's current space
// ID via mimi's existing private WindowServer connection. Unlike
// AllFocusableOnActiveSpace, this is not limited to the currently active
// space. Minimized windows are excluded when that can be determined (see
// MimiGetAllWindowsAcrossSpaces).
func AllAcrossSpaces() ([]AcrossSpacesEntry, error) {
	var count C.int

	infoPtr := C.MimiGetAllWindowsAcrossSpaces(&count)
	if infoPtr == nil || count == 0 {
		return nil, nil
	}
	defer C.MimiFreeLayoutWindowInfo(infoPtr, count) //nolint:nlreturn

	total := int(count)
	cSlice := (*[1 << 20]C.MimiLayoutWindowInfo)(unsafe.Pointer(infoPtr))[:total:total]

	result := make([]AcrossSpacesEntry, total)
	for i, info := range cSlice {
		result[i] = AcrossSpacesEntry{
			WindowID:   uint32(info.wid),
			BundleID:   C.GoString(info.bundleID),
			Title:      C.GoString(info.title),
			SpaceID:    uint64(info.sid),
			Fullscreen: info.fullscreen != 0,
		}
	}

	return result, nil
}
