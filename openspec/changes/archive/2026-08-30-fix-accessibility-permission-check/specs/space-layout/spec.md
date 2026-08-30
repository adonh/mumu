## MODIFIED Requirements

### Requirement: Permission status command

`mumu status` SHALL report whether Accessibility and Screen Recording permission are currently granted, without performing a save or restore and without reporting on any background process or daemon. The Accessibility check SHALL reflect the current, live grant state at the moment the command runs, not a value cached from an earlier point in the same process's lifetime or from a previous build of `mumu`.

#### Scenario: Status reports granted permissions

- **WHEN** a user runs `mumu status` after granting both Accessibility and Screen Recording permission
- **THEN** the command reports both permissions as granted and makes no changes to any window, Space, or saved layout

#### Scenario: Status reports missing permissions

- **WHEN** a user runs `mumu status` before granting one or both of Accessibility or Screen Recording permission
- **THEN** the command clearly reports which of the two permissions is missing

#### Scenario: Status reflects a permission grant made immediately beforehand

- **WHEN** a user grants Accessibility permission to `mumu` in System Settings and then runs `mumu status` (a new process)
- **THEN** the command reports Accessibility as granted, without requiring the user to relaunch, reboot, or take any action beyond granting the permission
