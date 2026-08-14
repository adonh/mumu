## Why

`mimi layout show`'s window list and `mimi layout restore`'s per-window move progress and skip summary currently order entries inconsistently: `show` sorts by logical (display-sequence) Space ordinal, while `restore` simply follows the saved file's original capture-time order — effectively arbitrary from a user's perspective, since capture order reflects whatever order the native window enumeration happened to return windows in, not anything meaningful. Users reviewing or troubleshooting a restore have no reliable way to predict or follow the order entries will print in, and have no way to ask for an order that better matches what they're trying to check (e.g. all of one app's windows together, or matching macOS's own Space numbering).

## What Changes

- Add a `--sort <display|macos|app>` flag to both `mimi layout show` and `mimi layout restore`, controlling the order entries are printed in:
  - `display` (default): logical left-to-right Space ordinal, ascending — mimi's existing display-sequence numbering.
  - `macos`: macOS Mission Control Space ordinal, ascending (the same number `mimi action space <n>` uses).
  - `app`: bundle identifier, ascending (alphabetical); mimi has no captured or resolvable human-readable app display name, so this sorts by the raw bundle ID string (e.g. `com.apple.Safari`), consistent with how mimi already identifies apps everywhere else in its output.
- **BREAKING** (minor, cosmetic): `mimi layout restore`'s default per-window move order changes from the saved file's original capture-time order to the same `display`-ordinal sort `show` already uses by default, fixing the "arbitrary" default rather than only offering it as an opt-in. This does not change which windows get matched or moved, or which Space each ends up on — only the order status lines print in.
- Whichever key is chosen (primary key), ties are broken by cascading through the other two keys in a fixed priority — Space ordinal, then bundle ID, then window title — so output ordering is always fully deterministic regardless of which sort is primary.
- Applies uniformly to: `mimi layout show`'s per-entry listing, `mimi layout restore`'s per-window "moving" progress lines, and `mimi layout restore`'s skip summary (ordering entries within each skip-reason group; the grouping by reason itself is unchanged).
- No change to what's persisted in a saved layout file, and no change to restore's matching or move logic — this is purely a display-ordering concern.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `space-layout`: adds a requirement that `mimi layout show` and `mimi layout restore` support a `--sort` flag controlling per-entry output order, and changes `mimi layout restore`'s default output order to the display-sequence ordinal (matching `show`'s existing default) instead of the saved file's original capture-time order.

## Impact

- `cmd/mimi/cmd/layout.go`: new `--sort` flag definition (shared between `layoutShowCmd` and `layoutRestoreCmd`), plus a shared sort function used by both commands' entry lists.
- `internal/layout/restore.go`: sort `toMove` before the move loop, using the same shared sort logic, before emitting per-window progress lines; sort skipped entries within each skip-reason group before printing the summary.
- No changes to the saved layout JSON schema, to `Capture`/`Restore`'s matching or move behavior, or to `mimi action space`/`move_window_to_space`.
- `docs/CLI.md`: document the new `--sort` flag and note the restore default-order change.
