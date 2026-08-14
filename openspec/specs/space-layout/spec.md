## Purpose

Lets a user save the current assignment of application windows to Mission Control Spaces for the connected display configuration, and later restore that arrangement with a single command, without requiring admin/root privileges.

## Requirements

### Requirement: Logical left-to-right Space numbering

The system SHALL compute a logical Space ordinal, scoped to this capability only, by sorting all connected displays by physical left-to-right position and concatenating each display's own Spaces in that order, regardless of which display is the primary (menu-bar) display. This numbering SHALL NOT alter the underlying macOS Mission Control Space ordinal (primary-first), which the system separately resolves and displays alongside the logical ordinal.

#### Scenario: Primary display is not the leftmost display

- **WHEN** the connected displays are physically arranged left to right as [Display B, Display A (primary), Display C], each with its own Spaces
- **THEN** the logical Space ordinals number Display B's Spaces first, then Display A's Spaces, then Display C's Spaces, regardless of Display A's primary status

#### Scenario: Existing space commands are unaffected

- **WHEN** the system resolves the macOS Mission Control Space ordinal for display alongside the logical ordinal
- **THEN** that resolution uses macOS's own primary-first ordering, unchanged by the logical left-to-right numbering

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

For each saved window entry belonging to a given application, the system SHALL first attempt to match it to a currently open window of that application by exact title. If no unambiguous title match exists, the system SHALL fall back to matching by the window's positional index within that application, using the same deterministic window ordering used at save time. If neither yields a match and exactly one of the application's currently open windows remains unclaimed, the system SHALL match the entry to that window regardless of title or position.

#### Scenario: Exact title match

- **WHEN** a saved entry for an application has a window title that exactly matches one of that application's currently open windows
- **THEN** that window is selected for the move

#### Scenario: Fallback to index when titles differ

- **WHEN** a saved entry's window title does not match any of that application's currently open window titles
- **THEN** the system matches by the entry's positional index among that application's currently open windows

#### Scenario: Fallback to sole remaining window when title and index both fail

- **WHEN** a saved entry's window title does not match any currently open window and its saved positional index is unavailable, but exactly one of that application's currently open windows remains unclaimed by any other entry
- **THEN** the system matches the entry to that sole remaining window regardless of its title or position

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

If a saved layout's logical Space ordinals exceed the number of Spaces currently available, the system SHALL NOT create new Spaces. It SHALL place window entries that fit within the currently available Spaces and SHALL report entries that could not be placed due to insufficient Spaces.

#### Scenario: Saved layout requires more Spaces than currently exist

- **WHEN** a saved layout references a logical Space ordinal higher than the number of Spaces currently available on the corresponding display
- **THEN** the system does not create a new Space, moves the windows that fit within existing Spaces, and reports the entries it could not place

### Requirement: Saved layouts contain no window geometry

Saved layouts SHALL contain only application identity, window title, and logical Space ordinal. Window position and size SHALL NOT be part of a saved layout.

#### Scenario: Saved layout omits position and size

- **WHEN** a layout is saved
- **THEN** the persisted data for each window entry contains no position or size information

### Requirement: Layout management commands

`mumu list` SHALL show all saved layouts along with the display count each is keyed to. `mumu show` SHALL display the contents of a saved layout without applying it. `mumu delete` SHALL remove a saved layout for the current (or explicitly specified) display count.

#### Scenario: Listing saved layouts

- **WHEN** a user runs `mumu list`
- **THEN** the system lists each saved layout with the display count it was saved for

#### Scenario: Previewing a saved layout

- **WHEN** a user runs `mumu show` for a display count that has a saved layout
- **THEN** the system displays that layout's window entries without moving any windows

#### Scenario: Deleting a saved layout

- **WHEN** a user runs `mumu delete` for a display count that has a saved layout
- **THEN** the system removes that saved layout and it no longer appears in `mumu list`

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

`mumu show` and `mumu restore` SHALL accept a `--sort` flag with values `display` (default), `macos`, and `app`, controlling the order in which per-window entries are printed: `display` orders by the logical left-to-right Space ordinal ascending; `macos` orders by the macOS Mission Control Space ordinal ascending (the same numbering macOS's own "Switch to Desktop `<n>`" shortcut uses); `app` orders by application bundle identifier ascending. This ordering SHALL apply to `mumu show`'s entry listing, `mumu restore`'s per-window move progress lines, and the ordering of entries within each reason group of `mumu restore`'s skip summary. Regardless of which key is selected as primary, entries with an equal primary-key value SHALL be ordered by cascading through the remaining keys in the fixed priority: Space ordinal, then bundle identifier, then window title. This SHALL NOT change which windows are matched, moved, or skipped, nor anything persisted in a saved layout file — it only affects display order.

#### Scenario: Default order is display sequence

- **WHEN** a user runs `mumu show` or `mumu restore` without passing `--sort`
- **THEN** entries are printed ordered by logical left-to-right Space ordinal ascending

#### Scenario: Sorting by macOS Mission Control Space number

- **WHEN** a user passes `--sort macos` to `mumu show` or `mumu restore`
- **THEN** entries are printed ordered by their current macOS Mission Control Space ordinal ascending, regardless of their logical Space ordinal

#### Scenario: Sorting by application

- **WHEN** a user passes `--sort app` to `mumu show` or `mumu restore`
- **THEN** entries are printed ordered by application bundle identifier ascending, grouping all of one application's windows together

#### Scenario: Tie-break when the primary sort key is equal

- **WHEN** two or more entries share the same value for the selected `--sort` key (e.g. two windows on the same Space when sorting by `display`)
- **THEN** those entries are ordered relative to each other by Space ordinal, then bundle identifier, then window title, in that order

#### Scenario: Sort order applies within restore's skip summary

- **WHEN** `mumu restore` reports multiple skipped entries sharing the same skip reason
- **THEN** those entries are listed within that reason group ordered according to the selected `--sort` key

### Requirement: Flat top-level command surface

`mumu`'s commands SHALL be top-level subcommands of the executable directly (`mumu save`, `mumu restore`, `mumu show`, `mumu list`, `mumu delete`, `mumu status`), with no intermediate command grouping. There SHALL be no `layout` (or other) subcommand namespace to pass through first.

#### Scenario: Commands are invoked directly on the executable

- **WHEN** a user runs `mumu save`, `mumu restore`, `mumu show`, `mumu list`, or `mumu delete`
- **THEN** the command runs directly, with no intermediate subcommand word required

#### Scenario: Top-level help lists the commands directly

- **WHEN** a user runs `mumu --help`
- **THEN** `save`, `restore`, `show`, `list`, and `delete` are listed as direct top-level commands, not nested under any grouping subcommand

### Requirement: Permission status command

`mumu status` SHALL report whether Accessibility and Screen Recording permission are currently granted, without performing a save or restore and without reporting on any background process or daemon.

#### Scenario: Status reports granted permissions

- **WHEN** a user runs `mumu status` after granting both Accessibility and Screen Recording permission
- **THEN** the command reports both permissions as granted and makes no changes to any window, Space, or saved layout

#### Scenario: Status reports missing permissions

- **WHEN** a user runs `mumu status` before granting one or both of Accessibility or Screen Recording permission
- **THEN** the command clearly reports which of the two permissions is missing
