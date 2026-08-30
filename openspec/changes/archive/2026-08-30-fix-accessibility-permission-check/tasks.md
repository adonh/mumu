## 1. Accessibility check API

- [x] 1.1 Swap `AXIsProcessTrusted()` for `AXIsProcessTrustedWithOptions(NULL)` in `internal/permissions/permissions.m`, and add a comment documenting the known per-process TCC trust-cache quirk (both functions share it) and that it's unrelated to the fix in section 2 below; verify with `just build` and a manual `mumu status` run.
- [x] 1.2 Re-audit `internal/native/*.m` and `internal/permissions/permissions.m` for any other deprecated/stale Apple API usage found during investigation (`CGWindowListCopyWindowInfo`, `CGPreflightScreenCaptureAccess`, `AXUIElement*` family); confirm in a code comment or PR note that none needed changes, so this isn't re-investigated later.

## 2. Stable local code-signing identity

- [x] 2.1 Document (in `docs/CODING_STANDARDS.md` or a new short doc section) the one-time step to create a local self-signed code-signing certificate for `mumu` development, and why it's needed (TCC pins Accessibility/Screen Recording grants to a code identity; ad-hoc signing pins that identity to the binary's own hash, which changes every rebuild).
- [x] 2.2 Add a `just setup-codesign-identity` recipe that creates the certificate (in a dedicated `mumu-dev.keychain-db`, not the login keychain — see design note below) if it doesn't already exist, and is a no-op (with a clear message) if it does.
- [x] 2.3 Update the `build` recipe in `justfile` to codesign `bin/mumu` with that identity when present, falling back to ad-hoc signing (today's unsigned behavior) with a printed warning when absent.
- [x] 2.4 Update the `bundle` recipe in `justfile` to use the same identity (replacing `--sign -`) when present, falling back to ad-hoc signing with a printed warning when absent.
- [x] 2.5 Verify the fix end-to-end: grant Accessibility to a locally built `mumu`, rebuild via `just build` (and separately via `just bundle`), and confirm `mumu status` still reports Accessibility as granted without re-granting, both with the identity present and (documenting the expected caveat) understood when absent. **Result:** cryptographic verification passed (`codesign -d -r-` shows an identical certificate-leaf-based designated requirement across repeated `just build`/`just bundle` runs). Live grant-persistence verification is blocked on this specific test machine by an MDM-enforced policy (see design.md's "Verification environment note") — confirmed via a minimal isolated test binary (no mumu code, same signing identity) that also fails `AXIsProcessTrustedWithOptions`/`AXIsProcessTrusted` after an identical grant flow, ruling out a mumu-side bug.

## 3. Spec and test verification

- [x] 3.1 Run `openspec validate --strict` for this change and confirm it passes.
- [x] 3.2 Run `just fmt`, `just lint`, `just test`, and `just build` and confirm all pass.
