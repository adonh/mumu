# CLI Usage

mimi is a macOS window and space utility. Use `mimi action` for immediate commands, or `mimi start` to run the hook daemon.

---

## Table of Contents

- [Global Flags](#global-flags)
- [Window & Space Actions](#window--space-actions)
- [Layout Save/Restore](#layout-saverestore)
- [Hook Daemon](#hook-daemon)
- [Service Management](#service-management)
- [Configuration Management](#configuration-management)

---

## Global Flags

| Flag            | Shorthand | Default | Description            |
| --------------- | --------- | ------- | ---------------------- |
| `--config, -c`  |           | auto    | Path to config file    |
| `--verbose, -v` |           | `false` | Verbose output         |
| `--version`     |           |         | Print version and exit |

---

## Window & Space Actions

These commands run directly in the CLI process when the daemon is not running. When the daemon **is** running, mimi routes actions over its Unix socket (`settings.socket_file`, default `~/.local/share/mimi/mimi.sock`) so hotkeys feel instant. **Accessibility permission is required.**

```bash
mimi action focus_window
mimi action focus_window --backward
mimi action focus_window --left
mimi action focus_window --right
mimi action focus_window --up
mimi action focus_window --down
mimi action space 1
mimi action space next
mimi action space prev
mimi action move_window_to_space 2
mimi action move_window_to_space next
mimi action move_window_to_space prev
mimi action resize_window left-half
mimi action resize_window center --width-percent 80 --height-percent 90
mimi action resize_window --width 1024 --height 768 --anchor cc
```

### `mimi action focus_window`

Cycle keyboard focus through all focusable windows on the current space, or move focus spatially with direction flags.

| Flag         | Description                                                      |
| ------------ | ---------------------------------------------------------------- |
| `--backward` | Cycle to the previous window instead of the next                 |
| `--up`       | Move focus to the nearest window above the current one           |
| `--down`     | Move focus to the nearest window below the current one           |
| `--left`     | Move focus to the nearest window to the left of the current one  |
| `--right`    | Move focus to the nearest window to the right of the current one |

### `mimi action space <number|next|prev>`

Focus a Mission Control space by its 1-based index, or cycle to the next/previous space with wrapping. Uses a synthetic dock-swipe gesture (no public macOS API exists for direct space switching).

### `mimi action move_window_to_space <number|next|prev>`

Move the frontmost window to a space by its 1-based index, or cycle to the next/previous space with wrapping. Uses private SkyLight APIs; does not require disabling SIP.

### `mimi action resize_window [preset] [flags]`

Resize and reposition the frontmost window using presets or custom flags. Respects the macOS tiled window margins setting (`com.apple.WindowManager.EnableTiledWindowMargins`), applying full margins on screen-facing edges and half margins on internal (split) edges.

**Presets** provide quick tiling layouts:

| Preset         | Effect                               |
| -------------- | ------------------------------------ |
| `left-half`    | Fill the left half of the screen     |
| `right-half`   | Fill the right half of the screen    |
| `top-half`     | Fill the top half of the screen      |
| `bottom-half`  | Fill the bottom half of the screen   |
| `top-left`     | Fill the top-left quadrant           |
| `top-right`    | Fill the top-right quadrant          |
| `bottom-left`  | Fill the bottom-left quadrant        |
| `bottom-right` | Fill the bottom-right quadrant       |
| `center`       | Center window at 60% × 80% of screen |
| `fill`         | Fill entire screen                   |

**Custom sizing flags:**

| Flag                     | Description                            |
| ------------------------ | -------------------------------------- |
| `--width, -w <pixels>`   | Absolute window width in points        |
| `--height, -h <pixels>`  | Absolute window height in points       |
| `--width-percent <pct>`  | Width as percentage of screen (0–100)  |
| `--height-percent <pct>` | Height as percentage of screen (0–100) |

**Positioning flags** (use anchors to align the window):

| Flag           | Description              |
| -------------- | ------------------------ |
| `--x <pixels>` | Absolute X position      |
| `--y <pixels>` | Absolute Y position      |
| `--anchor, -a` | Anchor point (see below) |

**Anchor system** places the window's anchor point at the computed or specified position. Valid anchors (use 2 letters: vertical + horizontal):

```
tl  tc  tr       (top-left, top-center, top-right)
cl  cc  cr       (center-left, center-center, center-right)
bl  bc  br       (bottom-left, bottom-center, bottom-right)
```

**Margin control:**

| Flag          | Effect                                          |
| ------------- | ----------------------------------------------- |
| `--margin`    | Enable tiled margins (overrides system setting) |
| `--no-margin` | Disable tiled margins                           |

**Examples:**

```bash
# Presets
mimi action resize_window left-half
mimi action resize_window right-half
mimi action resize_window top-left
mimi action resize_window center

# Fixed dimensions, centered
mimi action resize_window --width 800 --height 600 --anchor cc

# Percentage of screen, top-left
mimi action resize_window --width-percent 50 --height-percent 75 --anchor tl

# Absolute position, top-left anchor
mimi action resize_window --width 1024 --height 768 --x 100 --y 50 --anchor tl

# Override margins for a preset
mimi action resize_window left-half --no-margin

# Mix preset with custom size
mimi action resize_window center --width-percent 80 --height-percent 90
```

---

## Layout Save/Restore

`mimi layout` saves and restores the assignment of application windows to Mission Control spaces, keyed by the number of currently connected displays. **Requires both Accessibility and Screen Recording permissions** — Screen Recording is needed to read window titles reliably (see [Limitations](#limitations) below); no other mimi command requires it.

```bash
mimi layout save
mimi layout restore
mimi layout restore --yes
mimi layout list
mimi layout show
mimi layout show 2
mimi layout delete
mimi layout delete 2
```

### Space numbering

Space numbers used by `mimi layout` are counted **left to right across all connected displays**, independent of which display is primary — matching how a person visually counts spaces on screen. This is a separate numbering from `mimi action space`'s Mission Control ordering (which always lists the primary display's spaces first) and does not affect it.

### `mimi layout save`

Captures, for every non-fullscreen window on every space across all connected displays, its owning application's bundle identifier, window title, and logical (left-to-right) space number. Persists the result keyed by the current display count, overwriting any previous save for that same count.

### `mimi layout restore [--yes]`

Auto-detects the current display count, loads the layout saved for it, and moves each matching, already-running application's window back to its recorded space. Applications that aren't running are skipped (never launched). Windows are matched by exact title first, falling back to positional order within the same app; if exactly one of an app's windows remains unmatched after that, it's used regardless of title or position — there's no real ambiguity left once only one candidate remains, which matters for apps like browsers whose titles rarely match exactly across save and restore. Never creates or removes spaces — entries whose target space no longer exists are skipped and reported.

If the current per-display space-count arrangement doesn't match what was recorded at save time, you'll be prompted to confirm before any windows move. Pass `--yes` (or `-y`) to skip the prompt (e.g. for scripting).

### `mimi layout list`

Lists every saved layout with its display count, window count, and save timestamp.

### `mimi layout show [display-count]`

Prints a saved layout's window entries (space number, bundle ID, title) without moving anything. Defaults to the layout for the current display count.

### `mimi layout delete [display-count]`

Deletes the saved layout for the given display count (default: current display count).

### Limitations

- **Reflow-only**: applications that have quit since save time are skipped and reported, never relaunched. This is not a full session restore.
- **No window geometry**: only space assignment is saved — position and size are never captured or restored.
- **Fullscreen windows are excluded** entirely, both at save and restore.
- **Minimized-window detection is best-effort**: a minimized window on the space currently displayed on its screen is correctly excluded, but a minimized window on a space that isn't currently displayed on any screen may still be captured and restored as if it weren't minimized — there's no reliable per-window signal for that case.
- **Manual only**: layout save/restore only happens when you run these commands directly. It's never triggered automatically by the hook daemon, on a schedule, or at login.

---

## Hook Daemon

### `mimi start`

Start the background daemon that watches window and space events and runs your hooks.

```bash
mimi start
mimi start -c /path/to/config.toml
```

On first run without a config file, mimi offers to create one.

### `mimi stop`

Stop the running daemon via SIGTERM.

```bash
mimi stop
```

### `mimi status`

Show whether the daemon is running, whether Accessibility and Screen Recording permissions are granted, and whether the IPC socket is available. Screen Recording is only required by `mimi layout`.

```bash
mimi status
```

---

## Service Management

### `mimi services install`

Install mimi as a launchd user agent for automatic startup at login.

```bash
mimi services install
```

### `mimi services uninstall`

Remove the launchd agent.

### `mimi services start` / `stop` / `restart` / `status`

Control the launchd service directly.

---

## Configuration Management

### `mimi config init`

Create a default config at `~/.config/mimi/config.toml`.

### `mimi config validate`

Parse and validate the config file.

### `mimi config dump`

Print the resolved config as JSON.

### `mimi config reload`

Send SIGHUP to a running daemon to reload config without restart.
