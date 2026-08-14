## Why

Restore currently skips an application's open windows when title, saved index, and sole-candidate matching cannot identify them, even when other windows from that same application establish where the application normally belongs. This commonly leaves newly opened or retitled browser windows on an unexpected Space after a restore.

## What Changes

- Preserve the existing title, index, and sole-remaining-candidate matching behavior for saved layout entries.
- After direct matches are established for a running application, place that application's remaining unmatched open windows on the logical Space that is most prevalent among its valid matched assignments.
- When multiple target Spaces are equally prevalent, place the unmatched windows on the logical Space currently displayed on the primary (menu-bar) display, making them immediately visible to the user.
- Leave unmatched windows alone when their application has no valid matched assignment from which to infer a target.
- Include these fallback placements in restore progress reporting and document the behavior.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `space-layout`: Extend layout restore matching to place non-matching open windows using a per-application prevalent target Space, with the primary display's current Space as the deterministic tie-break.

## Impact

- Affected code: `internal/layout` restore matching and tests; potentially `internal/space` and native Space helpers to resolve the primary display's current logical Space; CLI restore documentation.
- No saved-layout schema, command syntax, external API, or dependency changes.
