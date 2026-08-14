#pragma once

int MumuCheckAccessibilityPermissions(void);

// Screen Recording permission is required because mumu reads window titles
// via CGWindowListCopyWindowInfo to match windows across Mission Control
// Spaces during restore. kCGWindowName is redacted to NULL system-wide
// without this permission.
int MumuCheckScreenRecordingPermissions(void);
