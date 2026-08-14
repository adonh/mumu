package cmd //nolint:testpackage // Tests unexported restore summary rendering.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

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
