package cmd

import (
	"bufio"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	derrors "github.com/y3owk1n/mimi/internal/errors"
	"github.com/y3owk1n/mimi/internal/layout"
	"github.com/y3owk1n/mimi/internal/space"
)

// layoutCmd saves and restores window-to-space layouts.
var layoutCmd = &cobra.Command{
	Use:   "layout",
	Short: "Save and restore window-to-space layouts",
	Long: `Save and restore window-to-Space layouts across display configurations.

Layouts are saved and looked up by the number of currently connected
displays. Restoring only ever moves already-open windows to already-existing
Spaces: it never launches applications that aren't running, and never
creates, removes, or reorders Spaces.

Space numbers used by this command group are counted left to right across
all connected displays, independent of which display is primary — matching
how a person visually counts Spaces on screen. This numbering is specific
to "mimi layout" and differs from "mimi action space"'s Mission Control
ordering.

Requires both Accessibility and Screen Recording permissions (see
'mimi status'); Screen Recording is needed to read window titles for
restore matching, and is not required by any other mimi command.

Subcommands:
  save       Save the current window-to-space layout
  restore    Restore the saved layout for the current display count
  list       List all saved layouts
  show       Show the contents of a saved layout
  delete     Delete a saved layout

Examples:
  mimi layout save
  mimi layout restore
  mimi layout list
  mimi layout show
  mimi layout delete 2`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return derrors.New(
			derrors.CodeInvalidInput,
			"layout subcommand required (e.g., mimi layout save, mimi layout restore)",
		)
	},
}

var layoutRestoreAssumeYes bool

var layoutSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save the current window-to-space layout",
	Long: `Save the current assignment of application windows to Mission Control
Spaces, keyed by the number of currently connected displays.

Fullscreen windows are always skipped — they have no meaningful Space
assignment for this feature. Minimized windows are skipped only on the
Space currently displayed on each display; minimized windows on other
Spaces cannot currently be detected and may be captured as if they were
not minimized.

Overwrites any previously saved layout for the same display count.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		captured, summary, err := layout.Capture(func(msg string) { cmd.Println(msg) })
		if err != nil {
			return err
		}

		err = layout.Save(captured)
		if err != nil {
			return err
		}

		cmd.Printf(
			"Saved layout for %d display(s): %d window(s) captured",
			captured.DisplayCount,
			summary.WindowsCaptured,
		)

		if summary.FullscreenSkipped > 0 {
			cmd.Printf(", %d fullscreen window(s) skipped", summary.FullscreenSkipped)
		}

		cmd.Println()

		return nil
	},
}

var layoutRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the saved layout for the current display count",
	Long: `Restore application windows to the Mission Control Spaces recorded by
"mimi layout save", auto-detected by the number of currently connected
displays.

Only already-running applications are affected: apps that aren't running
are skipped, never launched. Windows are matched to saved entries by exact
title first, falling back to positional order within the same app; if
exactly one of an app's windows remains unmatched, it's used regardless of
title or position, since there's no real ambiguity left (this matters for
apps like browsers, whose titles rarely match exactly). Restore never
creates or removes Spaces; entries whose saved Space no longer exists are
skipped and reported.

If the current per-display Space-count arrangement has changed since the
layout was saved, you'll be asked to confirm before any windows are moved.
Use --yes to skip the prompt.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		displayCount, err := currentDisplayCount()
		if err != nil {
			return err
		}

		saved, err := layout.Load(displayCount)
		if err != nil {
			return err
		}

		drift := layout.DetectDrift(saved)
		if drift.Mismatched() {
			cmd.Println("Warning: the current display arrangement doesn't match what was saved.")
			cmd.Printf("  Saved space-count sequence:   %v\n", drift.Saved)
			cmd.Printf("  Current space-count sequence: %v\n", drift.Current)

			if !layoutRestoreAssumeYes &&
				!promptConfirm(cmd, "Continue with a best-effort restore anyway?") {
				cmd.Println("Restore aborted; no windows were moved.")

				return nil
			}
		}

		summary, err := layout.Restore(saved, func(msg string) { cmd.Println(msg) })
		if err != nil {
			return err
		}

		printRestoreSummary(cmd, summary)

		return nil
	},
}

var layoutListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved layouts",
	RunE: func(cmd *cobra.Command, _ []string) error {
		counts, err := layout.List()
		if err != nil {
			return err
		}

		if len(counts) == 0 {
			cmd.Println("No saved layouts.")

			return nil
		}

		for _, dc := range counts {
			saved, loadErr := layout.Load(dc)
			if loadErr != nil {
				cmd.Printf("%d display(s): (error reading layout: %s)\n", dc, loadErr)

				continue
			}

			cmd.Printf(
				"%d display(s): %d window(s), saved %s\n",
				saved.DisplayCount,
				len(saved.Entries),
				saved.SavedAt.Format(time.RFC3339),
			)
		}

		return nil
	},
}

