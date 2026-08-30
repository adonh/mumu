## Why

`mumu`'s Accessibility permission check can disagree with what System Settings shows: a user grants Accessibility, but mumu still reports it as missing (or vice versa). Investigation found two contributing issues: (1) the runtime check uses the older `AXIsProcessTrusted()` entry point instead of the more current, forward-compatible `AXIsProcessTrustedWithOptions(NULL)`, and (2) mumu's dev/build tooling produces binaries with an inconsistent code identity across rebuilds (`bin/mumu` from `just build` is unsigned; `build/Mumu.app` from `just bundle` is ad-hoc signed via `codesign --sign -`), which is a well-documented way for macOS's TCC subsystem to silently invalidate a previously granted Accessibility permission on the next rebuild while System Settings still shows the old grant as checked.

## What Changes

- Switch the native Accessibility check in `internal/permissions/permissions.m` from `AXIsProcessTrusted()` to `AXIsProcessTrustedWithOptions(NULL)`, which is the currently documented entry point for a non-prompting trust check and keeps the door open for an explicit prompting variant later.
- Audit all other native (Objective-C) system calls under `internal/native` and `internal/permissions` for stale/deprecated API usage; no other changes found necessary (`CGWindowListCopyWindowInfo`, `CGPreflightScreenCaptureAccess`, and the `AXUIElement*` family are all current).
- Establish a consistent code-signing identity for locally built `mumu` binaries (`just build` and `just bundle`) so that Accessibility (and Screen Recording) grants survive routine rebuilds during development, instead of silently breaking on every `just build`/`just bundle`.
- Document the known macOS quirk (per-process TCC trust caching, and grants being pinned to a code identity) in code comments near the permission check, so this isn't rediscovered from scratch later.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `space-layout`: the "Permission status command" requirement gains an accuracy guarantee — `mumu status` (and any other command that checks Accessibility permission) must reflect the live, current Accessibility grant state rather than a stale cached value, including immediately after a user grants permission in System Settings or after mumu itself was rebuilt/resigned.

## Impact

- `internal/permissions/permissions.m`, `internal/permissions/permissions.h`: swap the Accessibility check call.
- `justfile`: `build` and `bundle` recipes gain consistent code-signing so the built binary/app keeps a stable identity across rebuilds.
- No changes to `internal/native/*` beyond the audit (no stale APIs found there).
- No changes to CLI output format or `internal/config`.
