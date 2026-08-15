## Why

`mumu delete` currently removes a saved layout immediately, with no confirmation and no way to recover it. A saved layout can represent significant manual setup effort, so an accidental or mistyped `mumu delete` (e.g. wrong display count) is easy to run and impossible to undo. `mumu restore` already prompts before a potentially surprising action (a display-arrangement mismatch) and offers `--yes` to skip it for scripting; `delete` should follow the same pattern.

## What Changes

- `mumu delete [display-count]` now prompts the user to confirm before removing a saved layout, showing which display count's layout is about to be deleted.
- Add a `--yes`/`-y` flag to `mumu delete` that skips the confirmation prompt (for scripting), mirroring `mumu restore --yes`.
- If the user declines the prompt, the layout is left untouched and the command exits without error, printing that the deletion was aborted.
- If no saved layout exists for the resolved display count, the command still fails the same way it does today (no layout to confirm deleting), without prompting.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `space-layout`: The "Layout management commands" requirement changes so `mumu delete` requires interactive confirmation before removing a saved layout, unless `--yes`/`-y` is passed.

## Impact

- Affected code: `cmd/mumu/cmd/layout.go` (`layoutDeleteCmd`), reusing the existing `promptConfirm` helper.
- Affected docs: `docs/CLI.md` (`mumu delete` section) to document the prompt and new flag.
- No changes to `internal/layout` persistence logic — `layout.Delete` behavior is unchanged, only when it's called changes.
