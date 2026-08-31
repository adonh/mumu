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

// Warnings returns a one-line warning for each permission mumu needs that
// Check reported as not granted, instead of a hard error. Callers that
// want to proceed even when a permission looks missing (rather than
// blocking via FriendlyError) use this: the check itself can report a
// false negative in some environments — for example, an enterprise
// MDM-managed Mac whose PPPC policy silently overrides a manual grant
// underneath an otherwise-correct-looking System Settings toggle (see
// docs/TROUBLESHOOTING.md) — so treating "not granted" as fatal can block
// an operation that would actually have succeeded. The real native calls
// mumu makes will fail on their own if the permission is genuinely
// missing, which is a more reliable signal than this preflight check.
func Warnings(result CheckResult) []string {
	var warnings []string

	if !result.Accessibility {
		warnings = append(warnings, "Accessibility permission does not appear to be granted; "+
			"continuing anyway, since this check can be wrong (see docs/TROUBLESHOOTING.md). "+
			"Window/Space moves will fail below if it's genuinely missing.")
	}

	if !result.ScreenRecording {
		warnings = append(warnings, "Screen Recording permission does not appear to be granted; "+
			"continuing anyway, since this check can be wrong. Window titles used for matching "+
			"may come back empty if it's genuinely missing, which can cause unmatched entries.")
	}

	return warnings
}
