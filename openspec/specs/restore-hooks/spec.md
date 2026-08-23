## Purpose

Lets a user configure external commands, split into "off" (run before) and "on" (run after) arrays, to execute automatically around every `mumu restore`, either globally or scoped to a specific connected-display-count.

## Requirements

### Requirement: Hooks are configured as global and per-display-count off/on arrays

The system SHALL let a user declare, in `config.yaml`, a global `off` command array and a global `on` command array that apply to every `mumu restore` regardless of display count, and SHALL let a user additionally declare, per connected-display-count, an `off` command array and an `on` command array that apply only when that number of displays is connected. Any of these four arrays MAY be absent or empty, independently of the others.

#### Scenario: Only global hooks configured

- **WHEN** a user configures global `off`/`on` command arrays and no per-display-count arrays
- **THEN** `mumu restore` runs the global arrays regardless of the current display count

#### Scenario: Only per-display-count hooks configured

- **WHEN** a user configures `off`/`on` command arrays for a specific display count and no global arrays
- **THEN** `mumu restore` runs the configured arrays only when that number of displays is connected, and runs no commands for any other display count

#### Scenario: No hooks configured

- **WHEN** a user has not configured any `off`/`on` command arrays, globally or per display count
- **THEN** `mumu restore` runs no external commands

### Requirement: Off commands run before restore, on commands run after, global brackets layout

For a `mumu restore` invocation targeting a given display count, the system SHALL run, in order: the global `off` array, then that display count's `off` array, then the restore's window-move phase, then that display count's `on` array, then the global `on` array. Within each array, commands SHALL run sequentially, one at a time, in the order listed.

#### Scenario: Full bracketing order

- **WHEN** both global and per-display-count `off`/`on` arrays are configured for the current display count
- **THEN** the system runs the global `off` array, then the display count's `off` array, then moves windows, then the display count's `on` array, then the global `on` array, in that order

#### Scenario: Commands within an array run in listed order

- **WHEN** an `off` or `on` array contains more than one command
- **THEN** the system runs them one at a time, starting the next command only after the previous one has finished, in the array's listed order

### Requirement: A command may be a shell string or an explicit argv list

The system SHALL accept each configured command as either a single string, executed through a shell (so it may use pipes, redirection, and shell expansion), or as a list of strings, executed directly as a program and its arguments with no shell involved.

#### Scenario: Shell string command

- **WHEN** a configured command is written as a single string
- **THEN** the system executes it through a shell

#### Scenario: Argv list command

- **WHEN** a configured command is written as a list of strings
- **THEN** the system executes the first element as the program and the remaining elements as its arguments, without invoking a shell

### Requirement: Command output streams live, failures are logged without aborting

While a configured command runs, the system SHALL stream its standard output and standard error directly to `mumu`'s own output. If a command exits with a non-zero status or fails to start, the system SHALL report that failure clearly (identifying which command failed) and SHALL continue running the remaining commands in that array; a command failure SHALL NOT abort the restore's window-move phase or prevent any subsequently scheduled `off`/`on` array from running.

#### Scenario: Command output is visible while running

- **WHEN** a configured command produces output while it runs
- **THEN** that output appears as part of `mumu restore`'s own output, without waiting for the command to finish

#### Scenario: A failing command does not stop later commands

- **WHEN** one command in an `off` or `on` array exits non-zero
- **THEN** the system reports the failure and proceeds to run the remaining commands in that array

#### Scenario: A failing off command does not prevent windows from being moved

- **WHEN** a command in the `off` phase fails
- **THEN** the restore's window-move phase still runs afterward

### Requirement: Hooks require restore and have no independent trigger

Configured hook commands SHALL only run as part of a `mumu restore` invocation that actually proceeds to (or past) its window-move phase. The system SHALL NOT run any hook command when `mumu restore` reports no saved layout exists for the current display count, and SHALL NOT run any hook command when the user declines the arrangement-drift confirmation prompt. The system SHALL NOT run hook commands as part of `mumu save` or any other command, and SHALL NOT run them automatically on a schedule, at login, or in response to any system event.

#### Scenario: No saved layout means no hooks run

- **WHEN** a user runs `mumu restore` for a display count that has hook commands configured but no saved layout
- **THEN** the system reports that no saved layout exists and runs no hook commands

#### Scenario: Declining the drift confirmation means no hooks run

- **WHEN** a user runs `mumu restore`, is prompted to confirm a display-arrangement mismatch, and declines
- **THEN** the system aborts the restore, moves no windows, and runs no hook commands

#### Scenario: Hooks have no effect on save

- **WHEN** a user runs `mumu save`
- **THEN** no hook command runs, regardless of what is configured

### Requirement: `--no-hooks` skips hook execution for one invocation

`mumu restore` SHALL accept a `--no-hooks` flag that, when passed, skips running any `off`/`on` hook commands — global or per-display-count — for that invocation only, while otherwise performing the restore normally. This flag SHALL NOT modify `config.yaml` or otherwise persist beyond that invocation.

#### Scenario: Skipping hooks for one restore

- **WHEN** a user runs `mumu restore --no-hooks` with hook commands configured for the current display count
- **THEN** the system restores windows normally and runs no hook commands

#### Scenario: Hooks resume on the next restore without the flag

- **WHEN** a user runs `mumu restore --no-hooks` and then later runs `mumu restore` without that flag
- **THEN** the second invocation runs the configured hook commands normally

### Requirement: `mumu show` previews the effective hook command order

`mumu show` SHALL, in addition to its existing saved-layout and pin previews, display the effective `off` and `on` command lists for the display count being shown, in the exact order `mumu restore` would run them (global `off`, then that display count's `off`, then that display count's `on`, then global `on`). This listing SHALL show the configured commands as written, without executing any of them.

#### Scenario: Show lists the effective hook order

- **WHEN** a user runs `mumu show` for a display count that has both global and per-display-count hook commands configured
- **THEN** the output includes the `off` commands in run order (global then per-display-count) and the `on` commands in run order (per-display-count then global)

#### Scenario: Show with no configured hooks

- **WHEN** a user runs `mumu show` for a display count that has no hook commands configured, globally or for that display count
- **THEN** the output's hook-command listing is empty or omitted, and the rest of `mumu show`'s output is unaffected
