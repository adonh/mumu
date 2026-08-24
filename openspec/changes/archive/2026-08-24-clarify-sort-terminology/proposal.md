## Why

mumu overloads the word "display" to mean two unrelated things: a physical monitor (`display-count`, `connected displays`, `--sort display`'s own doc text calls it "left-to-right across all connected **displays**") and, via `--sort display`, the name of the sort key for mumu's own logical left-to-right Space ordinal (the bare `#NN` shown everywhere alongside a Space's macOS number). A user reading `--sort display` reasonably guesses it sorts by monitor, when it actually sorts by mumu's own Space numbering. This is confusing on its own, and compounds the pre-existing `DualLabel` ambiguity (`#NN (space MM)`) fixed in `fix-show-sort-consistency`, since that format's un-labeled `#NN` is exactly what `--sort display` names.

## What Changes

- **BREAKING**: rename the `--sort display` value to `--sort logical` across `mumu show` and `mumu restore`. Behavior is unchanged — it's still the default, still orders by mumu's own left-to-right logical Space ordinal — only the flag value's name changes, so it stops colliding with "display" meaning a physical monitor.
- Rename the corresponding Go identifier `layout.SortByDisplay` to `layout.SortByLogical` (value `"logical"`, was `"display"`).
- Update `--sort`'s help text, error message (`ParseSortKey`'s "must be one of: ..."), and all prose in `docs/CLI.md` that names the sort key, so "display" is reserved exclusively for physical-monitor concepts (`display-count`, `connected displays`, `primary display`) and never for the logical Space ordinal or its sort key.
- No change to `mumu`'s existing `#NN (space MM)` output format itself (already fixed for sort-consistency in `fix-show-sort-consistency`) — `space` there already unambiguously means the macOS Mission Control Space number, and this change doesn't touch it.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `space-layout`: the "Configurable output ordering for layout entries" requirement's `--sort` flag values change from `display|macos|app` to `logical|macos|app` (`logical` remains the default, with identical ordering behavior).

## Impact

- `internal/layout/sort.go`: rename `SortByDisplay` → `SortByLogical` (value `"logical"`); update `ParseSortKey`'s accepted values and error message.
- `internal/layout/sort_test.go`: rename identifiers/expectations accordingly.
- `cmd/mumu/cmd/layout.go`: update the `--sort` flag's default value and help text (both `layoutShowCmd` and `layoutRestoreCmd`).
- `cmd/mumu/cmd/layout_test.go`: rename identifiers/expectations accordingly.
- `docs/CLI.md`: update all `--sort display|macos|app` mentions to `--sort logical|macos|app`, and any prose describing the sort key as "display" (e.g. "display-sequence order", "Default order is display sequence") to use "logical" instead, reserving "display" for physical-monitor concepts only.
- No config.yaml schema change, no change to saved-layout file format, no change to `mumu`'s window-matching or restore behavior.
