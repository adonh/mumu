## Why

Restore's title matching only accepts an exact, unambiguous title string. Any drift between a saved title and a window's current title — a renamed document, a reordered word, a changed unsaved-indicator — makes that tier useless, so restore falls through to positional index or the sole-remaining-candidate tier, both of which ignore the title entirely. When an application has multiple open windows and none of them matches a saved title exactly, restore currently gives up on a per-window basis rather than using the title similarity it does have.

## What Changes

- **BREAKING**: Replace the three sequential per-entry tiers (exact title → saved index → sole remaining candidate) with a single per-application batch match: for every application, score every still-unclaimed saved entry against every still-unclaimed open window using word-level (token) title similarity, then assign pairs globally from the highest score down, so one window can no longer be claimed as the "best" match for more than one saved entry.
- Similarity is computed via Jaccard overlap of lowercased, whitespace-split tokens — order-independent, no external dependency.
- No minimum similarity score is required: as long as at least one open window remains unclaimed for an application, every saved entry for it now finds a match. `SkipUnmatchedWindow` now only occurs when an application has more saved entries than currently open windows.
- Tie-breaking when two or more candidates share the same best score for an entry (or vice versa): prefer the candidate whose current position equals the entry's saved index; if that still doesn't resolve it, pick deterministically by list order. Ties are no longer left unresolved — this is an intentional, explicit behavior change from today's exact-title tier, which currently refuses to guess when a title matches more than one open window.
- "Exact" match is no longer a distinct tier: a saved title and a window title with the same token set (case-insensitive, regardless of word order) score 1.0, functionally the same outcome — this includes titles that merely have reordered words, which previously did not exact-match.
- Restore progress and skip-summary output mark entries resolved with a similarity score below 1.0 with a `(fuzzy)` marker, mirroring the existing `(fallback)` marker for app-level Space fallback placements. Score-1.0 matches remain unmarked, matching today's exact-match presentation.
- The application-level prevalent-Space fallback for windows left unclaimed after matching (added in `fallback-unmatched-windows-by-app-space`) is unchanged; it still only runs on whatever remains unclaimed after the new batch match.
- Documented as a known follow-up, not built now: this uses greedy highest-score-first assignment, an approximation that is not guaranteed to maximize total similarity across an application's whole window set. A future change could replace it with an optimal bipartite assignment (e.g. the Hungarian algorithm) if greedy assignment proves insufficient in practice.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `space-layout`: The "Restore matches windows by title, falling back to index" requirement is replaced with a single scored, per-application batch-matching requirement with defined tie-break and reporting behavior. The app-level prevalent-Space fallback requirement in the same capability is unaffected.

## Impact

- `internal/layout/restore.go`: replace `matchWindowIndex` and `soleRemainingCandidate` with a per-application batch matcher; restructure `Restore`'s main loop to group saved entries by bundle ID before matching (mirroring the existing `liveByBundle` grouping); add a `fuzzy` marker to `moveTarget` and `SkippedEntry`, propagated through `moveFailureSkip` and progress/skip output.
- `internal/layout/restore_test.go`: several existing test cases assert today's "leave genuinely ambiguous entries unmatched" behavior (e.g. `TestMatchWindowIndex_NoCandidateReturnsNegativeOne`); these must be rewritten to reflect that a candidate is now always chosen when one remains available. New tests cover token similarity scoring, the index/arbitrary tie-break cascade, and the global assignment's one-window-per-entry guarantee.
- `cmd/mumu/cmd/layout.go` and `docs/CLI.md`: restore command help text and documentation describing the old three-tier matching behavior need rewriting to describe the new scored batch match.
- No change to persisted layout schema (`Entry`) or any CLI flags.
