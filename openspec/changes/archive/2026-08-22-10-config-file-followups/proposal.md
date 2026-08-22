## Why

Two decisions from the `10-yaml-config-file` change need revisiting now that it's landed: saved layouts were made a single hand-editable YAML file, but they're internal data a user shouldn't be encouraged to edit by hand, and the YAML encoder used a 4-space indent when 2-space has become the more standard convention. Both were captured in the moment but don't reflect what should ship.

## What Changes

- **BREAKING**: Revert saved layouts from a single `layouts.yaml` (keyed by display count) back to one JSON file per display count (`<n>.json`), restoring the original pre-`10-yaml-config-file` on-disk shape and its "internal, not for hand-editing" framing. These files now live in a `layouts` subdirectory of the configured `data_dir` (assumption: preserves the `data_dir` configurability that was the point of `10-yaml-config-file`, e.g. `<data_dir>/layouts/<n>.json`), rather than the old hardcoded `~/.local/share/mumu/layouts`.
- Drop the "safe to hand-edit" framing and guidance for saved layouts from docs (`docs/CONFIG_SCHEMA.md`, `docs/examples/`) — layouts are internal state, not a user-facing editable file. `config.yaml` remains user-editable.
- Codify 2-space YAML indentation as mumu's project-wide convention for any YAML mumu writes (documented in `AGENTS.md`, applied via the YAML encoder's indent setting), so `config.yaml` and any future nested config settings follow it. There is no currently-shipping nested YAML output this visibly changes, since `config.yaml` is flat today.
- Update `AGENTS.md` with the guidelines and conventions this effort (the `10-yaml-config-file` change plus this follow-up) established: where config and data live, the config-file auto-creation/error-handling pattern, and the YAML-indent convention.

## Capabilities

### Modified Capabilities

- `space-layout`: "Saved layouts persisted as a single editable YAML file" reverts to per-display-count JSON files; hand-editability and YAML-specific error reporting requirements for saved layouts are removed. The requirement that saved layouts live under the configured `data_dir` is preserved (new: within a `layouts` subdirectory of it).
- `configuration`: no requirement text changes to `data_dir` resolution itself, but the description of what `data_dir` is used for (previously "layouts.yaml") is updated to reflect the reverted layout storage shape. Adds a project-wide YAML output convention (2-space indent) that any YAML file mumu writes — currently just `config.yaml` — must follow.

## Impact

- `internal/layout/persist.go`, `internal/layout/persist_test.go`, `internal/layout/types.go`: revert to per-display-count JSON persistence (informed by, not identical to, the pre-`10-yaml-config-file` version, since `data_dir` configurability is kept).
- `internal/config/config.go`: no behavior change expected, but its default-file-writing path becomes the reference implementation for the 2-space YAML convention if/when it starts using a YAML encoder for nested settings.
- `docs/CONFIG_SCHEMA.md`, `docs/examples/config.yaml`, `docs/examples/layouts.yaml` (removed or replaced with a `.json` example), `docs/ARCHITECTURE.md`: updated to match.
- `AGENTS.md`: new/updated guidance section.
- No CLI-facing behavior changes beyond the on-disk file format of saved layouts.
