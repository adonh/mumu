## Purpose

Lets a user save the current assignment of application windows to Mission Control Spaces for the connected display configuration, and later restore that arrangement with a single command, without requiring admin/root privileges.

## Requirements

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

### Requirement: Save current window-to-Space layout

`mumu save` SHALL capture, for every non-fullscreen window on every Space across all connected displays, the owning application's bundle identifier, the window's title, and the window's logical Space ordinal. The captured layout SHALL be persisted keyed by the current number of connected displays, overwriting any previously saved layout for that same display count.

#### Scenario: Save captures windows across multiple Spaces and displays

- **WHEN** a user runs `mumu save` with windows open across several Spaces on multiple displays
- **THEN** the saved layout contains one entry per non-fullscreen window, each with its app bundle identifier, window title, and logical Space ordinal

#### Scenario: Save excludes fullscreen windows

- **WHEN** an application window is in native fullscreen (occupying its own dedicated Space) at save time
- **THEN** that window is not included in the saved layout

#### Scenario: Save overwrites a prior layout for the same display count

- **WHEN** a layout was previously saved while 2 displays were connected, and the user runs `mumu save` again while still 2 displays are connected
- **THEN** the previous layout for "2 displays" is replaced by the new capture

### Requirement: Restore auto-detects the layout to apply

`mumu restore`, invoked with no name or identifier argument, SHALL detect the current number of connected displays and look up the saved layout keyed to that count. If no layout exists for the current display count, the system SHALL report this clearly and SHALL make no changes.

#### Scenario: Restore finds a matching saved layout

- **WHEN** a user runs `mumu restore` and a layout was previously saved for the current number of connected displays
- **THEN** the system proceeds to restore that layout

#### Scenario: Restore finds no matching saved layout

- **WHEN** a user runs `mumu restore` and no layout has been saved for the current number of connected displays
- **THEN** the system reports that no matching layout exists and makes no changes to any window or Space

### Requirement: Restore warns on display arrangement mismatch

Before moving any windows, if the current physical left-to-right display arrangement (per-display Space counts and order) does not match what was recorded when the layout was saved, the system SHALL warn the user describing the mismatch and SHALL require explicit confirmation before proceeding. Declining SHALL abort the restore with no windows moved.

#### Scenario: Arrangement matches saved state

- **WHEN** the current per-display Space arrangement matches exactly what was recorded at save time
- **THEN** the system proceeds with restore without prompting for confirmation

#### Scenario: Arrangement differs from saved state and user proceeds

- **WHEN** the current display arrangement differs from what was recorded (e.g. displays have been reordered) and the user confirms when prompted
- **THEN** the system proceeds with a best-effort restore using the current arrangement

#### Scenario: Arrangement differs from saved state and user declines

- **WHEN** the current display arrangement differs from what was recorded and the user declines when prompted
- **THEN** the system aborts the restore and moves no windows

### Requirement: Restore matches windows by title, falling back to index

For each application with saved window entries, the system SHALL match its saved entries to that application's currently open windows as a single batch: for every remaining saved entry and every remaining currently open window not yet claimed by another saved entry, the system SHALL measure title similarity as the proportion of shared words to total distinct words between the saved title and the window's title, comparing words case-insensitively and independent of word order. The system SHALL then assign entry-window pairs starting from the highest similarity score downward, skipping any entry or window already assigned, until no further pair can be assigned. When two or more candidate windows tie for an entry's highest score (or two or more entries tie for a window's highest score), the system SHALL prefer the candidate whose current position among the application's open windows equals the entry's saved positional index; if that does not resolve the tie, the system SHALL deterministically choose one of the tied candidates rather than leaving the entry unmatched. A saved entry SHALL remain unmatched only when no currently open window of that application remains unclaimed for it.

After completing those matching steps for an application, the system SHALL place every remaining unclaimed currently open window of that application according to the following order of precedence:

1. If the configuration file (see the `configuration` capability) has a `default_spaces` rule for that application's bundle identifier at the current display count, the fallback target SHALL be that rule's configured logical Space ordinal, regardless of the application's valid saved-entry assignments (if any) or whether they would otherwise tie.
2. Otherwise, if the application has at least one valid saved-entry assignment this restore, the fallback target SHALL be the logical Space ordinal occurring most often among those assignments; if two or more logical Space ordinals tie for most prevalent, the fallback target SHALL instead be the logical Space currently displayed on the primary (menu-bar) display.
3. Otherwise (no configured `default_spaces` rule and no valid saved-entry assignment), the system SHALL leave the application's remaining unclaimed open windows unchanged.

