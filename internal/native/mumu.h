#ifndef ACCESSIBILITY_H
#define ACCESSIBILITY_H

#import <ApplicationServices/ApplicationServices.h>
#import <Foundation/Foundation.h>

#pragma mark - Element Functions

void *MumuGetFocusedApplication(void);
void MumuReleaseElement(void *element);
int MumuAreElementsEqual(void *element1, void *element2);

#pragma mark - Window Functions

void **MumuGetAllFocusableWindowsOnActiveSpaceWithFocused(int *count, int *focusedIndex);
void *MumuGetFrontmostWindow(void);
int MumuActivateWindow(void *window);

#pragma mark - Screen Functions

bool MumuIsMissionControlActive(void);
double *MumuGetScreenFrameForPoint(double x, double y);
double *MumuGetScreenVisibleFrameForPoint(double x, double y);

#pragma mark - Window Frame Functions

double *MumuGetWindowFrame(void *window);
int MumuSetWindowFrame(void *window, double x, double y, double w, double h);

#pragma mark - Tiling Margins

bool MumuTiledWindowMarginsEnabled(void);
double MumuTiledWindowMarginSize(void);

#pragma mark - Space Functions

int MumuCountMissionControlSpaces(void);
uint64_t MumuMissionControlSpaceID(int index);

/// The 1-based Mission Control index for a given macOS Space ID (the same
/// numbering MumuMissionControlSpaceID uses, and the same ordinal macOS's
/// own "Switch to Desktop <n>" keyboard shortcut uses), or 0 if not found.
int MumuMissionControlIndexForSpace(uint64_t sid);

uint32_t MumuSpaceDisplayID(uint64_t sid);
uint64_t MumuActiveSpaceID(void);
int MumuFocusSpaceUsingGesture(uint32_t new_did, uint64_t new_sid);
int MumuMoveWindowToSpace(void *windowElement, uint64_t spaceID);

/// Moves the window with the given CGWindowID to the Mission Control space
/// with the given macOS space ID, without altering its position or size.
/// Unlike MumuMoveWindowToSpace, this does not require an AXUIElementRef,
/// which makes it usable with windows discovered via
/// MumuGetAllWindowsAcrossSpaces (CGWindowList-based, not AX-based).
int MumuMoveWindowIDToSpace(uint32_t windowID, uint64_t spaceID);

uint32_t MumuCursorDisplayID(void);
void MumuActivateDisplay(uint32_t did);

#pragma mark - Logical (Left-to-Right) Space Numbering
//
// A numbering scheme scoped to the layout save/restore capability only: all
// connected displays are sorted by physical left-to-right position
// (CGDisplayBounds.origin.x) and each display's own Spaces are concatenated
// in that order, regardless of which display is the primary (menu-bar)
// display. This differs from MumuMissionControlSpaceID's ordering (which is
// primary-display-first, per SLSCopyManagedDisplaySpaces) and must not
// alter it.

/// Total number of Spaces, counted in logical left-to-right order. Numerically
/// equal to MumuCountMissionControlSpaces(); only the ordering differs.
int MumuLogicalSpaceCount(void);

/// The macOS Space ID at the given 1-based logical left-to-right index, or 0
/// if out of range.
uint64_t MumuLogicalSpaceID(int logicalIndex);

/// The 1-based logical left-to-right index for a given macOS Space ID, or 0
/// if not found.
int MumuLogicalIndexForSpace(uint64_t sid);

/// Per-display Space-count sequence in left-to-right order, used to detect
/// display-arrangement drift between save and restore. *outCount is set to
/// the number of displays. Caller must free() the returned array.
int *MumuLeftToRightSpaceCounts(int *outCount);

#pragma mark - Layout: Window Enumeration Across All Spaces

//
// Unlike MumuGetAllFocusableWindowsOnActiveSpaceWithFocused (AX-based), this
// enumerates via CGWindowListCopyWindowInfo(kCGWindowListOptionAll, ...) because the AX
// APIs only ever expose windows belonging to the currently *displayed*
// Space on each connected display; windows sitting on any other Space are
// invisible to AX entirely, no matter how they're queried. CGWindowList
// covers every Space, at the cost of window titles requiring Screen
// Recording permission (kCGWindowName is otherwise redacted to NULL) and
// minimized-state only being reliably knowable for windows on a currently
// displayed Space (see .m for details).

/// One window discovered by MumuGetAllWindowsAcrossSpaces.
typedef struct {
	uint32_t wid;    // CGWindowID. Pass directly to
	                 // MumuMoveWindowIDToSpace; no AXUIElementRef needed.
	char *bundleID;  // malloc'd UTF-8 C string, never NULL (may be empty).
	char *title;     // malloc'd UTF-8 C string, never NULL (may be
	                 // empty). Populated from kCGWindowName, which is
	                 // redacted to empty without Screen Recording
	                 // permission.
	uint64_t sid;    // Resolved macOS Space ID, or 0 if unresolved.
	int fullscreen;  // 1 if the window's Space is a fullscreen Space.
} MumuLayoutWindowInfo;

/// Enumerates windows across ALL Spaces (not just the displayed ones) for
/// every running regular application, resolving each window's owning Space
/// via the private CGSCopySpacesForWindows call. Excludes minimized windows
/// when minimized state can be determined (windows on a currently displayed
/// Space); windows on other Spaces are included regardless of minimized
/// state, since that can't be reliably determined for them.
/// *count is set to the number of entries returned, or 0 on failure.
/// Caller must free the result with MumuFreeLayoutWindowInfo.
MumuLayoutWindowInfo *MumuGetAllWindowsAcrossSpaces(int *count);

/// Frees an array returned by MumuGetAllWindowsAcrossSpaces, including each
/// entry's strings.
void MumuFreeLayoutWindowInfo(MumuLayoutWindowInfo *info, int count);

#endif  // ACCESSIBILITY_H
