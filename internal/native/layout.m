//
//  layout.m
//  mumu
//
//  Window enumeration across all Mission Control Spaces for the layout
//  save/restore capability.
//
//  AX enumeration (kAXWindowsAttribute) only ever exposes windows on the
//  Space currently *displayed* on each connected display; windows parked on
//  any other Space are invisible to AX no matter how it's queried — this was
//  confirmed empirically (moving a window to a non-displayed Space made it
//  vanish from its owning app's kAXWindowsAttribute list entirely). Instead,
//  this enumerates via CGWindowListCopyWindowInfo(kCGWindowListOptionAll,
//  ...), which does cover every Space, and resolves each window's Space via
//  the private CGSCopySpacesForWindows call.
//
//  Trade-offs of this approach (see the layout-save-restore change's
//  design.md for the full writeup):
//   - Window titles (kCGWindowName) are redacted to NULL system-wide unless
//     the process has Screen Recording permission, which mumu's layout
//     capability therefore requires in addition to Accessibility.
//   - Minimized state can only be determined for windows on a currently
//     displayed Space, by checking membership in the
//     kCGWindowListOptionOnScreenOnly result set (which — unlike
//     kCGWindowListOptionAll — omits minimized windows). Windows on other
//     Spaces are always included regardless of minimized state, since
//     there's no reliable per-window minimized signal available for them.
//

#import "mumu.h"
#import "mumu_log.h"

#import <Cocoa/Cocoa.h>
#import <CoreFoundation/CoreFoundation.h>
#import <CoreGraphics/CoreGraphics.h>
#import <Foundation/Foundation.h>
#import <stdlib.h>
#import <string.h>

// Private SkyLight / WindowServer symbols (not in the public SDK). See
// docs/ARCHITECTURE.md for the private-API risk notes that already apply to
// mumu's existing Space primitives; this file follows the same pattern.
extern int SLSMainConnectionID(void);
extern CFArrayRef SLSCopyManagedDisplaySpaces(int cid);
extern CFArrayRef CGSCopySpacesForWindows(int cid, uint32_t mask, CFArrayRef windowIDs);
extern uint64_t SLSManagedDisplayGetCurrentSpace(int cid, CFStringRef uuid);

// Space type value for fullscreen Spaces, per SLSCopyManagedDisplaySpaces'
// "type" field (0 = standard/user, 4 = fullscreen). Undocumented; derived
// from empirical observation and prior art in other private-API tooling.
static const int kMumuSpaceTypeFullscreen = 4;

// Mask for CGSCopySpacesForWindows: include the window's current space plus
// other/user spaces. Matches prior art usage in other CGSInternal-based
// tooling.
static const uint32_t kMumuAllSpacesMask = 0x7;

#pragma mark - String Helpers

/// Copies a CFStringRef into a malloc'd, NUL-terminated UTF-8 C string.
/// Never returns NULL: an empty string is returned if str is NULL or the
/// conversion fails, so callers can always safely free() the result.
static char *mumuCopyUTF8String(CFStringRef str) {
	if (!str) {
		return strdup("");
	}

	CFIndex length = CFStringGetLength(str);
	CFIndex maxSize = CFStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8) + 1;

	char *buf = (char *)malloc((size_t)maxSize);
	if (!buf) {
		return strdup("");
	}

	if (!CFStringGetCString(str, buf, maxSize, kCFStringEncodingUTF8)) {
		buf[0] = '\0';
	}

	return buf;
}

#pragma mark - Fullscreen / Active Space Lookup

