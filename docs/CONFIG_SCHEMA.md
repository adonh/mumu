# Configuration & Layout File Schemas

Reference for `config.yaml`, the one file `mumu` expects a user to read or edit, and the internal JSON layout files it manages on its own. See [`docs/examples/config.yaml`](examples/config.yaml) for a full annotated sample of the former. For file locations and resolution rules, see [Architecture Guide](ARCHITECTURE.md).

## `config.yaml`

mumu's own settings.

| Key              | Type                                             | Required | Default                                                            | Notes                                                                                        |
| ---------------- | ------------------------------------------------- | -------- | ------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `data_dir`       | string                                             | yes      | `$XDG_DATA_HOME/mumu` if set, else `~/Library/Application Support/mumu` | Directory mumu's `layouts/` subdirectory lives in. A leading `~` is expanded to the home directory. Must be a non-empty string. |
| `pins`           | map of display count (int) to list of [pin rule](#pin-rule-object) | no | none (no pins configured) | Fixed application-window-to-Space assignments `mumu restore` applies, keyed by the number of connected displays. Different display counts can declare entirely different pins. |
| `pin_precedence` | string (`pin` or `layout`)                         | no       | `pin`                                                                | Whether pin rules (`pin`) or saved-layout entries (`layout`) win when both would claim the same open window during `mumu restore`. |

```yaml
data_dir: ~/Library/Application Support/mumu

pins:
  2:
    - bundle_id: com.tinyspeck.slackmacgap
      title: "general"
      space: 1
  4:
    - bundle_id: com.tinyspeck.slackmacgap
      title: "general"
      space: 1
    - bundle_id: com.google.Chrome
      title: "GitHub"
      space: 5

pin_precedence: pin
```

### Pin rule object

One entry under a `pins` display-count list.

| Key         | Type   | Notes                                                                                          |
| ----------- | ------ | ------------------------------------------------------------------------------------------------ |
| `bundle_id` | string | The pinned application's bundle identifier (e.g. `com.google.Chrome`). Must be a non-empty string. |
| `title`     | string | Approximate title pattern, matched the same way `mumu restore` matches saved-layout entries against open windows (shared-word similarity, not an exact match). Must be a non-empty string. |
| `space`     | int    | Target logical left-to-right Space number (see [CLI Guide — Space Numbering](CLI.md#space-numbering)). Must be a positive integer. |

Loading rules:

- If `config.yaml` doesn't exist yet, it's auto-created with commented defaults (see `defaultConfigYAML` in `internal/config/config.go`) and never overwritten afterward.
- Missing/empty `data_dir`, or a `data_dir` value that isn't a plain string, is a load error (`CodeInvalidConfig`) — the process exits rather than silently falling back to a default.
- A missing `pins` setting means no pins are configured for any display count; `mumu restore` proceeds using only its saved-layout matching.
- Any pin rule missing `bundle_id` or `title`, or with a `space` that isn't a positive integer, is a load error (`CodeInvalidConfig`) naming the config file path and the offending display count/app.
- A `pin_precedence` value other than `pin` or `layout` is a load error (`CodeInvalidConfig`).
- Malformed YAML is a load error (`CodeInvalidConfig`).
- Unrecognized top-level keys are ignored (forward-compatible), not rejected.
- Any YAML mumu writes, including `config.yaml`, uses two-space indentation.

## Saved layouts (`<data_dir>/layouts/<display-count>.json`)

Saved window-to-Space layouts are internal state, not a user-facing file: mumu makes no guarantee about their structure staying hand-editable, and doesn't document them as something to edit directly. One JSON file exists per display count, named `<display-count>.json`, inside a `layouts` subdirectory of `data_dir`.

### Layout file shape

| Key             | Type                            | Notes                                                        |
| --------------- | -------------------------------- | -------------------------------------------------------------- |
| `schemaVersion` | int                              | On-disk schema version. Currently always `1`.                |
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
| `ordinal`  | int    | The window's logical left-to-right Mission Control Space number (see [CLI Guide — Space Numbering](CLI.md#space-numbering)), independent of which display is primary. |

```json
{
  "schemaVersion": 1,
  "displayCount": 2,
  "spaceCounts": [3, 2],
  "savedAt": "2026-08-20T14:02:07-04:00",
  "entries": [
    {
      "bundleId": "com.google.Chrome",
      "title": "mumu/README.md at feature/10-config-file",
      "index": 0,
      "ordinal": 2
    }
  ]
}
```

Loading rules:

- A missing layout file for a given display count is a "no saved layout" condition (e.g. `mumu restore` reports it clearly), not treated as an empty layout.
- Malformed JSON is a load error identifying the file path.
