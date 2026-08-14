## MODIFIED Requirements

### Requirement: Restore matches windows by title, falling back to index

For each saved window entry belonging to a given application, the system SHALL first attempt to match it to a currently open window of that application by exact title. If no unambiguous title match exists, the system SHALL fall back to matching by the window's positional index within that application, using the same deterministic window ordering used at save time. If neither yields a match and exactly one of the application's currently open windows remains unclaimed, the system SHALL match the entry to that window regardless of title or position.

After completing those matching steps for an application, the system SHALL place every remaining unclaimed currently open window of that application only when the application has at least one valid saved-entry assignment. The fallback target SHALL be the logical Space ordinal occurring most often among that application's valid saved-entry assignments. If two or more logical Space ordinals tie for most prevalent, the fallback target SHALL instead be the logical Space currently displayed on the primary (menu-bar) display. The system SHALL report each fallback placement in restore progress output and SHALL move each currently open window at most once. If the application has no valid saved-entry assignment, the system SHALL leave its remaining unclaimed open windows unchanged.

#### Scenario: Exact title match

- **WHEN** a saved entry for an application has a window title that exactly matches one of that application's currently open windows
- **THEN** that window is selected for the move

#### Scenario: Fallback to index when titles differ

- **WHEN** a saved entry's window title does not match any of that application's currently open window titles
- **THEN** the system matches by the entry's positional index among that application's currently open windows

#### Scenario: Fallback to sole remaining window when title and index both fail

- **WHEN** a saved entry's window title does not match any currently open window and its saved positional index is unavailable, but exactly one of the application's currently open windows remains unclaimed by any other entry
- **THEN** the system matches the entry to that sole remaining window regardless of its title or position

#### Scenario: Fallback to an application's prevalent assigned Space

- **WHEN** one open Chrome window has a valid saved-entry assignment to logical Space 4 and another open Chrome window remains unclaimed after title, index, and sole-remaining-candidate matching
- **THEN** the system moves the unclaimed Chrome window to logical Space 4 and reports that placement in restore progress output

#### Scenario: Most prevalent Space wins

- **WHEN** an application's valid saved-entry assignments target logical Spaces 2, 2, and 5, and one of its currently open windows remains unclaimed after standard matching
- **THEN** the system moves the unclaimed window to logical Space 2

#### Scenario: Tied prevalent Spaces use the primary display's current Space

- **WHEN** an application's valid saved-entry assignments are evenly split between logical Spaces 2 and 5, one of its currently open windows remains unclaimed after standard matching, and the primary display currently shows logical Space 7
- **THEN** the system moves the unclaimed window to logical Space 7 and reports that placement in restore progress output

#### Scenario: No valid assignment leaves unmatched windows unchanged

- **WHEN** an application's currently open windows cannot be matched to any valid saved-entry assignment
- **THEN** the system does not move that application's remaining unclaimed windows through the application-level fallback
