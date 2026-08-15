# Manual Test: Fuzzy Title Matching in `mumu restore`

Steps to manually confirm `mumu restore`'s batch title-matching lands windows
on the right Spaces and shows `(fuzzy)` markers only for approximate matches.

## Setup

1. Open 2+ windows of the same application on different Mission Control
   Spaces. Terminal.app works well since its window title can be changed
   freely without closing/reopening the window:
   - Window A: `printf '\e]0;Deploy Script\a'`
   - Window B: `printf '\e]0;Log Viewer\a'`
2. Run `mumu save` to capture the current layout.

## Test cases

### Reordered words — should NOT be marked fuzzy

- Change Window A's title to the same words, different order:
  `printf '\e]0;Script Deploy\a'`
- Word sets still match exactly (order-independent), so this should restore
  with no marker.

### Renamed title — SHOULD be marked fuzzy

- Change Window B's title to different words:
  `printf '\e]0;Deployment Script\a'`
- This is an approximate (non-exact) match, so it should restore with a
  `(fuzzy)` marker.

## Run restore

1. Move both windows to different Spaces than where they were saved, so
   restore actually has to move them.
2. Run `mumu restore` (add `--yes` to skip the arrangement-mismatch prompt
   if it appears).
3. Check the printed progress lines:
   - Window A (reordered) → no marker.
   - Window B (renamed) → `(fuzzy)` marker.
4. Confirm each window landed back on its originally saved Space (shown as
   `#N (space M)` in the output).

## Optional: one-to-one guarantee under contention

1. Open 3 windows of the same app; save.
2. Rename two of them to titles that would both plausibly best-match the
   same saved title (e.g. similar generic wording).
3. Restore, and confirm no window is claimed twice — every window still
   gets its own distinct Space assignment, with `(fuzzy)` on the
   less-confident match(es).

## Wrap-up

Once satisfied, check off task 7.2 in
`openspec/changes/fuzzy-title-matching/tasks.md`. This change is then ready
to archive with `/opsx-archive`.
