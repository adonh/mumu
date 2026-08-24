## 1. Config schema

- [x] 1.1 Add `DefaultSpaceRule{BundleID string, Space int}` and a `map[int][]DefaultSpaceRule` `DefaultSpaces` field to `config.Config` and its `fileFormat` (`defaultSpaceRuleFileFormat`) in `internal/config/config.go`, mirroring `PinRule`/`pinRuleFileFormat`. Verify with a unit test that a config file with `default_spaces:` parses into the expected struct values.
- [x] 1.2 Add `validateDefaultSpaces`, mirroring `validatePins`: each rule has a non-empty `bundle_id` and a positive `space`; on violation return a `CodeInvalidConfig` error naming the config file path, display count, and offending entry. Verify with unit tests for each invalid case (missing `bundle_id`, non-positive `space`) and confirm no partial `Config` is returned.
- [x] 1.3 Update `defaultConfigYAML` with commented, empty-by-default `default_spaces` documentation (placed near `pins`, explaining the distinction: no `title`, application-level fallback only) so newly created config files explain the setting. Verify by inspecting the auto-created file's contents in a unit test.

## 2. Fallback resolution

- [x] 2.1 Add a helper that converts a display count's configured `DefaultSpaceRule`s into a `map[string]int` (bundle ID → target logical ordinal). Verify with a unit test asserting the conversion, including that a duplicate `bundle_id` for the same display count is handled deterministically (last-one-wins or a load-time validation error — pick one and cover it with a test).
- [x] 2.2 Extend `planFallbackMoves`'s bundle-ID iteration to include every bundle ID present in the configured-default map, not only those in `assignmentOrdinals`, and short-circuit `fallbackTargetForAssignments`'s prevalent/tie logic whenever a configured default exists for that bundle ID. Verify with unit tests: (a) a configured default overrides an unambiguous prevalent Space, (b) a configured default overrides a tied prevalent Space (no primary-display lookup happens), (c) a configured default activates a fallback for a bundle ID with zero valid assignments this restore.
- [x] 2.3 Add a `defaultConfigured` (or equivalently named) marker to `moveTarget`, distinct from `fallback`/`fuzzy`, so restore progress output and `SkippedEntry` reporting can distinguish a configured-default placement from a prevalent-Space placement. Verify with a unit test asserting the progress-line marker text differs between the two cases.
- [x] 2.4 Thread `cfg.DefaultSpaces[displayCount]` from `cmd/mumu/cmd/layout.go`'s `mumu restore` invocation of `layout.Restore` down through `planLayoutPhase` into `planFallbackMoves`. Verify by running `just build` and manually confirming `mumu restore` behavior is unchanged when no `default_spaces` are configured.

## 3. `mumu show` preview

- [x] 3.1 Add `printConfiguredDefaultSpaces` to `cmd/mumu/cmd/layout.go`, mirroring `printConfiguredPins`, that prints the resolved display count's configured `default_spaces` rules (bundle ID, target Space via `space.DualLabel`) without performing any window matching, called from the same `mumu show` path as `printConfiguredPins`. Verify with a CLI/unit test asserting the printed output for a display count with configured default spaces, and that output is unchanged (no empty section artifacts) when none are configured.

## 4. Documentation

- [x] 4.1 Update `docs/CONFIG_SCHEMA.md` with the `default_spaces` key, its type, default, an example, and its precedence over the prevalent-Space heuristic. Verify by re-reading the rendered doc against the actual struct/validation.
- [x] 4.2 Update `docs/CLI.md` describing how `mumu restore` applies a configured default space and how `mumu show` previews it. Verify by cross-checking against the `space-layout` spec's new scenarios.
- [x] 4.3 Update `AGENTS.md` - Configuration and data files if `default_spaces` introduces any convention worth capturing. Verify by reviewing the section reads coherently with the rest of the file.

## 5. Verification

- [x] 5.1 Run `just fmt`, `just lint`, `just build`, and `just test`; confirm all four pass with no new warnings.
- [x] 5.2 Manually exercise `mumu restore` with a `default_spaces` rule configured for an application that (a) has an unambiguous prevalent Space from saved entries, and (b) has zero saved-entry matches this restore; confirm the configured Space wins in both cases and `mumu show` lists the configured rule correctly. **Skipped/deferred**: user will verify manually later against their real desktop.
