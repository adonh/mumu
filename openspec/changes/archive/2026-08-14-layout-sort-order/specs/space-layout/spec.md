## ADDED Requirements

### Requirement: Configurable output ordering for layout entries

`mimi layout show` and `mimi layout restore` SHALL accept a `--sort` flag with values `display` (default), `macos`, and `app`, controlling the order in which per-window entries are printed: `display` orders by the logical left-to-right Space ordinal ascending; `macos` orders by the macOS Mission Control Space ordinal ascending (the same numbering `mimi action space <n>` uses); `app` orders by application bundle identifier ascending. This ordering SHALL apply to `mimi layout show`'s entry listing, `mimi layout restore`'s per-window move progress lines, and the ordering of entries within each reason group of `mimi layout restore`'s skip summary. Regardless of which key is selected as primary, entries with an equal primary-key value SHALL be ordered by cascading through the remaining keys in the fixed priority: Space ordinal, then bundle identifier, then window title. This SHALL NOT change which windows are matched, moved, or skipped, nor anything persisted in a saved layout file — it only affects display order.

#### Scenario: Default order is display sequence

- **WHEN** a user runs `mimi layout show` or `mimi layout restore` without passing `--sort`
- **THEN** entries are printed ordered by logical left-to-right Space ordinal ascending

#### Scenario: Sorting by macOS Mission Control Space number

- **WHEN** a user passes `--sort macos` to `mimi layout show` or `mimi layout restore`
- **THEN** entries are printed ordered by their current macOS Mission Control Space ordinal ascending, regardless of their logical Space ordinal

#### Scenario: Sorting by application

- **WHEN** a user passes `--sort app` to `mimi layout show` or `mimi layout restore`
- **THEN** entries are printed ordered by application bundle identifier ascending, grouping all of one application's windows together

#### Scenario: Tie-break when the primary sort key is equal

- **WHEN** two or more entries share the same value for the selected `--sort` key (e.g. two windows on the same Space when sorting by `display`)
- **THEN** those entries are ordered relative to each other by Space ordinal, then bundle identifier, then window title, in that order

#### Scenario: Sort order applies within restore's skip summary

- **WHEN** `mimi layout restore` reports multiple skipped entries sharing the same skip reason
- **THEN** those entries are listed within that reason group ordered according to the selected `--sort` key
