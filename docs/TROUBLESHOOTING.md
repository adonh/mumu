# Troubleshooting

## `mumu save`/`restore` fails or refuses to run

1. Run `mumu status` — confirm both Accessibility and Screen Recording are granted.
2. Make sure you granted permission to the exact binary you run (Homebrew cask path vs. a locally built `bin/mumu`) — granting one doesn't carry over to the other.
3. If you just rebuilt `mumu` (e.g. from source) after granting, macOS's permission cache (TCC) can take a few seconds to catch up even though the toggle shows enabled — re-run the command once more before assuming it's broken.

## Permission prompt keeps reappearing

Remove and re-add `mumu` in **System Settings → Privacy & Security → Accessibility** (and **Screen Recording**). Ensure you're granting the binary you actually execute.

## `mumu restore` reports Screen Recording denied after granting it

1. Confirm with `mumu status` — it reports Accessibility and Screen Recording separately.
2. Window titles (used for restore matching) come back empty without Screen Recording even when Accessibility is granted — the two permissions are independent, and both are required.

## Restore skips windows I expect it to move

- **App isn't running**: restore never launches apps — only windows belonging to already-running applications are moved. Start the app first, then restore again.
- **Target Space no longer exists**: restore never creates Spaces. If you removed a Space since saving, entries for it are skipped and listed in the post-restore summary.
- **Ambiguous title match**: if an app has multiple windows with the same (or blank) title, `mumu` falls back to positional matching; if more than one candidate remains ambiguous, the extras are skipped rather than guessed. Check the skip summary printed after restore for the specific bundle ID and title.

## Display count mismatch prompt

If restore detects your current display setup has a different Space count than what was saved, it asks for confirmation before moving anything (since a saved Space number may not map to what you expect on a different arrangement). Use `mumu restore --yes` to skip this prompt, or run `mumu save` again to capture a fresh layout for the current setup.
