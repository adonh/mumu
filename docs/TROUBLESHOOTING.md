# Troubleshooting

## `mumu save`/`restore` prints a permission warning, or moves no windows

`mumu save` and `mumu restore` print a `Warning: ... permission does not appear to be granted` line and then proceed anyway when their own preflight check thinks Accessibility or Screen Recording is missing — they don't stop, because that preflight check can itself be wrong (see below). If the permission really is missing, the warning is followed by the actual failure it causes (e.g. no windows captured, or entries skipped for lack of a title to match on); if it's a false alarm, the operation completes normally past the warning.

1. Run `mumu status` — confirm both Accessibility and Screen Recording are granted.
2. Make sure you granted permission to the exact binary you run (Homebrew cask path vs. a locally built `bin/mumu`) — granting one doesn't carry over to the other.
3. If you just rebuilt `mumu` from source and it worked before but not now, this is expected with the default build setup: `just build`/`just bundle` ad-hoc sign by default, and macOS's TCC ties an Accessibility grant to the binary's code identity — ad-hoc signing's identity changes on every rebuild, so a grant that worked before the rebuild can silently stop matching. This isn't a delay/cache-catchup issue and won't resolve itself by re-running the command. Run `just setup-codesign-identity` once (see [DEVELOPMENT.md](DEVELOPMENT.md#code-signing-and-accessibility-permissions)) so future rebuilds keep a stable identity, then re-grant Accessibility/Screen Recording one more time.
4. On a managed Mac (MDM/enterprise-enrolled), a Privacy Preferences Policy Control (PPPC) profile can override a manual grant even though the toggle in System Settings looks correctly enabled — check with your IT admin if the warning persists despite the steps above and (1)-(3) all check out.

## Permission prompt keeps reappearing

Remove and re-add `mumu` in **System Settings → Privacy & Security → Accessibility** (and **Screen Recording**). Ensure you're granting the binary you actually execute.

## `mumu restore` warns about Screen Recording after granting it

1. Confirm with `mumu status` — it reports Accessibility and Screen Recording separately.
2. Window titles (used for restore matching) come back empty without Screen Recording even when Accessibility is granted — the two permissions are independent, and both are required. Missing Screen Recording surfaces as unmatched/skipped entries in the restore summary rather than as a separate error, since restore now warns and continues instead of stopping outright (see above).

## Restore skips windows I expect it to move

- **App isn't running**: restore never launches apps — only windows belonging to already-running applications are moved. Start the app first, then restore again.
- **Target Space no longer exists**: restore never creates Spaces. If you removed a Space since saving, entries for it are skipped and listed in the post-restore summary.
- **Ambiguous title match**: if an app has multiple windows with the same (or blank) title, `mumu` falls back to positional matching; if more than one candidate remains ambiguous, the extras are skipped rather than guessed. Check the skip summary printed after restore for the specific bundle ID and title.

## Display count mismatch prompt

If restore detects your current display setup has a different Space count than what was saved, it asks for confirmation before moving anything (since a saved Space number may not map to what you expect on a different arrangement). Use `mumu restore --yes` to skip this prompt, or run `mumu save` again to capture a fresh layout for the current setup.
