# Architecture

`mumu` is a one-shot CLI: every command runs, does its work, and exits. There is no daemon, no background process, and no persistent state beyond the saved layout files themselves.

```
mumu <command>
  → internal/layout (capture, persist, restore)
  → internal/window / internal/space
  → internal/native (CGWindowList + private SkyLight APIs, Objective-C + CGO)
```

All paths use native macOS APIs via CGO. No SIP disable is required.

---

## Layout Save/Restore

Layout capture enumerates windows via `CGWindowListCopyWindowInfo(kCGWindowListOptionAll)` rather than the Accessibility API: AX only exposes windows on each display's _currently displayed_ Space, so windows on other Spaces are AX-invisible. The CGWindowList result is filtered to document-layer windows (`kCGWindowLayer == 0`) owned by regular (non-background) apps, then resolved to a Space ID via the private `CGSCopySpacesForWindows` API. This is also why layout commands require Screen Recording in addition to Accessibility — window titles from `kCGWindowName` are redacted without it.

Space numbers are a two-part `<display>:<space>` ordinal, scoped per display: displays are ordered left to right by physical position (`CGDisplayBounds.origin.x`), and each display's own Spaces (already primary-first via `SLSCopyManagedDisplaySpaces`) are numbered left to right within that display alone — independent of which display macOS considers primary, and independent of every other display's Space count, so adding or removing a Space on one display never renumbers another display's Spaces. This differs from the macOS Mission Control ordering, which always lists the primary display's Spaces first as one flat sequence; `mumu` resolves and displays both numbers side by side (see [CLI Guide — Space Numbering](CLI.md#space-numbering)).

Layouts are persisted as internal state: one JSON file per display count (`<display-count>.json`), inside a `layouts` subdirectory of mumu's data directory (`data_dir`). Restore only moves windows belonging to already-running applications to already-existing Spaces — it never launches apps, and never creates or removes Spaces. These files aren't meant for hand-editing — see [`docs/CONFIG_SCHEMA.md`](CONFIG_SCHEMA.md) for the full field-level schema.

`mumu`'s own settings live in an explicit, user-editable `config.yaml`, resolved as `$XDG_CONFIG_HOME/mumu/config.yaml` if `XDG_CONFIG_HOME` is set, otherwise `~/Library/Application Support/mumu/config.yaml`. It's auto-created with commented defaults on first use, including `data_dir` — the directory whose `layouts/` subdirectory holds saved layouts, which defaults to `$XDG_DATA_HOME/mumu` if `XDG_DATA_HOME` is set, otherwise `~/Library/Application Support/mumu` (colocated with `config.yaml` by default). Editing `data_dir` moves where mumu reads and writes saved layouts. Any YAML mumu writes, including `config.yaml`, uses two-space indentation. See [`docs/examples/config.yaml`](examples/config.yaml) for a sample and [`docs/CONFIG_SCHEMA.md`](CONFIG_SCHEMA.md) for the schema.

Space-to-Space window moves use the private SkyLight API (`SLSMoveWindowsToManagedSpace` and friends) for instant, animation-free relocation, without requiring an `AXUIElementRef` (layout-discovered windows come from CGWindowList, identified by `CGWindowID`, not AX).

---

## Package Layout

```
cmd/mumu/           CLI entry point and commands
internal/
  layout/           Layout save/restore: capture, JSON persistence, restore matching
  config/           config.yaml resolution, auto-creation, and loading
  window/           Go wrappers for AX window APIs and CGWindowList-based window moves
  space/            Mission Control space operations, logical left-to-right numbering
  native/           Objective-C + CGO bridge (window/space APIs, layout enumeration)
  permissions/      Accessibility and Screen Recording permission checks
  errors/           Structured error types
  paths/            Home-directory path expansion
```

---

## Permissions

`mumu save` and `mumu restore` check for both **Accessibility** and **Screen Recording** before doing any work, since both are required to enumerate windows and read their titles reliably. `mumu show`, `list`, and `delete` only read or write the local saved-layout JSON files and don't perform a permission check. `mumu status` reports both permissions without touching any window, Space, or layout file.

---

## Platform Notes

Space switching and window-to-Space moves use undocumented private APIs that may break on macOS updates. They are provided as-is for personal automation workflows, not as guaranteed-stable APIs.
