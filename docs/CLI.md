# CLI Usage

`mumu` saves and restores window-to-Space layouts on macOS. All commands are top-level — there is no subcommand grouping.

---

## Table of Contents

- [Global Flags](#global-flags)
- [Space Numbering](#space-numbering)
- [Output Ordering (`--sort`)](#output-ordering---sort)
- [Pinned Windows](#pinned-windows)
- [Default Spaces](#default-spaces)
- [Restore Hooks](#restore-hooks)
- [`mumu save`](#mumu-save)
- [`mumu restore`](#mumu-restore---yes---no-hooks---sort-logicalmacosapp)
- [`mumu list`](#mumu-list)
- [`mumu show`](#mumu-show-display-count---sort-logicalmacosapp)
- [`mumu delete`](#mumu-delete-display-count---yes)
- [`mumu status`](#mumu-status)
- [Limitations](#limitations)

---

## Global Flags

| Flag        | Description             |
| ----------- | ------------------------ |
| `--version` | Print version and exit   |
| `--help`    | Show help for any command |

---

## Space Numbering

`mumu`'s own ordinal is a two-part `<display>:<space>` pair (shown as `#2:01`, `#1:21`, etc.): the display ordinal counts connected displays **left to right**, and the space ordinal counts that display's Spaces left to right — independent of which display is primary, and independent of every other display's Space count, matching how a person visually counts Spaces on screen. Because numbering is scoped per display, adding or removing a Space on one display never renumbers another display's Spaces.

Because this can diverge from macOS's own Mission Control ordering whenever the primary display isn't the leftmost one, any output that names a specific window's Space (`show`, restore's per-window progress, and its skip summary) shows both numbers together, e.g. `#2:01 (space 21)` — mumu's own two-part ordinal first, then the macOS Mission Control Space number in parentheses, which is the same ordinal macOS's own "Switch to Desktop `<n>`" keyboard shortcut uses. The Mission Control number is resolved fresh against the current display arrangement each time it's printed, so it always reflects what's true right now (it's never saved to the layout file).

## Output ordering (`--sort`)

`mumu show` and `mumu restore` both accept `--sort <logical|macos|app>`, controlling the order per-window entries print in (and, for `restore`, the order windows are actually moved in — it has no effect on _which_ windows get matched, moved, or skipped):

- `logical` (default): mumu's own logical left-to-right Space number, ascending.
- `macos`: macOS Mission Control Space number, ascending.
- `app`: bundle identifier, ascending (alphabetical), grouping all of one application's windows together. mumu has no captured or resolvable human-readable app display name, so this sorts by the raw bundle ID string (e.g. `com.apple.Safari`), not a "friendly" app name.

Whichever key is primary, entries that tie on it are ordered by falling back through the other two keys, in the fixed priority Space number, then bundle identifier, then window title — so output is always fully deterministic. This ordering applies to `show`'s entry list, restore's per-window move progress lines, and the ordering of entries within each reason group of restore's skip summary.

## Pinned Windows

`config.yaml`'s `pins` setting lets you declare fixed application-window-to-Space assignments per display count — e.g. always keep Slack on Space 1 — without needing to re-save a layout whenever it drifts. See [Configuration Schema — `config.yaml`](CONFIG_SCHEMA.md#configyaml) for the full field reference.

Pins only take effect as part of `mumu restore` — there's no separate command to apply them, and they never launch an app or take effect without a saved layout for the current display count (restore still requires one). Restore matches a display count's configured pins against that application's currently open windows using the exact same approximate title-matching `mumu restore` already uses for saved layouts, then moves matched windows to their pinned Space, subject to the same "app must be running" and "target Space must exist" rules. `config.yaml`'s `pin_precedence` setting (`pin`, or `layout`, the default) controls which wins — pins or the saved layout — when both would otherwise claim the same open window.

`mumu show` also lists a display count's configured pins (as written, without matching them against any open window) alongside its saved layout.

## Default Spaces

`config.yaml`'s `default_spaces` setting lets you declare a fixed, application-level fallback Space per display count — e.g. always send any leftover Slack windows to Space 1 — for windows `mumu restore` doesn't otherwise place via a pin or a saved-layout match. See [Configuration Schema — `config.yaml`](CONFIG_SCHEMA.md#configyaml) for the full field reference.

Unlike a pin, a default-space rule has no title pattern — it's application-level, not per-window. Normally, an application's leftover unclaimed windows go to that application's most prevalent matched Space this restore (or the primary display's current Space on a tie), and only if it has at least one valid saved-entry match; an application with zero matches is left unchanged. A configured `default_spaces` rule for an application changes both of those: its target **always** wins over the prevalent-Space heuristic, even when that heuristic is unambiguous, and it activates a fallback placement even when the application has zero valid saved-entry matches this restore. Restore progress output marks a configured-default placement `(default)`, distinct from the heuristic's `(fallback)` marker, so you can tell which one placed a given window.

`mumu show` also lists a display count's configured default spaces (as written, without matching them against any open window) alongside its saved layout, pins, and hooks.

## Restore Hooks

`config.yaml`'s `hooks` setting lets you configure external commands to run automatically around every `mumu restore` — e.g. muting audio before windows move and unmuting it after, or notifying another tool that the arrangement changed. See [Configuration Schema — Hooks object](CONFIG_SCHEMA.md#hooks-object) for the full field reference.

Commands are split into an `off` array (run first, before any window moves) and an `on` array (run last, after the move phase completes), configurable both globally and per connected-display-count. For a given restore, the run order brackets the global arrays outermost: global `off` → that display count's `off` → **[windows restored]** → that display count's `on` → global `on`. Within an array, commands run one at a time, in the order listed. Each command may be written as a single shell string (run via `sh -c`) or as an explicit list of strings (run directly, no shell); its output streams live as part of `mumu restore`'s own output. A command that exits non-zero or fails to start is reported and logged, but doesn't stop the remaining commands in its array or abort the restore's window-move phase.

Hooks only run when `mumu restore` actually proceeds to (or past) its window-move phase — not when it reports no saved layout for the current display count, and not when you decline the arrangement-drift confirmation prompt. They have no effect on `mumu save`, and are never triggered automatically on a schedule, at login, or in response to any system event. Pass `--no-hooks` to skip running any configured hooks for one restore, without changing `config.yaml`.

`mumu show` also lists a display count's effective, ordered `off`/`on` hook commands (as configured, without executing them) alongside its saved layout and pins.

---

## `mumu save`

Captures, for every non-fullscreen window on every Space across all connected displays, its owning application's bundle identifier, window title, and logical (left-to-right) Space number. Persists the result keyed by the current display count, overwriting any previous save for that same count. Prints a status line before scanning starts, since enumerating windows can take a moment.

```bash
mumu save
```

## `mumu restore [--yes] [--no-hooks] [--sort logical|macos|app]`

Auto-detects the current display count, loads the layout saved for it, and moves each matching, already-running application's window back to its recorded Space. Applications that aren't running are skipped (never launched). For each app, its saved entries and currently open windows are matched as a single batch: every remaining entry is scored against every remaining window by title similarity (shared words, ignoring case and word order), and the closest-matching pairs are assigned first, so no open window is ever claimed by more than one saved entry — this matters for apps like browsers, whose titles rarely match exactly across save and restore. An entry goes unmatched only once its app has no open window left to claim. Ties are broken by the entry's saved position, then deterministically, rather than left unresolved. After matching, remaining open windows from an app go, in order of precedence: (1) to that app's [configured default Space](#default-spaces), if one is set for the current display count — always, regardless of any matched assignment; (2) otherwise, if the app has at least one valid saved-entry assignment this restore, to its most prevalent target Space, or the Space currently shown on the primary (menu-bar) display if its target Spaces are tied; (3) otherwise, left unchanged. Never creates or removes Spaces — entries whose target Space no longer exists are skipped and reported.

If the current per-display Space-count arrangement doesn't match what was recorded at save time, you'll be prompted to confirm before any windows move. Pass `--yes` (or `-y`) to skip the prompt (e.g. for scripting).

Also applies any pins configured for the current display count (see [Pinned Windows](#pinned-windows)) — matched the same way, and moved/skipped/reported alongside saved-layout entries in the output below.

If hooks are configured (see [Restore Hooks](#restore-hooks)), their `off` commands run before any window moves and their `on` commands run after the move phase completes. Pass `--no-hooks` to skip running any configured hooks for this invocation.

```bash
mumu restore
mumu restore --yes
mumu restore --no-hooks
mumu restore --sort macos
```

Since each window move is deliberately paced to let WindowServer catch up, restoring many windows can take a few seconds; restore prints a line for each window as it's moved (target Space — both numbers, see [Space Numbering](#space-numbering) — then bundle ID and title) so it's never a silent pause. Fallback placements are marked `(fallback)` — or `(default)` instead, when a [configured default Space](#default-spaces) is what placed them — and approximate (non-exact) title matches are marked `(fuzzy)` in progress output and failure summaries. A placement is never marked both `(default)` and `(fallback)`, and never both a fallback marker and `(fuzzy)`, since they come from different steps. Windows are moved, and progress lines printed, in the order set by `--sort` (default: logical Space sequence). Afterward, any skipped entries are listed grouped by reason — each group also ordered by `--sort` — each showing the bundle ID, title, and saved Space, including the specific windows that couldn't be matched to a currently open window.

## `mumu list`

Lists every saved layout with its display count, window count, and save timestamp.

```bash
mumu list
```

## `mumu show [display-count] [--sort logical|macos|app]`

Prints a saved layout's window entries (Space number — both numbers, see [Space Numbering](#space-numbering) — bundle ID, title) without moving anything. Defaults to the layout for the current display count. Entry order follows `--sort` (default: logical Space sequence). Also lists that display count's configured pins (see [Pinned Windows](#pinned-windows)), configured default spaces (see [Default Spaces](#default-spaces)), and effective, ordered hook commands (see [Restore Hooks](#restore-hooks)) as written in `config.yaml`, without matching them against any currently open window or executing anything.

```bash
mumu show
mumu show 2
mumu show --sort app
```

## `mumu delete [display-count] [--yes]`

Deletes the saved layout for the given display count (default: current display count). You'll be prompted to confirm before it's deleted. Pass `--yes` (or `-y`) to skip the prompt (e.g. for scripting). If no saved layout exists for the display count, the command reports that and exits without prompting.

```bash
mumu delete
mumu delete 2
mumu delete --yes
```

## `mumu status`

Reports whether Accessibility and Screen Recording permission are currently granted. Makes no changes to any window, Space, or saved layout.

```bash
mumu status
```

---

## Limitations

- **Reflow-only**: applications that have quit since save time are skipped and reported, never relaunched. This is not a full session restore.
- **No window geometry**: only Space assignment is saved — position and size are never captured or restored.
- **Fullscreen windows are excluded** entirely, both at save and restore.
- **Minimized-window detection is best-effort**: a minimized window on the Space currently displayed on its screen is correctly excluded, but a minimized window on a Space that isn't currently displayed on any screen may still be captured and restored as if it weren't minimized — there's no reliable per-window signal for that case.
- **Manual only**: save/restore only happens when you run these commands directly. It's never triggered automatically on a schedule, at login, or in response to any system event.
- **Requires both Accessibility and Screen Recording permissions** — Screen Recording is needed to read window titles reliably for restore matching; see [`mumu status`](#mumu-status).
