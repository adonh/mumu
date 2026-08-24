## Why

`mumu restore` can only place windows where they were when `mumu save` last ran. Some windows a user always wants on the same Space regardless of what was captured (e.g. Slack always on Space 1) currently require re-saving every time that arrangement drifts. A user-editable, per-display-count pin list in `config.yaml` lets a user declare those fixed assignments once, using the same approximate title matching `mumu restore` already relies on.

## What Changes

- Add a `pins` setting to `config.yaml`: a map from display count to a list of pin rules, each specifying an application bundle identifier, a window-title pattern, and a target logical Space ordinal (the same left-to-right numbering saved layouts already use). Different display counts (e.g. 2 vs. 4 displays) can declare entirely different pins.
- Add a `pin_precedence` setting to `config.yaml` (`pin` or `layout`, default `pin`) controlling which wins when a pin and a saved-layout entry would both claim the same open window during restore.
- `mumu restore` matches each display count's configured pins against that application's currently open windows using the exact same title-similarity heuristic (shared-word Jaccard score, greedy highest-score-first assignment, deterministic tie-break) `mumu restore` already uses for saved-layout entries, then moves matched windows to their pinned Space. Depending on `pin_precedence`, this pin-matching phase runs either before the saved-layout matching phase (claiming windows first) or after it (only considering windows the saved layout left unclaimed).
- `mumu show` additionally lists the configured pins for the display count being shown (application, title pattern, target Space, shown with both logical and Mission Control numbers), alongside the existing saved-layout entry listing. This is a config preview only — no live-window matching is performed for this listing.
- Pins have no effect outside `mumu restore`; there is no standalone "apply pins" command, and pins never trigger without a saved layout for the current display count (restore still requires one to run at all).

## Capabilities

### New Capabilities

- `window-pinning`: user-configured, per-display-count app+title-pattern-to-Space pin rules that participate in `mumu restore`'s window matching and moving, and are previewed by `mumu show`.

### Modified Capabilities

- `configuration`: `config.yaml` gains the `pins` and `pin_precedence` settings, with their own validation and default-value rules.
- `space-layout`: `mumu restore`'s matching/move algorithm gains a pin-matching phase interleaved with the existing saved-layout matching phase (order controlled by `pin_precedence`); `mumu show`'s output gains the pin-preview listing described above.

## Impact

- `internal/config/config.go`, `internal/config/config_test.go`: parse and validate `pins` (map[int][]PinRule) and `pin_precedence`; extend `defaultConfigYAML`.
- `internal/layout/restore.go`, `internal/layout/match.go`: new pin-matching phase reusing `matchEntries`/`titleSimilarity`, sequenced relative to `planDirectMoves`/`planFallbackMoves` per `pin_precedence`; likely a shared helper so pins and saved entries can both feed the existing move/skip-reporting pipeline.
- `cmd/mumu/cmd/layout.go`: `mumu show` prints the configured pins for the resolved display count.
- `docs/CONFIG_SCHEMA.md`, `docs/CLI.md`, `AGENTS.md`: document the new `config.yaml` keys and the restore precedence behavior.
