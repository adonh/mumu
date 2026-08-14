#import "permissions.h"

#import <ApplicationServices/ApplicationServices.h>
#import <Cocoa/Cocoa.h>

int MumuCheckAccessibilityPermissions(void) {
	Boolean trusted = AXIsProcessTrusted();
	return trusted ? 1 : 0;
}

int MumuCheckScreenRecordingPermissions(void) {
	// CGPreflightScreenCaptureAccess is a read-only check: it never prompts
	// and has no side effects.
	return CGPreflightScreenCaptureAccess() ? 1 : 0;
}
