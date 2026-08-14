# Development Guide

## Prerequisites

- macOS (required for CGO/Objective-C)
- Go 1.26+
- [just](https://github.com/casey/just) (build system)
- [devbox](https://www.jetify.com/devbox) (optional, recommended)

```bash
devbox shell
just build
```

---

## Project Layout

```
cmd/mumu/              CLI binary
internal/
  layout/              Layout save/restore: capture, JSON persistence, restore matching
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
just build          # build bin/mumu
just test           # run all tests
just test-unit      # unit tests only
just lint           # golangci-lint
just fmt            # format Go + Objective-C
just bundle         # build Mumu.app
just genman         # generate man pages
```

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
