## ADDED Requirements

### Requirement: Saved layouts persisted as a single editable YAML file

All saved layouts SHALL be persisted together in a single YAML file (rather than one file per display count), keyed by display count, located in the data directory resolved per the `configuration` capability's `data_dir` setting. This file SHALL be valid YAML that a user can open and edit directly.

#### Scenario: Multiple display-count layouts share one file

- **WHEN** a user has saved layouts for both 2 displays and 3 displays
- **THEN** both saved layouts exist as entries within the same YAML file, keyed by their respective display counts

#### Scenario: Saved layout file location follows the configured data directory

- **WHEN** the `configuration` capability's `data_dir` setting resolves to a given directory
- **THEN** the system reads and writes the saved-layouts YAML file within that directory

#### Scenario: A user can hand-edit a saved layout

- **WHEN** a user opens the saved-layouts YAML file in a text editor and makes a valid edit (e.g. correcting a window title) and saves it
- **THEN** a subsequent `mumu show`, `mumu list`, or `mumu restore` reflects the edited content

#### Scenario: Malformed saved-layouts file is reported clearly

- **WHEN** a user runs a `mumu` command that reads saved layouts and the saved-layouts YAML file exists but cannot be parsed as valid YAML
- **THEN** the system reports a clear error identifying the file path and makes no changes to any window, Space, or saved layout
