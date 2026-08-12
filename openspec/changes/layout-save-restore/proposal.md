## Why

mimi can move a window to a Space, but it has no memory of *where things go*. Every time a user reconnects external monitors (or works from a different physical setup), they re-sort every app's windows back onto the right Spaces by hand. This change adds the ability to save the current window-to-Space arrangement and restore it later — without requiring admin/root privileges. This capability additionally requires Screen Recording permission (alongside mimi's existing Accessibility requirement), scoped to the `mimi layout` command group only — see `design.md` for why.

## What Changes

- New `mimi layout` CLI command group: `save`, `restore`, `list`, `delete`, `show`.
- **Save**: for the current display setup, record which Space each window (identified by bundle ID + window title) is assigned to, then persist it keyed by the number of connected displays. Saving overwrites any existing layout for that display count.
- **Restore**: auto-detects the current display count, looks up the layout saved for it, and moves each already-running app's matching window to its recorded Space. Apps that aren't running are skipped and reported, not launched. If the saved arrangement needs more Space slots than currently exist, place what fits and warn about the rest. If the finer-grained per-display arrangement doesn't line up with what was saved (e.g. displays reordered), warn and ask the user whether to proceed anyway or abort.
- New **logical, left-to-right Space numbering**, scoped to this feature only: Spaces are ordinally numbered by sorting connected displays by physical position (left to right) and concatenating each display's own Spaces in that order — independent of primary-display status. This matches how a user visually perceives "the Nth Space, counting left to right," rather than macOS's internal (primary-first) numbering. This numbering is **not** used by and does not change the existing `mimi action space <n>` / `move_window_to_space` commands.
- Window position and size are **not** captured or restored — only Space assignment. Fullscreen windows/Spaces are ignored entirely (not captured, not restored).
- Manual CLI trigger only — no daemon/hook integration, no auto-save, no auto-restore-at-login in this change.

## Capabilities

### New Capabilities

- `space-layout`: Save and restore the assignment of windows to Mission Control Spaces, keyed by display configuration, using a left-to-right logical Space numbering scheme.

### Modified Capabilities

_None — existing `mimi action space` / `move_window_to_space` behavior and numbering are unchanged._

## Impact

- New Go package(s) under `internal/` (e.g. `internal/layout`) for capture, matching, and persistence logic.
- New native (Objective-C/CGO) functions in `internal/native` to: (a) resolve the logical left-to-right ordinal ↔ actual Space ID mapping from physical display positions, and (b) enumerate windows across *all* Spaces (not just the active one), most likely via the private `CGSCopySpacesForWindows` call layered on the existing SkyLight connection mimi already opens — same private-API risk profile already documented in `docs/ARCHITECTURE.md`.
- Reuses existing primitives: `move_window_to_space`'s underlying native call, window/app enumeration helpers, `AXUIElementGetPid`/title lookups.
- New CLI command group in `cmd/mimi` and new `docs/CLI.md` section.
- New persisted layout files under mimi's existing user-data directory convention (no admin/root required, consistent with current permission model).
