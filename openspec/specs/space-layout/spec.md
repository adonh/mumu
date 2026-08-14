## Purpose

Lets a user save the current assignment of application windows to Mission Control Spaces for the connected display configuration, and later restore that arrangement with a single command, without requiring admin/root privileges.

## Requirements

### Requirement: Logical left-to-right Space numbering

The system SHALL compute a logical Space ordinal, scoped to this capability only, by sorting all connected displays by physical left-to-right position and concatenating each display's own Spaces in that order, regardless of which display is the primary (menu-bar) display. This numbering SHALL NOT affect and SHALL NOT be used by the existing `mimi action space <n>` or `move_window_to_space <n>` commands.

#### Scenario: Primary display is not the leftmost display

- **WHEN** the connected displays are physically arranged left to right as [Display B, Display A (primary), Display C], each with its own Spaces
- **THEN** the logical Space ordinals number Display B's Spaces first, then Display A's Spaces, then Display C's Spaces, regardless of Display A's primary status

#### Scenario: Existing space commands are unaffected

- **WHEN** a user runs `mimi action space <n>` or `mimi action move_window_to_space <n>`
- **THEN** the Space numbering used is the existing (primary-first) numbering, unchanged by this capability

### Requirement: Save current window-to-Space layout

`mimi layout save` SHALL capture, for every non-fullscreen window on every Space across all connected displays, the owning application's bundle identifier, the window's title, and the window's logical Space ordinal. The captured layout SHALL be persisted keyed by the current number of connected displays, overwriting any previously saved layout for that same display count.

#### Scenario: Save captures windows across multiple Spaces and displays

- **WHEN** a user runs `mimi layout save` with windows open across several Spaces on multiple displays
- **THEN** the saved layout contains one entry per non-fullscreen window, each with its app bundle identifier, window title, and logical Space ordinal

#### Scenario: Save excludes fullscreen windows

- **WHEN** an application window is in native fullscreen (occupying its own dedicated Space) at save time
- **THEN** that window is not included in the saved layout

#### Scenario: Save overwrites a prior layout for the same display count

- **WHEN** a layout was previously saved while 2 displays were connected, and the user runs `mimi layout save` again while still 2 displays are connected
- **THEN** the previous layout for "2 displays" is replaced by the new capture

### Requirement: Restore auto-detects the layout to apply

`mimi layout restore`, invoked with no name or identifier argument, SHALL detect the current number of connected displays and look up the saved layout keyed to that count. If no layout exists for the current display count, the system SHALL report this clearly and SHALL make no changes.

#### Scenario: Restore finds a matching saved layout

- **WHEN** a user runs `mimi layout restore` and a layout was previously saved for the current number of connected displays
- **THEN** the system proceeds to restore that layout

#### Scenario: Restore finds no matching saved layout

- **WHEN** a user runs `mimi layout restore` and no layout has been saved for the current number of connected displays
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

For each successfully matched window, the system SHALL move it to the Space corresponding to its saved logical ordinal, resolved against the current display arrangement, using mimi's existing window-to-Space move capability. The system SHALL NOT modify the window's position or size during restore.

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

`mimi layout list` SHALL show all saved layouts along with the display count each is keyed to. `mimi layout show` SHALL display the contents of a saved layout without applying it. `mimi layout delete` SHALL remove a saved layout for the current (or explicitly specified) display count.

#### Scenario: Listing saved layouts

- **WHEN** a user runs `mimi layout list`
- **THEN** the system lists each saved layout with the display count it was saved for

#### Scenario: Previewing a saved layout

- **WHEN** a user runs `mimi layout show` for a display count that has a saved layout
- **THEN** the system displays that layout's window entries without moving any windows

#### Scenario: Deleting a saved layout

- **WHEN** a user runs `mimi layout delete` for a display count that has a saved layout
- **THEN** the system removes that saved layout and it no longer appears in `mimi layout list`

### Requirement: Manual invocation only

Layout save and restore SHALL only occur when explicitly invoked via the `mimi layout save` or `mimi layout restore` CLI commands. The system SHALL NOT automatically save or restore layouts through the hook daemon, on a schedule, or at login as part of this capability.

#### Scenario: No automatic save or restore

- **WHEN** the hook daemon is running, or the system starts up, or a display is connected or disconnected
- **THEN** no layout is automatically saved or restored as a result

### Requirement: No elevated privileges required

All layout save and restore operations SHALL function using only user-grantable macOS privacy permissions (Accessibility and Screen Recording) plus mimi's existing unprivileged private WindowServer connection. The system SHALL NOT require root/admin privileges or SIP modification. Screen Recording is required in addition to Accessibility specifically because window titles — needed for reliable restore matching — are inaccessible without it; this is scoped to the `mimi layout` command group and does not change the permission requirements of mimi's other actions.

#### Scenario: Layout operations require only Accessibility and Screen Recording

- **WHEN** a user has granted mimi both Accessibility and Screen Recording permission
- **THEN** `mimi layout save` and `mimi layout restore` function without requesting or requiring root/admin privileges or any other elevated access

#### Scenario: Screen Recording not granted

- **WHEN** a user runs `mimi layout save` or `mimi layout restore` without having granted Screen Recording permission
- **THEN** the system reports clearly that Screen Recording permission is required for this command group and makes no changes
