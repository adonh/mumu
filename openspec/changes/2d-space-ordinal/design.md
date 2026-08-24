## Context

See `proposal.md` - Why/What Changes for motivation. Today's flat logical ordinal is computed in `internal/native/space.m` (`MumuLogicalSpaceCount`/`MumuLogicalSpaceID`/`MumuLogicalIndexForSpace`, all built on `mumuCopyDisplaySpacesSortedLeftToRight`) and exposed to Go as `int` via `internal/space/logical.go`. That flat `int` then flows, unchanged in type, through `internal/layout` (`Entry.Ordinal`, saved-layout JSON, matching/fallback/sort logic) and `internal/config` (`PinRule.Ordinal`, `DefaultSpaceRule.Ordinal`, YAML `ordinal:` key). `internal/native/space.m` already has a private per-display-relative helper, `mumuLocalSpaceIndex(sid, did)`, used today only by the Dock-swipe gesture code — the per-display indexing this change needs already exists natively in that one place, it's just not exposed as a public ordinal.

`Layout.SpaceCounts []int` (per-display Space-count sequence, left to right) already exists for arrangement-drift detection, so the per-display structure this change needs is already present at the `Layout` level; it just isn't reflected in the per-`Entry` ordinal yet.

## Goals / Non-Goals

**Goals:**
- Replace the flat logical ordinal with a two-part `(display, space-within-display)` ordinal everywhere it's computed, stored, compared, or displayed.
- Keep the existing left-to-right, primary-independent display sort as the source of the "display" part — this change only adds a dimension, it doesn't change which display is "first."
- Preserve every existing observable ordering/tie-break behavior (`--sort logical/macos/app`, the fixed tie-break cascade) under the new two-part comparison.
- Make per-display bounds checking ("does this ordinal still exist?") independent per display, which is the whole point of the change.

**Non-Goals:**
- No migration of existing saved-layout JSON files or `config.yaml` `ordinal` values — both are explicitly recreated/hand-edited (see proposal.md - Impact). Writing a JSON-shape or YAML-string migrator is out of scope.
- No change to how displays themselves are sorted left-to-right, how Spaces are moved, or the window-matching algorithm (title similarity, index fallback) — only the ordinal's shape changes.
- No change to the macOS Mission Control ordinal (`MumuMissionControlIndexForSpace` and friends) — it stays flat, per existing scope boundaries.

## Decisions

### 1. New `space.Ordinal` type: `struct{ Display, Space int }`

Both fields are 1-based. `Ordinal` is comparable (usable as a map key, e.g. the `--sort macos` memoization cache) and gets:
- `Less(other Ordinal) bool`: compares `Display` first, then `Space` — this reproduces today's flat left-to-right ordering exactly, since the flat ordinal was already "every Display A space, then every Display B space, ..." in the same order.
- `String() string`: `"D:SS"` (space zero-padded to 2 digits, display unpadded — matches the proposal's `2:01` example; displays realistically number in the single digits).
- `ParseOrdinal(raw string) (Ordinal, error)`: parses `"D:S"` (padding on either side accepted, e.g. `"02:1"` and `"2:01"` both parse), used by `config.yaml`'s `ordinal:` field. Rejects a bare integer with an error explicitly naming the new `"display:space"` format, so a leftover pre-migration `config.yaml` fails loudly instead of silently misresolving (see proposal.md - Impact: flat ordinals can't be safely reinterpreted).

Zero value `Ordinal{}` continues to mean "unresolved" wherever the old code used `0` for that purpose (e.g. `DualLabel`'s "unknown" fallback, `LogicalSpaceID`'s "not found" return) — `Display == 0` (or `Space == 0`) is never valid since both are 1-based, so it's a safe sentinel without introducing a separate "found" bool everywhere.

Lives in `internal/space` (alongside `logical.go`/`label.go`), not `internal/layout` or `internal/config`, since both of those already import `internal/space` for logical-numbering concerns and this is squarely that concern.

**Alternative considered**: keep two parallel flat `int` fields (`DisplayOrdinal`, `SpaceOrdinal`) instead of a struct, avoiding a new named type. Rejected — every call site that currently threads one `int` (comparisons, map keys, sort cascades, function params) would have to thread two, doubling the diff surface for no benefit; a single comparable struct is what those call sites actually want to hold and compare as one value.

