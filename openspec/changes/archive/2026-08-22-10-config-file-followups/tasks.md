## 1. Revert saved layouts to per-display-count JSON

- [x] 1.1 Update `internal/layout/types.go`: drop the `Store`/single-file wrapper type if no longer needed, and restore a `DisplayCount` field on `Layout` (removed when layouts moved into a map keyed by display count) — verify `go build ./...` succeeds
- [x] 1.2 Rewrite `internal/layout/persist.go` to write/read one JSON file per display count (`<display-count>.json`) inside a `layouts` subdirectory of `config.Load().DataDir`, using `encoding/json` with indented output (matching the pre-`10-yaml-config-file` style) instead of `go.yaml.in/yaml/v3` — verify `Save`, `Load`, `Exists`, `Delete`, and `List` all operate against the new file shape
- [x] 1.3 Update `internal/layout/persist_test.go` to assert against per-file JSON persistence (file existence per display count, JSON parse error handling, no single shared file) — verify `go test ./internal/layout/...` passes
- [x] 1.4 Remove the now-unused single-file YAML marshal/unmarshal error codes or paths if `internal/errors` had anything YAML-specific to saved layouts only — verify no dead code remains (`just lint`) — confirmed `CodeSerializationFailed`/`CodeConfigIOFailed` are already format-agnostic; no YAML-specific error paths existed to remove

## 2. Update documentation for the reverted layout format

- [x] 2.1 Update `docs/CONFIG_SCHEMA.md` to describe the `layouts` directory of per-display-count JSON files instead of `layouts.yaml`, and remove the hand-editing guidance/loading rules that no longer apply
- [x] 2.2 Replace `docs/examples/layouts.yaml` with a JSON example (or remove it if an example is no longer warranted for internal state) and update any cross-references to it (e.g. `docs/examples/config.yaml`, `docs/ARCHITECTURE.md`) — removed; a JSON sample is inlined in `docs/CONFIG_SCHEMA.md` instead, since layouts are internal state rather than something to document as a standalone user-facing example
- [x] 2.3 Update `docs/ARCHITECTURE.md` and any other doc referencing `layouts.yaml` or its hand-editability — verified with `rg -i "layouts.yaml"`: only remaining references are in this change's own planning artifacts and the archived `10-yaml-config-file` change (historical, left as-is)

## 3. Codify the 2-space YAML indent convention

- [x] 3.1 Configure the YAML encoder used anywhere mumu marshals YAML (currently none, after task 1's revert — verify by checking `internal/config` and confirming it still writes `config.yaml` as a hand-written string) to use 2-space indentation if/when a `yaml.Marshal`/`Encoder` call is introduced; add a short code comment at the relevant spot (or in `internal/config/doc.go`) noting the convention for future authors — confirmed no encoder call exists today; documented the convention in `internal/config/doc.go`
- [x] 3.2 Verify `docs/examples/config.yaml` (and any remaining YAML example) already reflects 2-space indentation — file is flat (no nested/list content), no example remaining after removing `docs/examples/layouts.yaml`

## 4. Update AGENTS.md

- [x] 4.1 Add a "Configuration and data files" section to `AGENTS.md` covering: `config.yaml` location resolution ($XDG_CONFIG_HOME vs macOS fallback) and its auto-create-with-commented-defaults pattern, `data_dir` resolution and what lives under it (`layouts/<n>.json`), the JSON-for-internal-state vs YAML-for-user-facing-config split and why, and the 2-space YAML indent convention — verify the section is present and consistent with the final implementation from sections 1-3 — `AGENTS.md` didn't exist yet in this repo; created it with this section plus a short build/lint/test pointer

## 5. Verify

- [x] 5.1 Run `just fmt`, `just lint`, `just test`, `just build` and confirm all pass (or that any failures are pre-existing and unrelated to this change) — `just build` and `just test` pass; `just fmt`/`just lint` fail only on the pre-existing `exhaustruct_v5` tooling mismatch (`.golangci.yml` disables `exhaustruct`, installed golangci-lint v2.13.1 renamed it `exhaustruct_v5`), same 75 issues as before this change, none introduced by it
- [x] 5.2 Manually confirm `mumu save`/`mumu list`/`mumu show`/`mumu delete`/`mumu restore` work end-to-end against the new per-display-count JSON files — ran all five against an isolated `HOME`; confirmed `<data_dir>/layouts/4.json` is created with the expected shape, and each command reads/writes/removes it correctly
