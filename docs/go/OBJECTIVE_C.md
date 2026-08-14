# Objective-C Guidelines

## File Organization

### CGO and Go Files

Native implementations belong in `.m` / `.h` files:

- `internal/native/` — window/space enumeration and move APIs for layout save/restore (SkyLight, Accessibility, CGWindowList)
- `internal/permissions/` — Accessibility and Screen Recording permission checks

Go files use a minimal CGO preamble (`#include` headers, `#cgo` flags, and `extern` declarations for `//export` callbacks only).

Bridge `.m` files must `#include` their matching header and must **not** re-declare structs or typedefs already defined in that header (duplicate definitions cause `conflicting types` errors when CGO includes the same header).

### Header Files (.h)

- Minimal public interface
- Use `@class` forward declarations when possible
- Group related declarations with `#pragma mark`

```objc
#import <Foundation/Foundation.h>

bool MumuCheckAccessibilityPermissions(void);
bool MumuCheckScreenRecordingPermissions(void);
```

### Implementation Files (.m)

Standard structure:

1. Imports
2. `#pragma mark` sections
3. Interface declarations (private)
4. Implementation
5. C interface functions

```objc
#import "mumu.h"
#import <Cocoa/Cocoa.h>

#pragma mark - C Interface

uint64_t MumuActiveSpaceID(void) {
    // Implementation
}
```

## Naming Conventions

### C Bridge Exports

Functions declared in `.h` files and called from Go via CGO use the `Mumu` prefix:

```objc
uint64_t MumuActiveSpaceID(void);
int MumuMissionControlIndexForSpace(uint64_t sid);
bool MumuMoveWindowToSpace(uint32_t windowID, uint64_t spaceID);
```

### Objective-C Methods

- Use descriptive names with clear intent
- Follow Apple's naming conventions
- Start with lowercase letter, use camelCase

```objc
- (void)startObserving;
- (void)stopObserving;
```

## Property Attributes

- `strong` for object ownership
- `weak` for delegates and to avoid retain cycles
- `assign` for primitive types
- `copy` for NSString and blocks

```objc
@property(nonatomic, strong) NSWindow *window;
@property(nonatomic, weak) id<NSWorkspaceDelegate> delegate;
@property(nonatomic, assign) NSInteger eventCount;
```

## Memory Management

### ARC

mumu uses Automatic Reference Counting (ARC) for Objective-C code. The compiler handles `retain`/`release` automatically.

### C Interface Objects

For objects passed across the C/Go boundary, use toll-free bridging or `__bridge` casts:

```objc
CFArrayRef MumuCopySpaceIDs(void) {
    return (__bridge CFArrayRef)someNSArray;
}
```

## Comments

Use HeaderDoc-style comments for public API:

```objc
/// Returns the Mission Control ordinal (1-based) for the given space ID,
/// matching macOS's own "Switch to Desktop <n>" keyboard shortcut.
int MumuMissionControlIndexForSpace(uint64_t sid);
```

Inline comments for non-obvious logic:

```objc
// CGWindowList sees windows on every Space; AX only sees the currently
// displayed Space on each screen, so layout capture must use CGWindowList.
```

## Code Organization

Use `#pragma mark` to organize code:

```objc
#pragma mark - Space Lookup

#pragma mark - C Interface
```

## Threading

All Cocoa/UI code must run on the main thread:

```objc
if ([NSThread isMainThread]) {
    [self doWork];
} else {
    dispatch_async(dispatch_get_main_queue(), ^{
        [self doWork];
    });
}
```

## See Also

- [CONVENTIONS.md](./CONVENTIONS.md)
- [TESTING_PATTERNS.md](../testing/TESTING_PATTERNS.md)
