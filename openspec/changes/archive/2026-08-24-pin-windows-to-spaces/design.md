## Context

`internal/layout/restore.go` already batch-matches one application's saved `Entry` values against its currently open windows (`matchEntries` in `internal/layout/match.go`, scored by `titleSimilarity` in `internal/layout/similarity.go`), then moves matched windows and fills in an application-level fallback for anything left over (`planDirectMoves`/`planFallbackMoves`). `config.yaml` today exposes one flat setting, `data_dir` (`internal/config/config.go`). See `proposal.md` - Why for the motivation; this doc covers how pins plug into that existing matching/move pipeline without duplicating it.

## Goals / Non-Goals

**Goals:**
- Reuse the existing title-similarity matching and move/skip-reporting machinery for pins, rather than building a parallel implementation.
- Keep `config.yaml` parsing/validation additions consistent with the existing auto-create-with-defaults, never-silently-fall-back pattern (`AGENTS.md` - Configuration and data files).
- Make the pin-vs-layout precedence ordering a small, explicit seam so it's easy to reason about and to flip later.

**Non-Goals:**
- No standalone "apply pins" command, no launching of non-running applications, no new Space creation — pins reuse `space-layout`'s existing constraints on all of these (see the `window-pinning` spec).
- No CLI (`mumu pin add/rm`) for managing pins — `config.yaml` hand-editing is the only interface; `mumu show` only previews.
- No change to `mumu save`/capture — pins are read-only configuration, never written by mumu.

## Decisions

- **Model a pin rule as a `layout.Entry`-shaped value with `Index: -1`, and reuse `matchEntries` unchanged.** `Entry` already carries `BundleID`, `Title`, `Index`, `Ordinal` — exactly what a pin rule needs (bundle ID, title pattern, target ordinal). Setting `Index: -1` means the saved-positional-index tie-break in `candidateLess` can never fire for a pin (no live window has index `-1`), so ties between pins fall through to the same deterministic fixed order saved entries already use when their index doesn't help. This avoids writing a second matching implementation. Alternative considered: a distinct `PinRule`/matching path — rejected as needless duplication of already-tested logic.
- **Represent precedence as a two-phase live-window pool split, not a merged single match.** `Restore` currently builds one `liveByBundle` map from `window.AllAcrossSpaces()` and feeds it to `planDirectMoves`. With pins, `Restore` will run `planDirectMoves` twice — once for pins, once for saved entries — in the order `pin_precedence` dictates, with the second call's `liveByBundle` filtered to exclude window indices the first call claimed (via the same per-bundle `usedIndex` map `planDirectMoves`/`planFallbackMoves` already thread through, extended to be seeded from the higher-precedence phase's claims rather than starting empty). This keeps `planDirectMoves` itself unchanged; only `Restore`'s orchestration and the live-window filtering step are new. Application-level fallback (`planFallbackMoves`) only ever runs for the saved-layout phase, per the `window-pinning` spec's non-goal of pins having their own fallback — a pin that doesn't match a window is simply skipped, never given a prevalent-Space fallback.
- **`pins` YAML shape: `map[int][]PinRule]`, `PinRule{BundleID, Title, Space string/int}`.** Keyed directly by display count (an integer YAML key, e.g. `2:`, `4:`), each a list of `{bundle_id, title, space}` rules — mirrors how saved layouts are already keyed by display count, so it reads naturally next to the `space-layout` capability's own numbering. `pin_precedence: pin|layout` sits as its own flat top-level key (not nested under `pins`) since it's a single global toggle, not per-display-count data.
- **`mumu show`'s pin listing is unresolved config, not a live-matched preview.** Printing the raw `{bundle_id, title, space}` tuples (with `space.DualLabel` for the ordinal) avoids running window enumeration/matching just to preview, keeps `show` fast and side-effect-free the way it is today, and avoids the confusion of a preview match silently going stale by the time `restore` actually runs.
- **Config validation for `pins`/`pin_precedence` lives alongside `data_dir`'s existing validation in `internal/config/config.go`**, following the same "reject with a clear, path-naming error, never partially apply" pattern already established there, rather than deferring validation to `restore`/`show` call time.

## Risks / Trade-offs

- Splitting `liveByBundle` into two filtered pools per phase adds a small amount of orchestration complexity to `Restore` → Mitigated by keeping `planDirectMoves`/`planFallbackMoves` themselves untouched and covering the split with unit tests exercising both `pin_precedence` values directly (see `tasks.md`).
- A pin rule that never matches any window (typo'd title, app never running) fails silently unless a user reads `mumu restore`'s skip summary → Mitigated by reusing the existing skip-reporting pipeline (`SkippedEntry`) so pins show up there exactly like unmatched saved entries, with no new reporting surface to miss.
- Two independent settings (`pins`, `pin_precedence`) both gated behind restore-only activation could surprise a user who expects pins configured for a display count with no saved layout to still do something → Mitigated by making this explicit in both the proposal and the `window-pinning` spec's "Pins require restore and have no independent trigger" requirement.
