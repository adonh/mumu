## MODIFIED Requirements

### Requirement: Saved layouts persisted as per-display-count JSON files

Each saved layout SHALL be persisted as its own JSON file, named `<display-count>.json`, within a `layouts` subdirectory of the data directory resolved per the `configuration` capability's `data_dir` setting. Saved layout files are internal state, not a user-facing editable file: the system provides no guarantee that hand-edits to a saved layout file are preserved or honored correctly, and does not document their structure as something a user should edit directly.

#### Scenario: Each display count has its own file

- **WHEN** a user has saved layouts for both 2 displays and 3 displays
- **THEN** two separate JSON files exist in the `layouts` subdirectory, one named for each display count

#### Scenario: Saved layout files follow the configured data directory

- **WHEN** the `configuration` capability's `data_dir` setting resolves to a given directory
- **THEN** the system reads and writes saved-layout JSON files within a `layouts` subdirectory of that directory

#### Scenario: Malformed saved-layout file is reported clearly

- **WHEN** a user runs a `mumu` command that reads a saved layout and that display count's JSON file exists but cannot be parsed as valid JSON
- **THEN** the system reports a clear error identifying the file path and makes no changes to any window, Space, or saved layout
