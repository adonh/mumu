## Why

Where mumu stores its data today is invisible and inflexible: the layouts directory is a hardcoded constant (`~/.local/share/mumu/layouts`, a Linux-style path baked into a macOS-only app), saved layouts are opaque per-display-count JSON files, and there is no way for a user to see or change any of this without reading source code. This closes [GitHub Issue #10](https://github.com/adonh/mumu/issues/10) by giving mumu an explicit, user-editable YAML configuration file, and by making the saved-layout data itself a single human-readable YAML file rather than hidden JSON blobs.

## What Changes

- Add a `config.yaml` settings file, auto-created with commented defaults on first use, so its location and contents are discoverable without reading source.
  - Location: `$XDG_CONFIG_HOME/mumu/config.yaml` if `XDG_CONFIG_HOME` is set, otherwise `~/Library/Application Support/mumu/config.yaml`.
  - Initial format holds a single setting: `data_dir`, the directory mumu uses for its data (see below).
- **BREAKING**: Replace the per-display-count JSON layout files (`~/.local/share/mumu/layouts/<n>.json`) with a single `layouts.yaml` file containing all saved layouts, keyed by display count. No migration of old JSON files is provided (mumu has no installed user base to preserve continuity for, per existing precedent in `internal/layout/persist.go`).
- **BREAKING**: Change the default data directory from `~/.local/share/mumu` to `$XDG_DATA_HOME/mumu` if `XDG_DATA_HOME` is set, otherwise `~/Library/Application Support/mumu` (the native macOS convention), colocating `layouts.yaml` with `config.yaml` by default.
- `data_dir` in `config.yaml` lets a user redirect where `layouts.yaml` is stored, so the previously hardcoded path becomes user-editable.

## Capabilities

### New Capabilities
- `configuration`: introduces the `config.yaml` settings file — its location, auto-creation, format, and the `data_dir` setting.

### Modified Capabilities
- `space-layout`: saved layouts move from per-display-count JSON files to a single `layouts.yaml` file, and the default data directory changes to follow the new `configuration` capability's `data_dir` setting.

## Impact

- `internal/layout/persist.go`: rewritten to read/write a single `layouts.yaml` keyed by display count instead of one JSON file per display count; resolves its directory via the new configuration capability instead of the hardcoded `DefaultDir` constant.
- New `internal/config` package: resolves the config file path, auto-creates it with defaults/comments if missing, and loads `data_dir`.
- `go.mod`: promotes `go.yaml.in/yaml/v3` from an indirect to a direct dependency (already present transitively).
- No CLI flag or command changes; `mumu save`/`restore`/`list`/`show`/`delete` behavior is unchanged except for the on-disk format and default location of their data.
