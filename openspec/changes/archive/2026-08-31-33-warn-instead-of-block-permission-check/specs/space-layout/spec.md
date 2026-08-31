## MODIFIED Requirements

### Requirement: No elevated privileges required

All layout save and restore operations SHALL function using only user-grantable macOS privacy permissions (Accessibility and Screen Recording) plus mumu's unprivileged private WindowServer connection. The system SHALL NOT require root/admin privileges or SIP modification. Screen Recording is required in addition to Accessibility specifically because window titles — needed for reliable restore matching — are inaccessible without it.

#### Scenario: Layout operations require only Accessibility and Screen Recording

- **WHEN** a user has granted mumu both Accessibility and Screen Recording permission
- **THEN** `mumu save` and `mumu restore` function without requesting or requiring root/admin privileges or any other elevated access

#### Scenario: Screen Recording not granted

- **WHEN** a user runs `mumu save` or `mumu restore` and mumu's own preflight check reports Screen Recording as not granted
- **THEN** the system prints a warning that Screen Recording permission does not appear to be granted and proceeds with the operation anyway, rather than stopping — since this preflight check can itself be wrong in some environments (see [GitHub Issue #33](https://github.com/adonh/mumu/issues/33)) — and reports whatever the operation's actual outcome is (success, or the specific failures/skips that result if the permission is genuinely missing)

#### Scenario: Accessibility not granted

- **WHEN** a user runs `mumu save` or `mumu restore` and mumu's own preflight check reports Accessibility as not granted
- **THEN** the system prints a warning that Accessibility permission does not appear to be granted and proceeds with the operation anyway, rather than stopping, for the same reason as the Screen Recording case above
