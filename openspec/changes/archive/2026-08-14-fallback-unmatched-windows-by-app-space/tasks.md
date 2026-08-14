## 1. Primary-display Space resolution

- [x] 1.1 Add a native Space API that resolves the current Space ID of the primary (menu-bar) display without using cursor position.
- [x] 1.2 Expose a Go helper that converts the primary display's current Space ID to a logical Space ordinal and returns a clear error when it cannot be resolved.

## 2. Per-application fallback targeting

- [x] 2.1 Refactor restore planning into a direct-match phase that preserves title, index, and sole-candidate behavior while recording valid assignments and claimed live windows per application.
- [x] 2.2 Add unit-testable target selection that chooses each application's uniquely most prevalent valid logical Space and uses the primary-display logical Space when the mode is tied.
- [x] 2.3 Queue every remaining unclaimed live window only for applications with a valid direct assignment; leave windows untouched when no target is available or the tied primary target cannot be resolved.
- [x] 2.4 Mark fallback move targets as transient, include them in existing paced moves and ordering, and identify them in restore progress and failure reporting without persisting them.

## 3. Verification and documentation

- [x] 3.1 Add unit tests for unique and tied prevalent targets, primary-target resolution failure, no valid assignment, and one-move-per-live-window behavior; retain coverage for the existing direct matching tiers.
- [x] 3.2 Update restore CLI documentation to explain per-application fallback placement and the primary-display tie-break.
- [x] 3.3 Run formatting, focused layout tests, and the repository's Go validation commands.
