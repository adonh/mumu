package layout

import (
	"time"

	derrors "github.com/y3owk1n/mimi/internal/errors"
	"github.com/y3owk1n/mimi/internal/permissions"
	"github.com/y3owk1n/mimi/internal/space"
	"github.com/y3owk1n/mimi/internal/window"
)

// CaptureSummary reports counts from a Capture call.
type CaptureSummary struct {
	WindowsCaptured   int
	FullscreenSkipped int
}

func ensureLayoutPermissions() error {
	return permissions.FriendlyErrorLayout(permissions.CheckLayout())
}

// Capture records the current assignment of application windows to Mission
// Control spaces across all connected displays, using the logical
// left-to-right space numbering (see internal/space). Fullscreen windows
// are excluded entirely, per the space-layout capability's scope.
func Capture() (*Layout, CaptureSummary, error) {
	err := ensureLayoutPermissions()
	if err != nil {
		return nil, CaptureSummary{}, err
	}

	spaceCounts := space.LeftToRightSpaceCounts()

	displayCount := len(spaceCounts)
	if displayCount == 0 {
		return nil, CaptureSummary{}, derrors.New(
			derrors.CodeActionFailed,
			"failed to enumerate connected displays",
		)
	}

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

		ordinal := space.LogicalIndexForSpace(entry.SpaceID)
		if ordinal == 0 {
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
