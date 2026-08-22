## 1. Config schema

- [x] 1.1 Add `PinRule{BundleID, Title string, Space int}` and a `map[int][]PinRule` `Pins` field plus a `PinPrecedence` (`"pin"`/`"layout"`, default `"pin"`) field to `config.Config` and its `fileFormat` in `internal/config/config.go`; verify with a unit test that a config file with `pins:` and `pin_precedence:` parses into the expected struct values.
- [x] 1.2 Validate `pins`/`pin_precedence` on load: each rule has a non-empty `bundle_id`/`title` and a positive `space`; `pin_precedence`, if present, is exactly `pin` or `layout`; on violation return a `CodeInvalidConfig` error naming the config file path and the offending entry. Verify with unit tests for each invalid case (missing field, non-positive space, bad precedence value) and confirm no partial `Config` is returned.
- [x] 1.3 Update `defaultConfigYAML` with commented, empty-by-default `pins`/`pin_precedence` documentation so newly created config files explain the settings. Verify by inspecting the auto-created file's contents in a unit test.

## 2. Pin matching and restore integration

- [x] 2.1 Add a helper that converts a display count's configured `PinRule`s into `[]layout.Entry` (with `Index: -1`), grouped by bundle ID the same way saved entries are grouped. Verify with a unit test asserting the conversion and that `Index` is always `-1`.
- [x] 2.2 Extend `internal/layout/restore.go`'s `Restore` to run `planDirectMoves` twice — once for pin entries, once for saved entries — in the order given by `pin_precedence`, threading a shared `usedIndex` map so the second phase's `liveByBundle` excludes windows the first phase claimed. Verify with unit tests covering both precedence orders: a pin and a saved entry independently able to claim the same window, asserting only the higher-precedence phase's target is moved.
- [x] 2.3 Ensure pin-matched windows use the existing move/skip pipeline (`moveTarget`, `SkippedEntry`, fuzzy marker) unchanged, with no application-level fallback (`planFallbackMoves`) applied to unmatched pins. Verify with a unit test that an unmatched pin appears in `RestoreSummary.Skipped` and never triggers a fallback placement.
- [x] 2.4 Wire `config.Load()`'s `Pins`/`PinPrecedence` into the `mumu restore` command path (`cmd/mumu/cmd/layout.go` or wherever `layout.Restore` is invoked), resolving the current display count's pin rules before calling `Restore`. Verify by running `just build` and manually confirming `mumu restore --help`/behavior is otherwise unchanged when no pins are configured.

## 3. `mumu show` pin preview

- [x] 3.1 Add a pin-listing section to `layoutShowCmd` in `cmd/mumu/cmd/layout.go` that prints the resolved display count's configured pin rules (bundle ID, title pattern, target Space via `space.DualLabel`) without performing any window matching. Verify with a CLI/unit test asserting the printed output for a display count with configured pins, and that output is unchanged (no empty section artifacts) when no pins are configured.

## 4. Documentation

- [x] 4.1 Update `docs/CONFIG_SCHEMA.md` with the `pins` and `pin_precedence` keys, their types, defaults, and an example. Verify by re-reading the rendered doc against the actual struct/validation.
- [x] 4.2 Update `docs/CLI.md` (or equivalent) describing how `mumu restore` applies pins and how `mumu show` previews them. Verify by cross-checking against `window-pinning`'s spec scenarios.
- [x] 4.3 Update `AGENTS.md` - Configuration and data files if the pin settings introduce any new convention worth capturing (e.g. nested YAML list under a flat key). Verify by reviewing the section reads coherently with the rest of the file.

## 5. Verification

- [ ] 5.1 Run `just fmt`, `just lint`, `just build`, and `just test`; confirm all four pass with no new warnings. **BLOCKED**: see pause note below.
- [ ] 5.2 Manually exercise `mumu restore` with pins configured for the current display count against real open windows, confirming pinned windows land on their configured Space and `mumu show` lists the configured pins correctly.
