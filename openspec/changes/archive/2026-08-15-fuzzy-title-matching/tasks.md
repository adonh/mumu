## 1. Title similarity scoring

- [x] 1.1 Implement a word-set similarity function: lowercase and whitespace-split both titles into token sets, then score as intersection-over-union (both-empty scores 0).
- [x] 1.2 Unit test the similarity function directly: identical titles, reordered words, case-only differences, no shared words, one or both titles empty.

## 2. Per-application batch assignment

- [x] 2.1 Implement a batch matcher that, given one application's still-unmatched saved entries and still-unclaimed live windows, scores every remaining pair and commits pairs greedily from the highest score down, skipping any entry or window already committed.
- [x] 2.2 Implement the tie-break cascade within the matcher: prefer the candidate whose live position equals the entry's saved `Index`; if that doesn't uniquely resolve it, fall back to a fixed, repeatable ordering.
- [x] 2.3 Unit test: a unique highest-scoring pair is chosen over lower-scoring alternatives.
- [x] 2.4 Unit test: two entries that would each independently prefer the same window are resolved to a valid one-to-one assignment (the core "one window, many entries" guarantee), using a fixture where each entry has a distinct second-best alternative.
- [x] 2.5 Unit test: an entry with zero word overlap against every remaining candidate is still matched when at least one candidate remains.
- [x] 2.6 Unit test: a tie between candidates is broken by the entry's saved `Index` when one tied candidate sits at that position.
- [x] 2.7 Unit test: a tie with no candidate at the saved `Index` is still resolved deterministically (repeated calls with identical input produce the identical assignment).
- [x] 2.8 Unit test: an entry is left unmatched only when no live window remains unclaimed for it (more saved entries than open windows).

## 3. Restructure Restore's matching phase

- [x] 3.1 Add a helper that groups `saved.Entries` by `BundleID`, mirroring the existing `groupLiveByBundle` grouping of live windows.
- [x] 3.2 Replace `Restore`'s flat per-entry matching loop with: iterate bundles, run the batch matcher (Section 2) once per bundle against that bundle's live windows, then apply the existing per-entry downstream logic (ordinal-range validation, Space-ID resolution, `moveTarget` construction, `validAssignmentOrdinals` bookkeeping, `SkippedEntry{SkipUnmatchedWindow}` for anything the batch matcher left unmatched) to each entry in the bundle.
- [x] 3.3 Remove `matchWindowIndex` and `soleRemainingCandidate`, now fully superseded by the batch matcher.
- [x] 3.4 Update `Restore`'s doc comment to describe the new single batch-matching step in place of the old three-tier description.

## 4. Fuzzy marker propagation and reporting

- [x] 4.1 Add a `fuzzy bool` field to `moveTarget`, set when a committed pair's similarity score is below a perfect word-set match.
- [x] 4.2 Add a `Fuzzy bool` field to `SkippedEntry` and thread it through `moveFailureSkip` alongside the existing `fallback` propagation.
- [x] 4.3 Add a `(fuzzy)` marker to the per-window restore progress line (`internal/layout/restore.go`), alongside the existing `(fallback)` marker, shown only when `target.fuzzy` is set.
- [x] 4.4 Add a `(fuzzy)` marker to the skip-summary output (`cmd/mumu/cmd/layout.go`'s `printRestoreSummary`), alongside the existing `(fallback)` marker, shown only when `skipped.Fuzzy` is set.
- [x] 4.5 Confirm a single placement is never marked both `(fallback)` and `(fuzzy)` (they apply to disjoint move sets by construction) with a test or code comment asserting the invariant.

## 5. Update existing tests for removed tiers

- [x] 5.1 Remove or rewrite `internal/layout/restore_test.go` cases that assert the now-removed tiered/ambiguous-refusal behavior (`TestMatchWindowIndex_*` and `TestMatchWindowIndex_NoCandidateReturnsNegativeOne` in particular), replacing them with equivalent coverage against the Section 2 batch matcher's exported/testable entry points.
- [x] 5.2 Add a `Restore`-level test (or extend an existing one) proving the one-to-one guarantee end-to-end: one application with multiple saved entries and multiple live windows where naive per-entry matching would double-claim a window, asserting every live window ends up assigned to at most one saved entry.
- [x] 5.3 Review `internal/layout/restore_fallback_test.go` and `cmd/mumu/cmd`'s restore-summary tests for any fixture or assumption that depended on the removed tiers; update only what's actually affected.

## 6. Documentation

- [x] 6.1 Update `cmd/mumu/cmd/layout.go`'s `restore` command `Long` help text to describe the new batch-matching behavior in place of the old "exact title first, falling back to positional order..." description.
- [x] 6.2 Update `docs/CLI.md`'s restore section to describe the new matching behavior and the `(fuzzy)` output marker alongside the existing `(fallback)` marker.

## 7. Verification

- [x] 7.1 Run `just test-all`, `golangci-lint run`, `go vet ./...`, `go build ./...`, and `just fmt-check`; fix any failures.
- [ ] 7.2 Manually restore a saved layout with intentionally renamed and reordered-word window titles and confirm windows land in the expected Spaces with `(fuzzy)` markers where expected.
