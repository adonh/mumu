package permissions

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa -framework ApplicationServices
#include <stdlib.h>
#include "permissions.h"
*/
import "C"

import (
	"unsafe"

	derrors "github.com/y3owk1n/mimi/internal/errors"
)

// ConfigOnboardingChoice represents the user's choice in the config onboarding alert.
type ConfigOnboardingChoice int

// AccessibilityStartupChoice represents the user's choice in the startup permission alert.
type AccessibilityStartupChoice int

const (
	// ConfigOnboardingCreate indicates the user chose to create a config file.
	ConfigOnboardingCreate ConfigOnboardingChoice = 1
	// ConfigOnboardingQuit indicates the user chose to quit.
	ConfigOnboardingQuit ConfigOnboardingChoice = 2

	// AccessibilityStartupGranted indicates accessibility permission is granted.
	AccessibilityStartupGranted AccessibilityStartupChoice = 1
	// AccessibilityStartupQuit indicates the user chose to quit.
	AccessibilityStartupQuit AccessibilityStartupChoice = 2
	// AccessibilityStartupRestartRequired indicates user needs to restart the app after granting permission.
	AccessibilityStartupRestartRequired AccessibilityStartupChoice = 3
)

// CheckResult holds the results of a permissions check.
type CheckResult struct {
	Accessibility    bool
	AccessibilityMsg string

	// ScreenRecording and ScreenRecordingMsg are only populated by
	// CheckLayout; Check leaves them zero-valued since only the layout
	// save/restore capability requires Screen Recording.
	ScreenRecording    bool
	ScreenRecordingMsg string
}

// Check verifies macOS accessibility permissions.
func Check() CheckResult {
	trusted := C.MimiCheckAccessibilityPermissions() != 0
	res := CheckResult{Accessibility: trusted}
	if !trusted {
		res.AccessibilityMsg = `Accessibility permission is required for window/space actions and window focus events.

  Grant it in:
    System Settings -> Privacy & Security -> Accessibility -> enable "mimi"

  After granting, restart mimi (or re-run the action command).
  Window events (on_window_focus, on_window_title_change, etc.) and
  action commands (focus_window, space, move_window_to_space) will be
  unavailable until the permission is granted.`
	}

	return res
}

// CheckLayout verifies the permissions required by the layout save/restore
// capability ("mimi layout ..."): Accessibility (for window/space control,
// same as Check) plus Screen Recording. Screen Recording is required
// because window titles (used to match windows across Mission Control
// Spaces) come from CGWindowListCopyWindowInfo, which redacts them to empty
// system-wide unless the process holds that permission.
func CheckLayout() CheckResult {
	res := Check()

	trusted := C.MimiCheckScreenRecordingPermissions() != 0
	res.ScreenRecording = trusted
	if !trusted {
		res.ScreenRecordingMsg = `Screen Recording permission is required for "mimi layout" commands.

  Grant it in:
    System Settings -> Privacy & Security -> Screen Recording -> enable "mimi"

  After granting, restart mimi (or re-run the layout command).
  Without it, window titles can't be read for windows outside the
  currently displayed Space, which layout save/restore needs to reliably
  match windows.`
	}

	return res
}

// RequestAccessibility asks macOS to start the accessibility permission flow.
func RequestAccessibility() bool {
	return C.MimiRequestAccessibilityPermissions() != 0
}

// RequestScreenRecording asks macOS to start the Screen Recording
// permission flow (used only by the layout save/restore capability).
func RequestScreenRecording() bool {
	return C.MimiRequestScreenRecordingPermissions() != 0
}

// ShowConfigOnboardingAlert displays startup guidance for creating the first config file.
func ShowConfigOnboardingAlert(configPath string) ConfigOnboardingChoice {
	cPath := C.CString(configPath)
	defer C.free(unsafe.Pointer(cPath)) //nolint:nlreturn

	return ConfigOnboardingChoice(C.MimiShowConfigOnboardingAlert(cPath))
}

// ShowAccessibilityStartupAlert displays startup guidance for granting accessibility permission.
func ShowAccessibilityStartupAlert() AccessibilityStartupChoice {
	return AccessibilityStartupChoice(C.MimiShowAccessibilityPermissionStartupAlert())
}

// FriendlyError returns an error if accessibility permission is denied.
func FriendlyError(r CheckResult) error {
	if r.Accessibility {
		return nil
	}

	return derrors.New(derrors.CodeAccessibilityDenied, r.AccessibilityMsg)
}

// FriendlyErrorLayout returns an error if either permission required by the
// layout save/restore capability (Accessibility, Screen Recording) is
// denied.
func FriendlyErrorLayout(result CheckResult) error {
	if !result.Accessibility {
		return derrors.New(derrors.CodeAccessibilityDenied, result.AccessibilityMsg)
	}

	if !result.ScreenRecording {
		return derrors.New(derrors.CodeScreenRecordingDenied, result.ScreenRecordingMsg)
	}

	return nil
}
