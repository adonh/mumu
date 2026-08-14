## Why

The project has grown two largely independent halves: a background daemon (hooks, window/app/workspace observation, the systray, and immediate `action` commands routed over IPC) and a self-contained layout save/restore feature that never touches any of that infrastructure. Keeping the daemon half around only adds unused surface area, permission requirements, and documentation for a capability the project no longer wants to maintain — and the shared `mimi` name risks confusion with an unrelated, pre-existing `mimi` app. Narrowing the project to layout save/restore only, under a distinct name, lets the codebase, CLI surface, and packaging all shrink to match what's actually used.

## What Changes

- **BREAKING**: Remove the daemon and everything that only exists to support it: `mimi start`/`stop`/`status`/`services`/`config`/`action *` commands, the `--config`/`-c` global flag and config-file support, hooks, window/app/workspace event observation, the systray, and the IPC server/client. None of this is used by layout.
- **BREAKING**: Flatten the CLI: drop the `layout` subcommand grouping. `mimi layout save` → `mumu save`, `mimi layout restore` → `mumu restore`, `mimi layout show` → `mumu show`, `mimi layout list` → `mumu list`, `mimi layout delete` → `mumu delete`. Existing flags on each (`--yes`/`-y` on restore, `--sort` on show and restore) carry over unchanged.
- **BREAKING**: Rename the project and executable from `mimi` to `mumu`: new Go module path, renamed binary/build output, root command name and version banner, and a full sweep of README, CHANGELOG (going forward), docs, nix packaging, `justfile`, and the app bundle's `Info.plist.template` to match.
- Keep showing both Space numbers in layout output (mumu's own left-to-right ordinal and the macOS Mission Control ordinal) even though the `action space <n>` command that previously gave the second number an obvious purpose is gone — it's still the same ordinal macOS's own Ctrl+`<n>` "Switch to Desktop" shortcut uses, so it remains meaningful on its own. Its description no longer references a removed command name.
- A minimal permissions-only `mumu status` is kept (Accessibility + Screen Recording, no daemon/IPC info) since layout's own permission checks already need this logic and it's useful to check without attempting a real save/restore.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `space-layout`: command names and examples change from `mimi layout <verb>` to flat `mumu <verb>` commands; permission-error guidance now names `mumu` instead of `mimi`; the requirement describing the dual Space-number display no longer references the removed `mimi action space`/`move_window_to_space` commands, but keeps the same displayed behavior.

## Impact

- Go module path changes from `github.com/y3owk1n/mimi` to `github.com/adonh/mumu`, touching every internal import across the codebase.
- `cmd/mimi/` is renamed to `cmd/mumu/` (binary name, root `Use` field, version banner text, `cmd/genman` man-page generation).
- Deleted entirely: `internal/daemon`, `internal/hooks`, `internal/observe`, `internal/events`, `internal/systray`, `internal/ipc`, `internal/action`, `internal/config`, their native (`.m`/`.h`) counterparts, and their tests.
- Deleted from `cmd/mimi/cmd`: `action.go`, `action_runner.go`, `start.go`, `stop.go`, `services.go`, `config.go`, `configpath.go`, and any now-unused parts of `runtime_paths.go`.
- `layout.go`'s subcommands become top-level commands directly on the root command (no `layout` parent); `status.go` is slimmed to a permissions-only check.
- `internal/permissions`: remove the full `Check()`/config-onboarding/accessibility-startup-alert functions that only served the removed daemon path; keep (and likely rename, since it becomes the only check) the Accessibility + Screen Recording check layout already uses.
- `configs/default-config.toml` and hook/config-related docs sections are removed; `docs/CLI.md` and other docs are rewritten for the flattened command surface and new name.
- Full rename sweep: `README.md`, `CHANGELOG.md` (going forward), `docs/*.md`, `nix/*.nix`, `flake.nix`, `justfile`, `resources/Info.plist.template`, and any `.github/workflows/*` referencing the `mimi` binary name or paths.
- `openspec/specs/space-layout/spec.md`: requirement text/examples updated to `mumu` command names; the dual Space-number requirement's description rewritten to not name a removed command.
