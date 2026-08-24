## MODIFIED Requirements

### Requirement: Configurable output ordering for layout entries

`mumu show` and `mumu restore` SHALL accept a `--sort` flag with values `logical` (default), `macos`, and `app`, controlling the order in which per-window entries are printed: `logical` orders by mumu's own logical left-to-right Space ordinal ascending; `macos` orders by the macOS Mission Control Space ordinal ascending (the same numbering macOS's own "Switch to Desktop `<n>`" shortcut uses); `app` orders by application bundle identifier ascending. This ordering SHALL apply to `mumu show`'s entry listing, `mumu restore`'s per-window move progress lines, and the ordering of entries within each reason group of `mumu restore`'s skip summary. Regardless of which key is selected as primary, entries with an equal primary-key value SHALL be ordered by cascading through the remaining keys in the fixed priority: Space ordinal, then bundle identifier, then window title. This SHALL NOT change which windows are matched, moved, or skipped, nor anything persisted in a saved layout file — it only affects display order. The word "display" is reserved exclusively for physical-monitor concepts (e.g. the current display count, connected displays, the primary display) and SHALL NOT be used to name this sort key or its default value, to avoid confusion with logical Space ordering.

#### Scenario: Default order is logical Space sequence

- **WHEN** a user runs `mumu show` or `mumu restore` without passing `--sort`
- **THEN** entries are printed ordered by mumu's own logical left-to-right Space ordinal ascending

#### Scenario: Sorting by macOS Mission Control Space number

- **WHEN** a user passes `--sort macos` to `mumu show` or `mumu restore`
- **THEN** entries are printed ordered by their current macOS Mission Control Space ordinal ascending, regardless of their logical Space ordinal

#### Scenario: Sorting by application

- **WHEN** a user passes `--sort app` to `mumu show` or `mumu restore`
- **THEN** entries are printed ordered by application bundle identifier ascending, grouping all of one application's windows together

#### Scenario: Tie-break when the primary sort key is equal

- **WHEN** two or more entries share the same value for the selected `--sort` key (e.g. two windows on the same Space when sorting by `logical`)
- **THEN** those entries are ordered relative to each other by Space ordinal, then bundle identifier, then window title, in that order

#### Scenario: Sort order applies within restore's skip summary

- **WHEN** `mumu restore` reports multiple skipped entries sharing the same skip reason
- **THEN** those entries are listed within that reason group ordered according to the selected `--sort` key

#### Scenario: The old "display" sort value is no longer accepted

- **WHEN** a user passes `--sort display` (the value's pre-rename name) to `mumu show` or `mumu restore`
- **THEN** the system reports a clear error naming the accepted values (`logical`, `macos`, `app`) and makes no changes
