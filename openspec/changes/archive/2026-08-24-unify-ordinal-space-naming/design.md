## Context

See `proposal.md` - Why. This is a mechanical rename spanning `internal/config`, `internal/layout`, `cmd/mumu/cmd`, their tests, and two docs files, plus the user's own live `config.yaml`. It has no architectural ambiguity, but two small decisions are worth writing down since they affect existing configs and the shape of the change.

## Goals / Non-Goals

**Goals:**
- One vocabulary, everywhere a Space number appears: `ordinal` = mumu's own logical left-to-right number, `space` = the macOS Mission Control Space concept.
- Saved-layout JSON and `config.yaml` use the same field name (`ordinal`) for the same concept, since a user may hand-edit either.

**Non-Goals:**
- No change to `mumu show`/`restore`'s printed dual-label format (`#NN (space MM)`) — already correct once "ordinal"/"space" are the only two words in play (per the user's explicit decision).
- No backward-compatibility shim for the old `space:` config key — this is a pre-1.0 tool with no compatibility guarantee on `config.yaml`'s schema (see `clarify-sort-terminology` precedent, which hard-cut `--sort display` with no fallback).
- No change to the saved-layout JSON schema (`internal/layout/types.go`'s `Entry.Ordinal`/`"ordinal"` is already correctly named).

## Decisions

**Old `space:` key: rely on existing validation, no special-cased detection.** `config.Load` uses plain `yaml.Unmarshal` (not strict), so an old `space:` key under a renamed `ordinal` field is silently unrecognized, leaving `Ordinal` at its zero value. The existing "ordinal must be a positive integer, got 0" validation error (already required for a missing/zero value) then fires naming the offending app and display count — good enough to point a user at the problem without adding strict-mode YAML decoding or a dedicated "did you mean `ordinal`?" error path, which would be new scope beyond this rename.

**Go field rename mirrors the YAML rename exactly.** `PinRule.Space` → `PinRule.Ordinal`, `DefaultSpaceRule.Space` → `DefaultSpaceRule.Ordinal`, matching `layout.Entry.Ordinal`'s existing name. This keeps one Go-side name for "mumu's own logical ordinal" across all three structs, rather than leaving `PinRule`/`DefaultSpaceRule` as the odd ones out.

**Courtesy-edit the user's live config, not a migration tool.** The user's `~/.xdg/config/mumu/config.yaml` is outside the repo and has no automated migration path (mumu doesn't rewrite `config.yaml` after first creation, by existing design — see the `configuration` capability's "Existing config file is left untouched" scenario). It's edited by hand once, alongside this change, the same way pins/default_spaces were added to it by hand in the prior change.

## Risks / Trade-offs

- [Any other mumu user with an existing `config.yaml` using `pins`/`default_spaces` silently loses those rules on upgrade (ordinal defaults to 0 → load error) rather than a friendly one-line migration message] → Acceptable pre-1.0; the resulting "ordinal must be a positive integer, got 0" error is directly actionable (names the file, app, and display count) even without special-casing the old key name.
