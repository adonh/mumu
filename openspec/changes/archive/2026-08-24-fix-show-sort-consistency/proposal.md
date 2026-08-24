## Why

`mumu show`'s saved-layout entry list respects the `--sort` flag (`display`, `macos`, `app`), but the configured pin rules, configured `default_spaces` rules, and configured hook commands printed below it are always listed in whatever order they happen to appear in `config.yaml` (YAML map/list iteration order) — never sorted by Space, application, or anything else. Next to a coherently sorted entry list, these unsorted sections read as scattered and make it hard to scan a display count's full configuration at a glance, especially as the number of pins/default-spaces grows.

## What Changes

- `mumu show`'s "configured pin(s)" section is sorted using the same `--sort` key (and the same Space → bundle ID → title tie-break cascade) as the saved-layout entry list above it, instead of raw config-file order.
- `mumu show`'s "configured default space(s)" section is sorted the same way (Space → bundle ID, since there's no title).
- No change to `mumu restore`'s behavior, to what's persisted in `config.yaml`, or to which pins/defaults apply — this is purely a display-order fix for `mumu show`.
- Hook command previews are unaffected: `hooks.off`/`hooks.on` are inherently ordered (they execute in the order listed), so reordering them would be incorrect and is out of scope.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `space-layout`: the "Configurable output ordering for layout entries" requirement's scope is extended so `--sort` also governs the order of `mumu show`'s configured-pins and configured-default-spaces sections, not just the saved-layout entry list.

## Impact

- `cmd/mumu/cmd/layout.go`: `printConfiguredPins` and `printConfiguredDefaultSpaces` gain sorting before printing; `layoutShowCmd`'s `RunE` passes the resolved `sortKey` to both.
- `internal/layout/sort.go`: may need small additions (e.g. a title-less comparison path, or reusing `entryLess` with an empty title) to sort `config.PinRule`/`config.DefaultSpaceRule` values, which aren't `layout.Entry`.
- Tests in `cmd/mumu/cmd/layout_test.go` covering the current (unsorted) preview output will need updating.
