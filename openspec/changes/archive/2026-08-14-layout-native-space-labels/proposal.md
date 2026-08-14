## Why

`mimi layout` reports Space numbers using its own logical left-to-right ordinal, which is deliberately different from macOS's own Mission Control (primary-display-first) ordering — the same ordering `mimi action space <n>` and `mimi action move_window_to_space <n>` use. When a user's primary display isn't the leftmost one, these two numbers diverge (e.g. logical Space 3 might be macOS's Mission Control Space 21), so a user cross-referencing `mimi layout` output against Mission Control, or against `mimi action space <n>`, has no way to tell which physical Space a logical number actually refers to.

## What Changes

- Wherever `mimi layout` output currently shows a logical Space number for a specific window entry (`mimi layout restore`'s per-window move progress and skipped-entry summary, and `mimi layout show`), also show the corresponding macOS Mission Control Space number (the same numbering `mimi action space <n>` uses), so the two numbering systems are never shown in isolation.
- Add a Go-level reverse lookup — Mission Control Space ID → 1-based Mission Control ordinal — mirroring the existing logical reverse lookup (`space.LogicalIndexForSpace`), backed by a new native function alongside the existing `MimiMissionControlSpaceID`/`MimiLogicalIndexForSpace` pair.
- No change to what's persisted in a saved layout file (still just the logical ordinal) — this only affects what's printed to the terminal. The Mission Control number is resolved fresh at print time, since it depends on the live display arrangement.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `space-layout`: adds a requirement that `mimi layout` output displaying a logical Space number must also display the corresponding macOS Mission Control Space number.

## Impact

- `internal/native/space.m` / `internal/native/mimi.h`: new native function resolving a Space ID to its 1-based Mission Control (primary-first) ordinal.
- `internal/space`: new Go wrapper exposing that reverse lookup.
- `cmd/mimi/cmd/layout.go`: `mimi layout show` output and the restore skip summary formatting.
- `internal/layout/restore.go`: per-window move progress messages.
- No change to `mimi layout save`'s output — it reports counts only, with no per-window Space numbers.
- `docs/CLI.md`: note the dual-numbering in output examples/description.
- No changes to the saved layout JSON schema or to `mimi action space`/`move_window_to_space` behavior.
