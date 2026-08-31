## Context

See proposal.md - Why. `internal/layout/capture.go` (`Capture`, used by `mumu save`) and `internal/layout/restore.go` (`Restore`, used by `mumu restore`) both call a shared `ensureLayoutPermissions()` helper before doing anything else, which returns `permissions.FriendlyError(permissions.Check())` — a hard error that aborts the command immediately if either Accessibility or Screen Recording is reported as not granted. Both `Capture` and `Restore` already accept a `progress ProgressFunc` parameter used to stream human-readable status lines to the CLI (`cmd.Println(msg)`) while the operation runs.

## Goals / Non-Goals

**Goals:**
- Stop treating mumu's own permission preflight check as authoritative for whether to attempt the operation at all — let the real native window/Space calls be the source of truth.
- Keep the warning visible and specific (which permission, why it's not fatal) without turning it into a second competing error-reporting path.

**Non-Goals:**
- Changing `mumu status`'s behavior or its contract to report the raw, current permission state (including a false negative) — that command's whole purpose is to surface exactly what the OS-level check says, unfiltered.
- Diagnosing or fixing the underlying cause of a false negative (e.g. MDM/PPPC policy behavior) — that's an environmental concern tracked separately, not something this change can control.
- Adding a `--strict`/`--ignore-permissions` flag to toggle between old and new behavior — the old hard-block behavior is being replaced outright, not made optional, since the preflight check has no way to know for itself whether it's a false negative.

## Decisions

- **Replace the hard error with a warning line through the existing `ProgressFunc`, not a new output channel.** `Capture` and `Restore` already have a progress callback wired to `cmd.Println`; reusing it keeps warnings visible in the same stream as other status lines (e.g. "Scanning currently open windows...") instead of introducing a second mechanism. Alternative considered: return the warning text alongside the summary struct for the caller to print — rejected because it would require callers to remember to check and print it, whereas emitting it inline via `progress` happens automatically and immediately, matching how other status lines already work.
- **Add a new `permissions.Warnings(result CheckResult) []string` function rather than changing what `FriendlyError` returns.** `FriendlyError` becomes unused internally once `layout` stops calling it, but it's kept as-is (not deleted) since it's exported and expresses a coherent alternative policy (hard block) some future caller might still want; `Warnings` is additive, not a signature change to an existing exported function that could surprise another caller.
- **No new skip/error code is introduced.** A permission-related failure that actually occurs (e.g. no windows captured, or entries skipped for an empty title) surfaces through the same skip-reason/error paths `Capture`/`Restore` already have for other causes, rather than a dedicated "permission probably denied" code — from the caller's perspective there's no reliable way to distinguish "permission denied" from other native-call failures once the preflight check itself is admitted to be unreliable, so inventing a separate code would imply a certainty the change doesn't actually have.

## Risks / Trade-offs

- [A user with a genuinely correct "not granted" state now gets a warning instead of an immediate, clear stop] → Mitigation: the warning text explicitly names the affected permission and points at `docs/TROUBLESHOOTING.md`; the operation's subsequent failure (e.g. "0 windows captured") still surfaces, just one step later than before, and is not worse than today's behavior for a user who ignores errors.
- [Warn-and-continue could mask a real regression in the permission check itself, since a broken check no longer blocks anything] → Mitigation: `mumu status` is unchanged and remains the tool for actually confirming permission state; this change only affects whether `save`/`restore` treat a "not granted" preflight result as fatal.
