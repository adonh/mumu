# Go Conventions

## Package Organization

### Package Names

- Use short, lowercase, single-word names when possible
- Avoid underscores, hyphens, or mixed caps

```go
package layout
package space
package permissions
```

### Package Documentation

Every package should have a `doc.go` file with package-level documentation:

```go
// Package layout captures, persists, and restores window-to-Space layouts for mumu.
package layout
```

## File Structure

1. Package declaration
2. Imports (organized by `goimports`)
3. Constants
4. Type definitions
5. Constructor functions
6. Methods (grouped by receiver type)
7. Helper functions

## Imports

Organized by `goimports` into three groups:

1. Standard library
2. External packages
3. Internal packages

Use blank lines between groups:

```go
import (
  "context"
  "os"

  "github.com/spf13/cobra"

  "github.com/adonh/mumu/internal/layout"
)
```

## Naming

- Packages: lowercase, short, descriptive
- Variables: camelCase local, PascalCase exported
- Constants: PascalCase exported, camelCase unexported
- Receiver names: consistent single-letter (e.g., `c` for `Capture`, `r` for `Restorer`)

## Function Parameters

- `context.Context` first parameter when needed for cancellable operations
- Required parameters before optional

## Return Values

- Return errors as the last value
- Use named return values sparingly

```go
func Load(displayCount int) (*Layout, error) {
  data, err := os.ReadFile(pathFor(displayCount))
  if err != nil {
    return nil, err
  }
  return parse(data)
}
```

## Error Handling

Use the `derrors` package for structured errors:

```go
import derrors "github.com/adonh/mumu/internal/errors"

// Create new error
return derrors.New(derrors.CodeInvalidInput, "display count must be positive")

// Wrap existing error
return derrors.Wrapf(err, derrors.CodeSerializationFailed, "encoding layout")
```

## Context

- Accept `context.Context` as first parameter for cancellable operations
- Don't store context in structs

## Concurrency

### Mutex Usage

- Use `sync.RWMutex` for read-heavy workloads
- Use `sync.Mutex` for write-heavy or simple cases
- Always defer unlock immediately after lock

### Goroutines

- Use a semaphore pattern (`chan struct{}`) to limit concurrent goroutines
- Always provide a mechanism for graceful shutdown via context cancellation

## Comments

- Comment public APIs and exported symbols
- Use complete sentences with proper punctuation
- Explain _why_ for non-obvious code, not _what_

```go
// Pace window moves so WindowServer has time to catch up; moving many
// windows back-to-back can otherwise silently drop moves.
time.Sleep(moveDelay)
```

## Performance

### Pre-allocation

```go
entries := make([]Entry, 0, expectedCount)
```

## macOS-Specific Conventions

mumu is macOS-only, so no cross-platform build tags or platform factories are needed. CGo code lives in `internal/native/`.

## See Also

- [TESTING_PATTERNS.md](../testing/TESTING_PATTERNS.md)
- [OBJECTIVE_C.md](./OBJECTIVE_C.md)
