## 1. `internal/layout`: sort key and comparator

- [x] 1.1 Add `internal/layout/sort.go` with `type SortKey string`, constants `SortByDisplay = "display"`, `SortByMacOS = "macos"`, `SortByApp = "app"`, and `ParseSortKey(raw string) (SortKey, error)` returning a clear error for any other value.
- [x] 1.2 In the same file, add an unexported `entryLess(a, b Entry, key SortKey, mcOrdinal func(int) int) bool` implementing: primary check on `key` (`SortByMacOS` compares `mcOrdinal(a.Ordinal)` vs `mcOrdinal(b.Ordinal)`; `SortByApp` compares `BundleID`; `SortByDisplay` falls straight into the cascade), then falling through to the fixed cascade `Ordinal` → `BundleID` → `Title` whenever the primary comparison ties.
- [x] 1.3 Add exported `SortEntries(entries []Entry, key SortKey)`, building its own memoized `mcOrdinal` closure (a `map[int]int` keyed by logical `Ordinal`, populated via `space.LogicalSpaceID` + `space.MissionControlIndexForSpace` on first use per distinct ordinal) and sorting with `sort.SliceStable` + `entryLess`.

## 2. `internal/layout/restore.go`: apply sort to move order

- [x] 2.1 Change `Restore`'s signature to `Restore(saved *Layout, sortKey SortKey, progress ProgressFunc) (RestoreSummary, error)`, updating its doc comment.
- [x] 2.2 After building `toMove` and before the "Moving N window(s)..." progress line, sort `toMove` in place using `entryLess` (own memoized `mcOrdinal` closure) compared on each target's `.entry`, so both the progress lines and the actual move order follow the selected key.

## 3. CLI: `--sort` flag on `show` and `restore`

- [x] 3.1 In `cmd/mimi/cmd/layout.go`, add `addSortFlag(cmd *cobra.Command, dest *string)` registering a `--sort` string flag (default `"display"`) with shared help text listing the three values.
- [x] 3.2 Add package vars `layoutShowSort` and `layoutRestoreSort`; call `addSortFlag` for each command in `init()`.
- [x] 3.3 In `layoutShowCmd`'s `RunE`, parse `layoutShowSort` via `layout.ParseSortKey`, returning a `derrors.CodeInvalidInput` error on failure; replace the existing inline `sort.Slice(entries, ...)` (by `Ordinal`/`BundleID`) with `layout.SortEntries(entries, sortKey)`.
- [x] 3.4 In `layoutRestoreCmd`'s `RunE`, parse `layoutRestoreSort` the same way and pass the resulting `SortKey` through to `layout.Restore(saved, sortKey, ...)`.
- [x] 3.5 Thread the parsed `SortKey` into `printRestoreSummary` (new parameter), and call `layout.SortEntries` on each skip-reason group's entries before printing them.

## 4. Documentation

- [x] 4.1 Update `docs/CLI.md`'s `mimi layout show`/`restore` sections: document the new `--sort` flag and its three values, and note that restore's default per-window order is now the same display-sequence order `show` already defaults to (previously unspecified/capture-order).

## 5. Testing

- [x] 5.1 Add `internal/layout/sort_test.go` unit-testing `entryLess` directly with a stub `mcOrdinal` function (no native calls) covering: each of the three keys as primary, and the `Ordinal` → `BundleID` → `Title` tiebreak cascade when the primary key is equal.
- [x] 5.2 In the same file, unit-test `ParseSortKey` for all three valid values and at least one invalid value.
- [x] 5.3 Add a live self-consistency test (skipped on headless, mirroring `internal/space/logical_test.go`'s pattern) asserting `SortEntries` with `SortByMacOS` on a small set of synthetic entries produces a total order consistent with each entry's live-resolved Mission Control ordinal.
- [x] 5.4 Manually verify `mimi layout show --sort macos` and `--sort app` (and the bare default) on the multi-display setup already available, confirming: `app` groups each bundle ID's windows together alphabetically, `macos` matches the Mission Control numbers already shown in parentheses, and the bare default matches today's `show` output ordering unchanged.
