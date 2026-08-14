package layout

import (
	"sort"

	derrors "github.com/adonh/mumu/internal/errors"
	"github.com/adonh/mumu/internal/space"
)

// SortKey selects which field "mumu show"/"restore" order entries by.
// Whichever key is primary, entries that tie on it fall back to a fixed
// cascade: logical Space ordinal, then bundle identifier, then window
// title (see entryLess).
type SortKey string

// Supported SortKey values, matching the "--sort" flag's accepted strings.
const (
	// SortByDisplay orders by the logical left-to-right Space ordinal
	// (mumu's display-sequence numbering). This is the default.
	SortByDisplay SortKey = "display"
	// SortByMacOS orders by the current macOS Mission Control Space
	// ordinal — the same ordinal macOS's own "Switch to Desktop <n>"
	// keyboard shortcut uses.
	SortByMacOS SortKey = "macos"
	// SortByApp orders by application bundle identifier.
	SortByApp SortKey = "app"
)

// ParseSortKey validates a raw "--sort" flag value, returning the
// corresponding SortKey or an error naming the accepted values.
func ParseSortKey(raw string) (SortKey, error) {
	switch key := SortKey(raw); key {
	case SortByDisplay, SortByMacOS, SortByApp:
		return key, nil
	default:
		return "", derrors.Newf(
			derrors.CodeInvalidInput,
			"invalid --sort value %q; must be one of: display, macos, app",
			raw,
		)
	}
}

// entryLess reports whether entryA should sort before entryB for the
// given key. mcOrdinal resolves an entry's logical Ordinal to its current
// macOS Mission Control ordinal (only consulted for SortByMacOS), letting
// callers memoize the underlying native lookups across a whole sort.
func entryLess(entryA, entryB Entry, key SortKey, mcOrdinal func(int) int) bool {
	switch key {
	case SortByMacOS:
		if mcA, mcB := mcOrdinal(entryA.Ordinal), mcOrdinal(entryB.Ordinal); mcA != mcB {
			return mcA < mcB
		}
	case SortByApp:
		if entryA.BundleID != entryB.BundleID {
			return entryA.BundleID < entryB.BundleID
		}
	case SortByDisplay:
		// The cascade below already starts with Ordinal, so SortByDisplay
		// needs no separate primary check.
	}

	if entryA.Ordinal != entryB.Ordinal {
		return entryA.Ordinal < entryB.Ordinal
	}

	if entryA.BundleID != entryB.BundleID {
		return entryA.BundleID < entryB.BundleID
	}

	return entryA.Title < entryB.Title
}

// newMissionControlOrdinalLookup returns a function resolving a logical
// Space ordinal to its current macOS Mission Control ordinal, memoizing
// results across calls so a single sort's repeated comparisons don't
// repeat native lookups for entries sharing the same saved Space.
func newMissionControlOrdinalLookup() func(int) int {
	cache := map[int]int{}

	return func(logicalOrdinal int) int {
		if mcIndex, ok := cache[logicalOrdinal]; ok {
			return mcIndex
		}

		mcIndex := 0
		if sid := space.LogicalSpaceID(logicalOrdinal); sid != 0 {
			mcIndex = space.MissionControlIndexForSpace(sid)
		}

		cache[logicalOrdinal] = mcIndex

		return mcIndex
	}
}

// SortEntries sorts entries in place according to key. See SortKey and
// entryLess for the exact ordering and tie-break rules.
func SortEntries(entries []Entry, key SortKey) {
	mcOrdinal := newMissionControlOrdinalLookup()

	sort.SliceStable(entries, func(i, j int) bool {
		return entryLess(entries[i], entries[j], key, mcOrdinal)
	})
}

// SortSkippedEntries sorts skipped entries in place according to key while
// retaining their restore metadata.
func SortSkippedEntries(entries []SkippedEntry, key SortKey) {
	mcOrdinal := newMissionControlOrdinalLookup()

	sort.SliceStable(entries, func(i, j int) bool {
		return entryLess(entries[i].Entry, entries[j].Entry, key, mcOrdinal)
	})
}
