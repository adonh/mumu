## Purpose

Gives mumu an explicit, user-editable YAML settings file so a user can discover and control where mumu stores its data, instead of that location being an invisible hardcoded path.

## ADDED Requirements

### Requirement: Config file location

The system SHALL resolve its configuration file path as `$XDG_CONFIG_HOME/mumu/config.yaml` when the `XDG_CONFIG_HOME` environment variable is set to a non-empty value, and otherwise as `~/Library/Application Support/mumu/config.yaml`.

#### Scenario: XDG_CONFIG_HOME is set

- **WHEN** the `XDG_CONFIG_HOME` environment variable is set to `/custom/config`
- **THEN** the system resolves its configuration file path as `/custom/config/mumu/config.yaml`

#### Scenario: XDG_CONFIG_HOME is not set

- **WHEN** the `XDG_CONFIG_HOME` environment variable is unset or empty
- **THEN** the system resolves its configuration file path as `~/Library/Application Support/mumu/config.yaml`

### Requirement: Config file is auto-created with discoverable defaults

If the resolved configuration file does not exist when the system first needs configuration, the system SHALL create it (and any missing parent directories) containing the default settings and explanatory comments for each setting, before proceeding with the requested operation.

#### Scenario: First run with no existing config file

- **WHEN** a user runs any `mumu` command that reads configuration and no configuration file exists yet at the resolved path
- **THEN** the system creates the configuration file at the resolved path with commented default settings, and proceeds using those defaults

#### Scenario: Existing config file is left untouched

- **WHEN** a user runs a `mumu` command and a configuration file already exists at the resolved path
- **THEN** the system reads that file as-is and does not overwrite or regenerate it

### Requirement: Config file format

The configuration file SHALL be valid YAML understood by a human editor without external documentation, and SHALL support at minimum a `data_dir` setting: the directory mumu uses to store its data (see the `space-layout` capability for what that data contains). Values in `data_dir` SHALL support a leading `~` expanding to the user's home directory.

#### Scenario: Setting a custom data directory

- **WHEN** a user edits the configuration file's `data_dir` setting to `~/mumu-data` and saves the file
- **THEN** subsequent `mumu` commands use `~/mumu-data` (expanded to an absolute path) as the data directory

#### Scenario: Default data directory when unset

- **WHEN** the configuration file's `data_dir` setting is absent or the configuration file was just auto-created with defaults
- **THEN** the system uses `$XDG_DATA_HOME/mumu` as the data directory if `XDG_DATA_HOME` is set, and otherwise `~/Library/Application Support/mumu`

### Requirement: Invalid config file is reported clearly

If the configuration file exists but cannot be parsed as valid YAML, or contains a `data_dir` value that is not a non-empty string, the system SHALL report a clear error identifying the configuration file path and SHALL make no changes to any window, Space, or saved layout.

#### Scenario: Malformed YAML

- **WHEN** a user runs a `mumu` command and the configuration file contains invalid YAML syntax
- **THEN** the system reports an error naming the configuration file path and makes no changes

#### Scenario: Invalid data_dir value

- **WHEN** a user runs a `mumu` command and the configuration file's `data_dir` setting is present but empty or not a string
- **THEN** the system reports a clear error and makes no changes