/// Builds the set of Space IDs whose type is fullscreen, for O(1) lookup
/// while enumerating windows.
static NSSet<NSNumber *> *mumuCopyFullscreenSpaceIDs(void) {
	NSMutableSet<NSNumber *> *fullscreenSids = [NSMutableSet set];

	CFArrayRef displaySpaces = SLSCopyManagedDisplaySpaces(SLSMainConnectionID());
	if (!displaySpaces) {
		return fullscreenSids;
	}

	CFIndex displayCount = CFArrayGetCount(displaySpaces);
	for (CFIndex i = 0; i < displayCount; i++) {
		CFDictionaryRef displayRef = (CFDictionaryRef)CFArrayGetValueAtIndex(displaySpaces, i);
		CFArrayRef spacesRef = (CFArrayRef)CFDictionaryGetValue(displayRef, CFSTR("Spaces"));
		if (!spacesRef) {
			continue;
		}

		CFIndex spacesCount = CFArrayGetCount(spacesRef);
		for (CFIndex j = 0; j < spacesCount; j++) {
			CFDictionaryRef spaceRef = (CFDictionaryRef)CFArrayGetValueAtIndex(spacesRef, j);
			CFNumberRef typeRef = (CFNumberRef)CFDictionaryGetValue(spaceRef, CFSTR("type"));

			int type = 0;
			if (typeRef) {
				CFNumberGetValue(typeRef, kCFNumberIntType, &type);
			}

			if (type != kMumuSpaceTypeFullscreen) {
				continue;
			}

			CFNumberRef sidRef = (CFNumberRef)CFDictionaryGetValue(spaceRef, CFSTR("id64"));
			if (!sidRef) {
				continue;
			}

			uint64_t sid = 0;
			CFNumberGetValue(sidRef, CFNumberGetType(sidRef), &sid);
			[fullscreenSids addObject:@(sid)];
		}
	}

	CFRelease(displaySpaces);

	return fullscreenSids;
}

/// Builds the set of Space IDs currently displayed on some connected
/// display (i.e. each display's "current space"). Used to scope
/// minimized-window detection: only windows on one of these Spaces can be
/// reliably checked against the on-screen window list.
static NSSet<NSNumber *> *mumuCopyActiveSpaceIDs(void) {
	NSMutableSet<NSNumber *> *activeSids = [NSMutableSet set];

	int cid = SLSMainConnectionID();
	CFArrayRef displaySpaces = SLSCopyManagedDisplaySpaces(cid);
	if (!displaySpaces) {
		return activeSids;
	}

	CFIndex displayCount = CFArrayGetCount(displaySpaces);
	for (CFIndex i = 0; i < displayCount; i++) {
		CFDictionaryRef displayRef = (CFDictionaryRef)CFArrayGetValueAtIndex(displaySpaces, i);
		CFStringRef uuid = (CFStringRef)CFDictionaryGetValue(displayRef, CFSTR("Display Identifier"));
		if (!uuid) {
			continue;
		}

		uint64_t sid = SLSManagedDisplayGetCurrentSpace(cid, uuid);
		if (sid != 0) {
			[activeSids addObject:@(sid)];
		}
	}

	CFRelease(displaySpaces);

	return activeSids;
}

#pragma mark - Window-to-Space Resolution

/// Resolves a single window ID to its current Space ID via the private
/// CGSCopySpacesForWindows call. Windows are resolved one at a time (rather
/// than batched) because passing multiple window IDs in one call returns a
/// deduplicated, unordered union of Spaces rather than a 1:1 per-window
/// mapping.
static uint64_t mumuSpaceIDForWindow(int cid, CGWindowID wid) {
	CFNumberRef windowNumber = CFNumberCreate(NULL, kCFNumberIntType, &wid);
	if (!windowNumber) {
		return 0;
	}

	CFArrayRef windowIDs = CFArrayCreate(NULL, (const void **)&windowNumber, 1, &kCFTypeArrayCallBacks);
	CFRelease(windowNumber);
	if (!windowIDs) {
		return 0;
	}

	CFArrayRef spaces = CGSCopySpacesForWindows(cid, kMumuAllSpacesMask, windowIDs);
	CFRelease(windowIDs);

	if (!spaces) {
		return 0;
	}

	uint64_t sid = 0;
	if (CFArrayGetCount(spaces) > 0) {
		CFNumberRef sidRef = (CFNumberRef)CFArrayGetValueAtIndex(spaces, 0);
		CFNumberGetValue(sidRef, kCFNumberSInt64Type, &sid);
	}

	CFRelease(spaces);

	return sid;
}

