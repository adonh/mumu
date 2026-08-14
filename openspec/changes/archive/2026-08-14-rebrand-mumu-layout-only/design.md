## Context

Today's `internal/` layer splits cleanly into two groups by who depends on them:

- **Daemon-only** (nothing under `internal/layout` imports these): `internal/action`, `internal/config`, `internal/daemon`, `internal/events`, `internal/hooks`, `internal/ipc`, `internal/observe`, `internal/systray`, and `internal/logging` (which itself only exists to wire `config` + `events` into a file/console logger for the daemon).
- **Shared or layout-only**: `internal/layout`, `internal/space`, `internal/window`, `internal/errors`, `internal/paths` (a single `ExpandHome` helper), `internal/permissions` (layout needs `CheckLayout`; the daemon-only `Check`/onboarding-alert functions in the same file do not), and `internal/native` (a cgo bridge — some of its Objective-C sources back layout/space/window, others back AXObserver/workspace-notification hooks that only the daemon used).

`cmd/mimi/cmd/layout.go` already reads as a self-contained CLI surface: it imports only `internal/layout`, `internal/space`, and `internal/errors`, never `internal/config` or anything daemon-related. That is the concrete evidence this split is real, not just organizational.

The rename touches every Go import path (`github.com/y3owk1n/mimi` → `github.com/adonh/mumu`), the binary/build output, and a long tail of mechanical string replacements across docs, `justfile`, `flake.nix`, `nix/*.nix`, `resources/Info.plist.template`, `.github/workflows/*`, and release tooling config. See proposal.md for the full file-level impact list.

## Goals / Non-Goals

**Goals:**
- Delete every package/command that exists only to support the daemon, hooks, action, systray, IPC, or config-file system, with no dead code or placeholder stubs left behind.
- Reduce the CLI to five flat commands plus a minimal `status`, all operating directly on the root command.
- Complete the module/binary/docs/packaging rename to `mumu` in one change, so there's no intermediate state where the binary is still called `mimi` but the feature set is already layout-only (or vice versa).

**Non-Goals:**
- No new layout functionality — this change is a subtraction and a rename, not a feature change. Every requirement already in `space-layout` keeps its existing behavior; only command names and cross-references change (see specs delta).
- No decision here about publishing to a new GitHub repo, choosing a license header update, or migrating existing users' saved layout files — out of scope; saved layout files are read/written by path only and are unaffected by the rename (see Risks below for the one place this isn't quite free).
- No redesign of `internal/layout`, `internal/space`, `internal/window`, or `internal/native`'s Go-level APIs. They keep their current shapes; only their import path and (where native C symbols embed the name) some identifiers are renamed.

## Decisions

### Delete whole packages instead of trimming them

`internal/action`, `internal/config`, `internal/daemon`, `internal/events`, `internal/hooks`, `internal/ipc`, `internal/observe`, `internal/systray`, and `internal/logging` are deleted in full, along with `cmd/mimi/cmd/action.go`, `action_runner.go`, `start.go`, `stop.go`, `services.go`, `config.go`, `configpath.go`, and their tests. This was the explicit, confirmed choice over keeping them as dormant/disconnected code — dormant code with no CLI path to exercise it rots silently and still shows up in `go build`/lint/dependency graphs. If any of this is wanted back later, it's one `git revert` away.

### `internal/permissions`: split `Check()` from `CheckLayout()` behavior, keep one function

`Check()` (Accessibility only, with daemon/action-flavored error text) is deleted. `CheckLayout()` (Accessibility + Screen Recording) becomes the only permission check and is renamed to `Check()` now that it's the only one — call sites (`mumu save`/`restore`/`show`/`status`) stop saying "Layout" since there's nothing else to disambiguate from. `RequestAccessibility`/`RequestScreenRecording`, `ShowConfigOnboardingAlert`, and `ShowAccessibilityStartupAlert` are deleted along with the daemon-startup flows that called them; layout's permission errors are reported through `derrors`/`FriendlyErrorLayout` (also renamed to `FriendlyError`) exactly as today, not through an interactive alert.

### `internal/native`: prune by call-site, not by filename guess

Rather than pre-deciding which `.m`/`.h` files go, the implementation step greps for Go call sites (cgo `C.<Symbol>` references) from the retained Go packages (`layout`, `space`, `window`, `permissions`, `bridge.go`/`link.go`) and deletes any native source whose exported symbols have zero remaining callers. Expected to go: `axobserver.h/.m` (AX observation → hooks), `workspace.h/.m` (NSWorkspace launch/terminate notifications → daemon/observe), `eventkinds.h` (hook event enum). Expected to stay (renamed, not deleted): `screen.m`, `space.m`, `window.m`, `layout.m`, `element.m` (used for window title/attribute reads), `constants.h`, plus the umbrella header (`mimi.h` → `mumu.h`) and its logging header (`mimi_log.h` → `mumu_log.h`). This call-site-driven approach is safer than deleting by filename convention, since a couple of files (e.g. `element.m`) plausibly serve both the retained window-enumeration path and the removed AX-observer path.

