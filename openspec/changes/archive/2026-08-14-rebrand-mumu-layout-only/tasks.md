## 1. Module and binary rename (must land first so the tree keeps compiling)

- [x] 1.1 Update `go.mod`'s module path from `github.com/y3owk1n/mimi` to `github.com/adonh/mumu`
- [x] 1.2 Update every internal import across all `.go` files from `github.com/y3owk1n/mimi/...` to `github.com/adonh/mumu/...`
- [x] 1.3 Rename `cmd/mimi/` to `cmd/mumu/`; update `main.go`'s import of the `cmd` package
- [x] 1.4 Update `RootCmd.Use`, `Short`, `Long`, and the version banner text in `cmd/mumu/cmd/root.go` to `mumu`
- [x] 1.5 Update `cmd/genman/main.go`'s import path and `GenManHeader` (`Title`, `Manual`, `Source`)
- [x] 1.6 Confirm `go build ./...` still succeeds with the daemon/action code still in place, before starting deletions

## 2. Remove daemon/action/hooks/systray/ipc/config infrastructure

- [x] 2.1 Delete `internal/action`, `internal/config`, `internal/daemon`, `internal/events`, `internal/hooks`, `internal/ipc`, `internal/observe`, `internal/systray`, `internal/logging` (packages and tests)
- [x] 2.2 Delete `cmd/mumu/cmd/action.go`, `action_runner.go`, `start.go`, `stop.go`, `services.go`, `config.go`, `configpath.go`
- [x] 2.3 Remove the `--config`/`-c` and `--verbose`/`-v` persistent flags and their `resolveConfigPath`/`addConfigPreRun` wiring from `root.go`
- [x] 2.4 Update `root.go`'s `RootCmd.AddCommand` list and `Long` description to drop references to the removed commands
- [x] 2.5 Run `go build ./...` and fix compile errors surfaced by the removed packages (expected: leftover cmd wiring, `internal/permissions` call sites, `internal/native` call sites)

## 3. Flatten the CLI surface

- [x] 3.1 Remove the `layoutCmd` grouping command (and its "layout subcommand required" error path) from `cmd/mumu/cmd/layout.go`
- [x] 3.2 Register `save`, `restore`, `list`, `show`, `delete` directly on `RootCmd` instead of on `layoutCmd`
- [x] 3.3 Update each command's `Use`/`Short`/`Long` text, examples, and error strings to drop "layout" and use `mumu <verb>` instead of `mimi layout <verb>`
- [x] 3.4 Slim `status.go` to only Accessibility/Screen Recording reporting; remove PID-file, socket-file, and event-drop-count reporting
- [x] 3.5 Remove `cmd/mumu/cmd/runtime_paths.go` if nothing references it after 3.4, otherwise trim it to what's still used
- [x] 3.6 Update `layout.go`'s package-level doc comments to describe the flat command shape (no "Subcommands:" section)

## 4. Simplify `internal/permissions`

- [x] 4.1 Delete `Check()`, `RequestAccessibility`, `ShowConfigOnboardingAlert`, `ShowAccessibilityStartupAlert`, `ConfigOnboardingChoice`, `AccessibilityStartupChoice`, and their native counterparts in `permissions.h`/`.m`
- [x] 4.2 Rename `CheckLayout()` to `Check()` and `FriendlyErrorLayout()` to `FriendlyError()`; update call sites in `cmd/mumu/cmd` (save, restore, show, status)
- [x] 4.3 Update permission error message text to say "mumu" instead of "mimi" and drop references to "window/space actions" and "action commands"

## 5. Prune `internal/native` to what's still called

- [x] 5.1 Grep for `C.<Symbol>` call sites from the retained Go packages (`layout`, `space`, `window`, `permissions`, `bridge.go`/`link.go`) to build the list of still-used native symbols
- [x] 5.2 Delete native source/header files whose exported symbols have zero remaining callers per 5.1 (expected candidates: `axobserver.h/.m`, `workspace.h/.m`, `eventkinds.h` — verify, don't assume)
- [x] 5.3 Rename `mimi.h` → `mumu.h` and `mimi_log.h` → `mumu_log.h` (and any `Mimi`-prefixed C function/type names) across retained native sources; update `#include` directives and cgo preambles
- [x] 5.4 Run `go build ./...` and fix any remaining native linkage errors

## 6. Confirm layout behavior is unchanged and check the on-disk path

- [x] 6.1 Run `go test ./...` to confirm `internal/space`'s `DualLabel` and `internal/layout`'s sort/restore logic need no behavior change (the spec delta only changes command names/wording)
- [x] 6.2 Check `internal/layout/persist.go`'s on-disk storage path constant for an embedded "mimi" string; apply the rename per design.md's open question (rename the directory, or keep the path stable for continuity) and note the choice in a code comment

## 7. Full non-code rename sweep

- [x] 7.1 `README.md`, `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, and new `CHANGELOG.md` entries going forward (leave historical entries as-is)
- [x] 7.2 `docs/*.md` (`CLI.md`, `INSTALLATION.md`, `CONFIGURATION.md`, `TROUBLESHOOTING.md`, `ARCHITECTURE.md`, `DEVELOPMENT.md`, `CODING_STANDARDS.md`, `go/CONVENTIONS.md`, `go/OBJECTIVE_C.md`) — remove or rewrite sections describing removed daemon/hooks/action/config/systray functionality, not just rename the word
- [x] 7.3 `justfile`, `flake.nix`, `nix/darwin.nix`, `nix/home.nix`, `nix/package.nix`
- [x] 7.4 `resources/Info.plist.template`
- [x] 7.5 `.github/workflows/*.yml` (`ci.yml`, `release-please.yml`, `publish-artifacts.yml`, `flakehub-publish-rolling.yml`)
- [x] 7.6 `release-please-config.json`, `.release-please-manifest.json`
- [x] 7.7 Delete `configs/default-config.toml` (config-file support is removed) and any remaining reference to it
- [x] 7.8 Repo-wide case-insensitive grep for "mimi" (excluding `.git/` and `CHANGELOG.md`'s historical entries) and resolve every remaining match

## 8. Verification

- [x] 8.1 `go build ./...` succeeds
- [x] 8.2 `go vet ./...` and `golangci-lint run` pass with no new violations
- [x] 8.3 `go test ./...` passes
- [x] 8.4 Manually exercise `mumu save`, `mumu restore`, `mumu show`, `mumu list`, `mumu delete`, `mumu status`, and `mumu --help` to confirm the flattened surface and permission messaging
- [x] 8.5 `openspec validate rebrand-mumu-layout-only --strict` passes
