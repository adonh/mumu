## MODIFIED Requirements

### Requirement: Config file format

The configuration file SHALL be valid YAML understood by a human editor without external documentation, and SHALL support at minimum a `data_dir` setting: the directory mumu uses to store its data (see the `space-layout` capability for what that data contains). Values in `data_dir` SHALL support a leading `~` expanding to the user's home directory.

The configuration file SHALL also support a `pins` setting: a mapping from connected-display-count to a list of pin rules (see the `window-pinning` capability), each rule specifying an application bundle identifier, a window-title pattern, and a target `ordinal` (mumu's own two-part `"<display>:<space>"` logical Space number, per the `space-layout` capability's two-part numbering). The configuration file SHALL also support a `pin_precedence` setting with the value `pin` or `layout`, controlling pin-vs-saved-layout precedence during `mumu restore` (see the `window-pinning` capability), defaulting to `pin` when absent.

The configuration file SHALL also support a `default_spaces` setting: a mapping from connected-display-count to a list of default-space rules, each rule specifying an application bundle identifier and a target `ordinal` (mumu's own two-part `"<display>:<space>"` logical Space number) (see the `space-layout` capability's application-level fallback). A display count with no configured `default_spaces` entries SHALL behave as if none are configured.

The configuration file SHALL also support a `hooks` setting (see the `restore-hooks` capability): an `off` command array and an `on` command array applied to every `mumu restore`, plus a `layouts` mapping from connected-display-count to its own `off`/`on` command arrays applied only for that display count. Each command entry SHALL be either a single string or a list of strings (see the `restore-hooks` capability for how each form is executed). Any of `hooks.off`, `hooks.on`, or a given display count's `layouts` entry MAY be absent or empty.

A pin rule's or default-space rule's target field SHALL be named `ordinal`, not `space`, in both the configuration file and any error message describing it: `ordinal` names mumu's own two-part logical Space number, written `"<display>:<space>"` (matching the CLI's own `#D:SS` display convention and the saved-layout JSON file's `ordinal` field), while `space` is reserved exclusively for the macOS Mission Control Space concept, avoiding the collision where the same word would otherwise mean two different numbers depending on context.

#### Scenario: Setting a custom data directory

- **WHEN** a user edits the configuration file's `data_dir` setting to `~/mumu-data` and saves the file
- **THEN** subsequent `mumu` commands use `~/mumu-data` (expanded to an absolute path) as the data directory

#### Scenario: Default data directory when unset

- **WHEN** the configuration file's `data_dir` setting is absent or the configuration file was just auto-created with defaults
- **THEN** the system uses `$XDG_DATA_HOME/mumu` as the data directory if `XDG_DATA_HOME` is set, and otherwise `~/Library/Application Support/mumu`

#### Scenario: Configuring pins for a display count

- **WHEN** a user adds a display-count entry under `pins` in the configuration file, listing one or more rules each with an application bundle identifier, a title pattern, and a target `ordinal` written as `"<display>:<space>"` (e.g. `"2:1"`)
- **THEN** those rules are available to `mumu restore` and `mumu show` for that display count

#### Scenario: Pins setting absent

- **WHEN** the configuration file has no `pins` setting
- **THEN** the system behaves as if no pin rules are configured for any display count, and `mumu restore` proceeds using only saved-layout matching

#### Scenario: Default pin precedence when unset

- **WHEN** the configuration file has no `pin_precedence` setting
- **THEN** the system treats pin rules as taking precedence over saved-layout entries during `mumu restore`

#### Scenario: Configuring a default space for a display count

- **WHEN** a user adds a display-count entry under `default_spaces` in the configuration file, listing one or more rules each with an application bundle identifier and a target `ordinal` written as `"<display>:<space>"` (e.g. `"2:1"`)
- **THEN** those rules are available to `mumu restore` and `mumu show` for that display count

#### Scenario: Default spaces setting absent

- **WHEN** the configuration file has no `default_spaces` setting
- **THEN** the system behaves as if no default-space rule is configured for any application at any display count

#### Scenario: Configuring global hooks

- **WHEN** a user adds one or more commands under `hooks.off` and/or `hooks.on` in the configuration file
- **THEN** those commands are available to run around every `mumu restore`, regardless of display count

#### Scenario: Configuring per-display-count hooks

- **WHEN** a user adds a display-count entry under `hooks.layouts` with its own `off`/`on` command arrays
- **THEN** those commands are available to run around `mumu restore` only when that number of displays is connected

#### Scenario: Hooks setting absent

- **WHEN** the configuration file has no `hooks` setting
- **THEN** the system behaves as if no hook commands are configured, globally or for any display count, and `mumu restore` runs no external commands

