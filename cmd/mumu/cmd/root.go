package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set via ldflags at build time.
	Version = "dev"
	// GitCommit is set via ldflags at build time.
	GitCommit = "unknown"
	// BuildDate is set via ldflags at build time.
	BuildDate = "unknown"
)

// RootCmd is the root cobra command for the mumu CLI.
var RootCmd = &cobra.Command{
	Use:   "mumu",
	Short: "Save and restore window-to-space layouts on macOS",
	Long: `mumu saves and restores window-to-Space layouts across display
configurations, without disabling SIP.

Layouts are saved and looked up by the number of currently connected
displays. Restoring only ever moves already-open windows to already-existing
Spaces: it never launches applications that aren't running, and never
creates, removes, or reorders Spaces.`,
}

// Execute runs the root command and returns any error.
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.Version = Version
	RootCmd.SetVersionTemplate(
		fmt.Sprintf(
			"Mumu version %s\nGit commit: %s\nBuild date: %s\n",
			Version,
			GitCommit,
			BuildDate,
		),
	)

	RootCmd.AddCommand(statusCmd)
}