#pragma mark - On-Screen Window Set

/// Builds the set of window IDs currently in the "on-screen" window list
/// (kCGWindowListOptionOnScreenOnly), which — unlike kCGWindowListOptionAll
/// — omits minimized windows. Membership only means something for windows
/// whose Space is currently displayed: windows on other Spaces are never
/// "on-screen" regardless of minimized state, so absence from this set
/// can't by itself be used to infer minimized state for them.
static NSSet<NSNumber *> *mumuCopyOnscreenWindowIDs(void) {
	NSMutableSet<NSNumber *> *onscreen = [NSMutableSet set];

	CFArrayRef windowList = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly, kCGNullWindowID);
	if (!windowList) {
		return onscreen;
	}

	CFIndex count = CFArrayGetCount(windowList);
	for (CFIndex i = 0; i < count; i++) {
		CFDictionaryRef info = (CFDictionaryRef)CFArrayGetValueAtIndex(windowList, i);
		CFNumberRef widRef = (CFNumberRef)CFDictionaryGetValue(info, kCGWindowNumber);
		if (!widRef) {
			continue;
		}

		int wid = 0;
		CFNumberGetValue(widRef, kCFNumberIntType, &wid);
		[onscreen addObject:@(wid)];
	}

	CFRelease(windowList);

	return onscreen;
}

#pragma mark - Regular Application Lookup

/// Builds a PID -> bundle identifier map covering every running regular
/// (Dock-visible) application. Used both to filter CGWindowList's
/// system-wide results down to real user application windows (excluding the
/// Dock, menu bar items, background agents, etc.) and to attach a bundle ID
/// to each window without a second AX round-trip.
static NSDictionary<NSNumber *, NSString *> *mumuCopyRegularAppBundleIDsByPID(void) {
	NSMutableDictionary<NSNumber *, NSString *> *byPID = [NSMutableDictionary dictionary];

	for (NSRunningApplication *app in [NSWorkspace sharedWorkspace].runningApplications) {
		if (app.activationPolicy != NSApplicationActivationPolicyRegular) {
			continue;
		}

		byPID[@(app.processIdentifier)] = app.bundleIdentifier ?: @"";
	}

	return byPID;
}

#pragma mark - Public Layout Window Enumeration API