The system SHALL report each fallback placement in restore progress output, distinguishing a configured-default placement from a prevalent-Space placement, and SHALL move each currently open window at most once.

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

- **WHEN** one open Chrome window has a valid saved-entry assignment to logical Space 4 and another open Chrome window remains unclaimed after batch matching, and Chrome has no configured `default_spaces` rule for the current display count
- **THEN** the system moves the unclaimed Chrome window to logical Space 4 and reports that placement in restore progress output

#### Scenario: Most prevalent Space wins

- **WHEN** an application's valid saved-entry assignments target logical Spaces 2, 2, and 5, one of its currently open windows remains unclaimed after standard matching, and that application has no configured `default_spaces` rule for the current display count
- **THEN** the system moves the unclaimed window to logical Space 2

#### Scenario: Tied prevalent Spaces use the primary display's current Space

- **WHEN** an application's valid saved-entry assignments are evenly split between logical Spaces 2 and 5, one of its currently open windows remains unclaimed after standard matching, that application has no configured `default_spaces` rule for the current display count, and the primary display currently shows logical Space 7
- **THEN** the system moves the unclaimed window to logical Space 7 and reports that placement in restore progress output

#### Scenario: No valid assignment leaves unmatched windows unchanged

- **WHEN** an application's currently open windows cannot be matched to any valid saved-entry assignment, and that application has no configured `default_spaces` rule for the current display count
- **THEN** the system does not move that application's remaining unclaimed windows through the application-level fallback

#### Scenario: Configured default space overrides the prevalent-Space heuristic

- **WHEN** an application has a configured `default_spaces` rule targeting logical Space 6 for the current display count, and its valid saved-entry assignments this restore unambiguously target logical Space 2
- **THEN** the system moves that application's remaining unclaimed windows to logical Space 6, not logical Space 2, and reports the placement as a configured-default placement

#### Scenario: Configured default space activates a fallback with zero valid saved-entry assignments

- **WHEN** an application has a configured `default_spaces` rule targeting logical Space 8 for the current display count, and none of its currently open windows were matched to any valid saved-entry assignment this restore
- **THEN** the system moves that application's currently open windows to logical Space 8 and reports the placement as a configured-default placement

#### Scenario: Windows claimed by a higher-precedence pin are excluded from saved-layout matching

- **WHEN** pin rules take precedence (the default) and a currently open window has already been claimed by a matched pin rule
- **THEN** that window is not considered by saved-layout matching or its application-level fallback, as if it were not currently open

### Requirement: Restore skips applications that are not running

If a saved window entry's application is not currently running, the system SHALL skip that entry, SHALL NOT launch the application, and SHALL report the skipped entry to the user.

#### Scenario: Saved application is not running

- **WHEN** a saved layout references an application that is not currently running
- **THEN** the system skips all window entries for that application, does not launch it, and reports the skip

### Requirement: Restore moves matched windows without altering their frame

For each successfully matched window, the system SHALL move it to the Space corresponding to its saved logical ordinal, resolved against the current display arrangement, using mumu's existing window-to-Space move capability. The system SHALL NOT modify the window's position or size during restore.

#### Scenario: Matched window is moved to its recorded Space

- **WHEN** a saved window entry is matched to a currently open window
- **THEN** that window is moved to the Space corresponding to the entry's logical Space ordinal, and its position and size are left unchanged

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

### Requirement: Saved layouts contain no window geometry

Saved layouts SHALL contain only application identity, window title, and logical Space ordinal. Window position and size SHALL NOT be part of a saved layout.

#### Scenario: Saved layout omits position and size

- **WHEN** a layout is saved
- **THEN** the persisted data for each window entry contains no position or size information

### Requirement: Layout management commands

`mumu list` SHALL show all saved layouts along with the display count each is keyed to. `mumu show` SHALL display the contents of a saved layout, without applying it, plus that display count's effective hook-command preview (see the `restore-hooks` capability), configured pin rules (see the `window-pinning` capability), and configured `default_spaces` rules (see the `configuration` capability). `mumu delete` SHALL remove a saved layout for the current (or explicitly specified) display count, but only after the user confirms the deletion; passing `--yes` (or `-y`) SHALL skip the confirmation prompt.

