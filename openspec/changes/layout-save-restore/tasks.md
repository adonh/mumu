## 1. Native: logical left-to-right Space numbering

- [x] 1.1 Add a native function that enumerates connected displays sorted by physical left-to-right position (`CGDisplayBounds.origin.x`)
- [x] 1.2 Add a native function that builds the logical-ordinal ↔ Space-ID mapping by concatenating each display's `Spaces` (from the existing `SLSCopyManagedDisplaySpaces` call) in that sorted display order
- [x] 1.3 Add a native function that returns the per-display Space-count sequence (left-to-right) for arrangement-drift comparison at restore time
- [x] 1.4 Expose 1.1–1.3 via Go wrappers (new package or extension of `internal/space`)

## 2. Native: window enumeration across all Spaces

- [x] 2.1 Add a native function listing all windows via `CGWindowListCopyWindowInfo(kCGWindowListOptionAll)`, filtered to `kCGWindowLayer == 0` and owner PIDs belonging to running `NSApplicationActivationPolicyRegular` apps (superseded AX-based enumeration; see design.md Decisions for why)
- [x] 2.2 Fetch each window's title from `kCGWindowName` (requires Screen Recording permission; see design.md) and exclude minimized windows where determinable — a window on a currently displayed Space that's absent from a `kCGWindowListOptionOnScreenOnly` query is minimized; windows on non-displayed Spaces are always included (no reliable signal available for them)
- [x] 2.3 Add a native wrapper around `CGSCopySpacesForWindows` to map each window ID to its Space ID
- [x] 2.4 Detect and exclude fullscreen windows/Spaces (via the Space `type` field from `SLSCopyManagedDisplaySpaces`) from the enumeration
- [x] 2.5 Combine 2.1–2.4 into a single native entry point: "all non-fullscreen windows with their Space ID" (`MimiGetAllWindowsAcrossSpaces` in `internal/native/layout.m`)

> **Redesigned from AX-based to CGWindowList-based enumeration** after empirical testing showed AX only ever exposes windows on each
> display's *currently displayed* Space — confirmed by moving a window to a non-displayed Space and observing it vanish from
> `kAXWindowsAttribute` entirely, even though `CGSCopySpacesForWindows` still correctly reported its Space ID. `CGWindowListCopyWindowInfo`
> filtered to `layer == 0`, regular-app owner PIDs, and a resolved (non-zero) Space ID reliably reaches windows on non-displayed Spaces where
> AX cannot. This required making Screen Recording a required permission (window titles are redacted without it) — see design.md.

## 3. Go: layout capture (save)

- [x] 3.1 New `internal/layout` package: define `Entry` (bundle ID, window title, logical Space ordinal) and `Layout` (display count, per-display Space-count sequence, entries) types
- [x] 3.2 Implement `Capture()`: call the native enumeration (Task 2), resolve each window's Space ID to a logical ordinal (Task 1), using the same deterministic per-app window ordering as `internal/native/window.m`
- [x] 3.3 Implement persistence: write `Layout` as JSON keyed by display count under mimi's existing user-data directory convention, overwriting any existing file for that display count
- [x] 3.4 Report a summary at the end of save (windows captured, fullscreen windows skipped)

## 4. Go: restore

- [x] 4.1 Implement `Load(displayCount)` returning a clear "no saved layout" error when none exists for the current display count
- [x] 4.2 Implement arrangement-drift detection: compare the current per-display Space-count sequence against the one recorded in the saved layout
- [x] 4.3 Implement an interactive confirmation prompt shown only when drift is detected; abort cleanly with no changes on decline (implemented in `cmd/mimi/cmd/layout.go`'s `restore` command, with a `--yes` flag to skip)
- [x] 4.4 Implement per-app discovery of currently running apps' live windows (bundle ID → titles, same deterministic ordering as save)
- [x] 4.5 Implement matching: exact title match first, positional-index fallback, per app
- [x] 4.6 Implement per-window move: resolve the saved logical ordinal to a live Space ID (Task 1) and call mimi's existing move-to-Space primitive; skip and record if the target ordinal doesn't currently exist
- [x] 4.7 Aggregate skip reasons (app not running, unmatched window, ordinal out of range) into a restore summary reported to the user

## 5. CLI

- [x] 5.1 Add the `mimi layout` command group (`save`, `restore`, `list`, `show`, `delete`) in `cmd/mimi`
- [x] 5.2 Implement `mimi layout save`
- [x] 5.3 Implement `mimi layout restore` (auto-detects display count, shows confirmation prompt on drift, prints restore summary)
- [x] 5.4 Implement `mimi layout list`
- [x] 5.5 Implement `mimi layout show [display-count]`
- [x] 5.6 Implement `mimi layout delete [display-count]`

## 6. Documentation

- [x] 6.1 Add a "Layout" section to `docs/CLI.md` documenting all `mimi layout` subcommands
- [x] 6.2 Document the reflow-only, no-fullscreen, no-geometry limitations clearly (per `design.md` Risks / Trade-offs)
- [x] 6.3 Update `README.md` to surface layout save/restore alongside mimi's other window/space actions

## 7. Testing

- [x] 7.1 Regression test for the logical-ordinal ↔ Space-ID mapping (`internal/space/logical_test.go`) — asserts round-trip and count invariants against whatever arrangement is live on the test machine; skips on headless CI. **Deviation from original wording**: a *synthetic* multi-display arrangement (incl. non-leftmost-primary) isn't feasible without adding a mocking seam to native SkyLight calls, which was judged out of scope for this change — this was instead validated manually earlier in this change's development against a real non-leftmost-primary 4-display setup (see design.md Decisions)
- [x] 7.2 Unit tests for the title-match/index-fallback matching logic (`internal/layout/restore_test.go`)
- [x] 7.3 Unit tests for arrangement-drift detection (`internal/layout/restore_test.go`)
- [x] 7.4 Manual test plan covering real multi-display save/restore (see design.md "Manual Test Plan"; native SkyLight/AX/CGWindowList behavior cannot be fully unit tested) — executed against a real 4-display, 25-Space session with 32 real windows during this change's implementation
