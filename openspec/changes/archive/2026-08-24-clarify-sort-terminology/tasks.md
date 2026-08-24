## 1. Rename the sort key

- [x] 1.1 In `internal/layout/sort.go`, rename `SortByDisplay` to `SortByLogical` with value `"logical"` (was `"display"`); update its doc comment, `ParseSortKey`'s accepted-values switch and error message (`"must be one of: logical, macos, app"`), and every `case SortByDisplay` branch (including in `entryLess`, `pinRuleLess`, `defaultSpaceRuleLess`). Verify with `go build ./...` and by running the existing `ParseSortKey` tests after renaming their expectations.
- [x] 1.2 Update `internal/layout/sort_test.go`: rename all `SortByDisplay` references to `SortByLogical`, rename `TestEntryLess_SortByDisplay` to `TestEntryLess_SortByLogical`, and change `TestParseSortKey`'s `"display"` key to `"logical"`. Add a case (or extend `TestParseSortKey_Invalid`) asserting `ParseSortKey("display")` now fails, since it's no longer an accepted value. Verify with `go test ./internal/layout/...`.

## 2. Update the CLI surface

- [x] 2.1 In `cmd/mumu/cmd/layout.go`, update the `--sort` flag's default value (`string(layout.SortByLogical)`) and help text (currently `` `"display"` (logical left-to-right Space number, default) ``) to name `logical` instead of `display`, for both `layoutShowCmd` and `layoutRestoreCmd` (via `addSortFlag`/wherever the shared help text lives). Verify with `just build` and `./bin/mumu show --help` / `./bin/mumu restore --help` showing `logical` as the flag's documented default.
- [x] 2.2 Update `cmd/mumu/cmd/layout_test.go`: rename all `layout.SortByDisplay` references to `layout.SortByLogical`. Verify with `go test ./cmd/mumu/cmd/...`.

## 3. Documentation

- [x] 3.1 Update `docs/CLI.md`: every `--sort display|macos|app` (or similar) mention becomes `--sort logical|macos|app`; "Default order is display sequence" and similar prose becomes "Default order is logical Space sequence" (or equivalent); the "Output ordering" section's `display` bullet is renamed `logical` with unchanged behavior text. Confirm no remaining prose uses "display" to mean anything other than a physical monitor by re-reading the file. Verify by cross-referencing against the updated `space-layout` spec scenarios.

## 4. Verification

- [x] 4.1 Run `just fmt`, `just lint`, `just build`, and `just test`; confirm all four pass with no new warnings.
- [x] 4.2 Manually run `mumu show` and `mumu show --sort display` (expect a clear "must be one of: logical, macos, app" error) and `mumu show --sort logical` (expect it to behave exactly as the old default did) to confirm the rename took effect end-to-end.
