## 1. Native layer: per-display Space ordinal resolution

- [x] 1.1 In `internal/native/mumu.h`, replace the "Logical (Left-to-Right) Space Numbering" declarations (`MumuLogicalSpaceCount`, `MumuLogicalSpaceID`, `MumuLogicalIndexForSpace`) with `MumuLogicalDisplayCount(void)`, `MumuOrdinalSpaceID(int displayOrdinal, int spaceOrdinal)`, and `MumuSpaceOrdinal(uint64_t sid, int *outDisplay, int *outSpace)`, documented per design.md - Decision 2; verify the header still compiles as part of `just build`.
- [x] 1.2 In `internal/native/space.m`, implement the three new functions on top of the existing `mumuCopyDisplaySpacesSortedLeftToRight()`, reusing the per-display walk pattern already written for `mumuLocalSpaceIndex` (export or reuse that logic for `MumuSpaceOrdinal` rather than duplicating it); remove the now-unused `MumuLogicalSpaceCount`/`MumuLogicalSpaceID`/`MumuLogicalIndexForSpace` implementations.
- [x] 1.3 Update `internal/native/space.m`'s gesture-navigation code (`MumuFocusSpaceUsingGesture`) to keep using `mumuLocalSpaceIndex` (or the newly exported equivalent) unchanged — verify it still compiles and existing gesture behavior is untouched.
- [x] 1.4 Update or add native-adjacent Go tests (if any exist under `internal/space`) that exercised the old flat functions, and verify `just build` succeeds (this package requires CGO, so a full build is the practical verification step).

## 2. `internal/space`: the `Ordinal` type

- [x] 2.1 Add `internal/space/ordinal.go` defining `type Ordinal struct { Display, Space int }` with `Less(other Ordinal) bool`, `String() string` (`"D:SS"`, space zero-padded to 2 digits, per design.md - Decision 1), and `ParseOrdinal(raw string) (Ordinal, error)` (accepts optional zero-padding on either part, rejects a bare integer with an error naming the `"display:space"` format); add unit tests in `internal/space/ordinal_test.go` covering `Less` ordering (including cross-display comparisons), `String` formatting, and `ParseOrdinal` success/failure cases (bare integer, missing colon, non-positive parts, non-numeric parts).
- [x] 2.2 Rewrite `internal/space/logical.go`'s Go wrappers around the new native functions: `IDForOrdinal(o Ordinal) uint64`, `OrdinalForSpace(sid uint64) Ordinal` (zero value on not-found), `LogicalDisplayCount() int`; keep `LeftToRightSpaceCounts() []int` unchanged. Remove `LogicalCount`, `LogicalSpaceID`, `LogicalIndexForSpace`. Verify with `internal/space/logical_test.go` updates.
- [x] 2.3 Update `internal/space/label.go`'s `DualLabel` to accept an `Ordinal` and print `"#D:SS (space MM)"` per design.md - Decision 1; update its doc comment and any existing test to match the new format.

## 3. `internal/layout`: entries, capture, sort

- [x] 3.1 In `internal/layout/types.go`, change `Entry.Ordinal` from `int` to `space.Ordinal`; add `Ordinal.MarshalJSON`/`UnmarshalJSON` (or field-tag-based nested-object encoding, `{"display":N,"space":N}`) per design.md - Decision 3; bump `SchemaVersion` from `1` to `2`.
- [x] 3.2 In `internal/layout/capture.go`, replace the `space.LogicalIndexForSpace(entry.SpaceID)` call with `space.OrdinalForSpace(entry.SpaceID)`, treating a zero-value `Ordinal` the same way the old zero-ordinal case was skipped.
- [x] 3.3 In `internal/layout/sort.go`, change `entryLess`, `pinRuleLess`, and `defaultSpaceRuleLess` to compare `Ordinal` values via `!=` (equality) and `.Less(...)` instead of raw `int` comparison; change `newMissionControlOrdinalLookup`'s cache to `map[space.Ordinal]int` and its parameter/return types accordingly; update `internal/layout/sort_test.go` for the new type and to verify cross-display ordering is preserved.
- [x] 3.4 In `internal/layout/persist.go`, add the pre-parse schema-version check described in design.md - Decision 3: read `schemaVersion` before the full unmarshal, and return a clear "saved by an incompatible mumu version, run `mumu save` again" error (naming the file path) when it doesn't match the current `SchemaVersion`; leave the existing malformed-JSON error path for same-version files that still fail to parse. Add a unit test in `internal/layout/persist_test.go` covering both the version-mismatch error and the existing malformed-file error remaining distinct.

## 4. `internal/layout`: restore matching, fallback, and bounds checking

