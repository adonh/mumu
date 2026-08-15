# CLI Usage

`mumu` saves and restores window-to-Space layouts on macOS. All commands are top-level — there is no subcommand grouping.

---

## Table of Contents

- [Global Flags](#global-flags)
- [Space Numbering](#space-numbering)
- [Output Ordering (`--sort`)](#output-ordering---sort)
- [`mumu save`](#mumu-save)
- [`mumu restore`](#mumu-restore---yes---sort-displaymacosapp)
- [`mumu list`](#mumu-list)
- [`mumu show`](#mumu-show-display-count---sort-displaymacosapp)
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

`mumu`'s own ordinal (shown as `#3`, `#21`, etc.) is counted **left to right across all connected displays**, independent of which display is primary — matching how a person visually counts Spaces on screen.

Because this can diverge from macOS's own Mission Control ordering whenever the primary display isn't the leftmost one, any output that names a specific window's Space (`show`, restore's per-window progress, and its skip summary) shows both numbers together, e.g. `#03 (space 21)` — mumu's own ordinal first, then the macOS Mission Control Space number in parentheses, which is the same ordinal macOS's own "Switch to Desktop `<n>`" keyboard shortcut uses. The Mission Control number is resolved fresh against the current display arrangement each time it's printed, so it always reflects what's true right now (it's never saved to the layout file).

## Output ordering (`--sort`)

`mumu show` and `mumu restore` both accept `--sort <display|macos|app>`, controlling the order per-window entries print in (and, for `restore`, the order windows are actually moved in — it has no effect on _which_ windows get matched, moved, or skipped):

- `display` (default): logical left-to-right Space number, ascending.
- `macos`: macOS Mission Control Space number, ascending.
- `app`: bundle identifier, ascending (alphabetical), grouping all of one application's windows together. mumu has no captured or resolvable human-readable app display name, so this sorts by the raw bundle ID string (e.g. `com.apple.Safari`), not a "friendly" app name.

Whichever key is primary, entries that tie on it are ordered by falling back through the other two keys, in the fixed priority Space number, then bundle identifier, then window title — so output is always fully deterministic. This ordering applies to `show`'s entry list, restore's per-window move progress lines, and the ordering of entries within each reason group of restore's skip summary.

---

## `mumu save`

Captures, for every non-fullscreen window on every Space across all connected displays, its owning application's bundle identifier, window title, and logical (left-to-right) Space number. Persists the result keyed by the current display count, overwriting any previous save for that same count. Prints a status line before scanning starts, since enumerating windows can take a moment.

```bash
mumu save
```

## `mumu restore [--yes] [--sort display|macos|app]`

Auto-detects the current display count, loads the layout saved for it, and moves each matching, already-running application's window back to its recorded Space. Applications that aren't running are skipped (never launched). For each app, its saved entries and currently open windows are matched as a single batch: every remaining entry is scored against every remaining window by title similarity (shared words, ignoring case and word order), and the closest-matching pairs are assigned first, so no open window is ever claimed by more than one saved entry — this matters for apps like browsers, whose titles rarely match exactly across save and restore. An entry goes unmatched only once its app has no open window left to claim. Ties are broken by the entry's saved position, then deterministically, rather than left unresolved. After matching, remaining open windows from an app with a valid assignment move to that app's most prevalent target Space. If its target Spaces are tied, they move to the Space currently shown on the primary (menu-bar) display, so they are immediately visible; apps with no valid assignment are left unchanged. Never creates or removes Spaces — entries whose target Space no longer exists are skipped and reported.

If the current per-display Space-count arrangement doesn't match what was recorded at save time, you'll be prompted to confirm before any windows move. Pass `--yes` (or `-y`) to skip the prompt (e.g. for scripting).

```bash
mumu restore
mumu restore --yes
mumu restore --sort macos
```

Since each window move is deliberately paced to let WindowServer catch up, restoring many windows can take a few seconds; restore prints a line for each window as it's moved (target Space — both numbers, see [Space Numbering](#space-numbering) — then bundle ID and title) so it's never a silent pause. Fallback placements are marked `(fallback)` and approximate (non-exact) title matches are marked `(fuzzy)` in progress output and failure summaries — a single placement is never both, since they come from different steps. Windows are moved, and progress lines printed, in the order set by `--sort` (default: display-sequence order). Afterward, any skipped entries are listed grouped by reason — each group also ordered by `--sort` — each showing the bundle ID, title, and saved Space, including the specific windows that couldn't be matched to a currently open window.

## `mumu list`

Lists every saved layout with its display count, window count, and save timestamp.

```bash
mumu list
```

## `mumu show [display-count] [--sort display|macos|app]`

Prints a saved layout's window entries (Space number — both numbers, see [Space Numbering](#space-numbering) — bundle ID, title) without moving anything. Defaults to the layout for the current display count. Entry order follows `--sort` (default: display-sequence order).

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
