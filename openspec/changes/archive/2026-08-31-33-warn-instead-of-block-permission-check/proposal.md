## Why

`mumu save` and `mumu restore` used to hard-block with an error whenever mumu's own Accessibility/Screen Recording preflight check reported a permission as not granted. On at least one MDM-managed Mac, that check can report a false negative — `mumu status` shows Accessibility as not granted even after granting it correctly in System Settings (toggle stays on, approval prompt granted) — which meant `mumu save`/`restore` refused to run at all, even though the underlying window/Space operations would actually have succeeded. See [GitHub Issue #33](https://github.com/adonh/mumu/issues/33).

## What Changes

- `mumu save` and `mumu restore` no longer stop when their permission preflight check reports Accessibility and/or Screen Recording as not granted. They print a warning naming the permission(s) that appear missing and proceed with the operation anyway.
- The real native window/Space calls remain the source of truth: if a permission really is missing, the operation now fails (or degrades, e.g. unmatched entries from empty window titles) through its normal error/skip reporting instead of a separate preflight error.
- `mumu status` is unaffected — it still exists specifically to report the raw permission state, including a false negative if the underlying OS check itself is wrong.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `space-layout`: the "No elevated privileges required" requirement's contract for a missing Accessibility/Screen Recording permission changes from "reports clearly that permission is required and makes no changes" to "prints a warning and proceeds with the operation anyway."

## Impact

- `internal/permissions/check.go`: add a `Warnings(result CheckResult) []string` helper alongside the existing `FriendlyError`.
- `internal/layout/capture.go`, `internal/layout/restore.go`: replace the hard-blocking permission-check call with a non-blocking one that emits a warning through the existing progress-reporting mechanism instead of returning an error.
- `openspec/specs/space-layout/spec.md`, `docs/CLI.md`, `docs/TROUBLESHOOTING.md`: update to describe warn-and-continue instead of hard-block behavior.
- No change to `mumu status`, saved layout file format, or `config.yaml` schema.
