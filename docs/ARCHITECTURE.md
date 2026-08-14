# Architecture

`mumu` is a one-shot CLI: every command runs, does its work, and exits. There is no daemon, no background process, and no persistent state beyond the saved layout files themselves.

```
mumu <command>
  → internal/layout (capture, persist, restore)
  → internal/window / internal/space
  → internal/native (CGWindowList + private SkyLight APIs, Objective-C + CGO)
```

All paths use native macOS APIs via CGO. No SIP disable is required.

---

## Layout Save/Restore

Layout capture enumerates windows via `CGWindowListCopyWindowInfo(kCGWindowListOptionAll)` rather than the Accessibility API: AX only exposes windows on each display's _currently displayed_ Space, so windows on other Spaces are AX-invisible. The CGWindowList result is filtered to document-layer windows (`kCGWindowLayer == 0`) owned by regular (non-background) apps, then resolved to a Space ID via the private `CGSCopySpacesForWindows` API. This is also why layout commands require Screen Recording in addition to Accessibility — window titles from `kCGWindowName` are redacted without it.

Space numbers are "logical left-to-right": each display's Spaces (already primary-first via `SLSCopyManagedDisplaySpaces`) are concatenated in physical left-to-right display order (`CGDisplayBounds.origin.x`), independent of which display macOS considers primary. This differs from the macOS Mission Control ordering, which always lists the primary display's Spaces first; `mumu` resolves and displays both numbers side by side (see [CLI Guide — Space Numbering](CLI.md#space-numbering)).

Layouts are persisted as JSON under `~/.local/share/mumu/layouts/<display-count>.json`, keyed by the number of currently connected displays. Restore only moves windows belonging to already-running applications to already-existing Spaces — it never launches apps, and never creates or removes Spaces.

Space-to-Space window moves use the private SkyLight API (`SLSMoveWindowsToManagedSpace` and friends) for instant, animation-free relocation, without requiring an `AXUIElementRef` (layout-discovered windows come from CGWindowList, identified by `CGWindowID`, not AX).

---

## Package Layout

```
cmd/mumu/           CLI entry point and commands
internal/
  layout/           Layout save/restore: capture, JSON persistence, restore matching
  window/           Go wrappers for AX window APIs and CGWindowList-based window moves
  space/            Mission Control space operations, logical left-to-right numbering
  native/           Objective-C + CGO bridge (window/space APIs, layout enumeration)
  permissions/      Accessibility and Screen Recording permission checks
  errors/           Structured error types
  paths/            Home-directory path expansion
```

---

## Permissions

`mumu save` and `mumu restore` check for both **Accessibility** and **Screen Recording** before doing any work, since both are required to enumerate windows and read their titles reliably. `mumu show`, `list`, and `delete` only read or write the local JSON layout file and don't perform a permission check. `mumu status` reports both permissions without touching any window, Space, or layout file.

---

## Platform Notes

Space switching and window-to-Space moves use undocumented private APIs that may break on macOS updates. They are provided as-is for personal automation workflows, not as guaranteed-stable APIs.
