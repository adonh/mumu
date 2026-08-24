## Why

`mumu restore`'s application-level fallback only ever activates when an application has at least one valid saved-entry match this restore, and its target is always computed from that application's most-prevalent assigned Space (or the primary display's current Space on a tie). A user who wants a fixed, predictable landing spot for an application's leftover windows — including an application that currently has zero saved-entry matches, which today gets no fallback placement at all — has no way to configure that.

## What Changes

- Add a `default_spaces` setting to `config.yaml`: a map from connected-display-count to a list of `{bundle_id, space}` entries (mirrors `pins`' per-display-count keying), each declaring a fixed target logical Space for an application's leftover unclaimed windows during `mumu restore`.
- When an application has a configured `default_spaces` entry for the current display count, that target **always** wins for its leftover unclaimed windows this restore, regardless of whether the prevalent-Space heuristic would otherwise have produced a clear (non-tied) target.
- A configured `default_spaces` entry also activates the leftover-window fallback for an application with **zero** valid saved-entry matches this restore — a case that currently receives no fallback placement at all.
- Applications with no configured `default_spaces` entry are unaffected: they keep the existing prevalent-Space-with-primary-display-tie-break heuristic, gated on having at least one valid saved-entry match, exactly as today.
- `mumu show` additionally lists the configured `default_spaces` entries for the display count being shown, alongside the existing pin and saved-layout listings — a config preview only, no live-window matching.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `configuration`: `config.yaml` gains the `default_spaces` setting, with its own per-display-count shape, validation, and default-value (absent/empty) rules, documented the same way `pins` is.
- `space-layout`: `mumu restore`'s application-level fallback gains a per-application configured-override path that both supersedes the prevalent-Space heuristic when configured and activates a fallback for applications with zero valid saved-entry matches; `mumu show`'s output gains the `default_spaces` preview listing.

## Impact

- `internal/config/config.go`, `internal/config/config_test.go`: parse and validate `default_spaces` (`map[int][]DefaultSpaceRule]`); extend `defaultConfigYAML`.
- `internal/layout/restore.go`: `planFallbackMoves` (or its caller) gains a per-bundle configured-target lookup that short-circuits the prevalent-Space computation and removes the "at least one valid assignment" gate for bundles with a configured default.
- `cmd/mumu/cmd/layout.go`: `mumu show` prints the resolved display count's configured `default_spaces` entries.
- `docs/CONFIG_SCHEMA.md`, `docs/CLI.md`, `AGENTS.md`: document the new `config.yaml` key and its precedence over the prevalent-Space heuristic.