#### Scenario: Listing saved layouts

- **WHEN** a user runs `mumu list`
- **THEN** the system lists each saved layout with the display count it was saved for

#### Scenario: Previewing a saved layout

- **WHEN** a user runs `mumu show` for a display count that has a saved layout
- **THEN** the system displays that layout's window entries, that display count's configured pin rules, and that display count's effective hook-command preview, without moving any windows or running any command

#### Scenario: Previewing configured default spaces

- **WHEN** a user runs `mumu show` for a display count that has one or more `default_spaces` rules configured
- **THEN** the system displays each rule's application bundle identifier and target Space, without performing any live-window matching

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

### Requirement: Layout output displays the native Mission Control Space number alongside the logical number

Wherever `mumu` output displays a logical Space number for a specific window entry, the system SHALL also display the corresponding macOS Mission Control Space number, resolved against the current display arrangement at the time of output. This number is the same 1-based ordinal macOS's own "Switch to Desktop `<n>`" keyboard shortcut uses, so it remains meaningful as a cross-reference on its own. This SHALL NOT change what is persisted in a saved layout file, which continues to record only the logical ordinal.

#### Scenario: Previewing a saved layout shows both numbers

- **WHEN** a user runs `mumu show` and a saved entry's logical Space ordinal differs from its current Mission Control ordinal (e.g. because the primary display isn't the leftmost one)
- **THEN** the displayed entry shows both the logical Space number and the corresponding Mission Control Space number

#### Scenario: Restore progress shows both numbers per window moved

- **WHEN** `mumu restore` moves a matched window to its recorded Space
- **THEN** the progress line for that move shows both the logical Space number and the corresponding Mission Control Space number

#### Scenario: Restore skip summary shows both numbers per skipped entry

- **WHEN** `mumu restore` reports a skipped entry that has a resolvable saved Space ordinal
- **THEN** the skip summary line for that entry shows both the logical Space number and the corresponding Mission Control Space number

#### Scenario: Logical and Mission Control numbers coincide

- **WHEN** the current display arrangement's primary display is also the leftmost display, so the logical and Mission Control orderings are identical
- **THEN** the two numbers shown are equal, and both are still displayed

### Requirement: Manual invocation only

Layout save and restore SHALL only occur when explicitly invoked via the `mumu save` or `mumu restore` CLI commands. The system SHALL NOT automatically save or restore layouts on a schedule, at login, or in response to any system event.

#### Scenario: No automatic save or restore

- **WHEN** the system starts up, or a display is connected or disconnected
- **THEN** no layout is automatically saved or restored as a result

### Requirement: No elevated privileges required

All layout save and restore operations SHALL function using only user-grantable macOS privacy permissions (Accessibility and Screen Recording) plus mumu's unprivileged private WindowServer connection. The system SHALL NOT require root/admin privileges or SIP modification. Screen Recording is required in addition to Accessibility specifically because window titles — needed for reliable restore matching — are inaccessible without it.

#### Scenario: Layout operations require only Accessibility and Screen Recording

- **WHEN** a user has granted mumu both Accessibility and Screen Recording permission
- **THEN** `mumu save` and `mumu restore` function without requesting or requiring root/admin privileges or any other elevated access

#### Scenario: Screen Recording not granted

- **WHEN** a user runs `mumu save` or `mumu restore` without having granted Screen Recording permission
- **THEN** the system reports clearly that Screen Recording permission is required and makes no changes

### Requirement: Configurable output ordering for layout entries

`mumu show` and `mumu restore` SHALL accept a `--sort` flag with values `logical` (default), `macos`, and `app`, controlling the order in which per-window entries are printed: `logical` orders by mumu's own logical left-to-right Space ordinal ascending; `macos` orders by the macOS Mission Control Space ordinal ascending (the same numbering macOS's own "Switch to Desktop `<n>`" shortcut uses); `app` orders by application bundle identifier ascending. This ordering SHALL apply to `mumu show`'s entry listing, `mumu show`'s configured-pins and configured-default-spaces previews, `mumu restore`'s per-window move progress lines, and the ordering of entries within each reason group of `mumu restore`'s skip summary. Regardless of which key is selected as primary, entries with an equal primary-key value SHALL be ordered by cascading through the remaining keys in the fixed priority: Space ordinal, then bundle identifier, then window title (pin and default-space rules have no title, so they cascade only through Space ordinal then bundle identifier). This SHALL NOT change which windows are matched, moved, or skipped, which pin or default-space rule applies, nor anything persisted in a saved layout file or `config.yaml` — it only affects display order. The word "display" is reserved exclusively for physical-monitor concepts (e.g. the current display count, connected displays, the primary display) and SHALL NOT be used to name this sort key or its default value, to avoid confusion with logical Space ordering.

