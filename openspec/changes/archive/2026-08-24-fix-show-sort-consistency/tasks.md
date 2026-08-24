## 1. Sorting helpers

- [x] 1.1 Add a `SortPinRules(pins []config.PinRule, key SortKey)` helper to `internal/layout/sort.go` that orders by the same cascade as `entryLess` (Space ordinal, then bundle identifier, then title), reusing `SortByMacOS`'s Mission-Control lookup for the `macos` key. Verify with a unit test asserting `display`, `macos`, and `app` all produce the expected order, including a tie broken by bundle identifier then title.
- [x] 1.2 Add a `SortDefaultSpaceRules(rules []config.DefaultSpaceRule, key SortKey)` helper mirroring 1.1 but without a title component in the tie-break (Space ordinal, then bundle identifier only). Verify with a unit test covering all three sort keys and a bundle-identifier tie-break.

## 2. Wire sorting into `mumu show`

- [x] 2.1 Update `printConfiguredPins` in `cmd/mumu/cmd/layout.go` to accept the resolved `sortKey` and sort a copy of the input slice via `layout.SortPinRules` before printing (do not mutate `cfg.Pins[displayCount]`). Verify by updating `TestPrintConfiguredPins_ListsEachRule` to pass pins in a scrambled order and assert the printed order matches the sort key.
- [x] 2.2 Update `printConfiguredDefaultSpaces` the same way using `layout.SortDefaultSpaceRules`. Verify by updating `TestPrintConfiguredDefaultSpaces_ListsEachRule` similarly.
- [x] 2.3 Update `layoutShowCmd`'s `RunE` to pass `sortKey` to both `printConfiguredPins` and `printConfiguredDefaultSpaces`. Verify with `just build` and a CLI-level test asserting `mumu show --sort app` orders the configured-pins section by bundle identifier when the config lists them out of order.

## 3. Verification

- [x] 3.1 Run `just fmt`, `just lint`, `just build`, and `just test`; confirm all four pass with no new warnings.
- [x] 3.2 Manually run `mumu show` (and `mumu show --sort app`, `mumu show --sort macos`) against a `config.yaml` with pins and `default_spaces` listed out of Space order, and confirm every section (entries, pins, default spaces) reads in a single coherent order for the chosen `--sort` key.
