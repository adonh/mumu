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

- **`config.yaml`** — mumu's own settings, explicit and user-editable. Resolved as `$XDG_CONFIG_HOME/mumu/config.yaml` if `XDG_CONFIG_HOME` is set, otherwise `~/Library/Application Support/mumu/config.yaml` (see `internal/config`). It's auto-created with commented defaults the first time it's needed, and never overwritten afterward — if it exists, it's read as-is. It exposes `data_dir` plus the window-pinning settings `pins` (a map from display count to a list of pin rules), `pin_precedence`, the application-level fallback setting `default_spaces` (a map from display count to a list of `{bundle_id, space}` rules — like `pins` but with no `title`, since it targets an application's leftover unclaimed windows as a group rather than one specific window; see `internal/layout/pins.go`'s `defaultSpacesByBundle` and `restore.go`'s `planFallbackMoves`), and `hooks` (global and per-display-count `off`/`on` external-command arrays run around `mumu restore`; see `internal/hooks`). This pattern (auto-create with commented defaults on first use, never silently overwrite, report a clear error rather than ignore malformed input) is the template to follow for any future user-facing settings file. `pins` is the first nested/list-shaped setting `config.yaml` has, keyed directly by an integer display count (mirroring how saved layouts are already keyed) rather than introducing a new naming scheme — follow that precedent for future per-display-count settings. `hooks` commands are user-authored, arbitrary external processes (shell strings or argv lists) that `mumu restore` runs synchronously with no timeout and live-streamed output — the same trust boundary the rest of `config.yaml` sits inside, but the only setting that actually executes what's written in it, so keep them non-interactive and fast (see `docs/CLI.md#restore-hooks`).
- **Saved layouts (`<data_dir>/layouts/<display-count>.json`)** — internal state, not a user-facing file. One JSON file per display count. `data_dir` resolves from `config.yaml`, defaulting to `$XDG_DATA_HOME/mumu` if set, otherwise `~/Library/Application Support/mumu` (colocated with `config.yaml` by default).

The split is deliberate: **JSON for internal state, YAML for user-facing config.** JSON doesn't invite hand-editing the way YAML does (no comments, less forgiving to hand-edit), which matches saved layouts being data mumu manages for itself. YAML's comment support and readability are what make it worth the extra format for something a user is meant to open and edit, like `config.yaml`.

**YAML indentation convention:** any YAML mumu writes uses two-space indentation, not the go-yaml library's 4-space default. `config.yaml` is currently a flat hand-written string, so this has no visible effect yet, but if a future setting needs an actual YAML encoder (nested structure, lists), configure its indent width to 2 explicitly rather than accepting the library default. See `internal/config/doc.go`.

See [`docs/CONFIG_SCHEMA.md`](docs/CONFIG_SCHEMA.md) for the full field-level schema of both files.

## Learned User Preferences

- When asked to prepare a clean branch/PR for one change while the working tree already has unrelated, tangled, uncommitted edits sharing the same files/lines: create the new branch first (carries over uncommitted changes as-is), then isolate the unrelated work into its own first commit by inverting just the session's own edits (replay each edit's old_string/new_string pair in reverse), then re-apply and commit the session's edits separately. This only works cleanly when edits were made as precisely-scoped `StrReplace` pairs rather than full-file rewrites, since only those are exactly invertible.

## Learned Workspace Facts

- `pin_precedence` (in `config.yaml`) defaults to `"layout"`, not `"pin"`: when a `pins:` rule and a saved-layout entry would both claim the same open window during `mumu restore`, the saved layout wins unless the user opts into pin-first via `pin_precedence: pin`. Rationale: pins are documented as filling in for windows a saved layout doesn't cover, so the richer/primary saved-layout artifact should win by default — pins overriding it by default was surprising and could silently starve saved-layout entries of their windows, misreporting a genuinely-running app as `application is not running`.
- `internal/layout/restore.go`'s `planDirectMoves`/`planLayoutPhase` take an `allLiveByBundle` parameter (the full, unfiltered live-window map) separately from whichever filtered pool that phase actually matches against, so restore can distinguish "app genuinely not running" (`SkipAppNotRunning`) from "app is running but every window was already claimed by the other precedence phase this restore" (`SkipWindowsClaimedElsewhere`, also listed in the CLI's ordered skip-reason list in `cmd/mumu/cmd/layout.go`). Don't collapse these two cases back into one skip reason.
- `gh` CLI is not authenticated in this dev environment (`gh repo view` / `gh issue create` / `gh pr create` fail with `HTTP 401`); when asked to create issues or PRs, write out the intended issue/PR text for the user to paste manually instead of assuming `gh` will succeed.
