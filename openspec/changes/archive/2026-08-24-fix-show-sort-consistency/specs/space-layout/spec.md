## MODIFIED Requirements

### Requirement: Configurable output ordering for layout entries

`mumu show` and `mumu restore` SHALL accept a `--sort` flag with values `display` (default), `macos`, and `app`, controlling the order in which per-window entries are printed: `display` orders by the logical left-to-right Space ordinal ascending; `macos` orders by the macOS Mission Control Space ordinal ascending (the same numbering macOS's own "Switch to Desktop `<n>`" shortcut uses); `app` orders by application bundle identifier ascending. This ordering SHALL apply to `mumu show`'s entry listing, `mumu show`'s configured-pins and configured-default-spaces previews, `mumu restore`'s per-window move progress lines, and the ordering of entries within each reason group of `mumu restore`'s skip summary. Regardless of which key is selected as primary, entries with an equal primary-key value SHALL be ordered by cascading through the remaining keys in the fixed priority: Space ordinal, then bundle identifier, then window title (pin and default-space rules have no title, so they cascade only through Space ordinal then bundle identifier). This SHALL NOT change which windows are matched, moved, or skipped, which pin or default-space rule applies, nor anything persisted in a saved layout file or `config.yaml` — it only affects display order.

#### Scenario: Default order is display sequence

- **WHEN** a user runs `mumu show` or `mumu restore` without passing `--sort`
- **THEN** entries are printed ordered by logical left-to-right Space ordinal ascending

#### Scenario: Sorting by macOS Mission Control Space number

- **WHEN** a user passes `--sort macos` to `mumu show` or `mumu restore`
- **THEN** entries are printed ordered by their current macOS Mission Control Space ordinal ascending, regardless of their logical Space ordinal

#### Scenario: Sorting by application

- **WHEN** a user passes `--sort app` to `mumu show` or `mumu restore`
- **THEN** entries are printed ordered by application bundle identifier ascending, grouping all of one application's windows together

#### Scenario: Tie-break when the primary sort key is equal

- **WHEN** two or more entries share the same value for the selected `--sort` key (e.g. two windows on the same Space when sorting by `display`)
- **THEN** those entries are ordered relative to each other by Space ordinal, then bundle identifier, then window title, in that order

#### Scenario: Sort order applies within restore's skip summary

- **WHEN** `mumu restore` reports multiple skipped entries sharing the same skip reason
- **THEN** those entries are listed within that reason group ordered according to the selected `--sort` key

#### Scenario: Sort order applies to mumu show's configured-pins preview

- **WHEN** a user runs `mumu show --sort app` for a display count with multiple configured pin rules
- **THEN** the "configured pin(s)" section lists those rules ordered by application bundle identifier ascending, rather than their order in `config.yaml`

#### Scenario: Sort order applies to mumu show's configured-default-spaces preview

- **WHEN** a user runs `mumu show` (default `--sort display`) for a display count with multiple configured `default_spaces` rules
- **THEN** the "configured default space(s)" section lists those rules ordered by target Space ordinal ascending, rather than their order in `config.yaml`

#### Scenario: Hook command previews remain unsorted

- **WHEN** a user runs `mumu show` for a display count with configured `hooks.off`/`hooks.on` commands
- **THEN** those commands are still listed in their configured execution order, unaffected by `--sort`
