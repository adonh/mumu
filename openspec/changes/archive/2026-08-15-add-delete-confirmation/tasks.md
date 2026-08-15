## 1. CLI implementation

- [x] 1.1 Add a `--yes`/`-y` bool flag to `layoutDeleteCmd` (mirroring `layoutRestoreCmd`'s `restoreAssumeYes`).
- [x] 1.2 In `layoutDeleteCmd`'s `RunE`, load the saved layout for the resolved display count before deleting, so the existing "no layout" error still surfaces without prompting.
- [x] 1.3 When a layout exists and `--yes` was not passed, prompt via `promptConfirm` (e.g. `Delete saved layout for N display(s)?`) before calling `layout.Delete`.
- [x] 1.4 If the user declines, print an "aborted" message and return `nil` without calling `layout.Delete`.
- [x] 1.5 If `--yes` was passed or the user confirms, call `layout.Delete` and print the existing "Deleted saved layout..." message.

## 2. Docs

- [x] 2.1 Update `docs/CLI.md`'s `mumu delete` section to document the confirmation prompt and the `--yes`/`-y` flag, matching the style of the `mumu restore` section.

## 3. Verification

- [x] 3.1 Run `go build ./...` and existing tests (`go test ./...`) to confirm nothing is broken.
- [x] 3.2 Manually verify: `mumu delete` prompts and aborts on "n", `mumu delete --yes` deletes without prompting, and `mumu delete` for a non-existent layout still errors without prompting.
