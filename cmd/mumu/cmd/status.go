package cmd

import (
	"github.com/spf13/cobra"

	"github.com/adonh/mumu/internal/permissions"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Accessibility and Screen Recording permission status",
	RunE: func(cmd *cobra.Command, _ []string) error {
		perm := permissions.Check()
		if perm.Accessibility {
			cmd.Println("accessibility: granted")
		} else {
			cmd.Println("accessibility: not granted (required for window/space control)")
		}

		if perm.ScreenRecording {
			cmd.Println("screen recording: granted")
		} else {
			cmd.Println(
				"screen recording: not granted (required to read window titles for restore matching)",
			)
		}

		return nil
	},
}