### CLI structure: commands attach directly to root, `status` is trimmed not deleted

`cmd/mimi/cmd/layout.go`'s five subcommands (`save`, `restore`, `list`, `show`, `delete`) are registered directly on `RootCmd` instead of on an intermediate `layoutCmd`; the `layoutCmd` grouping command and its "layout subcommand required" error path are deleted. `status.go` is kept but stripped to only the two permission lines — the PID-file/socket-file/event-drop-count reporting it currently does is 100% daemon/IPC-specific and gets deleted along with `internal/daemon`/`internal/ipc`. This was a judgment call (not explicitly requested) balanced against deleting `status` outright; keeping a zero-argument way to check permissions before running a real save/restore was worth the ~15 lines it costs, and the spec delta captures it as a first-class requirement so it's reviewable on its own rather than smuggled in.

The root command's `PersistentFlags` drop `--config`/`-c` (no config file exists anymore); `--verbose`/`-v` is dropped too since nothing left reads it (verbosity in the daemon was about log level, which no longer exists — layout commands' output is unconditional, not leveled).

### Rename mechanics: module path first, then mechanical sweep

1. Change `module github.com/y3owk1n/mimi` to `module github.com/adonh/mumu` in `go.mod`, then update every internal import in one pass (`gofmt -r` or a scripted find/replace across `*.go`, since Go has no "rename module" tool that also fixes call sites) — this is the one step that must be complete before the code compiles again, so it happens before any deletions in the task order.
2. Move `cmd/mimi/` → `cmd/mumu/`, update `main.go`'s import of `cmd`, and change `RootCmd.Use`, the version banner string, and `cmd/genman`'s `GenManHeader` (`Title`, `Manual`, `Source`).
3. Sweep remaining `mimi`/`Mimi`/`MIMI` occurrences (case-sensitive, three passes) across docs, `justfile`, `flake.nix`, `nix/*.nix`, `resources/Info.plist.template`, `.github/workflows/*`, `release-please-config.json`, `.release-please-manifest.json`, and native C identifiers/comments (`Mimi*` function prefixes in `permissions.h`/`.m`, `mimi.h`, `mimi_log.h`) — each of these is a mechanical rename with no behavior change, so no per-file design decision is needed; tasks.md enumerates them for tracking, not because any single one is risky.

### Spec delta approach: rename in place, one new "flat command surface" requirement, one new "status" requirement

Existing `space-layout` requirements are updated in place (command names, permission-check wording) rather than removed-and-re-added, since the underlying behavior is unchanged — only the command users type and a couple of cross-references to now-deleted commands change. Two requirements get genuinely new content: the flat top-level command surface (previously implicit in "there's a `layout` subcommand," now an explicit, testable property) and the `status` command (previously undocumented in OpenSpec at all). The Mission Control dual-number requirement keeps the same displayed behavior but its rationale changes from "matches `mimi action space <n>`" to "matches macOS's own Ctrl+`<n>` shortcut," per the confirmed decision to keep showing both numbers.

## Risks / Trade-offs

- **[Risk]** Saved layout files live at a path derived from `paths.ExpandHome` plus a fixed sub-path that today likely embeds `mimi` (e.g. `~/.local/share/mimi/...` or similar) → **Mitigation**: tasks.md includes a step to check the actual constant in `internal/layout/persist.go`; if it embeds the old name, decide (during implementation, not here, since it's a one-line constant change with no spec impact) whether to rename the directory (users re-save once) or keep the on-disk path stable for continuity. Either choice is compatible with every requirement in the spec delta, so it's left as an implementation-time pick rather than a design decision blocking this document.
- **[Risk]** `internal/native`'s call-site-driven pruning (see Decisions) could miss a symbol that's referenced only from a `_test.go` file for a package otherwise being deleted, giving a false "still used" signal → **Mitigation**: run the grep after deleting the daemon-only Go packages, not before, so their test files are already gone and can't produce false positives.
- **[Risk]** A full rename sweep across ~15+ non-Go files in one change is easy to partially miss → **Mitigation**: tasks.md's rename step ends with a repo-wide case-insensitive `mimi` grep (excluding `CHANGELOG.md`'s historical entries and `.git/`) that must return zero matches before the change is considered done.
- **[Trade-off]** Deleting `internal/logging` means layout commands have no structured/file logging at all, only the `cmd.Println`-based progress output already in place → accepted: this matches current behavior (layout commands never used `internal/logging`) and adding logging back would be new scope, not a rename/removal side effect.

## Open Questions

- Exact on-disk saved-layout path/constant naming (see first Risk) — deferred to implementation; doesn't change any requirement or task shape, only a string literal.
