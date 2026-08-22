# Security Policy

## Supported Versions

Only the **latest release** receives security fixes.

| Version        | Supported |
| -------------- | --------- |
| Latest release | Yes       |
| Older releases | No        |

---

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Report privately via [GitHub Security Advisories](https://github.com/adonh/mumu/security/advisories/new) or contact [@adonh](https://github.com/adonh).

---

## Security Model

### Permissions

mumu requires **macOS Accessibility** and **Screen Recording** permission to save and restore window-to-Space layouts (`mumu save`, `mumu restore`). With these granted, mumu can enumerate windows, read their titles, and move them between Spaces. It does not record, transmit, or log window content — window titles are only used in-memory for restore matching and are written to local JSON files under `data_dir/layouts/`.

### No Network Access

mumu makes no outbound network connections, sends no telemetry, and does not phone home.

### No Background Process

mumu is a plain, one-shot CLI. It runs only when a command is invoked and exits when the command finishes — there is no daemon, background process, or auto-start mechanism.

### CGo / Objective-C

Native code lives in `internal/native/`. Window enumeration and window-to-Space moves use undocumented SkyLight and CGWindowList private APIs. Report memory-safety issues in this layer promptly.

### Private APIs

Space-related native calls use reverse-engineered private macOS APIs. They may break on OS updates and are not security-reviewed by Apple.
