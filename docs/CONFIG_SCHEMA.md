# Configuration & Layout File Schemas

Reference for `config.yaml`, the one file `mumu` expects a user to read or edit, and the internal JSON layout files it manages on its own. See [`docs/examples/config.yaml`](examples/config.yaml) for a full annotated sample of the former. For file locations and resolution rules, see [Architecture Guide](ARCHITECTURE.md).

## `config.yaml`

mumu's own settings.

| Key              | Type                                             | Required | Default                                                            | Notes                                                                                        |
| ---------------- | ------------------------------------------------- | -------- | ------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `data_dir`       | string                                             | yes      | `$XDG_DATA_HOME/mumu` if set, else `~/Library/Application Support/mumu` | Directory mumu's `layouts/` subdirectory lives in. A leading `~` is expanded to the home directory. Must be a non-empty string. |
| `pins`           | map of display count (int) to list of [pin rule](#pin-rule-object) | no | none (no pins configured) | Fixed application-window-to-Space assignments `mumu restore` applies, keyed by the number of connected displays. Different display counts can declare entirely different pins. |
| `pin_precedence` | string (`pin` or `layout`)                         | no       | `layout`                                                             | Whether pin rules (`pin`) or saved-layout entries (`layout`) win when both would claim the same open window during `mumu restore`. |
| `default_spaces` | map of display count (int) to list of [default-space rule](#default-space-rule-object) | no | none (no default spaces configured) | Fixed application-level fallback Spaces `mumu restore` applies to an app's leftover unclaimed windows, keyed by the number of connected displays. Overrides the prevalent-Space heuristic (see [CLI Guide — Default Spaces](CLI.md#default-spaces)) for any configured application. |
| `hooks`          | [hooks object](#hooks-object)                      | no       | none (no hooks configured) | External commands run automatically around every `mumu restore`. See [Hooks object](#hooks-object). |

```yaml
data_dir: ~/Library/Application Support/mumu

pins:
  2:
    - bundle_id: com.tinyspeck.slackmacgap
      title: "general"
      ordinal: "1:1"
  4:
    - bundle_id: com.tinyspeck.slackmacgap
      title: "general"
      ordinal: "1:1"
    - bundle_id: com.google.Chrome
      title: "GitHub"
      ordinal: "2:1"

pin_precedence: layout

default_spaces:
  2:
    - bundle_id: com.tinyspeck.slackmacgap
      ordinal: "1:1"

hooks:
  off:
    - osascript -e 'set volume output muted true'
  on:
    - [osascript, -e, "set volume output muted false"]
  layouts:
    2:
      off:
        - echo switching to 2-display layout
```

### Pin rule object

One entry under a `pins` display-count list.

| Key         | Type   | Notes                                                                                          |
| ----------- | ------ | ------------------------------------------------------------------------------------------------ |
| `bundle_id` | string | The pinned application's bundle identifier (e.g. `com.google.Chrome`). Must be a non-empty string. |
| `title`     | string | Approximate title pattern, matched the same way `mumu restore` matches saved-layout entries against open windows (shared-word similarity, not an exact match). Must be a non-empty string. |
| `ordinal`   | string | Target Space's two-part `"<display>:<space>"` ordinal — mumu's own numbering, not the macOS Mission Control Space number (see [CLI Guide — Space Numbering](CLI.md#space-numbering)). Both parts must be positive integers (e.g. `"2:1"`). |

### Default-space rule object

One entry under a `default_spaces` display-count list.

| Key         | Type   | Notes                                                                                          |
| ----------- | ------ | ------------------------------------------------------------------------------------------------ |
| `bundle_id` | string | The application's bundle identifier (e.g. `com.google.Chrome`). Must be a non-empty string. |
| `ordinal`   | string | Target Space's two-part `"<display>:<space>"` ordinal — mumu's own numbering, not the macOS Mission Control Space number (see [CLI Guide — Space Numbering](CLI.md#space-numbering)) — for that application's leftover unclaimed windows. Both parts must be positive integers (e.g. `"2:1"`). |

Unlike a [pin rule](#pin-rule-object), there's no `title`: a default-space rule is application-level, not per-window. When configured for an application at the current display count, it always wins over the usual "most prevalent assigned Space" fallback for that application's leftover windows — even when that heuristic would otherwise produce an unambiguous (non-tied) target — and it also activates a fallback placement for an application with zero valid saved-entry matches this restore, a case that otherwise receives no fallback at all. See `mumu restore --help` and the `space-layout` capability spec for the full precedence order.

### Hooks object

`config.yaml`'s `hooks` setting.

| Key       | Type                                                        | Required | Notes                                                                                          |
| --------- | ------------------------------------------------------------ | -------- | ------------------------------------------------------------------------------------------------ |
| `off`     | list of [command](#command)                                  | no       | Run before every `mumu restore` moves any window (see `layouts.<n>.off` for per-display-count commands run after this array). |
| `on`      | list of [command](#command)                                  | no       | Run after every `mumu restore`'s window-move phase completes (see `layouts.<n>.on` for per-display-count commands run before this array). |
| `layouts` | map of display count (int) to `{off, on}` command lists       | no       | Per-display-count `off`/`on` arrays, applied only when that number of displays is connected. Bracketing order for a given restore is: global `off`, that display count's `off`, **[windows restored]**, that display count's `on`, global `on`. |

#### Command

Each entry in an `off`/`on` list is either:

- a single string, executed through a shell (`sh -c`), so it may use pipes, redirection, and shell expansion; or
- a list of strings, executed directly as a program and its arguments, with no shell involved — the first element is the program, the rest are its arguments.

Any of `hooks.off`, `hooks.on`, or a given display count's `layouts` entry may be absent or empty, independently of the others. A command's stdout/stderr stream directly to `mumu`'s own output as it runs; a command that exits non-zero or fails to start is reported and logged but does not stop the remaining commands in its array or abort the restore. Hooks only run as part of a `mumu restore` invocation that actually proceeds to its window-move phase — never for `mumu save`, and never on a schedule or system event. Pass `--no-hooks` to `mumu restore` to skip running any configured hooks for that one invocation.

Loading rules:

- If `config.yaml` doesn't exist yet, it's auto-created with commented defaults (see `defaultConfigYAML` in `internal/config/config.go`) and never overwritten afterward.
- Missing/empty `data_dir`, or a `data_dir` value that isn't a plain string, is a load error (`CodeInvalidConfig`) — the process exits rather than silently falling back to a default.
- A missing `pins` setting means no pins are configured for any display count; `mumu restore` proceeds using only its saved-layout matching.
- Any pin rule missing `bundle_id` or `title`, or with an `ordinal` that isn't a valid `"<display>:<space>"` string (both parts positive integers), is a load error (`CodeInvalidConfig`) naming the config file path and the offending display count/app.
- A `pin_precedence` value other than `pin` or `layout` is a load error (`CodeInvalidConfig`).
- A missing `default_spaces` setting means no application has a configured default Space for any display count; `mumu restore` proceeds using only its prevalent-Space fallback heuristic.
- Any default-space rule missing `bundle_id`, or with an `ordinal` that isn't a valid `"<display>:<space>"` string (both parts positive integers), is a load error (`CodeInvalidConfig`) naming the config file path and the offending display count/app.
- A missing `hooks` setting means no hooks are configured, globally or for any display count; `mumu restore` runs no external commands.
- Any command entry that's neither a non-empty string nor a non-empty list of non-empty strings, or an `off`/`on` value that isn't a list, is a load error (`CodeInvalidConfig`) naming the config file path and the offending entry.
- Malformed YAML is a load error (`CodeInvalidConfig`).
- Unrecognized top-level keys are ignored (forward-compatible), not rejected.
- Any YAML mumu writes, including `config.yaml`, uses two-space indentation.

## Saved layouts (`<data_dir>/layouts/<display-count>.json`)

Saved window-to-Space layouts are internal state, not a user-facing file: mumu makes no guarantee about their structure staying hand-editable, and doesn't document them as something to edit directly. One JSON file exists per display count, named `<display-count>.json`, inside a `layouts` subdirectory of `data_dir`.

### Layout file shape

| Key             | Type                            | Notes                                                        |
| --------------- | -------------------------------- | -------------------------------------------------------------- |
| `schemaVersion` | int                              | On-disk schema version. Currently always `2`. A layout file written by an incompatible version fails to load with a clear error asking you to run `mumu save` again — there's no automatic migration. |
| `displayCount`  | int                               | The number of connected displays this layout was captured for; matches the file name. |
| `spaceCounts`   | list of int                       | Per-display Space count, left to right, recorded at save time. Used at restore to detect arrangement drift (a different Space count per display than when saved) and prompt for confirmation. |
| `entries`       | list of [Entry](#entry-object)   | The saved windows for this display count.                    |
| `savedAt`       | timestamp (RFC 3339)             | When this layout was last saved. Purely informational — shown by `mumu list`/`show`, not used for matching. |

### Entry object

One saved window.

| Key        | Type   | Notes                                                                                                                                    |
| ---------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `bundleId` | string | The owning application's bundle identifier (e.g. `com.google.Chrome`).                                                                     |
| `title`    | string | The window's title at save time. Primary signal for restore matching.                                                                     |
| `index`    | int    | 0-based position among the app's other captured (non-fullscreen) windows, in save-time enumeration order. Restore-time fallback when title matching is ambiguous. |
| `ordinal`  | object | The window's two-part Space ordinal, `{"display": D, "space": S}` — mumu's own numbering (see [CLI Guide — Space Numbering](CLI.md#space-numbering)), not the macOS Mission Control Space number. `display` counts connected displays left to right; `space` counts that display's Spaces left to right, independent of every other display's Space count. |

```json
{
  "schemaVersion": 2,
  "displayCount": 2,
  "spaceCounts": [3, 2],
  "savedAt": "2026-08-20T14:02:07-04:00",
  "entries": [
    {
      "bundleId": "com.google.Chrome",
      "title": "mumu/README.md at feature/10-config-file",
      "index": 0,
      "ordinal": {"display": 1, "space": 2}
    }
  ]
}
```

Loading rules:

- A missing layout file for a given display count is a "no saved layout" condition (e.g. `mumu restore` reports it clearly), not treated as an empty layout.
- A layout file whose `schemaVersion` doesn't match the current version (e.g. one saved by an older mumu build) is a load error identifying the file path and instructing you to run `mumu save` again — checked before the rest of the file is parsed, so it's reported instead of a raw JSON type-mismatch error.
- Malformed JSON is a load error identifying the file path.