#### Scenario: Default order is logical Space sequence

- **WHEN** a user runs `mumu show` or `mumu restore` without passing `--sort`
- **THEN** entries are printed ordered by mumu's own logical left-to-right Space ordinal ascending

#### Scenario: Sorting by macOS Mission Control Space number

- **WHEN** a user passes `--sort macos` to `mumu show` or `mumu restore`
- **THEN** entries are printed ordered by their current macOS Mission Control Space ordinal ascending, regardless of their logical Space ordinal

#### Scenario: Sorting by application

- **WHEN** a user passes `--sort app` to `mumu show` or `mumu restore`
- **THEN** entries are printed ordered by application bundle identifier ascending, grouping all of one application's windows together

#### Scenario: Tie-break when the primary sort key is equal

- **WHEN** two or more entries share the same value for the selected `--sort` key (e.g. two windows on the same Space when sorting by `logical`)
- **THEN** those entries are ordered relative to each other by Space ordinal, then bundle identifier, then window title, in that order

#### Scenario: Sort order applies within restore's skip summary

- **WHEN** `mumu restore` reports multiple skipped entries sharing the same skip reason
- **THEN** those entries are listed within that reason group ordered according to the selected `--sort` key

#### Scenario: The old "display" sort value is no longer accepted

- **WHEN** a user passes `--sort display` (the value's pre-rename name) to `mumu show` or `mumu restore`
- **THEN** the system reports a clear error naming the accepted values (`logical`, `macos`, `app`) and makes no changes

#### Scenario: Sort order applies to mumu show's configured-pins preview

- **WHEN** a user runs `mumu show --sort app` for a display count with multiple configured pin rules
- **THEN** the "configured pin(s)" section lists those rules ordered by application bundle identifier ascending, rather than their order in `config.yaml`

#### Scenario: Sort order applies to mumu show's configured-default-spaces preview

- **WHEN** a user runs `mumu show` (default `--sort logical`) for a display count with multiple configured `default_spaces` rules
- **THEN** the "configured default space(s)" section lists those rules ordered by target Space ordinal ascending, rather than their order in `config.yaml`

#### Scenario: Hook command previews remain unsorted

- **WHEN** a user runs `mumu show` for a display count with configured `hooks.off`/`hooks.on` commands
- **THEN** those commands are still listed in their configured execution order, unaffected by `--sort`

### Requirement: Flat top-level command surface

`mumu`'s commands SHALL be top-level subcommands of the executable directly (`mumu save`, `mumu restore`, `mumu show`, `mumu list`, `mumu delete`, `mumu status`), with no intermediate command grouping. There SHALL be no `layout` (or other) subcommand namespace to pass through first.

#### Scenario: Commands are invoked directly on the executable

- **WHEN** a user runs `mumu save`, `mumu restore`, `mumu show`, `mumu list`, or `mumu delete`
- **THEN** the command runs directly, with no intermediate subcommand word required

#### Scenario: Top-level help lists the commands directly

- **WHEN** a user runs `mumu --help`
- **THEN** `save`, `restore`, `show`, `list`, and `delete` are listed as direct top-level commands, not nested under any grouping subcommand

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

### Requirement: Permission status command

`mumu status` SHALL report whether Accessibility and Screen Recording permission are currently granted, without performing a save or restore and without reporting on any background process or daemon.

#### Scenario: Status reports granted permissions

- **WHEN** a user runs `mumu status` after granting both Accessibility and Screen Recording permission
- **THEN** the command reports both permissions as granted and makes no changes to any window, Space, or saved layout

#### Scenario: Status reports missing permissions

- **WHEN** a user runs `mumu status` before granting one or both of Accessibility or Screen Recording permission
- **THEN** the command clearly reports which of the two permissions is missing
