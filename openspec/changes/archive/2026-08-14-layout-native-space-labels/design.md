## Context

See `proposal.md` for motivation. This is a small, cross-cutting change (native → `internal/space` → CLI formatting), and it introduces one new technical piece — a reverse Space-ID-to-Mission-Control-ordinal lookup — so it's worth a short design pass before implementation.

`internal/space/logical.go` already has the mirror-image piece for the *logical* numbering: `LogicalIndexForSpace(sid) int`, backed by `MimiLogicalIndexForSpace` in `internal/native/space.m`, which linear-scans the left-to-right-sorted `SLSCopyManagedDisplaySpaces()` result looking for a matching Space ID. `internal/space/space.go`'s `ActiveIndex()` does the same kind of scan, but inline and only for the *raw* (primary-first) `SLSCopyManagedDisplaySpaces()` order — it was never factored into a reusable, general-purpose function.

## Goals / Non-Goals

**Goals:**
- Let any Space ID be resolved to its current 1-based Mission Control ordinal (the number `mimi action space <n>` uses), the same way it can already be resolved to its logical ordinal.
- Show both numbers together everywhere `mimi layout` currently prints a per-window Space number, with no change to the saved layout file format.

**Non-Goals:**
- Changing `mimi action space`/`move_window_to_space` behavior or numbering.
- Persisting the Mission Control ordinal in saved layouts — it's resolved fresh at print time because it depends on the live display arrangement, which can differ between save and a later restore/show.

## Decisions

**Add `MimiMissionControlIndexForSpace(sid) int` (native) / `space.MissionControlIndexForSpace(sid) int` (Go), mirroring the existing logical pair.** Same shape as `MimiLogicalIndexForSpace`/`LogicalIndexForSpace`, but scanning the raw (unsorted) `SLSCopyManagedDisplaySpaces()` order instead of the left-to-right-sorted one. Returns 0 if the Space ID isn't found (consistent with the logical version's existing convention).
_Alternative considered_: extract `ActiveIndex()`'s inline scan into this new function and have `ActiveIndex()` call it. Rejected only in the sense of scope — that's a reasonable follow-up cleanup, but not required for this change; `ActiveIndex()` is left as-is to keep the diff focused on the new capability.

**Display format: `space <logical> (macOS space <mission-control>)`.** e.g. `space 3 (macOS space 21)`. Kept as plain, greppable text consistent with the rest of `mimi layout`'s output style (no color/tables); the parenthetical makes clear which number is which without needing a legend. Applied consistently to `mimi layout show`, restore's per-window move progress, and restore's skip summary.
_Alternative considered_: a compact `3/21` form. Rejected as more compact but more ambiguous to a first-time reader with no established convention for which side is which.

**Resolution happens at print time, not at capture/restore time.** The Mission Control ordinal for a given Space ID can change if displays are reconnected/reordered between when a layout was saved and when it's shown or restored — the whole point of surfacing it is to reflect what's true *right now*. Computing it lazily, only for the entries actually being printed, avoids doing unnecessary work for entries that end up skipped before printing (e.g. app not running) and avoids ever persisting a number that could go stale.

## Risks / Trade-offs

- **Extra native round-trip per printed line**: `MissionControlIndexForSpace` linear-scans `SLSCopyManagedDisplaySpaces()`'s result (typically a handful of displays × a handful of Spaces each) once per printed entry. Negligible relative to the existing per-window-move pacing delay (~0.2s) restore already has.
- **Space ID not found (returns 0)**: can happen if the target Space no longer exists (already a known, reported case — `SkipOrdinalOutOfRange`) or in the rare event of a transient WindowServer state during a display change. Display as `space <logical> (macOS space unknown)` rather than a bare `0`, so the output never looks like it's claiming Space "0" exists.