### 2. Native layer: extend, don't replace, the existing left-to-right sort helper

`mumuCopyDisplaySpacesSortedLeftToRight()` (already sorts `SLSCopyManagedDisplaySpaces`' per-display array by physical origin) stays as the single source of truth for display order. Three native entry points replace today's flat trio:

- `int MumuLogicalDisplayCount(void)` — number of connected displays in left-to-right order (already computable via the existing `MumuLeftToRightSpaceCounts`'s `outCount`, but a dedicated call avoids allocating/freeing the counts array when only the display count is needed).
- `uint64_t MumuOrdinalSpaceID(int displayOrdinal, int spaceOrdinal)` — walks the sorted display array to `displayOrdinal`, then that display's `Spaces` array to `spaceOrdinal`; returns 0 if either is out of range. Replaces `MumuLogicalSpaceID`.
- `int MumuSpaceOrdinal(uint64_t sid, int *outDisplay, int *outSpace)` — walks the sorted display array, and for the display containing `sid`, returns that display's 1-based left-to-right position via `*outDisplay` and the space's 1-based position within it via `*outSpace`; returns 0 (leaving the out-params untouched) if `sid` isn't found. Replaces `MumuLogicalIndexForSpace` and folds in the logic already written once for the gesture code's `mumuLocalSpaceIndex` (that static helper is reused/exported rather than duplicated).

`MumuLogicalSpaceCount` (flat total) is dropped: every caller that used it for a whole-arrangement bound now uses `MumuLeftToRightSpaceCounts` (already existing) and checks the specific display's count instead — that's the per-display bounds check the whole change is for.

`internal/space/logical.go`'s Go wrappers change 1:1 with the above: `IDForOrdinal(o Ordinal) uint64`, `OrdinalForSpace(sid uint64) Ordinal`, `LogicalDisplayCount() int`. `LeftToRightSpaceCounts() []int` is unchanged (still per-display counts, still used directly for arrangement-drift and per-display bounds checks).

**Alternative considered**: return the two ordinal components by encoding them into a single `int` at the native boundary (e.g. `display*1000 + space`) and decoding in Go, keeping the cgo signature simpler (single return value instead of out-params). Rejected — it reintroduces exactly the kind of magic flat encoding this change is trying to get away from, and caps space-per-display at an arbitrary encoding width for no real benefit over a clean two-out-param cgo call (already an established pattern in this file, e.g. `MumuLeftToRightSpaceCounts(int *outCount)`).

### 3. `internal/layout.Entry.Ordinal` becomes `space.Ordinal`, JSON shape becomes a nested object

`Entry.Ordinal space.Ordinal` with `json:"ordinal"` and `Ordinal` implementing `MarshalJSON`/`UnmarshalJSON` as `{"display":2,"space":1}` (default struct field marshaling with lowercase JSON tags on `Display`/`Space` is sufficient — no custom marshaling needed beyond field tags). This keeps the JSON key named `ordinal` (unchanged) while changing its shape from a number to an object, so a pre-migration saved layout fails `Entry` unmarshaling with a type-mismatch error from `encoding/json` on that exact field.

`internal/layout.SchemaVersion` bumps from `1` to `2`. `persist.go`'s load path pre-parses just `{"schemaVersion": int}` before the full unmarshal (a second, minimal `json.Unmarshal` pass into a tiny struct) and, if it doesn't match current `SchemaVersion`, returns a clear "this saved layout was written by an older mumu version; run `mumu save` again to recreate it" error naming the file path — instead of letting the full unmarshal fail with a raw `encoding/json` type error. If the schema version does match but the full unmarshal still fails, that's a genuinely malformed file and keeps today's existing "malformed saved-layout file" error path unchanged.

**Alternative considered**: represent `Ordinal` in JSON as a two-element array (`"ordinal": [2, 1]`) instead of an object. Rejected — a bare pair of numbers with no field names is exactly the kind of undocumented internal shape AGENTS.md's data-files guidance is fine with for this file, but `{"display":2,"space":1}` costs nothing extra and is self-describing if anyone (a future migration script, a bug report attachment) ever has to read the raw JSON.

### 4. `config.yaml`'s `ordinal:` stays a scalar, now a `"D:S"` string instead of a bare int

`PinRule.Ordinal` / `DefaultSpaceRule.Ordinal` become `space.Ordinal`; their YAML file-format structs (`pinRuleFileFormat`, `defaultSpaceRuleFileFormat`) change the field's Go type from `int` to `string`, parsed via `space.ParseOrdinal` during `validatePins`/`validateDefaultSpaces` (replacing today's `rule.Ordinal <= 0` check with a parse-and-validate step: parse failure or either part non-positive both produce the existing style of clear, path-and-rule-naming error). `config.yaml`'s generated comments/examples (`ordinal: 1` → `ordinal: "2:1"`) update to match.

**Alternative considered**: nested YAML mapping (`ordinal: {display: 2, space: 1}`), matching the JSON shape exactly. Rejected — `pins`/`default_spaces` rules are meant for a human to type by hand (unlike saved-layout JSON), and a single quoted `"2:1"` reads the same as the CLI's own `#2:01` display convention (matching AGENTS.md's YAML-is-for-humans framing), whereas a 3-line nested mapping per rule is unnecessary ceremony for two small numbers a user is meant to glance at and edit directly.

