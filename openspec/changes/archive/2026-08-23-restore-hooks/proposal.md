## Why

`mumu restore` only moves windows to Spaces. Some display-arrangement changes a user wants to trigger alongside a restore — muting/pausing media, switching audio output, toggling a KVM or monitor-power script, notifying another tool — currently require a separate manual step or a wrapper script around `mumu restore` itself. Letting a user configure external commands to run automatically around a restore removes that manual step and keeps the arrangement-switch workflow to a single `mumu restore` invocation.

## What Changes

- Add a `hooks` setting to `config.yaml` with `off` and `on` command arrays run around every `mumu restore`: `off` commands run first (before any window is moved), `on` commands run last (after the restore's move phase completes). Commands within an array run sequentially in the order listed.
- Add a per-display-count override under `hooks.layouts.<display_count>.off`/`.on`, keyed the same way `pins` already is, for arrangement-specific commands (e.g. only for a 2-display setup).
- Both the global and per-layout arrays apply together, bracketed with the global arrays outermost: global `off` → layout `off` → **[windows restored]** → layout `on` → global `on`.
- Each command entry may be written as a single shell string (run via `sh -c`) or as an explicit argv list (run directly, no shell). A command's stdout/stderr stream directly to `mumu`'s own output as it runs.
- A command that exits non-zero or fails to start is reported and logged, but does not stop the remaining commands in its array and does not abort the restore's window-moving phase.
- `mumu restore` gains a `--no-hooks` flag to skip running any configured commands for that one invocation.
- `mumu show` additionally previews the effective, ordered `off`/`on` command lists (global + per-layout, in actual run order) for the display count being shown, alongside the existing saved-layout and pin previews. This is a config preview only — no commands are executed.
- Hooks only run as part of `mumu restore` actually proceeding: they do **not** run when restore reports no saved layout for the current display count, or when the user declines the arrangement-drift confirmation prompt. They have no effect on `mumu save`.

## Capabilities

### New Capabilities

- `restore-hooks`: user-configured external commands, split into `off`/`on` arrays and optionally scoped per display count, run sequentially around every `mumu restore`.

### Modified Capabilities

- `configuration`: `config.yaml` gains the `hooks` setting (global `off`/`on` arrays plus a `layouts.<display_count>.off`/`.on` map), with its own validation rules for command entries.
- `space-layout`: `mumu restore` gains a hook-execution phase bracketing its existing window-move phase, plus a `--no-hooks` flag; `mumu show`'s output gains the effective hook-command preview described above.

## Impact

- `internal/config/config.go`, `internal/config/config_test.go`: parse and validate the new `hooks` setting (global `Hooks{Off, On []Command}` plus `map[int]Hooks` for per-display-count overrides); extend `defaultConfigYAML`.
- New `internal/hooks` package: runs an ordered list of `config.Command` values sequentially, streaming output and reporting per-command success/failure without aborting the run.
- `cmd/mumu/cmd/layout.go`: `mumu restore` resolves the effective off/on command lists for the current display count and runs them bracketing the existing `layout.Restore` call (skipped entirely by `--no-hooks`, and never reached when restore aborts before that call); `mumu show` prints the effective ordered hook-command preview.
- `docs/CONFIG_SCHEMA.md`, `docs/CLI.md`, `AGENTS.md`: document the new `config.yaml` keys, the bracketing execution order, and the `--no-hooks` flag.