- [x] 4.1 In `internal/layout/restore.go`, replace the flat `currentSpaceCount`/`space.LogicalCount()`/`space.LogicalSpaceID` usages in `planDirectMoves` with a per-display bounds check against `space.LeftToRightSpaceCounts()` (display ordinal within `1..len(counts)`, space ordinal within `1..counts[display-1]`) and `space.IDForOrdinal`, per design.md - Decision 5; verify `SkipOrdinalOutOfRange` still fires correctly for both an out-of-range display and an out-of-range space-within-display.
- [x] 4.2 Update `fallbackTarget`, `fallbackTargetForAssignments`, and `planFallbackMoves` to carry `space.Ordinal` instead of `int` for the fallback target's ordinal, including the "most prevalent ordinal" tally (map keyed by `space.Ordinal`) and the tie-break primary-display lookup (`primaryDisplayFallbackTarget`/`space.PrimaryDisplayCurrentSpace`, updated to return an `Ordinal`).
- [x] 4.3 Update `internal/layout/restore_test.go` and `internal/layout/restore_fallback_test.go` for the new `Ordinal` type throughout, adding cases that specifically exercise the per-display bounds check (e.g. an entry valid on display 1 but referencing an out-of-range space on display 2, and vice versa) per the space-layout spec delta's new scenarios.

## 5. `internal/config`: pin and default-space rule ordinals

- [x] 5.1 In `internal/config/config.go`, change `PinRule.Ordinal` and `DefaultSpaceRule.Ordinal` from `int` to `space.Ordinal`; change `pinRuleFileFormat.Ordinal` and `defaultSpaceRuleFileFormat.Ordinal` from `int` to `string`.
- [x] 5.2 Update `validatePins` and `validateDefaultSpaces` to parse each rule's `ordinal` string via `space.ParseOrdinal`, replacing the `rule.Ordinal <= 0` check with parse-failure and non-positive-part handling, producing the config spec delta's updated error messages (naming the field as `ordinal` and stating the expected `"<display>:<space>"` format); add/update `internal/config/config_test.go` cases for: valid `"D:S"`, valid with zero-padding, bare integer (rejected), missing colon, non-positive display or space part, non-numeric parts.
- [x] 5.3 Update `config.go`'s generated default-`config.yaml` comments/examples (the `pins`/`default_spaces` example blocks) to show `ordinal: "2:1"` instead of `ordinal: 1`.

## 6. CLI output

- [x] 6.1 Search `cmd/mumu/cmd` for every place a logical ordinal is printed, parsed from a flag, or otherwise surfaced (e.g. `layout.go`), and update each to use `space.Ordinal`/`DualLabel`'s new signature; verify with `cmd/mumu/cmd/layout_test.go` updates and a manual `mumu show`/`mumu restore --sort logical` smoke check against a saved test layout.
- [x] 6.2 Verify `mumu list`, `mumu show`, and `mumu restore` progress/skip-summary output all display the new `#D:SS (space MM)` format end to end (manual verification is acceptable here if no existing test asserts exact CLI output strings; otherwise update those tests).

## 7. Documentation

- [x] 7.1 Update `docs/CLI.md` wherever the flat `#NN` ordinal format or its examples appear, to the new `#D:SS` two-part format.
- [x] 7.2 Update `docs/CONFIG_SCHEMA.md`'s `pins`/`default_spaces` `ordinal` field documentation and examples to the `"<display>:<space>"` string format.
- [x] 7.3 Update `docs/ARCHITECTURE.md` if it describes the flat logical-ordinal numbering scheme, to describe the two-part scheme instead.
- [x] 7.4 Add a note (CHANGELOG entry or equivalent the project already uses, if any) calling out the breaking change: saved layouts must be recreated with `mumu save`, and `config.yaml` `ordinal` values must be rewritten in `"<display>:<space>"` form. (This project auto-generates `CHANGELOG.md` from conventional-commit messages via release-please; the breaking-change note is carried in the commit message's `feat!:`/`BREAKING CHANGE:` footer when this change is committed, not hand-written into `CHANGELOG.md`.)

## 8. Verification

- [x] 8.1 Run `just fmt`, `just lint`, `just test`, and `just build`, and fix any failures, per AGENTS.md's pre-commit checklist.
- [ ] 8.2 Manually verify the core motivating scenario end to end on a real multi-display setup (or the closest available approximation): save a layout with 2+ displays, add a Space on the leftmost display, run `mumu show`/`mumu restore`, and confirm the second display's entries still resolve to the same Spaces they did before the Space was added. (Blocked: this machine currently has only 1 display connected — needs verification on real multi-display hardware.)
- [x] 8.3 Manually verify that an old-schema saved-layout JSON file (e.g. a copy from before this change) produces the clear "run `mumu save` again" error rather than a raw parse error, and that an old-format bare-integer `ordinal:` in `config.yaml` produces the clear format-mismatch validation error.
