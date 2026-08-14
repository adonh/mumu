# Contributing to mumu

Thanks for your interest in contributing! mumu is a macOS window-to-Space layout tool with an approachable, layout-only codebase. We welcome contributions of all kinds — code, docs, bug reports, or ideas.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Commit Messages](#commit-messages)
- [Pull Requests](#pull-requests)
- [Testing](#testing)
- [Code Style](#code-style)
- [Good First Contributions](#good-first-contributions)
- [Reporting Bugs](#reporting-bugs)
- [Feature Requests](#feature-requests)

---

## Code of Conduct

This project follows our [Code of Conduct](CODE_OF_CONDUCT.md). By participating you agree to uphold it. Please report unacceptable behavior via [GitHub Issues](https://github.com/adonh/mumu/issues) or by contacting [@adonh](https://github.com/adonh) directly.

---

## Getting Started

1. **Search existing issues** — check if someone is already working on the same thing or if there's a related discussion.
2. **Open an issue first** for non-trivial changes — this avoids wasted effort and lets us align on approach before you write code.
3. **Small, focused PRs** are preferred over large, sweeping changes.

---

## Development Setup

### Prerequisites

- **Go 1.26+** — [Install Go](https://golang.org/dl/)
- **Xcode Command Line Tools** — `xcode-select --install`
- **Just** — command runner — `brew install just`
- **golangci-lint** — linter — `brew install golangci-lint`

### Clone & Verify

```bash
git clone https://github.com/adonh/mumu.git
cd mumu
go version          # Should be 1.26+
just --version
golangci-lint --version
just --list         # See all available commands
```

For full details see:

- [Development Guide](docs/DEVELOPMENT.md)
- [System Architecture](docs/ARCHITECTURE.md)

---

## Making Changes

1. **Fork** the repository and clone your fork.
2. **Create a branch** from `main`:

    ```bash
    git checkout -b feat/my-feature
    ```

3. **Make your changes** following the [Coding Standards](docs/CODING_STANDARDS.md).
4. **Add or update tests** for any new or changed functionality.
5. **Run the pre-commit checklist**:

    ```bash
    just fmt            # Format code
    just lint           # Run linters
    just test           # Run unit tests
    just build          # Verify build
    ```

6. **Commit** using [conventional commits](#commit-messages).
7. **Push** and open a pull request.

---

## Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/).

**Format:**

```
<type>(<optional scope>): <subject>
<optional body>
<optional footer>
```

**Types:**

| Type       | When to use                            |
| ---------- | -------------------------------------- |
| `feat`     | New feature                            |
| `fix`      | Bug fix                                |
| `docs`     | Documentation only                     |
| `style`    | Formatting, no logic change            |
| `refactor` | Code restructuring, no behavior change |
| `perf`     | Performance improvement                |
| `test`     | Adding or updating tests               |
| `chore`    | Build, CI, dependencies, tooling       |

**Examples:**

```
feat(layout): support restoring across mismatched display counts
fix(restore): correct positional fallback when titles are blank
docs: clarify space numbering in the CLI guide
```

---

## Pull Requests

- **Title** should follow the same conventional commit format (e.g. `feat(layout): add app sort order`).
- **Description** should explain _what_ changed and _why_.
- **Keep PRs focused** — one logical change per PR.
- **Link related issues** (e.g. `Closes #123`).
- All CI checks (lint, test, build) must pass before merge.
- A maintainer will review your PR. Be open to feedback and iterate.

---

## Testing

mumu separates tests into unit and integration tests:

| Type              | File pattern            | Command                 | Build tag     |
| ----------------- | ----------------------- | ----------------------- | ------------- |
| Unit tests        | `*_test.go`             | `just test-unit`        | —             |
| Integration tests | `*_integration_test.go` | `just test-integration` | `integration` |

**Guidelines:**

- All new code requires tests.
- Use **table-driven tests** where possible.
- Unit tests should be fast with no system dependencies.
- Integration tests use real macOS APIs and are tagged `//go:build integration && darwin`.

For detailed patterns see [Testing Patterns](docs/testing/TESTING_PATTERNS.md).

---

## Code Style

All code must follow the [Coding Standards](docs/CODING_STANDARDS.md):

- **Go**: [Go Conventions](docs/go/CONVENTIONS.md) — imports, naming, error handling, receiver conventions.
- **Objective-C**: [Objective-C Guidelines](docs/go/OBJECTIVE_C.md) — `.h`/`.m` files, memory management, naming.
- Format with `just fmt` (uses `golangci-lint` + `clang-format`).
- Lint with `just lint` (uses `golangci-lint`).
- Add godoc comments for all exported symbols.

---

## Good First Contributions

Not sure where to start? These are great entry points:

- 🐛 Bug fixes — check [open issues](https://github.com/adonh/mumu/issues)
- 📝 Documentation improvements or typo fixes
- ⚡ Performance improvements
- 🧪 Additional test coverage

---

## Reporting Bugs

Open a [GitHub Issue](https://github.com/adonh/mumu/issues/new) with:

1. **macOS version** and **mumu version** (`mumu --version`).
2. **Steps to reproduce** — minimal and specific.
3. **Expected vs actual behavior**.
4. **Output of `mumu status`** — confirms whether this is a permissions issue.

See also: [Troubleshooting Guide](docs/TROUBLESHOOTING.md) and [Security Policy](SECURITY.md) (for vulnerability reports).

---

## Feature Requests

Open a [GitHub Issue](https://github.com/adonh/mumu/issues/new) or start a [Discussion](https://github.com/adonh/mumu/discussions) describing:

- **What** you'd like to see.
- **Why** it would be useful (your use case).
- **How** you envision it working (optional but helpful).

---

Thank you for helping make mumu better!
