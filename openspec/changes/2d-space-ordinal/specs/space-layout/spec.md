## MODIFIED Requirements

### Requirement: Logical two-part Space numbering

The system SHALL compute a logical Space ordinal, scoped to this capability only, as two parts: a display ordinal and a space-within-display ordinal. The display ordinal SHALL be assigned by sorting all connected displays by physical left-to-right position, numbering them 1-based left to right, regardless of which display is the primary (menu-bar) display. The space-within-display ordinal SHALL be assigned by numbering each display's own Spaces 1-based in that display's own native order, independent of any other display's Spaces or Space count. This numbering SHALL NOT alter the underlying macOS Mission Control Space ordinal (primary-first, flat across all displays), which the system separately resolves and displays alongside the logical ordinal.

Because the space-within-display ordinal is scoped to a single display, adding or removing a Space on one display SHALL NOT change the logical ordinal of any Space on a different display.

#### Scenario: Primary display is not the leftmost display

- **WHEN** the connected displays are physically arranged left to right as [Display B, Display A (primary), Display C], each with its own Spaces
- **THEN** the logical display ordinals number Display B as 1, Display A as 2, and Display C as 3, regardless of Display A's primary status, and each display's Spaces are numbered 1-based independently within that display

#### Scenario: Existing space commands are unaffected

- **WHEN** the system resolves the macOS Mission Control Space ordinal for display alongside the logical ordinal
- **THEN** that resolution uses macOS's own primary-first flat ordering, unchanged by the logical two-part numbering

#### Scenario: Adding a Space on one display does not renumber another display's Spaces

- **WHEN** a display's logical ordinal is 2 and a user adds a new Mission Control Space on a different display (logical ordinal 1)
- **THEN** every Space's logical ordinal on display 2 remains unchanged, since the space-within-display part is scoped to display 2 alone

#### Scenario: Removing a Space on one display does not renumber another display's Spaces

- **WHEN** a display's logical ordinal is 2 and a user removes a Mission Control Space from a different display (logical ordinal 1)
- **THEN** every Space's logical ordinal on display 2 remains unchanged, since the space-within-display part is scoped to display 2 alone

### Requirement: Restore never creates or deletes Spaces

If a saved layout's logical Space ordinal references a display ordinal that exceeds the number of displays currently connected, or a space-within-display ordinal that exceeds that specific display's currently available Space count, the system SHALL NOT create a new Space or a new display. It SHALL place window entries that fit within the currently available Spaces on their target display and SHALL report entries that could not be placed due to an out-of-range display ordinal or an out-of-range space-within-display ordinal.

#### Scenario: Saved layout references a display that no longer exists

- **WHEN** a saved layout references a display ordinal higher than the number of displays currently connected
- **THEN** the system does not create a new display, moves the windows that fit within currently connected displays, and reports the entries it could not place

#### Scenario: Saved layout requires more Spaces than currently exist

- **WHEN** a saved layout references a space-within-display ordinal higher than the number of Spaces currently available on the corresponding display, but that display itself still exists
- **THEN** the system does not create a new Space, moves the windows that fit within that display's existing Spaces, and reports the entries it could not place

#### Scenario: An out-of-range ordinal on one display does not affect placement on other displays

- **WHEN** a saved layout has one entry whose space-within-display ordinal is out of range for its target display, and other entries whose ordinals are in range on different displays
- **THEN** the system places the in-range entries normally and reports only the out-of-range entry as unplaceable

### Requirement: Saved layouts persisted as per-display-count JSON files

Each saved layout SHALL be persisted as its own JSON file, named `<display-count>.json`, within a `layouts` subdirectory of the data directory resolved per the `configuration` capability's `data_dir` setting. Saved layout files are internal state, not a user-facing editable file: the system provides no guarantee that hand-edits to a saved layout file are preserved or honored correctly, and does not document their structure as something a user should edit directly.

Each saved layout file SHALL record the schema version it was written with. If a `mumu` command reads a saved layout file whose recorded schema version does not match the version the running `mumu` writes, the system SHALL report a clear error identifying the file path and stating that the layout was saved by an incompatible mumu version and must be recreated with `mumu save`, rather than surfacing a raw parse error, and SHALL make no changes to any window, Space, or saved layout.

#### Scenario: Each display count has its own file

- **WHEN** a user has saved layouts for both 2 displays and 3 displays
- **THEN** two separate JSON files exist in the `layouts` subdirectory, one named for each display count

#### Scenario: Saved layout files follow the configured data directory

- **WHEN** the `configuration` capability's `data_dir` setting resolves to a given directory
- **THEN** the system reads and writes saved-layout JSON files within a `layouts` subdirectory of that directory

#### Scenario: Malformed saved-layout file is reported clearly

- **WHEN** a user runs a `mumu` command that reads a saved layout and that display count's JSON file exists but cannot be parsed as valid JSON
- **THEN** the system reports a clear error identifying the file path and makes no changes to any window, Space, or saved layout

#### Scenario: Saved layout from an incompatible schema version is reported clearly

- **WHEN** a user runs a `mumu` command that reads a saved layout whose recorded schema version does not match the running `mumu`'s schema version (for example, a layout saved before the logical ordinal became two-part)
- **THEN** the system reports a clear error identifying the file path and instructing the user to run `mumu save` again, rather than a raw parse error, and makes no changes to any window, Space, or saved layout