var layoutShowCmd = &cobra.Command{
	Use:   "show [display-count]",
	Short: "Show the contents of a saved layout",
	Long: `Display a saved layout's window entries without moving any windows.

If display-count is omitted, shows the layout for the current number of
connected displays.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		displayCount, err := resolveDisplayCountArg(args)
		if err != nil {
			return err
		}

		saved, err := layout.Load(displayCount)
		if err != nil {
			return err
		}

		cmd.Printf(
			"Layout for %d display(s), saved %s\n",
			saved.DisplayCount,
			saved.SavedAt.Format(time.RFC3339),
		)
		cmd.Printf("Space-count sequence: %v\n", saved.SpaceCounts)
		cmd.Printf("%d window(s):\n", len(saved.Entries))

		entries := make([]layout.Entry, len(saved.Entries))
		copy(entries, saved.Entries)
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Ordinal != entries[j].Ordinal {
				return entries[i].Ordinal < entries[j].Ordinal
			}

			return entries[i].BundleID < entries[j].BundleID
		})

		for _, entry := range entries {
			cmd.Printf(
				"  space %d: %s — %q\n",
				entry.Ordinal,
				entry.BundleID,
				displayTitle(entry.Title),
			)
		}

		return nil
	},
}

var layoutDeleteCmd = &cobra.Command{
	Use:   "delete [display-count]",
	Short: "Delete a saved layout",
	Long: `Delete the saved layout for the given display count.

If display-count is omitted, deletes the layout for the current number of
connected displays.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		displayCount, err := resolveDisplayCountArg(args)
		if err != nil {
			return err
		}

		err = layout.Delete(displayCount)
		if err != nil {
			return err
		}

		cmd.Printf("Deleted saved layout for %d display(s)\n", displayCount)

		return nil
	},
}

func currentDisplayCount() (int, error) {
	counts := space.LeftToRightSpaceCounts()
	if len(counts) == 0 {
		return 0, derrors.New(derrors.CodeActionFailed, "failed to enumerate connected displays")
	}

	return len(counts), nil
}

func resolveDisplayCountArg(args []string) (int, error) {
	if len(args) == 0 {
		return currentDisplayCount()
	}

	raw := strings.TrimSpace(args[0])

	displayCount, err := strconv.Atoi(raw)
	if err != nil || displayCount < 1 {
		return 0, derrors.Newf(
			derrors.CodeInvalidInput,
			"display-count must be a positive integer, got %q",
			args[0],
		)
	}

	return displayCount, nil
}

func promptConfirm(cmd *cobra.Command, question string) bool {
	cmd.Printf("%s [y/N]: ", question)

	reader := bufio.NewReader(cmd.InOrStdin())

	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return false
	}

	answer := strings.ToLower(strings.TrimSpace(line))

	return answer == "y" || answer == "yes"
}

func displayTitle(title string) string {
	if title == "" {
		return "(untitled)"
	}

	return title
}

func printRestoreSummary(cmd *cobra.Command, summary layout.RestoreSummary) {
	cmd.Printf("Restored %d window(s)\n", summary.Moved)

	if len(summary.Skipped) == 0 {
		return
	}

	byReason := map[layout.SkipReason][]layout.Entry{}
	for _, s := range summary.Skipped {
		byReason[s.Reason] = append(byReason[s.Reason], s.Entry)
	}

	cmd.Printf("Skipped %d entry(ies):\n", len(summary.Skipped))

	for _, reason := range []layout.SkipReason{
		layout.SkipAppNotRunning,
		layout.SkipUnmatchedWindow,
		layout.SkipOrdinalOutOfRange,
		layout.SkipMoveFailed,
	} {
		entries := byReason[reason]
		if len(entries) == 0 {
			continue
		}

		cmd.Printf("  %s (%d):\n", reason, len(entries))

		for _, entry := range entries {
			cmd.Printf(
				"    - %s — %q (space %d)\n",
				entry.BundleID,
				displayTitle(entry.Title),
				entry.Ordinal,
			)
		}
	}
}

func init() {
	layoutRestoreCmd.Flags().
		BoolVarP(&layoutRestoreAssumeYes, "yes", "y", false, "Skip the arrangement-mismatch confirmation prompt")

	layoutCmd.AddCommand(layoutSaveCmd)
	layoutCmd.AddCommand(layoutRestoreCmd)
	layoutCmd.AddCommand(layoutListCmd)
	layoutCmd.AddCommand(layoutShowCmd)
	layoutCmd.AddCommand(layoutDeleteCmd)
}