#### Scenario: The old "space" key is no longer recognized

- **WHEN** a user's configuration file has a `pins` or `default_spaces` rule using the pre-rename `space:` key instead of `ordinal:`
- **THEN** the system treats that rule as if its `ordinal` were absent (unrecognized key, ignored) and reports a clear validation error naming the offending rule, rather than silently misinterpreting `space:`'s value

#### Scenario: The old bare-integer ordinal format is no longer recognized

- **WHEN** a user's configuration file has a `pins` or `default_spaces` rule whose `ordinal` value is a bare integer (e.g. `4`) rather than a `"<display>:<space>"` string, left over from before the logical ordinal became two-part
- **THEN** the system reports a clear validation error naming the offending rule and stating the expected `"<display>:<space>"` format, rather than guessing which display the bare integer referred to

### Requirement: Invalid config file is reported clearly

If the configuration file exists but cannot be parsed as valid YAML, or contains a `data_dir` value that is not a non-empty string, the system SHALL report a clear error identifying the configuration file path and SHALL make no changes to any window, Space, or saved layout.

If the configuration file's `pins` setting is present but is not a mapping from display count to a list of pin rules, or any pin rule is missing its application bundle identifier, title pattern, or `ordinal`, or its `ordinal` is not a valid `"<display>:<space>"` string with both parts positive integers, the system SHALL report a clear error identifying the configuration file path and the offending entry, naming the field as `ordinal` and stating the expected format, and SHALL make no changes to any window, Space, or saved layout. If the configuration file's `pin_precedence` setting is present but is not `pin` or `layout`, the system SHALL report a clear error identifying the configuration file path, and SHALL make no changes.

If the configuration file's `default_spaces` setting is present but is not a mapping from display count to a list of default-space rules, or any rule is missing its application bundle identifier or `ordinal`, or its `ordinal` is not a valid `"<display>:<space>"` string with both parts positive integers, the system SHALL report a clear error identifying the configuration file path and the offending entry, naming the field as `ordinal` and stating the expected format, and SHALL make no changes to any window, Space, or saved layout.

If the configuration file's `hooks` setting is present but `hooks.off`, `hooks.on`, or any `hooks.layouts.<display_count>.off`/`.on` entry is not a list, or any command entry in one of those lists is neither a non-empty string nor a non-empty list of non-empty strings, the system SHALL report a clear error identifying the configuration file path and the offending entry, and SHALL make no changes to any window, Space, or saved layout.

#### Scenario: Malformed YAML

- **WHEN** a user runs a `mumu` command and the configuration file contains invalid YAML syntax
- **THEN** the system reports an error naming the configuration file path and makes no changes

#### Scenario: Invalid data_dir value

- **WHEN** a user runs a `mumu` command and the configuration file's `data_dir` setting is present but empty or not a string
- **THEN** the system reports a clear error and makes no changes

#### Scenario: Invalid pin rule

- **WHEN** a user runs a `mumu` command and the configuration file's `pins` setting contains a rule missing its title pattern, or with an `ordinal` that isn't a valid `"<display>:<space>"` string, or whose display or space part isn't a positive integer
- **THEN** the system reports a clear error identifying the configuration file path and the offending rule, naming the field as `ordinal` and stating the expected format, and makes no changes

#### Scenario: Invalid pin_precedence value

- **WHEN** a user runs a `mumu` command and the configuration file's `pin_precedence` setting is present but is neither `pin` nor `layout`
- **THEN** the system reports a clear error identifying the configuration file path, and makes no changes

#### Scenario: Invalid default-space rule

- **WHEN** a user runs a `mumu` command and the configuration file's `default_spaces` setting contains a rule missing its application bundle identifier, or with an `ordinal` that isn't a valid `"<display>:<space>"` string, or whose display or space part isn't a positive integer
- **THEN** the system reports a clear error identifying the configuration file path and the offending rule, naming the field as `ordinal` and stating the expected format, and makes no changes

#### Scenario: Invalid hook command entry

- **WHEN** a user runs a `mumu` command and the configuration file's `hooks` setting contains a command entry that is neither a non-empty string nor a non-empty list of non-empty strings
- **THEN** the system reports a clear error identifying the configuration file path and the offending entry, and makes no changes

#### Scenario: Hooks setting is not the expected shape

- **WHEN** a user runs a `mumu` command and the configuration file's `hooks.off`, `hooks.on`, or a `hooks.layouts` display-count entry's `off`/`on` is present but is not a list
- **THEN** the system reports a clear error identifying the configuration file path, and makes no changes
