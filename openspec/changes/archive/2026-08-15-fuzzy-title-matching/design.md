## Context

`Restore` currently matches saved entries one at a time, in saved order, inside a single loop: for each entry it tries an exact title match, then the saved positional index, then whether exactly one live window remains unclaimed. All three tiers only ever look at the one entry currently being processed and the live windows not yet claimed by an earlier entry — none of them can see entries later in the loop, so an earlier entry can claim a window that would have been a better (or only) match for a later one. See `proposal.md` for motivation and the delta spec for the new behavior contract.

## Goals / Non-Goals

**Goals:**

- Guarantee that within one application, no currently open window is assigned to more than one saved entry, and that assignment isn't sensitive to saved-entry iteration order.
- Make matching deterministic: identical entries, windows, and used-state always produce the identical assignment.
- Keep the existing app-level Space fallback (used-index map, `moveTarget`/`SkippedEntry` shapes, progress/skip reporting) working unchanged against whatever the new matcher leaves unclaimed.
- No new external dependency and no persisted-schema or CLI-flag change.

**Non-Goals:**

- Guaranteeing a globally-optimal (maximum-total-similarity) assignment. Greedy highest-score-first is an approximation; see Risks.
- Any similarity threshold or matching behavior configurable via a flag.
- Matching across application bundle identifiers, or changing save-time capture.

## Decisions

### Restructure the matching phase into a per-bundle batch

Group `saved.Entries` by `BundleID` before matching, the same way live windows are already grouped by `groupLiveByBundle`, then run the new batch matcher once per bundle against that bundle's still-unclaimed live windows — replacing the single flat per-entry loop for the matching step specifically.

_Alternative considered_: keep the existing single-pass per-entry loop and add scoring as one more sequential tier within it. Rejected — a sequential pass can't see entries not yet reached, so it can't prevent the exact "one window claimed by many entries" bug this change exists to fix.

### Score every remaining pair, then assign greedily from the highest score down

Per application, compute a similarity score for every (still-unmatched entry, still-unclaimed window) pair, then repeatedly commit the single highest-scoring remaining pair and remove both sides from further consideration, until no entries or no windows remain.

_Alternative considered_: an optimal bipartite assignment (e.g. the Hungarian algorithm) that maximizes total similarity across the whole set. Provably better in pathological cases — greedy can occasionally leave a worse total pairing than an optimal solver would choose — but adds real implementation and testing complexity for what is normally a handful of windows per application. Per explicit direction, this is deferred as a documented future improvement rather than built now.

### Word-set similarity over lowercased, whitespace-split tokens

Split both titles on whitespace, lowercase each token, and score as the intersection-over-union ratio of the two token sets (both-empty scores 0, not 1 — no title information is "no signal," not a perfect match, so it falls through to the tie-break cascade like any other zero-scoring pair rather than being spuriously preferred). This is order-independent, so reordered-word titles score identically to a literal match, and needs no external dependency.

_Alternative considered_: normalized Levenshtein edit-distance ratio. Rejected in favor of word-overlap, which tolerates reordered words that edit distance penalizes, per explicit direction.

### No minimum similarity threshold

Any remaining live window is an eligible match for any remaining entry regardless of score, as long as the greedy assignment reaches it. Per explicit direction, always placing something is preferred over refusing to guess, even when the best available score is low or zero.

### Tie-break cascade: score, then saved index, then a fixed deterministic order

When multiple pairs share an entry's (or window's) top score, prefer the pair whose window's current position equals the entry's saved `Index`; if that doesn't uniquely resolve it, fall back to a fixed, repeatable ordering (ascending live-window index, then ascending entry index) instead of leaving the tie unresolved. This intentionally removes today's exact-title tier's refusal to guess when a title matches more than one open window — per explicit direction, a deterministic pick is preferred over leaving a placeable window unmatched.

### Mark non-exact matches with a transient `fuzzy` field, mirroring `fallback`

Add `fuzzy bool` to `moveTarget` and `Fuzzy bool` to `SkippedEntry`, set when a committed pair's score is below a perfect word-set match, and thread it through `moveFailureSkip` exactly as `fallback` already is. Progress and skip-summary output show a `(fuzzy)` marker alongside the existing `(fallback)` marker. The two markers apply to disjoint sets of moves: fallback placements are synthesized for windows never matched to any saved entry, so a move is never both.

## Risks / Trade-offs

- [Greedy assignment is an approximation, not a guaranteed-optimal one] → Documented now as a known limitation; an optimal bipartite solver (e.g. Hungarian algorithm) is a candidate follow-up if greedy proves insufficient in practice. Not built now, per explicit direction.
- [Removing the minimum-threshold and tie-refusal guards means restore will now confidently move a window based on a low- or zero-similarity score, or an arbitrarily broken tie] → Intentional, explicit behavior change (see proposal's "What Changes"). The `(fuzzy)` marker keeps these placements visible and distinguishable from confident matches so a user can spot and correct a bad guess.
- [Restructuring the main loop from a flat per-entry pass to a grouped per-bundle batch pass changes iteration order] → Downstream logic (ordinal validation, `moveTarget` construction, `validAssignmentOrdinals` bookkeeping) doesn't depend on iteration order today — `toMove` is always re-sorted by the caller-selected sort key before use — so this is safe, but any test asserting on raw iteration order needs review.
- [Several existing tests assert today's "leave genuinely ambiguous entries unmatched" behavior] → These are intentionally being replaced, not preserved; see tasks.md.

## Migration Plan

No persisted-data or CLI-surface changes; the new matching behavior applies to the next `mumu restore` invocation. Rollback is a plain code revert, with no transitional state to clean up.
