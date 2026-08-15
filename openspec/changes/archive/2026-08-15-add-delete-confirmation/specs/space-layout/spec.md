## MODIFIED Requirements

### Requirement: Layout management commands

`mumu list` SHALL show all saved layouts along with the display count each is keyed to. `mumu show` SHALL display the contents of a saved layout without applying it. `mumu delete` SHALL remove a saved layout for the current (or explicitly specified) display count, but only after the user confirms the deletion; passing `--yes` (or `-y`) SHALL skip the confirmation prompt.

#### Scenario: Listing saved layouts

- **WHEN** a user runs `mumu list`
- **THEN** the system lists each saved layout with the display count it was saved for

#### Scenario: Previewing a saved layout

- **WHEN** a user runs `mumu show` for a display count that has a saved layout
- **THEN** the system displays that layout's window entries without moving any windows

#### Scenario: Deleting a saved layout

- **WHEN** a user runs `mumu delete` for a display count that has a saved layout and confirms the prompt
- **THEN** the system removes that saved layout and it no longer appears in `mumu list`

#### Scenario: Declining the delete confirmation prompt

- **WHEN** a user runs `mumu delete` for a display count that has a saved layout and does not confirm the prompt
- **THEN** the system leaves the saved layout untouched, reports that the deletion was aborted, and exits without error

#### Scenario: Skipping the delete confirmation with --yes

- **WHEN** a user runs `mumu delete --yes` (or `-y`) for a display count that has a saved layout
- **THEN** the system removes that saved layout immediately without prompting

#### Scenario: Deleting a display count with no saved layout

- **WHEN** a user runs `mumu delete` for a display count that has no saved layout
- **THEN** the system reports that no saved layout exists for that display count and does not prompt for confirmation
