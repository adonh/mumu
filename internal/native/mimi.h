#ifndef ACCESSIBILITY_H
#define ACCESSIBILITY_H

#import <ApplicationServices/ApplicationServices.h>
#import <Foundation/Foundation.h>

#pragma mark - Element Functions

void *MimiGetFocusedApplication(void);
void MimiReleaseElement(void *element);
void MimiRetainElement(void *element);
int MimiAreElementsEqual(void *element1, void *element2);

#pragma mark - Window Functions

void **MimiGetAllFocusableWindowsOnActiveSpace(int *count);
void **MimiGetAllFocusableWindowsOnActiveSpaceWithFocused(int *count, int *focusedIndex);
void *MimiGetFrontmostWindow(void);
int MimiActivateWindow(void *window);

#pragma mark - Screen Functions

bool MimiIsMissionControlActive(void);
double *MimiGetScreenFrameForPoint(double x, double y);
double *MimiGetScreenVisibleFrameForPoint(double x, double y);

#pragma mark - Window Frame Functions

double *MimiGetWindowFrame(void *window);
int MimiSetWindowFrame(void *window, double x, double y, double w, double h);

#pragma mark - Tiling Margins

bool MimiTiledWindowMarginsEnabled(void);
double MimiTiledWindowMarginSize(void);

#pragma mark - Space Functions

int MimiCountMissionControlSpaces(void);
uint64_t MimiMissionControlSpaceID(int index);
uint32_t MimiSpaceDisplayID(uint64_t sid);
uint64_t MimiActiveSpaceID(void);
int MimiFocusSpaceUsingGesture(uint32_t new_did, uint64_t new_sid);
int MimiMoveWindowToSpace(void *windowElement, uint64_t spaceID);

/// Moves the window with the given CGWindowID to the Mission Control space
/// with the given macOS space ID, without altering its position or size.
/// Unlike MimiMoveWindowToSpace, this does not require an AXUIElementRef,
/// which makes it usable with windows discovered via
/// MimiGetAllWindowsAcrossSpaces (CGWindowList-based, not AX-based).
int MimiMoveWindowIDToSpace(uint32_t windowID, uint64_t spaceID);

uint32_t MimiCursorDisplayID(void);
void MimiActivateDisplay(uint32_t did);

#pragma mark - Logical (Left-to-Right) Space Numbering
//
// A numbering scheme scoped to the layout save/restore capability only: all
// connected displays are sorted by physical left-to-right position
// (CGDisplayBounds.origin.x) and each display's own Spaces are concatenated
// in that order, regardless of which display is the primary (menu-bar)
// display. This differs from MimiMissionControlSpaceID's ordering (which is
// primary-display-first, per SLSCopyManagedDisplaySpaces) and must not be
// used by the existing `mimi action space` / `move_window_to_space` commands.

/// Total number of Spaces, counted in logical left-to-right order. Numerically
/// equal to MimiCountMissionControlSpaces(); only the ordering differs.
int MimiLogicalSpaceCount(void);

/// The macOS Space ID at the given 1-based logical left-to-right index, or 0
/// if out of range.
uint64_t MimiLogicalSpaceID(int logicalIndex);

/// The 1-based logical left-to-right index for a given macOS Space ID, or 0
/// if not found.
int MimiLogicalIndexForSpace(uint64_t sid);

/// Per-display Space-count sequence in left-to-right order, used to detect
/// display-arrangement drift between save and restore. *outCount is set to
/// the number of displays. Caller must free() the returned array.
int *MimiLeftToRightSpaceCounts(int *outCount);

#pragma mark - Layout: Window Enumeration Across All Spaces

//
// Unlike MimiGetAllFocusableWindowsOnActiveSpace (AX-based), this enumerates
// via CGWindowListCopyWindowInfo(kCGWindowListOptionAll, ...) because the AX
// APIs only ever expose windows belonging to the currently *displayed*
// Space on each connected display; windows sitting on any other Space are
// invisible to AX entirely, no matter how they're queried. CGWindowList
// covers every Space, at the cost of window titles requiring Screen
// Recording permission (kCGWindowName is otherwise redacted to NULL) and
// minimized-state only being reliably knowable for windows on a currently
// displayed Space (see .m for details).

/// One window discovered by MimiGetAllWindowsAcrossSpaces.
typedef struct {
	uint32_t wid;    // CGWindowID. Pass directly to
	                 // MimiMoveWindowIDToSpace; no AXUIElementRef needed.
	char *bundleID;  // malloc'd UTF-8 C string, never NULL (may be empty).
	char *title;     // malloc'd UTF-8 C string, never NULL (may be
	                 // empty). Populated from kCGWindowName, which is
	                 // redacted to empty without Screen Recording
	                 // permission.
	uint64_t sid;    // Resolved macOS Space ID, or 0 if unresolved.
	int fullscreen;  // 1 if the window's Space is a fullscreen Space.
} MimiLayoutWindowInfo;

/// Enumerates windows across ALL Spaces (not just the displayed ones) for
/// every running regular application, resolving each window's owning Space
/// via the private CGSCopySpacesForWindows call. Excludes minimized windows
/// when minimized state can be determined (windows on a currently displayed
/// Space); windows on other Spaces are included regardless of minimized
/// state, since that can't be reliably determined for them.
/// *count is set to the number of entries returned, or 0 on failure.
/// Caller must free the result with MimiFreeLayoutWindowInfo.
MimiLayoutWindowInfo *MimiGetAllWindowsAcrossSpaces(int *count);

/// Frees an array returned by MimiGetAllWindowsAcrossSpaces, including each
/// entry's strings.
void MimiFreeLayoutWindowInfo(MimiLayoutWindowInfo *info, int count);

#endif  // ACCESSIBILITY_H
