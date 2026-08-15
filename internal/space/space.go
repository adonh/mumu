package space

/*
#cgo CFLAGS: -x objective-c
#include "../native/mumu.h"
*/
import "C"

import (
	derrors "github.com/adonh/mumu/internal/errors"
	_ "github.com/adonh/mumu/internal/native"
)

// Focus focuses the Mission Control space at the given 1-based index.
func Focus(index int) error {
	count := int(C.MumuCountMissionControlSpaces())
	if count == 0 {
		return derrors.New(derrors.CodeActionFailed, "failed to enumerate Mission Control spaces")
	}

	if index < 1 || index > count {
		return derrors.Newf(
			derrors.CodeInvalidInput,
			"space number %d is out of range; valid range is 1..%d",
			index,
			count,
		)
	}

	sid := uint64(C.MumuMissionControlSpaceID(C.int(index)))
	if sid == 0 {
		return derrors.Newf(
			derrors.CodeActionFailed,
			"failed to resolve Mission Control space at index %d",
			index,
		)
	}

	did := uint32(C.MumuSpaceDisplayID(C.uint64_t(sid)))
	if did == 0 {
		return derrors.Newf(
			derrors.CodeActionFailed,
			"failed to resolve display for Mission Control space at index %d",
			index,
		)
	}

	if C.MumuFocusSpaceUsingGesture(C.uint32_t(did), C.uint64_t(sid)) == 0 {
		return derrors.New(derrors.CodeActionFailed, "failed to focus Mission Control space")
	}

	return nil
}

// Count returns the total number of Mission Control spaces.
func Count() int {
	return int(C.MumuCountMissionControlSpaces())
}

// MissionControlIndexForSpace returns the 1-based Mission Control index for
// a given macOS space ID — the same numbering Focus/MoveWindow accept, and
// the same ordinal macOS's own "Switch to Desktop <n>" keyboard shortcut
// uses — or 0 if the space ID is not currently known.
func MissionControlIndexForSpace(sid uint64) int {
	return int(C.MumuMissionControlIndexForSpace(C.uint64_t(sid)))
}

// ActiveIndex returns the 1-based index of the currently active space.
func ActiveIndex() (int, error) {
	count := Count()
	if count == 0 {
		return 0, derrors.New(
			derrors.CodeActionFailed,
			"failed to enumerate Mission Control spaces",
		)
	}

	activeID := uint64(C.MumuActiveSpaceID())
	if activeID == 0 {
		return 0, derrors.New(derrors.CodeActionFailed, "failed to resolve active space ID")
	}

	for i := 1; i <= count; i++ {
		sid := uint64(C.MumuMissionControlSpaceID(C.int(i)))
		if sid == activeID {
			return i, nil
		}
	}

	return 0, derrors.New(derrors.CodeActionFailed, "active space not found in space enumeration")
}

// PrimaryDisplayCurrentSpace returns the macOS space ID and logical
// left-to-right index of the Space currently shown on the configured primary
// display.
func PrimaryDisplayCurrentSpace() (uint64, int, error) {
	return primaryDisplayCurrentSpace(
		uint64(C.MumuPrimaryDisplaySpaceID()),
		LogicalIndexForSpace,
	)
}

// PrimaryDisplayCurrentLogicalIndex returns the logical left-to-right index
// of the Space currently shown on the configured primary display.
func PrimaryDisplayCurrentLogicalIndex() (int, error) {
	_, logicalIndex, err := PrimaryDisplayCurrentSpace()

	return logicalIndex, err
}

func primaryDisplayCurrentSpace(
	spaceID uint64,
	indexForSpace func(uint64) int,
) (uint64, int, error) {
	if spaceID == 0 {
		return 0, 0, derrors.New(
			derrors.CodeActionFailed,
			"failed to resolve the primary display's current space",
		)
	}

	logicalIndex := indexForSpace(spaceID)
	if logicalIndex == 0 {
		return 0, 0, derrors.New(
			derrors.CodeActionFailed,
			"primary display's current space is not in the logical space ordering",
		)
	}

	return spaceID, logicalIndex, nil
}

// MoveWindow moves the frontmost window to the space at the given 1-based index.
func MoveWindow(index int) error {
	count := int(C.MumuCountMissionControlSpaces())
	if count == 0 {
		return derrors.New(derrors.CodeActionFailed, "failed to enumerate Mission Control spaces")
	}

	if index < 1 || index > count {
		return derrors.Newf(
			derrors.CodeInvalidInput,
			"space number %d is out of range; valid range is 1..%d",
			index,
			count,
		)
	}

	sid := uint64(C.MumuMissionControlSpaceID(C.int(index)))
	if sid == 0 {
		return derrors.Newf(
			derrors.CodeActionFailed,
			"failed to resolve Mission Control space at index %d",
			index,
		)
	}

	frontmost := C.MumuGetFrontmostWindow()
	if frontmost == nil {
		return derrors.New(
			derrors.CodeActionFailed,
			"no active window found to move",
		)
	}

	defer C.MumuReleaseElement(frontmost) //nolint:nlreturn

	if C.MumuMoveWindowToSpace(frontmost, C.uint64_t(sid)) == 0 { //nolint:nlreturn
		return derrors.New(derrors.CodeActionFailed, "failed to move window to space")
	}

	targetDid := uint32(C.MumuSpaceDisplayID(C.uint64_t(sid)))
	if targetDid != 0 && targetDid != uint32(C.MumuCursorDisplayID()) {
		C.MumuActivateDisplay(C.uint32_t(targetDid))
	}

	return nil
}