### 5. Sorting and matching switch from `int` comparison to `Ordinal.Less`

`internal/layout/sort.go`'s `entryLess`/`pinRuleLess`/`defaultSpaceRuleLess` replace every `ordinalA != ordinalB` / `ordinalA < ordinalB` pair with `ordinalA != ordinalB` (struct equality, still cheap) / `ordinalA.Less(ordinalB)`. `newMissionControlOrdinalLookup`'s cache becomes `map[space.Ordinal]int` (valid since `Ordinal` is a comparable struct of two `int`s). `restore.go`'s out-of-range check (`entry.Ordinal < 1 || entry.Ordinal > currentSpaceCount`) becomes: resolve `space.LeftToRightSpaceCounts()`, confirm `entry.Ordinal.Display` is within `1..len(counts)`, then confirm `entry.Ordinal.Space` is within `1..counts[entry.Ordinal.Display-1]` — this is the per-display bounds check that's the actual point of the change, and it's the only place where "does this ordinal still exist" logic needs to change shape rather than just type.

## Risks / Trade-offs

- **[Risk]** Every existing saved layout and every `config.yaml` with `pins`/`default_spaces` becomes unusable without user action the moment this ships (no migration, per proposal.md - Impact). → Mitigation: both failure paths produce specific, actionable error text (re-run `mumu save`; edit `ordinal:` to the new `"D:S"` format) rather than a generic parse failure, and this is called out explicitly in the PR/changelog as a breaking change.
- **[Risk]** The native `MumuSpaceOrdinal`/`MumuOrdinalSpaceID` two-out-param cgo calls are new surface area in security-sensitive private-API territory (`internal/native/space.m` already uses undocumented SkyLight calls). → Mitigation: both new functions are pure re-reads of the same `SLSCopyManagedDisplaySpaces` data `mumuCopyDisplaySpacesSortedLeftToRight` already parses today for the flat functions being replaced; no new private symbols are resolved, only new indexing logic over data already fetched safely elsewhere in this file (and already proven out via `mumuLocalSpaceIndex`'s per-display walk, used today for gesture navigation).
- **[Trade-off]** A Space added or removed *before* another Space within the same display still renumbers that display's later Spaces (this change only isolates one display's numbering from another's, not within-display reordering). → Accepted: this is the scope the proposal and issue #26 describe; no requirement in this change claims full ordinal stability within a display.

## Migration Plan

No automated migration (see Non-Goals). On upgrade:
1. `mumu show`/`mumu restore`/`mumu list` reading an existing saved-layout JSON file report the clear "outdated schema, run `mumu save` again" error described in Decision 3, for every display count that has one.
2. Any `config.yaml` with `pins` or `default_spaces` entries fails validation with a clear per-rule error (Decision 4) until the user rewrites each `ordinal:` value in `"display:space"` form.
3. No flag or command bridges the old and new shapes; this is a clean cut, consistent with the project's small user base and `config.yaml`/saved-layouts both being explicitly non-guaranteed-stable internal/user-owned files per AGENTS.md.

Rollback is a plain revert: nothing on disk is rewritten by this change (validation only ever reads and rejects old-shape files; it never rewrites them), so reverting the binary restores the old behavior against the same still-present old-shape files.
