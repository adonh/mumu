## 1. Permission check

- [x] 1.1 Add `permissions.Warnings(result CheckResult) []string` in `internal/permissions/check.go`, returning a one-line warning for each of Accessibility/Screen Recording that `Check` reports as not granted; verify with a quick unit test covering both-granted (empty slice), one-missing, and both-missing cases.

## 2. Layout save/restore

- [x] 2.1 Replace `ensureLayoutPermissions()` in `internal/layout/capture.go` with a `warnMissingPermissions(progress ProgressFunc)` helper that calls `permissions.Warnings` and emits each via `progress.emit("Warning: " + ...)`, and update `Capture` to call it without returning early; verify with `just build` and a manual `mumu save` run.
- [x] 2.2 Update `Restore` in `internal/layout/restore.go` to call the same `warnMissingPermissions` helper instead of `ensureLayoutPermissions()`; verify with a manual `mumu restore` run.
- [x] 2.3 Confirm `permissions.FriendlyError` and the `derrors.CodeAccessibilityDenied`/`CodeScreenRecordingDenied` codes are left intact (no longer called from `internal/layout`, but not removed) and that `mumu status` (`cmd/mumu/cmd/status.go`) is untouched, since neither is in scope for this change.

## 3. Docs and spec verification

- [x] 3.1 Update `docs/CLI.md` and `docs/TROUBLESHOOTING.md` to describe the warn-and-continue behavior in place of the old hard-block description.
- [x] 3.2 Run `openspec validate --strict` for this change and confirm it passes.
- [x] 3.3 Run `just fmt`, `just lint`, `just test`, and `just build` and confirm all pass.
