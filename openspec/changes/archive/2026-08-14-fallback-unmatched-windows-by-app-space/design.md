## Context

`Restore` currently processes saved entries one at a time. Within an application's live-window list, it claims an exact-title, saved-index, or sole remaining candidate; any other open windows have no restore target. The existing active-Space helper resolves the Space beneath the cursor, while this change's tie-break must resolve the current Space of the primary menu-bar display. See `proposal.md` for motivation and the delta spec for the behavior contract.

## Goals / Non-Goals

**Goals:**

- Preserve every existing direct saved-entry matching tier and its one-window-per-move guarantee.
- Derive fallback destinations only from successful, currently resolvable assignments for the same application.
- Make a tied destination visibly recoverable by routing it to the primary display's current logical Space.
- Keep fallback placement diagnosable in normal restore progress.

**Non-Goals:**

- Changing saved-layout data, matching across application bundle identifiers, or launching applications.
- Treating the cursor's display as the primary display.
- Recreating unavailable Spaces or guessing a destination for an application with no valid direct assignment.

## Decisions

### Collect direct assignments before placing unmatched live windows

Restore will retain the current title, index, and sole-candidate match logic as the first phase, grouped by bundle identifier. For each direct match whose target logical Space is still resolvable, it will retain the matched live window, its target ordinal, and the used-live-window state.

After this phase, restore will examine every still-unclaimed live window per application. If that application has valid direct assignments, it will derive one fallback target and enqueue each remaining live window for that target. This makes the result independent of the saved-entry iteration order and ensures a directly matched window always retains its own saved destination.

_Alternative considered_: assign a fallback while processing each unmatched saved entry. Rejected because early entries may not yet establish the app's prevalent destination, and an unmatched live window need not correspond one-to-one with a saved entry.

### Choose a unique per-application mode, then use the primary display for ties

The fallback target will be the logical ordinal with the highest count among an application's valid direct assignments. A single valid assignment is therefore sufficient to establish a destination. If two or more ordinals share the highest count, restore will use the logical ordinal of the Space currently displayed on the primary display (the menu-bar display), as explicitly selected by the user.

If that primary-display Space cannot be resolved to a current logical ordinal, fallback windows will not be moved and restore will report the inability to place them rather than choosing an arbitrary saved Space.

_Alternative considered_: choose the lowest tied ordinal. Rejected because it can hide the affected windows on an arbitrary non-visible Space.

### Resolve the primary display's current Space explicitly

The native Space layer will expose the current Space ID for the primary display, based on the menu-bar display identity, and the Go Space layer will convert it through the existing logical Space-ID-to-ordinal mapping. Restore will use this dedicated helper only for tied fallback targets.

It will not reuse the existing active-Space API because that API intentionally resolves the display under the cursor, which can differ from the primary display and would violate the required tie-break behavior.

### Make fallback moves identifiable without persisting synthetic entries

Fallback moves will use transient restore targets containing the live window identity, bundle ID, title, target ordinal, and a fallback marker. They will participate in the existing paced move ordering and error handling, but their progress output will identify them as fallback placements. These transient targets will never be written to a saved layout.

## Risks / Trade-offs

- [A new window can be grouped with the application's dominant workflow rather than its historical location] → This is intentional best-effort behavior and only applies when the app already has a valid direct assignment.
- [A tie routes several windows to one visible Space] → The primary-display target is deliberate so the user can immediately inspect and organize those windows.
- [Primary-display Space resolution can fail] → Skip only the affected fallback placements with clear progress or summary reporting; never substitute a hidden or arbitrary target.
- [Two-phase matching adds restore bookkeeping] → Isolate target selection into unit-testable helpers and retain the existing native move path.

## Migration Plan

No migration is required: persisted layouts and command syntax remain unchanged. The behavior applies to the next restore. Rollback is a code revert, after which unmatched live windows return to the prior no-move behavior.