MumuLayoutWindowInfo *MumuGetAllWindowsAcrossSpaces(int *count) {
	if (!count) {
		return NULL;
	}

	*count = 0;

	@autoreleasepool {
		int cid = SLSMainConnectionID();

		NSSet<NSNumber *> *fullscreenSids = mumuCopyFullscreenSpaceIDs();
		NSSet<NSNumber *> *activeSids = mumuCopyActiveSpaceIDs();
		NSSet<NSNumber *> *onscreenWids = mumuCopyOnscreenWindowIDs();
		NSDictionary<NSNumber *, NSString *> *bundleIDsByPID = mumuCopyRegularAppBundleIDsByPID();

		CFArrayRef windowList = CGWindowListCopyWindowInfo(kCGWindowListOptionAll, kCGNullWindowID);
		if (!windowList) {
			return NULL;
		}

		// Collected as parallel NSMutableArrays (rather than growing a C
		// array) since the final count isn't known up front; converted to
		// the flat MumuLayoutWindowInfo array once enumeration finishes.
		NSMutableArray<NSNumber *> *collectedWids = [NSMutableArray array];
		NSMutableArray<NSString *> *collectedBundleIDs = [NSMutableArray array];
		NSMutableArray<NSString *> *collectedTitles = [NSMutableArray array];
		NSMutableArray<NSNumber *> *collectedSids = [NSMutableArray array];
		NSMutableArray<NSNumber *> *collectedFullscreen = [NSMutableArray array];

		CFIndex total = CFArrayGetCount(windowList);
		for (CFIndex i = 0; i < total; i++) {
			CFDictionaryRef info = (CFDictionaryRef)CFArrayGetValueAtIndex(windowList, i);

			// Layer 0 is the normal document-window layer; everything else
			// (menu bar items, tooltips, the Dock, overlays, ...) is noise
			// for layout purposes.
			CFNumberRef layerRef = (CFNumberRef)CFDictionaryGetValue(info, kCGWindowLayer);
			int layer = 0;
			if (layerRef) {
				CFNumberGetValue(layerRef, kCFNumberIntType, &layer);
			}
			if (layer != 0) {
				continue;
			}

			CFNumberRef pidRef = (CFNumberRef)CFDictionaryGetValue(info, kCGWindowOwnerPID);
			if (!pidRef) {
				continue;
			}
			int pid = 0;
			CFNumberGetValue(pidRef, kCFNumberIntType, &pid);

			NSString *bundleID = bundleIDsByPID[@(pid)];
			if (bundleID == nil) {
				// Not a regular (Dock-visible) application.
				continue;
			}

			CFNumberRef widRef = (CFNumberRef)CFDictionaryGetValue(info, kCGWindowNumber);
			if (!widRef) {
				continue;
			}
			int wid = 0;
			CFNumberGetValue(widRef, kCFNumberIntType, &wid);
			if (wid == 0) {
				continue;
			}

			uint64_t sid = mumuSpaceIDForWindow(cid, (CGWindowID)wid);
			if (sid == 0) {
				// Unresolvable (e.g. a phantom/offscreen helper window that
				// isn't actually attached to any Space) — skip rather than
				// record a bogus Space assignment.
				continue;
			}

			BOOL onActiveSpace = [activeSids containsObject:@(sid)];
			BOOL onscreen = [onscreenWids containsObject:@(wid)];
			if (onActiveSpace && !onscreen) {
				// On a currently displayed Space but absent from the
				// on-screen window list: minimized. See file header for why
				// this check only applies to windows on displayed Spaces.
				continue;
			}

			NSString *title = @"";
			CFStringRef titleRef = (CFStringRef)CFDictionaryGetValue(info, kCGWindowName);
			if (titleRef) {
				char *titleCStr = mumuCopyUTF8String(titleRef);
				title = [NSString stringWithUTF8String:titleCStr] ?: @"";
				free(titleCStr);
			}

			BOOL isFullscreen = [fullscreenSids containsObject:@(sid)];

			[collectedWids addObject:@(wid)];
			[collectedBundleIDs addObject:bundleID];
			[collectedTitles addObject:title];
			[collectedSids addObject:@(sid)];
			[collectedFullscreen addObject:@(isFullscreen)];
		}

		CFRelease(windowList);

		NSUInteger resultCount = collectedWids.count;
		if (resultCount == 0) {
			return NULL;
		}

		MumuLayoutWindowInfo *result = (MumuLayoutWindowInfo *)calloc(resultCount, sizeof(MumuLayoutWindowInfo));
		if (!result) {
			return NULL;
		}

		for (NSUInteger i = 0; i < resultCount; i++) {
			result[i].wid = (uint32_t)collectedWids[i].unsignedIntValue;
			result[i].bundleID = mumuCopyUTF8String((__bridge CFStringRef)collectedBundleIDs[i]);
			result[i].title = mumuCopyUTF8String((__bridge CFStringRef)collectedTitles[i]);
			result[i].sid = collectedSids[i].unsignedLongLongValue;
			result[i].fullscreen = collectedFullscreen[i].boolValue ? 1 : 0;
		}

		*count = (int)resultCount;

		return result;
	}
}

void MumuFreeLayoutWindowInfo(MumuLayoutWindowInfo *info, int count) {
	if (!info) {
		return;
	}

	for (int i = 0; i < count; i++) {
		if (info[i].bundleID) {
			free(info[i].bundleID);
		}

		if (info[i].title) {
			free(info[i].title);
		}
	}

	free(info);
}
