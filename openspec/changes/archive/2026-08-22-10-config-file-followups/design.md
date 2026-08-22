## Context

`10-yaml-config-file` (already implemented and archived) introduced `config.yaml` with a `data_dir` setting, and consolidated saved layouts into a single hand-editable `layouts.yaml` inside `data_dir`. See `proposal.md` - Why for why two of those decisions are being revisited.

## Goals / Non-Goals

**Goals:**
- Revert saved-layout persistence to JSON, one file per display count, while keeping `data_dir` as the configurable root for where that data lives.
- Establish and document a 2-space YAML indentation convention for any YAML mumu writes.
- Fold the guidelines this effort settled on into `AGENTS.md`.

**Non-Goals:**
- No change to `config.yaml`'s own format, resolution rules, or the `data_dir` setting's semantics.
- No migration of existing `layouts.yaml` content into the new JSON files — consistent with how `10-yaml-config-file` itself treated the prior on-disk format as abandoned.
- No change to save/restore/match/sort behavior — only the on-disk persistence shape.

## Decisions

- **Per-display-count JSON files under `<data_dir>/layouts/`, not a single JSON file.** The proposal asked "as it was before"; the original (pre-`10-yaml-config-file`) shape was one file per display count under a hardcoded directory. We keep that per-file shape (rather than a single `layouts.json` keyed by display count) but nest it under the now-configurable `data_dir`, as `<data_dir>/layouts/<display-count>.json`, so `data_dir` still controls where all of mumu's data lives. Alternative considered: single `layouts.json` — rejected because the user explicitly asked for the original per-file shape.
- **No migration from `layouts.yaml`.** Matches the precedent `10-yaml-config-file` set for the JSON-files-to-YAML-file transition it made (no migration, "run `mumu save` again"). A leftover `layouts.yaml` is simply ignored by the new code (it doesn't look for that filename).
- **2-space YAML indent enforced via the encoder, not hand-formatting.** Configure the YAML library's indent setting (e.g. `yaml.Encoder.SetIndent(2)` / equivalent) wherever mumu marshals YAML, rather than relying on manually-written strings staying consistent. `config.yaml`'s current default-file content is a hand-written string (no nested structure), so this has no visible effect today, but it's the mechanism to use once/if `config.yaml` gains nested settings.
- **`AGENTS.md` gets a "Configuration and data files" section**, covering: where `config.yaml` and layout data live and how each resolves ($XDG_* vs macOS fallback), the auto-create-with-commented-defaults pattern for user-facing config, the JSON-for-internal-state vs YAML-for-user-facing-config split, and the 2-space YAML indent convention. This is written directly into `AGENTS.md` as part of implementation (tasks.md), not as a planning artifact.

## Risks / Trade-offs

- Anyone who ran a `mumu save` under the now-archived `10-yaml-config-file` change and relied on `layouts.yaml` being hand-editable loses that layout silently (it's simply not read by the reverted code) → Mitigated the same way the original change mitigated its own breaking change: call it out as **BREAKING** in the proposal/commit and note that `mumu save` needs to be re-run. This branch hasn't shipped to end users yet, so the practical blast radius is limited to this repo's own history.
- Introducing a `layouts` subdirectory under `data_dir` (rather than writing files directly into `data_dir`) is a small new nesting decision not present in either the original or the `10-yaml-config-file` shape → Documented explicitly above and in the spec so it's not implicit; low risk since it's additive path structure, easy to change again if it proves wrong.
