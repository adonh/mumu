package cmd

import (
	"bufio"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	derrors "github.com/adonh/mumu/internal/errors"
	"github.com/adonh/mumu/internal/layout"
	"github.com/adonh/mumu/internal/space"
)

var restoreAssumeYes bool

var (
	showSort       string
	restoreSort    string
	layoutJSONFlag bool
)

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
"mumu save", auto-detected by the number of currently connected displays.

Only already-running applications are affected: apps that aren't running
are skipped, never launched. For each app, its saved entries and open
windows are matched as a batch by title similarity — shared words, ignoring
case and word order — assigning the closest-matching pairs first so no
window is ever claimed by more than one saved entry; an entry goes
unmatched only once its app has no open window left to claim. Matches
that aren't exact are marked "(fuzzy)" in output (this matters for apps
like browsers, whose titles rarely match exactly). Other open windows of
an app are placed in its most prevalent matched Space; tied targets use
the Space currently shown on the primary (menu-bar) display. Apps without a
matching assignment are left unchanged. Restore never creates or removes
Spaces; entries whose saved Space no longer exists are skipped and reported.

If the current per-display Space-count arrangement has changed since the
layout was saved, you'll be asked to confirm before any windows are moved.
Use --yes to skip the prompt.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		sortKey, err := layout.ParseSortKey(restoreSort)
		if err != nil {
			return err
		}

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

			if !restoreAssumeYes &&
				!promptConfirm(cmd, "Continue with a best-effort restore anyway?") {
				cmd.Println("Restore aborted; no windows were moved.")

				return nil
			}
		}

		summary, err := layout.Restore(saved, sortKey, func(msg string) { cmd.Println(msg) })
		if err != nil {
			return err
		}

		printRestoreSummary(cmd, summary, sortKey)

		return nil
	},
}

var layoutLayoutCmd = &cobra.Command{
	Use:   "layout",
	Short: "Show the current saved layout for connected displays",
	Long: `Display the saved layout for the current number of connected displays
in a single-line summary, or as JSON with the --json flag.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		displayCount, err := currentDisplayCount()
		if err != nil {
			return err
		}

		saved, err := layout.Load(displayCount)
		if err != nil {
			return err
		}

		if layoutJSONFlag {
			data := struct {
				DisplayCount int       `json:"display_count"` //nolint:tagliatelle // Stable CLI JSON field name.
				WindowCount  int       `json:"window_count"`  //nolint:tagliatelle // Stable CLI JSON field name.
				SavedAt      time.Time `json:"saved_at"`      //nolint:tagliatelle // Stable CLI JSON field name.
			}{
				DisplayCount: saved.DisplayCount,
				WindowCount:  len(saved.Entries),
				SavedAt:      saved.SavedAt,
			}

			bytes, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return derrors.Wrapf(
					err,
					derrors.CodeSerializationFailed,
					"encoding layout to JSON",
				)
			}

			cmd.Println(string(bytes))
		} else {
			cmd.Printf("%d display saved %s\n",
				saved.DisplayCount,
				saved.SavedAt.Format(time.RFC3339),
			)
		}

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
		sortKey, err := layout.ParseSortKey(showSort)
		if err != nil {
			return err
		}

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
		layout.SortEntries(entries, sortKey)

		for idx, entry := range entries {
			cmd.Printf(
				"  %s %s — %s — %q\n",
				layout.FormatIndex(idx+1, len(entries)),
				space.DualLabel(entry.Ordinal),
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

func printRestoreSummary(
	cmd *cobra.Command,
	summary layout.RestoreSummary,
	sortKey layout.SortKey,
) {
	cmd.Printf("Restored %d window(s)\n", summary.Moved)

	if len(summary.Skipped) == 0 {
		return
	}

	byReason := map[layout.SkipReason][]layout.SkippedEntry{}
	for _, s := range summary.Skipped {
		byReason[s.Reason] = append(byReason[s.Reason], s)
	}

	cmd.Printf("Skipped %d entry(ies):\n", len(summary.Skipped))

	for _, reason := range []layout.SkipReason{
		layout.SkipAppNotRunning,
		layout.SkipUnmatchedWindow,
		layout.SkipOrdinalOutOfRange,
		layout.SkipFallbackTargetUnavailable,
		layout.SkipMoveFailed,
	} {
		skippedEntries := byReason[reason]
		if len(skippedEntries) == 0 {
			continue
		}

		layout.SortSkippedEntries(skippedEntries, sortKey)

		cmd.Printf("  %s (%d):\n", reason, len(skippedEntries))

		for _, skipped := range skippedEntries {
			marker := ""
			if skipped.Fallback {
				marker += " (fallback)"
			}

			if skipped.Fuzzy {
				marker += " (fuzzy)"
			}

			entry := skipped.Entry
			cmd.Printf(
				"    - %s — %q (%s)%s\n",
				entry.BundleID,
				displayTitle(entry.Title),
				space.DualLabel(entry.Ordinal),
				marker,
			)
		}
	}
}

// addSortFlag registers the shared "--sort" flag, used identically by
// "mumu show" and "mumu restore" to control per-entry output order (see
// layout.SortKey).
func addSortFlag(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVar(
		dest,
		"sort",
		string(layout.SortByDisplay),
		`Order entries by: "display" (logical left-to-right Space number, default), `+
			`"macos" (macOS Mission Control Space number), or "app" (bundle identifier)`,
	)
}

func init() {
	layoutRestoreCmd.Flags().
		BoolVarP(&restoreAssumeYes, "yes", "y", false, "Skip the arrangement-mismatch confirmation prompt")

	addSortFlag(layoutShowCmd, &showSort)
	addSortFlag(layoutRestoreCmd, &restoreSort)

	layoutLayoutCmd.Flags().
		BoolVar(&layoutJSONFlag, "json", false, "Output as JSON instead of a single-line summary")

	RootCmd.AddCommand(layoutSaveCmd)
	RootCmd.AddCommand(layoutRestoreCmd)
	RootCmd.AddCommand(layoutLayoutCmd)
	RootCmd.AddCommand(layoutListCmd)
	RootCmd.AddCommand(layoutShowCmd)
	RootCmd.AddCommand(layoutDeleteCmd)
}
