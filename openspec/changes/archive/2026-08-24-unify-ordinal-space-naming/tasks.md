## 1. Rename the config schema (Go + YAML)

- [x] 1.1 In `internal/config/config.go`, rename `PinRule.Space` → `PinRule.Ordinal` and `DefaultSpaceRule.Space` → `DefaultSpaceRule.Ordinal` (update their doc comments too), rename `pinRuleFileFormat.Space`/`defaultSpaceRuleFileFormat.Space` fields and their `yaml:"space"` tags to `Ordinal`/`yaml:"ordinal"`, and update `validatePins`/`validateDefaultSpaces`'s error messages from "space must be a positive integer, got %d" to "ordinal must be a positive integer, got %d" (and any other message naming the field as "space"). Verify with `go build ./...`.
- [x] 1.2 In `internal/config/config.go`'s `defaultConfigYAML`, update the commented pin/default-space examples' `space: 1` to `ordinal: 1` and the surrounding prose ("a target space (mumu's logical...)") to say "ordinal" instead of "space". Verify by reading the generated comment block.
- [x] 1.3 Update `internal/config/config_test.go`: rename all `PinRule{Space: ...}`/`DefaultSpaceRule{Space: ...}` construction and `.Space` field reads to `.Ordinal`, and update any YAML fixture strings using `space:` to `ordinal:`. Verify with `go test ./internal/config/...`.

## 2. Update consumers of the renamed fields

- [x] 2.1 In `internal/layout/pins.go`, update `pinEntriesByBundle`'s `pin.Space` and `defaultSpacesByBundle`'s `rule.Space` field accesses to `.Ordinal`. Verify with `go build ./...`.
- [x] 2.2 In `internal/layout/sort.go`, update `pinRuleLess` and `defaultSpaceRuleLess`'s `ruleA.Space`/`ruleB.Space` field accesses to `.Ordinal`. Verify with `go build ./...`.
- [x] 2.3 In `cmd/mumu/cmd/layout.go`, update `printConfiguredPins`'s `pin.Space` and `printConfiguredDefaultSpaces`'s `rule.Space` field accesses (passed to `space.DualLabel`) to `.Ordinal`. Verify with `go build ./...`.
- [x] 2.4 Update `internal/layout/pins_test.go`, `internal/layout/default_spaces_test.go`, `internal/layout/sort_test.go`, and `cmd/mumu/cmd/layout_test.go`: rename all `config.PinRule{Space: ...}`/`config.DefaultSpaceRule{Space: ...}` construction and `.Space` field reads to `.Ordinal`. Verify with `go test ./internal/layout/... ./cmd/mumu/cmd/...`.

## 3. Documentation

- [x] 3.1 In `docs/CONFIG_SCHEMA.md`, rename the "Pin rule object" and "Default-space rule object" tables' `space` row to `ordinal` (keep the existing "target logical left-to-right Space number" description), update the `pins`/`default_spaces` example YAML blocks' `space: N` to `ordinal: N`, and fix the "Entry object" table's `ordinal` row description to drop the incorrect "Mission Control" wording (it should read as mumu's own logical left-to-right Space number only, not conflate it with the macOS/Mission Control number). Verify by re-reading the file for any remaining `space:` example under `pins`/`default_spaces` or any description naming `ordinal` as a Mission Control number.
- [x] 3.2 Re-read `docs/CLI.md`'s Pinned Windows and Default Spaces sections; confirm no prose or example needs updating (no literal `space:` config key appears there today) and leave a note in the PR/commit if none is needed. Confirmed: no literal `space:` key or `ordinal`-vs-`space` prose appears in `docs/CLI.md`; no edit needed.

## 4. Verification

- [x] 4.1 Run `just fmt`, `just lint`, `just build`, and `just test`; confirm all four pass with no new warnings.
- [x] 4.2 Manually edit a throwaway `config.yaml` to use the old `space:` key under `pins` or `default_spaces`, run a `mumu` command that loads config, and confirm it fails with a clear "ordinal must be a positive integer, got 0" error naming the offending app/display count (per the spec's "old space key" scenario) rather than silently misbehaving.
- [x] 4.3 Update the user's own `~/.xdg/config/mumu/config.yaml` (outside the repo): rename every `space:` key under `pins` and `default_spaces` to `ordinal:`, then run `mumu show` and confirm it loads without error and prints the same pins/default-spaces list as before.
