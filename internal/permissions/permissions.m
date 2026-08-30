#import "permissions.h"

#import <ApplicationServices/ApplicationServices.h>
#import <Cocoa/Cocoa.h>

int MumuCheckAccessibilityPermissions(void) {
	// AXIsProcessTrustedWithOptions(NULL) is the currently documented
	// non-prompting trust check (equivalent to the older AXIsProcessTrusted()
	// for a read-only check, but the entry point Apple's current guidance
	// points to, and it leaves room for an explicit prompting variant later
	// via kAXTrustedCheckOptionPrompt).
	//
	// Note: both functions share the same per-process TCC trust cache, which
	// can go stale mid-lifetime for a long-running process that polls this
	// repeatedly while permissions change underneath it. That's not the
	// cause of mumu reporting stale Accessibility state after a rebuild —
	// mumu is a short-lived CLI (fresh process per invocation), so each call
	// here gets a fresh TCC lookup regardless of which function is used. The
	// actual cause of that symptom is TCC pinning the grant to the binary's
	// code identity, which changes on every unsigned/ad-hoc rebuild; see the
	// codesigning setup in the justfile (`build`/`bundle` recipes).
	Boolean trusted = AXIsProcessTrustedWithOptions(NULL);
	return trusted ? 1 : 0;
}

int MumuCheckScreenRecordingPermissions(void) {
	// CGPreflightScreenCaptureAccess is a read-only check: it never prompts
	// and has no side effects.
	return CGPreflightScreenCaptureAccess() ? 1 : 0;
}
