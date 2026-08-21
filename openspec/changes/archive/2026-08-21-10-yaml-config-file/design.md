## Context

Today, `internal/layout/persist.go` hardcodes `DefaultDir = "~/.local/share/mumu/layouts"` and writes one `<displayCount>.json` file per saved layout there. There is no settings file of any kind, and `go.yaml.in/yaml/v3` is already present in `go.mod` as an indirect dependency (pulled in transitively), so adding YAML support needs no new external dependency, just promoting it to direct. See `proposal.md` for motivation and `specs/configuration/spec.md` / `specs/space-layout/spec.md` for the resulting behavior contract.

## Goals / Non-Goals

**Goals:**
- Give users a discoverable, editable `config.yaml` for settings, starting with where mumu stores its data.
- Replace opaque per-display-count JSON layout files with a single, hand-editable `layouts.yaml`.
- Default both files to a sensible, macOS-appropriate location without requiring any user action.

**Non-Goals:**
- Migrating existing `~/.local/share/mumu/layouts/*.json` files to the new format or location (no installed user base to preserve, per existing precedent in `persist.go`'s doc comment).
- Adding any config settings beyond `data_dir` in this change.
- Changing any CLI command, flag, or output format beyond what's needed to reflect the new storage format.
- Supporting Linux/XDG as a first-class target — the `$XDG_*` checks are a courtesy for developers/testers who set them, not a claim of Linux support (mumu remains macOS-only per its use of Accessibility/Mission Control/native bridges).

## Decisions

**One new `internal/config` package owns both settings and data-path resolution.** It exposes a `Load() (*Config, error)` that: resolves the config file path (`$XDG_CONFIG_HOME/mumu/config.yaml` or `~/Library/Application Support/mumu/config.yaml`), creates it with commented defaults if absent, parses it, validates `data_dir`, and returns a `Config{DataDir string}` with `~` already expanded. `internal/layout` depends on `internal/config` for its directory instead of a hardcoded constant.
_Alternative considered_: keep path resolution logic inside `internal/layout` and just add YAML parsing there. Rejected — config resolution (XDG rules, auto-creation, comments) is generic and will be reused by any future setting, whereas `internal/layout` should only know it needs "a data directory."

**`layouts.yaml` lives inside the resolved `data_dir`** (e.g. `~/Library/Application Support/mumu/layouts.yaml`), not in a separate `layouts/` subdirectory as before. A single file matches the new "one editable YAML file" requirement and avoids the awkwardness of a directory containing a single file per key when a top-level map does the same job more simply.
_Alternative considered_: keep a `layouts/` subdirectory but with one `.yaml` file per display count (mirroring the old JSON layout). Rejected — the issue specifically asks for something a user can look at and edit as a whole, and a single file is easier to reason about and back up than N files.

**`layouts.yaml` schema is a top-level map keyed by display count (as a string key, since YAML map keys round-trip more predictably as strings across encoders), each value being the existing `Layout` struct's fields minus the now-redundant `displayCount` field** (it's the map key):

```yaml
schemaVersion: 1
layouts:
  "2":
    spaceCounts: [2, 3]
    savedAt: 2026-08-20T10:15:00Z
    entries:
      - bundleId: com.google.Chrome
        title: "Inbox - user@example.com - Gmail"
        index: 0
        ordinal: 1
      - bundleId: com.apple.Terminal
        title: "zsh - 80x24"
        index: 0
        ordinal: 2
  "3":
    spaceCounts: [2, 2, 3]
    savedAt: 2026-08-19T09:02:00Z
    entries: []
```

A top-level `schemaVersion` (matching the existing constant in `internal/layout/types.go`) is kept at the file level rather than per-entry, since the whole file is rewritten together on every `Save` and there's no scenario where different keyed layouts in the same file would be on different schema versions.
_Alternative considered_: keep `displayCount` duplicated inside each entry (as the JSON version did, since it doubled as the JSON filename-derived value). Rejected as redundant once it's also the map key — a mismatch between the two would just be a new class of bug to guard against for no benefit.

**`config.yaml` schema is intentionally minimal for this change** — one key:

```yaml
# mumu configuration.
#
# data_dir: the directory mumu uses to store its data (currently just
# layouts.yaml, the saved window-to-Space layouts). Supports a leading
# "~" for your home directory. Defaults to $XDG_DATA_HOME/mumu if
# XDG_DATA_HOME is set, otherwise ~/Library/Application Support/mumu.
data_dir: ~/Library/Application Support/mumu
```

Auto-creation always writes the *resolved* default (expanding `$XDG_DATA_HOME` if applicable) as a literal value, not a placeholder — so the file is immediately valid, and a user editing it sees exactly where their data currently lives before changing it.
_Alternative considered_: leave `data_dir` commented out by default and only apply it when uncommented. Rejected — writing the resolved value directly better serves the issue's core complaint ("not sure where the current save data is going"): the answer is right there in the file, active, not commented out.

**Both `config.yaml`'s own directory and `data_dir`'s default value use the same XDG-aware-with-macOS-fallback rule, but resolved independently** (`XDG_CONFIG_HOME`/`~/Library/Application Support` for the config file itself; `XDG_DATA_HOME`/`~/Library/Application Support` for the default data directory). In the common case (no XDG env vars set) both land in the same `~/Library/Application Support/mumu/` directory, colocating `config.yaml` and `layouts.yaml`, which directly answers "where is my data" by making it sit next to the settings file that controls it.

**No migration of old data.** `~/.local/share/mumu/layouts/*.json` is simply abandoned; mumu never reads it after this change. This is safe per the existing precedent set when the app was rebranded (see `persist.go`'s current doc comment: "no installed user base to preserve continuity for").

## Risks / Trade-offs

- **BREAKING for any existing tester's saved layouts** → they silently stop appearing (`mumu list` reports none) after upgrading, needing a fresh `mumu save`. Mitigation: explicitly called out as a breaking change in the proposal and CHANGELOG; acceptable given no installed user base, per existing precedent.
- **Auto-creating `config.yaml` on first read is a side effect of what might otherwise look like a read-only command** (e.g. `mumu status` or `mumu list` before any `mumu save`). Mitigation: this is the whole point (issue asks for the file to be discoverable) and is explicitly specified in `specs/configuration/spec.md`; the created file only ever contains defaults, never surprises the user with unexpected values.
- **A malformed hand-edited `config.yaml` or `layouts.yaml` blocks every command that needs it**, not just the one being edited. Mitigation: fail with a clear, file-path-specific error per the spec's "Invalid config/malformed file" scenarios, rather than silently falling back to defaults (which would mask the user's edit being ignored).

## Migration Plan

No automated migration (see Decisions/Non-Goals). Rollout is: ship the new `internal/config` package and rewritten `internal/layout/persist.go`, update docs (`README.md`/`docs/CLI.md` if they mention the old path) to reference the new locations. Rollback is reverting the commit; no data-format downgrade path is needed since no data is migrated forward in the first place.

## Open Questions

- Whether future settings beyond `data_dir` belong in the same `config.yaml` or a namespaced structure — deferrable; the current one-key format doesn't foreclose either option.
