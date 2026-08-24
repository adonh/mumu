## Why

`config.yaml`'s `pins` and `default_spaces` rules use a field literally named `space:` for their target, but that field actually holds mumu's own logical left-to-right ordinal (the bare `#NN` number `mumu show`/`restore` print) — not the macOS Mission Control Space number that `mumu show`'s output labels `(space MM)`. A user reading `space: 4` in their config naturally assumes it means the same "space" as the output's `(space MM)` parenthetical, but it actually means the output's unlabeled `#04`. This is the opposite of what the field name implies, and it's the same "reused word, different meaning" problem `clarify-sort-terminology` fixed for `--sort display`, just hiding in the config schema instead of the CLI flag. `docs/CONFIG_SCHEMA.md` compounds this by describing the saved-layout JSON's `ordinal` field as the "logical left-to-right **Mission Control** Space number" — a self-contradictory phrase naming both "logical" and "Mission Control" for what is actually only the logical number.

## What Changes

- Adopt one fixed vocabulary everywhere a Space number appears: **`ordinal`** names mumu's own logical left-to-right number (today's bare `#NN`); **`space`** stays reserved for the macOS Mission Control Space concept/number (today's `(space MM)`). `mumu show`/`restore`'s dual-label output format itself is unchanged (`#NN (space MM)`), since it already matches this vocabulary once "ordinal" and "space" are the only two words in play.
- **BREAKING**: rename `config.yaml`'s `pins[].space` and `default_spaces[].space` keys to `pins[].ordinal` and `default_spaces[].ordinal`. The value and semantics are unchanged (still the target logical left-to-right Space number) — only the YAML key name changes, so it matches the saved-layout JSON's existing `ordinal` field instead of colliding with the macOS-Space meaning of "space" used in `mumu show`'s output.
- Rename the corresponding Go identifiers: `config.PinRule.Space` → `config.PinRule.Ordinal`, `config.DefaultSpaceRule.Space` → `config.DefaultSpaceRule.Ordinal` (and their on-disk `yaml:"space"` tags → `yaml:"ordinal"`).
- Update `config.yaml`'s validation error messages (currently "space must be a positive integer, got %d") to name `ordinal` instead of `space`.
- Update `docs/CONFIG_SCHEMA.md`, `docs/CLI.md`, and `config.yaml`'s auto-generated comments/examples to use `ordinal:` instead of `space:` for pin and default-space rules, and fix the saved-layout `ordinal` field's description to drop the incorrect "Mission Control" wording (it is the logical ordinal only).
- No change to the saved-layout JSON schema itself — its `ordinal` field is already named correctly and needs no rename, confirming internal (JSON) and external (YAML) config now use the same vocabulary for the same concept, since a user may hand-edit either file directly.
- No change to `mumu show`/`restore`'s printed output format, to `--sort`'s accepted values, or to which windows get matched, moved, or skipped.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `configuration`: the "Config file format" and "Invalid config file is reported clearly" requirements' `pins`/`default_spaces` rule shape changes from a `space` key to an `ordinal` key (same semantics: mumu's own logical left-to-right Space number).

## Impact

- `internal/config/config.go`: rename `PinRule.Space` → `PinRule.Ordinal`, `DefaultSpaceRule.Space` → `DefaultSpaceRule.Ordinal`, their `yaml:"space"` tags → `yaml:"ordinal"`, validation error text, and `defaultConfigYAML`'s example comments.
- `internal/config/config_test.go`: rename identifiers/expectations accordingly.
- `internal/layout/pins.go`: update `pin.Space`/`rule.Space` field accesses to `.Ordinal`.
- `internal/layout/pins_test.go`, `internal/layout/default_spaces_test.go`: rename identifiers/expectations accordingly.
- `internal/layout/sort.go`: update `pinRuleLess`/`defaultSpaceRuleLess`'s `ruleA.Space`/`ruleB.Space` field accesses to `.Ordinal`.
- `internal/layout/sort_test.go`: rename identifiers/expectations accordingly.
- `cmd/mumu/cmd/layout.go`: update `pin.Space`/`rule.Space` field accesses (passed to `space.DualLabel`) to `.Ordinal`.
- `cmd/mumu/cmd/layout_test.go`: rename identifiers/expectations accordingly.
- `docs/CONFIG_SCHEMA.md`: rename the pin-rule and default-space-rule object tables' `space` row to `ordinal`, update the example YAML blocks, and fix the saved-layout `ordinal` field's description.
- `docs/CLI.md`: no literal `space:` config examples exist there today, but re-read for any prose that would now read inconsistently.
- The user's own `~/.xdg/config/mumu/config.yaml` (outside the repo) needs the same `space:` → `ordinal:` rename under `pins` and `default_spaces` to keep loading correctly; this is done alongside the repo change as a courtesy, not tracked as a repo task.
