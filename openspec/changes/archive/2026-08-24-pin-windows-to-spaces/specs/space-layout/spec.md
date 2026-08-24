## MODIFIED Requirements

### Requirement: Restore matches windows by title, falling back to index

For each application with saved window entries, the system SHALL match its saved entries to that application's currently open windows as a single batch: for every remaining saved entry and every remaining currently open window not yet claimed by another saved entry, the system SHALL measure title similarity as the proportion of shared words to total distinct words between the saved title and the window's title, comparing words case-insensitively and independent of word order. The system SHALL then assign entry-window pairs starting from the highest similarity score downward, skipping any entry or window already assigned, until no further pair can be assigned. When two or more candidate windows tie for an entry's highest score (or two or more entries tie for a window's highest score), the system SHALL prefer the candidate whose current position among the application's open windows equals the entry's saved positional index; if that does not resolve the tie, the system SHALL deterministically choose one of the tied candidates rather than leaving the entry unmatched. A saved entry SHALL remain unmatched only when no currently open window of that application remains unclaimed for it.

After completing those matching steps for an application, the system SHALL place every remaining unclaimed currently open window of that application only when the application has at least one valid saved-entry assignment. The fallback target SHALL be the logical Space ordinal occurring most often among that application's valid saved-entry assignments. If two or more logical Space ordinals tie for most prevalent, the fallback target SHALL instead be the logical Space currently displayed on the primary (menu-bar) display. The system SHALL report each fallback placement in restore progress output and SHALL move each currently open window at most once. If the application has no valid saved-entry assignment, the system SHALL leave its remaining unclaimed open windows unchanged.

When the `window-pinning` capability's pin rules are configured to take precedence (the default), any currently open window already claimed by a matched pin rule SHALL be excluded from this saved-layout matching entirely, as if it were not currently open. When saved-layout entries are configured to take precedence instead, this saved-layout matching (including its application-level fallback placement) SHALL run unaffected by pin rules, and pin matching SHALL only run afterward against windows this process left unclaimed.

#### Scenario: Exact title match

- **WHEN** a saved entry's title exactly matches exactly one of that application's currently open window titles, and no other open window shares that same set of words
- **THEN** that window is selected for the move, and the placement is not marked approximate in output

#### Scenario: Titles with the same words in a different order or case match fully

- **WHEN** a saved entry's title and one of the application's currently open window titles contain exactly the same words, regardless of word order or letter case
- **THEN** the system matches the entry to that window and does not mark the placement as approximate in output

#### Scenario: Highest word-overlap match wins when no other candidate scores higher

- **WHEN** a saved entry's title shares more words with one of the application's currently open window titles than with any other currently open window title
- **THEN** the system matches the entry to that highest-scoring window

#### Scenario: One window is not claimed as the best match for more than one saved entry

- **WHEN** two saved entries for the same application would each independently score highest against the same currently open window, and each of those entries also scores highest, though lower, against a different one of the application's other currently open windows
- **THEN** the system assigns at most one of the two entries to the contested window and assigns the other entry to a different currently open window, so no currently open window is assigned to more than one saved entry

#### Scenario: Fallback to sole remaining window when title and index both fail

- **WHEN** a saved entry's title shares no words with any currently open window title and its saved positional index does not identify an unclaimed window, but exactly one of that application's currently open windows remains unclaimed by any other entry
- **THEN** the system matches the entry to that sole remaining window

#### Scenario: An entry matches even when no candidate title is similar

- **WHEN** a saved entry's title shares no words with any of the application's currently open window titles, and at least one of those windows remains unclaimed by any other entry
- **THEN** the system matches the entry to one of the remaining unclaimed windows rather than leaving it unmatched

#### Scenario: Fallback to index when titles differ

- **WHEN** a saved entry's title shares no words with any of that application's currently open window titles, so every remaining candidate ties for that entry's highest similarity score, and one of those tied candidates' current position matches the entry's saved positional index
- **THEN** the system matches the entry to the window at that position

#### Scenario: Tied similarity scores are broken by the saved positional index

- **WHEN** two or more of an application's currently open windows tie for a saved entry's highest similarity score, and one of those tied windows' current position matches the entry's saved positional index
- **THEN** the system matches the entry to the window at that position

#### Scenario: Tied similarity scores with no matching index are still resolved

- **WHEN** two or more of an application's currently open windows tie for a saved entry's highest similarity score, and none of the tied windows' positions match the entry's saved positional index
- **THEN** the system deterministically matches the entry to one of the tied windows rather than leaving it unmatched

#### Scenario: Entries are unmatched only when open windows run out

- **WHEN** an application has more saved entries than currently open windows
- **THEN** the system matches every currently open window to some saved entry and reports the remaining saved entries as unmatched

#### Scenario: Approximate matches are marked in restore output

- **WHEN** a saved entry is matched to a currently open window whose title does not contain exactly the same words as the saved title
- **THEN** the progress line or skip-summary entry for that window is marked to indicate the match was approximate

#### Scenario: Fallback to an application's prevalent assigned Space

- **WHEN** one open Chrome window has a valid saved-entry assignment to logical Space 4 and another open Chrome window remains unclaimed after batch matching
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

#### Scenario: Windows claimed by a higher-precedence pin are excluded from saved-layout matching

- **WHEN** pin rules take precedence (the default) and a currently open window has already been claimed by a matched pin rule
- **THEN** that window is not considered by saved-layout matching or its application-level fallback, as if it were not currently open

### Requirement: Layout management commands

`mumu list` SHALL show all saved layouts along with the display count each is keyed to. `mumu show` SHALL display the contents of a saved layout, plus that display count's configured pin rules (see the `window-pinning` capability) and effective hook-command preview (see the `restore-hooks` capability), without applying any of them. `mumu delete` SHALL remove a saved layout for the current (or explicitly specified) display count, but only after the user confirms the deletion; passing `--yes` (or `-y`) SHALL skip the confirmation prompt.

#### Scenario: Listing saved layouts

- **WHEN** a user runs `mumu list`
- **THEN** the system lists each saved layout with the display count it was saved for

#### Scenario: Previewing a saved layout

- **WHEN** a user runs `mumu show` for a display count that has a saved layout
- **THEN** the system displays that layout's window entries, that display count's configured pin rules, and that display count's effective hook-command preview, without moving any windows or running any command

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
