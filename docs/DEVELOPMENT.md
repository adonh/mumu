# Development Guide

## Prerequisites

- macOS (required for CGO/Objective-C)
- Go 1.26+
- [just](https://github.com/casey/just) (build system)
- [devbox](https://www.jetify.com/devbox) (optional, recommended)

```bash
devbox shell
just setup-codesign-identity  # one-time, keeps Accessibility grants stable across rebuilds
just build
```

---

## Project Layout

```
cmd/mumu/              CLI binary
internal/
  layout/              Layout save/restore: capture, YAML persistence, restore matching
  config/              config.yaml resolution, auto-creation, and loading
  window/              AX window wrappers and CGWindowList-based window moves
  space/               Mission Control operations, logical left-to-right numbering
  native/              Obj-C + CGO bridge (window/space APIs, layout enumeration)
  permissions/         Accessibility and Screen Recording checks
  errors/              Structured error types
  paths/               Home-directory path expansion
```

---

## Build Commands

```bash
just build                    # build bin/mumu
just test                     # run all tests
just test-unit                # unit tests only
just lint                     # golangci-lint
just fmt                      # format Go + Objective-C
just bundle                   # build Mumu.app
just genman                   # generate man pages
just setup-codesign-identity  # one-time: stable local code-signing identity
```

---

## Code Signing and Accessibility Permissions

macOS's TCC subsystem pins a granted Accessibility/Screen Recording permission to the binary's *code identity*, not just its bundle ID or path. `build`/`bundle` sign with an ad-hoc identity by default (`codesign --sign -`), and ad-hoc signing derives that identity from the binary's own hash — which changes on every rebuild. That means a routine `just build` can silently invalidate a permission you already granted, even though System Settings still shows the old grant as checked (it just no longer matches the binary you're running).

Run `just setup-codesign-identity` once per machine to create a stable, self-signed code-signing identity (`mumu-dev-signing`, stored in a dedicated `mumu-dev.keychain-db` — not your login keychain, to avoid a blocking macOS authorization prompt). `build`/`bundle` use it automatically whenever it's present, keying the permission grant to the certificate instead of the binary hash, so it survives rebuilds. Without running this step, both recipes keep working exactly as before (ad-hoc signing, with a printed warning) — the setup is optional but recommended for anyone doing repeated local `mumu save`/`restore` testing.

You'll need to re-grant Accessibility/Screen Recording once after running `just setup-codesign-identity` for the first time (the new identity is different from whatever ad-hoc identity was previously granted); after that, rebuilds keep the grant.

---

## Adding a CLI Command

1. Create `cmd/mumu/cmd/<name>.go` with a Cobra command using `RunE`, registered directly on `RootCmd` (there is no subcommand grouping)
2. Put business logic in `internal/`
3. Document in `docs/CLI.md`

---

## Native Code

Objective-C lives in:

- `internal/native/` — window/space enumeration and moves for layout save/restore
- `internal/permissions/` — accessibility and screen recording permission checks

Format with `just fmt`. See [OBJECTIVE_C.md](go/OBJECTIVE_C.md).

---

## Testing

```bash
just test-unit      # pure Go tests (no macOS APIs)
```

Integration tests for native APIs are not yet implemented. When adding them, use `//go:build integration` and name files `*_integration_test.go`.
