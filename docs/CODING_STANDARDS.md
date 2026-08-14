# mumu Coding Standards

This document defines the coding standards and conventions for the mumu project. Following these standards ensures the codebase appears written by a single developer and maintains consistency across all files.

---

## Table of Contents

- [Quick Reference](#quick-reference)
- [General Standards](#general-standards)
- [Error Handling](#error-handling)
- [Documentation Standards](#documentation-standards)
- [Git Commit Standards](#git-commit-standards)
- [Pre-commit Checklist](#pre-commit-checklist)
- [References](#references)

---

## Quick Reference

- [Go CONVENTIONS.md](./go/CONVENTIONS.md) — Go code style, imports, naming, error handling
- [Go OBJECTIVE_C.md](./go/OBJECTIVE_C.md) — .h/.m files, naming, memory management
- [TESTING_PATTERNS.md](./testing/TESTING_PATTERNS.md) — Test file naming, unit vs integration, table-driven tests

---

## General Standards

### File Formatting

All files must follow these basic formatting rules (enforced by `.editorconfig`):

- **Character encoding**: UTF-8
- **Line endings**: LF (Unix-style)
- **Indentation**: Tabs (width 4 spaces when displayed)
- **Trailing whitespace**: None
- **Final newline**: Required

### File Organization

```
mumu/
├── cmd/
│   └── mumu/           # Application entry point and CLI commands
├── internal/
│   ├── errors/         # Structured error types
│   ├── layout/         # Layout save/restore: capture, persistence, restore matching
│   ├── native/         # Objective-C + CGO bridge
│   ├── paths/          # Home-directory path expansion
│   ├── window/         # AX window wrappers
│   ├── space/          # Mission Control operations
│   └── permissions/    # Accessibility and Screen Recording checks
├── docs/               # Documentation
└── nix/                # Nix packaging
```

### Naming Conventions

- **Directories**: lowercase, underscore-separated
- **Files**: lowercase, underscore-separated
- **Test files**: `*_test.go`, `*_integration_test.go`

---

## Error Handling

Use the `derrors` package for structured errors:

```go
import derrors "github.com/adonh/mumu/internal/errors"

// Create new error
return derrors.New(derrors.CodeInternal, "something went wrong")

// Wrap existing error
return derrors.Wrapf(err, derrors.CodeSerializationFailed, "encoding layout")
```

Available error codes: `CodeAccessibilityDenied`, `CodeAccessibilityFailed`, `CodeScreenRecordingDenied`, `CodeInvalidInput`, `CodeActionFailed`, `CodeContextCanceled`, `CodeTimeout`, `CodeInternal`, `CodeSerializationFailed`, `CodeBridgeFailed`, `CodeNotSupported`.

---

## Documentation Standards

### Code Comments

**Do comment:**
- Complex algorithms or logic
- Non-obvious performance optimizations
- Workarounds for bugs or limitations
- Public APIs and exported symbols

**Don't comment:**
- Obvious code
- Redundant information already in the code
- Outdated information (update or remove)

### Package Documentation

Every package should have a `doc.go` file with package-level documentation.

---

## Git Commit Standards

### Format

```
<type>: <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Build process, dependencies, etc.

---

## Pre-commit Checklist

- [ ] Code formatted (`just fmt`)
- [ ] Linters pass (`just lint`)
- [ ] Tests pass (`just test`)
- [ ] Build succeeds (`just build`)
- [ ] Documentation updated if needed
- [ ] Commit message follows standards

---

## References

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Apple Coding Guidelines for Cocoa](https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/CodingGuidelines/CodingGuidelines.html)
