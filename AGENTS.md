# Agent Guidelines

Guidance for AI agents (and humans) working in this repository. See [`docs/CODING_STANDARDS.md`](docs/CODING_STANDARDS.md) for detailed code style, and [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for how the pieces fit together.

## Build, lint, test

Use the `just` recipes rather than invoking `go build`/`go test` directly, so version metadata and CGO flags stay consistent:

- `just build` — build the `mumu` binary
- `just fmt` — format Go and Objective-C files
- `just lint` — run `golangci-lint`
- `just test` — run the test suite

Run all four before considering a change done; see the pre-commit checklist in `docs/CODING_STANDARDS.md`.

## Configuration and data files

mumu splits its on-disk state into two categories with deliberately different formats:

- **`config.yaml`** — mumu's own settings, explicit and user-editable. Resolved as `$XDG_CONFIG_HOME/mumu/config.yaml` if `XDG_CONFIG_HOME` is set, otherwise `~/Library/Application Support/mumu/config.yaml` (see `internal/config`). It's auto-created with commented defaults the first time it's needed, and never overwritten afterward — if it exists, it's read as-is. It currently exposes one setting, `data_dir`. This pattern (auto-create with commented defaults on first use, never silently overwrite, report a clear error rather than ignore malformed input) is the template to follow for any future user-facing settings file.
- **Saved layouts (`<data_dir>/layouts/<display-count>.json`)** — internal state, not a user-facing file. One JSON file per display count. `data_dir` resolves from `config.yaml`, defaulting to `$XDG_DATA_HOME/mumu` if set, otherwise `~/Library/Application Support/mumu` (colocated with `config.yaml` by default).

The split is deliberate: **JSON for internal state, YAML for user-facing config.** JSON doesn't invite hand-editing the way YAML does (no comments, less forgiving to hand-edit), which matches saved layouts being data mumu manages for itself. YAML's comment support and readability are what make it worth the extra format for something a user is meant to open and edit, like `config.yaml`.

**YAML indentation convention:** any YAML mumu writes uses two-space indentation, not the go-yaml library's 4-space default. `config.yaml` is currently a flat hand-written string, so this has no visible effect yet, but if a future setting needs an actual YAML encoder (nested structure, lists), configure its indent width to 2 explicitly rather than accepting the library default. See `internal/config/doc.go`.

See [`docs/CONFIG_SCHEMA.md`](docs/CONFIG_SCHEMA.md) for the full field-level schema of both files.
