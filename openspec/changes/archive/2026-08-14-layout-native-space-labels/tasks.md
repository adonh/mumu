## 1. Native: reverse lookup for the Mission Control ordinal

- [x] 1.1 Add `MimiMissionControlIndexForSpace(uint64_t sid)` to `internal/native/space.m`, scanning the raw (unsorted) `SLSCopyManagedDisplaySpaces()` order — mirrors `MimiLogicalIndexForSpace`'s structure but without the left-to-right sort. Returns 0 if not found.
- [x] 1.2 Declare it in `internal/native/mimi.h` alongside the existing `MimiMissionControlSpaceID`/`MimiLogicalIndexForSpace` declarations.

## 2. Go: expose the reverse lookup

- [x] 2.1 Add `MissionControlIndexForSpace(sid uint64) int` to `internal/space` (likely `space.go`, next to `Focus`/`Count`/`ActiveIndex`), wrapping the new native function.

## 3. CLI: display both numbers

- [x] 3.1 Add a small shared formatting helper (e.g. in `cmd/mimi/cmd/layout.go`) that takes a logical ordinal and renders `space <logical> (macOS space <mission-control>)`, falling back to `(macOS space unknown)` when the reverse lookup returns 0.
- [x] 3.2 Use it in `mimi layout show`'s per-entry output.
- [x] 3.3 Use it in `internal/layout/restore.go`'s per-window move progress message (or pass the resolved label through if formatting must stay in the CLI layer — keep native/Go-level `internal/layout` package free of presentation-only concerns where reasonably possible, consistent with existing style).
- [x] 3.4 Use it in `printRestoreSummary`'s per-skipped-entry listing in `cmd/mimi/cmd/layout.go`.

## 4. Documentation

- [x] 4.1 Update `docs/CLI.md`'s `mimi layout show`/`restore` descriptions and any example output to mention the dual Space numbering.

## 5. Testing

- [x] 5.1 Add a test alongside `internal/space/logical_test.go` (or extend it) asserting `MissionControlIndexForSpace` round-trips against `space.Focus`'s existing index range on the live test machine, mirroring the existing self-consistency check for the logical mapping.
- [x] 5.2 Manually verify `mimi layout show` and `mimi layout restore` output on a multi-display setup where the primary display isn't leftmost, confirming the two numbers differ as expected and match `mimi action space <n>`'s own numbering.
