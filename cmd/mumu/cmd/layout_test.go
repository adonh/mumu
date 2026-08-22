package cmd //nolint:testpackage // Tests unexported restore summary rendering.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/adonh/mumu/internal/config"
	"github.com/adonh/mumu/internal/layout"
)

func TestPrintRestoreSummaryMarksFallbackFailures(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)

	printRestoreSummary(
		cmd,
		layout.RestoreSummary{
			Skipped: []layout.SkippedEntry{
				{
					Entry: layout.Entry{
						BundleID: "com.example.chrome",
						Title:    "window",
						Ordinal:  1,
					},
					Reason:   layout.SkipMoveFailed,
					Fallback: true,
				},
			},
		},
		layout.SortByDisplay,
	)

	if !strings.Contains(output.String(), "(fallback)") {
		t.Fatalf("restore summary = %q, want fallback marker", output.String())
	}
}

func TestPrintRestoreSummaryMarksFuzzyMatches(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)

	printRestoreSummary(
		cmd,
		layout.RestoreSummary{
			Skipped: []layout.SkippedEntry{
				{
					Entry: layout.Entry{
						BundleID: "com.example.chrome",
						Title:    "window",
						Ordinal:  1,
					},
					Reason: layout.SkipMoveFailed,
					Fuzzy:  true,
				},
			},
		},
		layout.SortByDisplay,
	)

	if !strings.Contains(output.String(), "(fuzzy)") {
		t.Fatalf("restore summary = %q, want fuzzy marker", output.String())
	}
}

func TestPrintConfiguredPins_ListsEachRule(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)

	printConfiguredPins(cmd, []config.PinRule{
		{BundleID: "com.tinyspeck.slackmacgap", Title: "general", Space: 1},
	})

	got := output.String()
	for _, want := range []string{"1 configured pin(s)", "com.tinyspeck.slackmacgap", "general"} {
		if !strings.Contains(got, want) {
			t.Fatalf("printConfiguredPins output = %q, want it to contain %q", got, want)
		}
	}
}

func TestPrintConfiguredPins_NoPinsPrintsNothing(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)

	printConfiguredPins(cmd, nil)

	if output.Len() != 0 {
		t.Fatalf(
			"printConfiguredPins output = %q, want empty output when no pins are configured",
			output.String(),
		)
	}
}
