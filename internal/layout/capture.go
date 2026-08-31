package layout

import (
	"time"

	derrors "github.com/adonh/mumu/internal/errors"
	"github.com/adonh/mumu/internal/permissions"
	"github.com/adonh/mumu/internal/space"
	"github.com/adonh/mumu/internal/window"
)

// CaptureSummary reports counts from a Capture call.
type CaptureSummary struct {
	WindowsCaptured   int
	FullscreenSkipped int
}

// warnMissingPermissions reports (via progress) any of the permissions
// mumu needs that Check found not granted, without stopping the caller —
// see permissions.Warnings for why a missing permission isn't treated as
// fatal here.
func warnMissingPermissions(progress ProgressFunc) {
	for _, warning := range permissions.Warnings(permissions.Check()) {
		progress.emit("Warning: " + warning)
	}
}

// Capture records the current assignment of application windows to Mission
// Control spaces across all connected displays, using the logical
// left-to-right space numbering (see internal/space). Fullscreen windows
// are excluded entirely, per the space-layout capability's scope.
//
// progress, if non-nil, receives status updates while the (potentially
// slow) native window enumeration runs; pass nil to discard them.
func Capture(progress ProgressFunc) (*Layout, CaptureSummary, error) {
	warnMissingPermissions(progress)

	spaceCounts := space.LeftToRightSpaceCounts()

	displayCount := len(spaceCounts)
	if displayCount == 0 {
		return nil, CaptureSummary{}, derrors.New(
			derrors.CodeActionFailed,
			"failed to enumerate connected displays",
		)
	}

	progress.emit("Scanning windows across all Spaces...")

	entries, err := window.AllAcrossSpaces()
	if err != nil {
		return nil, CaptureSummary{}, derrors.Wrapf(
			err,
			derrors.CodeActionFailed,
			"failed to enumerate windows across spaces",
		)
	}

	summary := CaptureSummary{}
	layoutEntries := make([]Entry, 0, len(entries))
	bundleIndex := map[string]int{}

	for _, entry := range entries {
		if entry.Fullscreen {
			summary.FullscreenSkipped++

			continue
		}

		ordinal := space.OrdinalForSpace(entry.SpaceID)
		if ordinal == (space.Ordinal{}) {
			// Defensive: the native enumeration only returns resolved
			// windows, but skip rather than record a bogus ordinal.
			continue
		}

		idx := bundleIndex[entry.BundleID]
		bundleIndex[entry.BundleID] = idx + 1

		layoutEntries = append(layoutEntries, Entry{
			BundleID: entry.BundleID,
			Title:    entry.Title,
			Index:    idx,
			Ordinal:  ordinal,
		})
		summary.WindowsCaptured++
	}

	captured := &Layout{
		SchemaVersion: SchemaVersion,
		DisplayCount:  displayCount,
		SpaceCounts:   spaceCounts,
		Entries:       layoutEntries,
		SavedAt:       time.Now(),
	}

	return captured, summary, nil
}
