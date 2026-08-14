## ADDED Requirements

### Requirement: Layout output displays the native Mission Control Space number alongside the logical number

Wherever `mimi layout` output displays a logical Space number for a specific window entry, the system SHALL also display the corresponding macOS Mission Control Space number — the same 1-based ordinal used by `mimi action space <n>` and `mimi action move_window_to_space <n>` — resolved against the current display arrangement at the time of output. This SHALL NOT change what is persisted in a saved layout file, which continues to record only the logical ordinal.

#### Scenario: Previewing a saved layout shows both numbers

- **WHEN** a user runs `mimi layout show` and a saved entry's logical Space ordinal differs from its current Mission Control ordinal (e.g. because the primary display isn't the leftmost one)
- **THEN** the displayed entry shows both the logical Space number and the corresponding Mission Control Space number

#### Scenario: Restore progress shows both numbers per window moved

- **WHEN** `mimi layout restore` moves a matched window to its recorded Space
- **THEN** the progress line for that move shows both the logical Space number and the corresponding Mission Control Space number

#### Scenario: Restore skip summary shows both numbers per skipped entry

- **WHEN** `mimi layout restore` reports a skipped entry that has a resolvable saved Space ordinal
- **THEN** the skip summary line for that entry shows both the logical Space number and the corresponding Mission Control Space number

#### Scenario: Logical and Mission Control numbers coincide

- **WHEN** the current display arrangement's primary display is also the leftmost display, so the logical and Mission Control orderings are identical
- **THEN** the two numbers shown are equal, and both are still displayed
