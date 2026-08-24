package layout

import (
	"time"

	"github.com/adonh/mumu/internal/space"
)

// SchemaVersion is the current on-disk layout schema version. Bumped
// whenever the persisted JSON shape changes incompatibly (most recently:
// Entry.Ordinal changed from a flat number to a {"display","space"}
// object) — see persist.go's pre-parse version check.
const SchemaVersion = 2

// Entry represents a single window's saved space assignment.
type Entry struct {
	// BundleID is the owning application's bundle identifier.
	BundleID string `json:"bundleId"`
	// Title is the window's title at save time, used for restore matching.
	Title string `json:"title"`
	// Index is the window's 0-based position among its application's other
	// captured (non-fullscreen) windows, in save-time enumeration order.
	// Used as the restore-time matching fallback when title matching is
	// ambiguous or unavailable.
	Index int `json:"index"`
	// Ordinal is the window's logical two-part (display, space-within-display)
	// ordinal (see internal/space's logical numbering), independent of
	// which display is primary.
	Ordinal space.Ordinal `json:"ordinal"`
}

// Layout is a saved window-to-space arrangement for a specific display count,
// persisted as its own internal JSON file (see internal/layout/persist.go).
type Layout struct {
	SchemaVersion int `json:"schemaVersion"`
	// DisplayCount is the number of connected displays this layout was
	// captured for, and the file it's persisted as is named for.
	DisplayCount int `json:"displayCount"`
	// SpaceCounts is the per-display space-count sequence, left to right,
	// recorded at save time. Used to detect arrangement drift at restore.
	SpaceCounts []int     `json:"spaceCounts"`
	Entries     []Entry   `json:"entries"`
	SavedAt     time.Time `json:"savedAt"`
}
