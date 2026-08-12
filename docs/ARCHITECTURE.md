# Architecture

mimi is a macOS window and space utility with four execution paths:

1. **CLI actions (direct)** — immediate one-shot commands (`mimi action …`)
2. **CLI actions (via daemon IPC)** — same commands routed over a Unix socket when the daemon is running
3. **Hook daemon** — background process that fires shell hooks on app, window, and space events
4. **Layout save/restore** — one-shot commands that snapshot and reflow window-to-Space assignments (`mimi layout …`)

All paths use native macOS APIs via CGO. No SIP disable is required.

---

## CLI Actions

```
mimi action <subcommand>
  → internal/action
  → internal/window / internal/space
  → internal/native (Objective-C + SkyLight)
```

| Action | API |
| ------ | --- |
| `focus_window` | Accessibility (`AXUIElement`) |
| `space` | Synthetic dock-swipe gesture via `CGEvent` |
| `move_window_to_space` | Private SkyLight (`SLSMoveWindowsToManagedSpace`) |

CLI actions pump the run loop briefly after posting events so gestures complete before the process exits.

When the daemon is running, `mimi action` first tries the Unix socket at `settings.socket_file`. The daemon executes the action on a dedicated OS thread and returns the result. If the socket is unavailable, the CLI falls back to direct execution.

---

## Hook Daemon

```
NSWorkspace + AX observers (workspace.m, axobserver.m)
  → internal/native (Go exports)
  → internal/observe (event router)
  → events.Bus
  → hooks.Executor
  → shell commands
```

### Observers

- **App lifecycle** — subscribes to `NSWorkspace` app notifications (activate, deactivate, launch, quit, hide, unhide) for both app hooks and AX observer management
- **AX window events** — focus, title change, create, close, resize (debounced)
- **Workspace polling** — detects Mission Control space changes when `on_workspace_changed` hooks are configured

### Event Bus

Non-blocking pub-sub fan-out. Subscribers: hook executor and optional event log writer.

### Hook Executor

Matches events against configured hooks, applies filters (`app`, `bundle_id`, `title`), runs shell commands with `mimi_*` environment variables.

---

## Layout Save/Restore

```
mimi layout <subcommand>
  → internal/layout (capture, persist, restore)
  → internal/window / internal/space
  → internal/native (CGWindowList + private SkyLight APIs)
```

Unlike `mimi action`, layout capture enumerates windows via `CGWindowListCopyWindowInfo(kCGWindowListOptionAll)` rather than the Accessibility API: AX only exposes windows on each display's *currently displayed* Space, so windows on other Spaces are AX-invisible. The CGWindowList result is filtered to document-layer windows (`kCGWindowLayer == 0`) owned by regular (non-background) apps, then resolved to a Space ID via the private `CGSCopySpacesForWindows` API. This is also why layout commands additionally require Screen Recording — window titles from `kCGWindowName` are redacted without it, even though Accessibility is already granted.

Space numbers for layout are "logical left-to-right": each display's Spaces (already primary-first via `SLSCopyManagedDisplaySpaces`) are concatenated in physical left-to-right display order (`CGDisplayBounds.origin.x`), independent of which display macOS considers primary. This differs from `mimi action space`'s numbering, which always lists the primary display's Spaces first, and is scoped entirely to `mimi layout`.

Layouts are persisted as JSON under `~/.local/share/mimi/layouts/<display-count>.json`, keyed by the number of currently connected displays. Restore only moves windows belonging to already-running applications to already-existing Spaces — it never launches apps, and never creates or removes Spaces.

---

## Package Layout

```
cmd/mimi/           CLI entry point and commands
internal/
  action/           Action dispatch (focus_window, space, move_window_to_space)
  window/           Go wrappers for AX window APIs
  space/            Mission Control space operations, logical left-to-right numbering
  layout/           Layout save/restore: capture, JSON persistence, restore matching
  native/           All Objective-C + CGO (actions, observers, layout enumeration)
  observe/          Hook daemon event routing
  hooks/            Hook registry and executor
  config/           TOML config loading
  daemon/           Daemon lifecycle
  permissions/      Accessibility and Screen Recording permission checks
  systray/          Optional menu bar UI
```

---

## Permissions

**Accessibility** is required for:

- All `mimi action` commands
- Window hooks (`on_window_*`)
- All `mimi layout` commands

App lifecycle hooks (`on_app_*`) and workspace hooks (`on_workspace_changed`) do not require Accessibility.

**Screen Recording** is required only for `mimi layout` commands, in addition to Accessibility — window titles used for restore matching come back redacted without it. `mimi status` reports both permissions.

---

## Platform Notes

Space switching and window-to-space moves use undocumented private APIs that may break on macOS updates. They are provided as-is for personal automation workflows, not as guaranteed-stable APIs.
