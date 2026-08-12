#pragma once

int MimiCheckAccessibilityPermissions(void);
int MimiRequestAccessibilityPermissions(void);
int MimiShowAccessibilityPermissionStartupAlert(void);
int MimiShowConfigOnboardingAlert(const char *configPath);

// Screen Recording permission is required only by the layout save/restore
// capability (`mimi layout ...`), which reads window titles via
// CGWindowListCopyWindowInfo to match windows across Mission Control
// Spaces. kCGWindowName is redacted to NULL system-wide without this
// permission, so mimi treats it as required (not optional) for that
// capability specifically — other commands never check it.
int MimiCheckScreenRecordingPermissions(void);
int MimiRequestScreenRecordingPermissions(void);
