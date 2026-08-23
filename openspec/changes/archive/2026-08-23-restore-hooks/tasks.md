## 1. Config schema

- [x] 1.1 Add a `Command{Shell string; Argv []string}` type to `internal/config/config.go` with a custom `UnmarshalYAML` that decodes a scalar YAML node into `Shell` and a sequence node into `Argv`; verify with a unit test that both a plain string and a list-of-strings entry in the same YAML array decode correctly, and that an empty string, an empty list, or an unsupported node kind (e.g. a mapping) is rejected.
- [x] 1.2 Add `Hooks{Off, On []Command}` and, to `config.Config`, a global `Hooks Hooks` field and a `LayoutHooks map[int]Hooks` field (plus matching `fileFormat` entries for `hooks.off`/`hooks.on`/`hooks.layouts.<n>.off`/`.on`); verify with a unit test that a config file with global and per-display-count `hooks` sections parses into the expected struct values.
- [x] 1.3 Validate `hooks` on load: every command entry in every array must be a non-empty string or a non-empty list of non-empty strings; on violation return a `CodeInvalidConfig` error naming the config file path and the offending array/entry. Verify with unit tests for each invalid case (empty string, empty list, list containing an empty string, non-list value for `off`/`on`) and confirm no partial `Config` is returned.
- [x] 1.4 Update `defaultConfigYAML` with commented, empty-by-default `hooks` documentation (global and per-display-count shape) so newly created config files explain the setting. Verify by inspecting the auto-created file's contents in a unit test.

## 2. Hook resolution and execution

- [x] 2.1 Add a `resolveHooks(cfg *config.Config, displayCount int) (off, on []config.Command)` helper (in `cmd/mumu/cmd` or `internal/config`) that concatenates global + per-display-count arrays in the bracketing order (`off`: global then layout; `on`: layout then global). Verify with a unit test covering: only global configured, only per-display-count configured, both configured, and neither configured.
- [x] 2.2 Create `internal/hooks` package with a function that runs an ordered `[]config.Command` sequentially — `sh -c` for a `Shell` value, direct `exec.Command(Argv[0], Argv[1:]...)` for an `Argv` value — streaming each command's stdout/stderr to the given writers and invoking a progress callback per command (start, and success/failure with exit status). A failing or unstartable command is reported via the callback and does not stop the remaining commands. Verify with unit tests using real short-lived commands (e.g. `true`, `false`, `echo`) asserting: successful commands run in order, a failing command doesn't block subsequent ones, and both `Shell` and `Argv` forms execute correctly.

## 3. `mumu restore` integration

- [x] 3.1 In `layoutRestoreCmd`, add a `--no-hooks` bool flag (following the existing `--yes`/`--sort` flag registration pattern in `init()`). Verify with a test asserting the flag is registered and defaults to `false`.
- [x] 3.2 In `layoutRestoreCmd.RunE`, after the no-saved-layout and drift-decline early returns and unless `--no-hooks` is set, resolve the current display count's effective `off`/`on` arrays via `resolveHooks` and run the `off` array (via the `internal/hooks` runner) before calling `layout.Restore`, then run the `on` array after it returns successfully. Verify with a test (using fake/echo commands configured via a temp config file) that `off` commands run before any window-move progress output and `on` commands run after, and that `--no-hooks` suppresses both.
- [x] 3.3 Confirm a failing `off` command does not prevent `layout.Restore` from being called, and a failing `on` command does not change `mumu restore`'s exit code contributed by the restore itself. Verify with a test using a command that exits non-zero.

## 4. `mumu show` preview

- [x] 4.1 Add a hook-preview section to `layoutShowCmd` that calls `resolveHooks` for the resolved display count and prints the effective `off` array followed by the effective `on` array (each command shown as configured — the shell string, or the argv list joined for display), without executing anything. Verify with a test asserting the printed output for a display count with configured global and per-display-count hooks, and that no hook section (or an empty one) appears when none are configured.

## 5. Documentation

- [x] 5.1 Update `docs/CONFIG_SCHEMA.md` with the `hooks` key (global `off`/`on`, per-display-count `layouts.<n>.off`/`.on`), the two command-entry forms (shell string vs. argv list), and an example. Verify by re-reading the rendered doc against the actual struct/validation.
- [x] 5.2 Update `docs/CLI.md` describing the off/on bracketing order, `--no-hooks`, and the `mumu show` hook preview. Verify by cross-checking against the `restore-hooks` spec's scenarios.
- [x] 5.3 Update `AGENTS.md` if hook commands introduce a new convention worth capturing (e.g. running arbitrary user-configured external commands, non-interactive/fast expectations). Verify by reviewing the section reads coherently with the rest of the file.

## 6. Verification

- [x] 6.1 Run `just fmt`, `just lint`, `just build`, and `just test`; confirm all four pass with no new warnings.
- [x] 6.2 Manually exercise `mumu restore` with global and per-display-count hooks configured (e.g. simple `echo` commands) against a real saved layout, confirming the run order matches `mumu show`'s preview and that `--no-hooks` suppresses execution.
