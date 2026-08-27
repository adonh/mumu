## Purpose

Lets a user declare, per connected-display-count, fixed application-window-to-Space assignments in `config.yaml` that `mumu restore` applies on every run, independent of and without requiring re-saving a layout.

## Requirements

### Requirement: Pin rules are configured per display count

The system SHALL let a user declare, in `config.yaml`, a list of pin rules for each connected-display-count, where each pin rule specifies an application bundle identifier, a window-title pattern, and a target `ordinal` (the same two-part `"<display>:<space>"` numbering the `space-layout` capability uses for saved layouts). Pin rules declared for one display count SHALL have no effect when a different number of displays is connected.

#### Scenario: Pins differ by display count

- **WHEN** a user configures one set of pin rules for 2 displays and a different set of pin rules for 4 displays
- **THEN** with 2 displays connected only the 2-display pin rules take effect, and with 4 displays connected only the 4-display pin rules take effect

#### Scenario: No pin rules configured for the current display count

- **WHEN** `mumu restore` runs and no pin rules are configured for the current number of connected displays
- **THEN** restore proceeds using only its existing saved-layout matching, unaffected by this capability

### Requirement: Pin title matching reuses the existing approximate-match heuristic

The system SHALL match a pin rule's application windows using the exact same title-similarity method the `space-layout` capability's restore matching uses: shared-word proportion (case-insensitive, order-independent) scored across every candidate open window of the pin's bundle identifier, with pairs assigned greedily from the highest score down and deterministic tie-breaking, so a pin's title pattern need not exactly match a window's current title.

#### Scenario: Approximate title match resolves a pin

- **WHEN** a pin rule's configured title pattern shares some but not all words with one of that application's currently open window titles, and no other open window of that application scores higher
- **THEN** the system matches the pin to that window

#### Scenario: Multiple pins for the same application resolve independently

- **WHEN** two pin rules target the same application with different title patterns and different open windows of that application each score highest against a different one of those two patterns
- **THEN** the system matches each pin to its own highest-scoring window, and no open window is matched to more than one pin

### Requirement: Pin precedence relative to saved-layout matching is configurable

The system SHALL provide a `config.yaml` setting controlling whether pin rules or saved-layout entries take precedence when both would otherwise match the same currently open window during `mumu restore`: when pins take precedence, pin matching SHALL run first and any window it claims SHALL be excluded from saved-layout matching; when saved-layout entries take precedence, saved-layout matching (including its own application-level fallback placement) SHALL run first and pin matching SHALL only consider windows left unclaimed afterward. This setting SHALL default to pins taking precedence.

#### Scenario: Pins take precedence (default)

- **WHEN** the precedence setting is unset or set to prefer pins, and both a pin rule and a saved-layout entry for the same application would match the same currently open window
- **THEN** the system moves that window to the pin's target Space, and the saved-layout entry is matched against a different remaining window or reported unmatched

#### Scenario: Saved-layout entries take precedence

- **WHEN** the precedence setting is set to prefer the saved layout, and both a pin rule and a saved-layout entry for the same application would match the same currently open window
- **THEN** the system moves that window to the saved-layout entry's recorded Space, and the pin rule is matched against a different remaining window or left unapplied

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

### Requirement: Pins require restore and have no independent trigger

Pin rules SHALL only take effect as part of running `mumu restore`. The system SHALL NOT provide a separate command to apply pins on their own, and SHALL NOT apply pins automatically on any schedule, at login, or in response to a system event. If no saved layout exists for the current display count, `mumu restore` SHALL continue to report that condition and make no changes, regardless of whether pin rules are configured for that display count.

#### Scenario: Pins have no effect without running restore

- **WHEN** pin rules are configured for the current display count but the user has not run `mumu restore`
- **THEN** no window is moved as a result of those pin rules

#### Scenario: Restore still requires a saved layout

- **WHEN** a user runs `mumu restore` for a display count that has pin rules configured but no saved layout
- **THEN** the system reports that no saved layout exists for that display count and makes no changes, including not applying the configured pins

### Requirement: `mumu show` previews configured pins

`mumu show` SHALL, in addition to its existing saved-layout entry listing, display the pin rules configured for the display count being shown: each rule's application bundle identifier, title pattern, and target Space (shown with both the logical and macOS Mission Control Space numbers, per the `space-layout` capability's dual-numbering display). This listing SHALL show the configured rules as written, without matching them against any currently open window.

#### Scenario: Show lists configured pins alongside the saved layout

- **WHEN** a user runs `mumu show` for a display count that has both a saved layout and configured pin rules
- **THEN** the output includes the existing saved-layout entry listing plus a separate listing of that display count's pin rules, each showing its application, title pattern, and target Space with both numbering systems

#### Scenario: Show with no configured pins

- **WHEN** a user runs `mumu show` for a display count that has no pin rules configured
- **THEN** the output's pin listing is empty or omitted, and the existing saved-layout listing is unaffected
