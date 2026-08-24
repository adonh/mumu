## Why

mumu's logical Space ordinal is a single flat number: every Space across every connected display is numbered sequentially left to right (Display A's Spaces, then Display B's, then Display C's). This numbering isn't stable under the exact kind of change users make routinely: adding or removing a Mission Control Space on any display other than the rightmost one shifts the flat ordinal of every Space on every display to its right. Any saved layout, pin rule, or `default_spaces` rule that referenced one of those now-shifted ordinals silently targets the wrong Space after the next restore — with no warning, since the ordinal is still a valid, in-range integer. [Issue #26](https://github.com/adonh/mumu/issues/26) tracks this.

Numbering ordinals two-dimensionally — `<display ordinal>:<space ordinal>`, e.g. `2:01` for the 1st Space on the 2nd display from the left — scopes each Space's number to its own display. Adding or removing a Space on one display no longer renumbers Spaces on any other display, so existing pins, `default_spaces` rules, and saved layouts referencing other displays keep working.

## What Changes

- **BREAKING**: mumu's logical Space ordinal changes from a flat integer (e.g. `4`) to a two-part ordinal, display-then-space (e.g. `2:01`), scoped per display instead of per whole arrangement.
- **BREAKING**: saved-layout JSON files' `ordinal` field changes shape to carry both parts; existing saved layouts (any `<data_dir>/layouts/<display-count>.json` written by a prior mumu version) can no longer be read as-is and must be recreated with `mumu save`. The system SHALL detect this outdated shape and report a clear "run `mumu save` again" message rather than a raw parse error.
- **BREAKING**: `config.yaml`'s `ordinal` field (under `pins` and `default_spaces`) changes from a bare positive integer to a `"<display>:<space>"` string (e.g. `"2:1"`). The old bare-integer form is rejected with a clear validation error naming the expected format, since a flat ordinal can't be reinterpreted as a display+space pair without knowing the arrangement it was captured against.
- Every place mumu currently prints a logical ordinal (`mumu show`, `mumu restore` progress and skip summaries) switches from `#NN` to `#D:SS` (e.g. `#2:01`), still paired with the existing macOS Mission Control Space number.
- `--sort logical`/`--sort macos`/`--sort app` ordering and their tie-break cascades continue to work, now comparing the two-part ordinal (display first, then space within display) wherever they previously compared the flat integer — producing the same left-to-right ordering as before for any single arrangement.
- Out-of-range detection (`SkipOrdinalOutOfRange`, "saved space no longer exists") becomes a per-display bounds check (display ordinal within the connected display count, space ordinal within that specific display's current Space count) instead of a single whole-arrangement bound.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `space-layout`: the "Logical left-to-right Space numbering" requirement changes from a single flat ordinal to a two-part (display, space-within-display) ordinal; every downstream requirement referencing "logical Space ordinal" (save, restore matching/fallback, sort ordering, dual-number display, saved-layout persistence, out-of-range handling) is updated to use the two-part form.
- `configuration`: the `pins` and `default_spaces` settings' `ordinal` field changes from a bare positive integer to a `"<display>:<space>"` string, including its validation error messages.
- `window-pinning`: pin rules' target `ordinal` changes from a bare positive integer to the two-part ordinal, matching `space-layout`'s and `configuration`'s new form; pin-target out-of-range detection becomes a per-display bounds check.

## Impact

- `internal/native/mumu.h` / `internal/native/space.m`: the flat "logical left-to-right" native API (`MumuLogicalSpaceCount`, `MumuLogicalSpaceID`, `MumuLogicalIndexForSpace`) is replaced with per-display-aware equivalents that resolve/accept a (display ordinal, space-within-display ordinal) pair, built on the existing left-to-right display sort and per-display `Spaces` arrays already used by `MumuLeftToRightSpaceCounts`.
- `internal/space`: `logical.go` and `label.go` — new `Ordinal` type (display + space-within-display, comparable, sortable, with `"D:SS"` string form) replacing the flat `int` ordinal; `LogicalSpaceID`/`LogicalIndexForSpace`/`LogicalCount` replaced or rebuilt around `Ordinal`; `DualLabel` updated to print `#D:SS (space MM)`.
- `internal/layout`: `types.go` (`Entry.Ordinal`), `capture.go`, `restore.go` (matching, fallback targeting, out-of-range bounds checks), `sort.go` (`entryLess`, `pinRuleLess`, `defaultSpaceRuleLess`, the macOS-ordinal memoization cache) — all switch from flat-int comparisons/lookups to the new `Ordinal` type. `SchemaVersion` bumps; a saved layout whose stored shape doesn't match the current schema is reported clearly instead of a raw JSON error.
- `internal/config`: `PinRule.Ordinal` / `DefaultSpaceRule.Ordinal` and their YAML file-format structs and validation switch from `int` to a parsed `"D:S"` string, with updated error messages and `config.yaml`'s generated comments/examples.
- CLI output across `cmd/mumu/cmd` wherever a logical ordinal is printed or accepted.
- Docs: `docs/CLI.md`, `docs/CONFIG_SCHEMA.md`, `docs/ARCHITECTURE.md` wherever the flat ordinal or its examples are described.
- No migration path for existing saved layouts or `config.yaml` `ordinal` values — both are explicitly reported as needing to be recreated/edited by hand (saved layouts are internal state the user re-generates with `mumu save`; `config.yaml`'s `ordinal` values are few enough per the pinning/default-spaces use case to hand-edit, consistent with `config.yaml` never being silently rewritten).
