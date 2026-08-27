## MODIFIED Requirements

### Requirement: Pin rules are configured per display count

The system SHALL let a user declare, in `config.yaml`, a list of pin rules for each connected-display-count, where each pin rule specifies an application bundle identifier, a window-title pattern, and a target `ordinal` (the same two-part `"<display>:<space>"` numbering the `space-layout` capability uses for saved layouts). Pin rules declared for one display count SHALL have no effect when a different number of displays is connected.

#### Scenario: Pins differ by display count

- **WHEN** a user configures one set of pin rules for 2 displays and a different set of pin rules for 4 displays
- **THEN** with 2 displays connected only the 2-display pin rules take effect, and with 4 displays connected only the 4-display pin rules take effect

#### Scenario: No pin rules configured for the current display count

- **WHEN** `mumu restore` runs and no pin rules are configured for the current number of connected displays
- **THEN** restore proceeds using only its existing saved-layout matching, unaffected by this capability

### Requirement: Pin matching moves windows using the same rules as saved-layout matching

Once a pin rule is matched to a currently open window, the system SHALL move that window to the Space corresponding to the pin's target `ordinal` following the same rules `mumu restore` applies to saved-layout entries: the application must currently be running (mumu SHALL NOT launch it), the target ordinal's display part must be within the number of displays currently connected and its space part must be within that specific display's currently available Space count, and the window's position and size SHALL be left unchanged. A pin whose application is not running, whose target ordinal's display or space part no longer exists, or whose window move fails SHALL be reported in `mumu restore`'s output the same way an unmatched or failed saved-layout entry is.

#### Scenario: Pinned application is not running

- **WHEN** `mumu restore` runs and a pin rule's application is not currently running
- **THEN** the system does not launch that application and reports the pin as skipped

#### Scenario: Pinned target display no longer exists

- **WHEN** a pin rule's target `ordinal` references a display ordinal higher than the number of displays currently connected
- **THEN** the system does not create a new display, does not move the pinned window, and reports the pin as skipped

#### Scenario: Pinned target Space no longer exists

- **WHEN** a pin rule's target `ordinal` references a space-within-display ordinal that exceeds the number of Spaces currently available on that specific display, but the display itself still exists
- **THEN** the system does not create a new Space, does not move the pinned window, and reports the pin as skipped

#### Scenario: Pin match is approximate

- **WHEN** a pin is matched to a currently open window whose title does not contain exactly the same words as the pin's configured title pattern
- **THEN** the move is reported as an approximate ("fuzzy") match, the same way an approximate saved-layout match is reported
