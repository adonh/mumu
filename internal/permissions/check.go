package permissions

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa -framework ApplicationServices
#include "permissions.h"
*/
import "C"

import (
	derrors "github.com/adonh/mumu/internal/errors"
)

// CheckResult holds the results of a permissions check.
type CheckResult struct {
	Accessibility    bool
	AccessibilityMsg string

	ScreenRecording    bool
	ScreenRecordingMsg string
}

// Check verifies the macOS permissions mumu needs: Accessibility (for
// window/Space control) and Screen Recording. Screen Recording is required
// because window titles (used to match windows across Mission Control
// Spaces during restore) come from CGWindowListCopyWindowInfo, which
// redacts them to empty system-wide unless the process holds that
// permission.
func Check() CheckResult {
	res := CheckResult{
		Accessibility: C.MumuCheckAccessibilityPermissions() != 0,
	}
	if !res.Accessibility {
		res.AccessibilityMsg = `Accessibility permission is required for window/Space control.

  Grant it in:
    System Settings -> Privacy & Security -> Accessibility -> enable "mumu"

  After granting, re-run the command.`
	}

	res.ScreenRecording = C.MumuCheckScreenRecordingPermissions() != 0
	if !res.ScreenRecording {
		res.ScreenRecordingMsg = `Screen Recording permission is required.

  Grant it in:
    System Settings -> Privacy & Security -> Screen Recording -> enable "mumu"

  After granting, re-run the command. Without it, window titles can't be
  read for windows outside the currently displayed Space, which save/restore
  needs to reliably match windows.`
	}

	return res
}

// FriendlyError returns an error if either permission mumu needs
// (Accessibility, Screen Recording) is denied.
func FriendlyError(result CheckResult) error {
	if !result.Accessibility {
		return derrors.New(derrors.CodeAccessibilityDenied, result.AccessibilityMsg)
	}

	if !result.ScreenRecording {
		return derrors.New(derrors.CodeScreenRecordingDenied, result.ScreenRecordingMsg)
	}

	return nil
}
