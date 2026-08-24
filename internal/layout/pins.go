package layout

import (
	"github.com/adonh/mumu/internal/config"
	"github.com/adonh/mumu/internal/space"
	"github.com/adonh/mumu/internal/window"
)

// pinEntriesByBundle converts a display count's configured pin rules into
// synthetic Entry values grouped by bundle ID, so they can feed the same
// matching/move pipeline (matchEntries, planDirectMoves) saved-layout
// entries already use. Index is always -1 so the saved-positional-index
// tie-break in candidateLess can never fire for a pin — no live window
// ever has index -1 — leaving pins to fall through to the same
// deterministic fixed order saved entries use when their index doesn't
// help.
func pinEntriesByBundle(pins []config.PinRule) map[string][]Entry {
	byBundle := map[string][]Entry{}

	for _, pin := range pins {
		byBundle[pin.BundleID] = append(byBundle[pin.BundleID], Entry{
			BundleID: pin.BundleID,
			Title:    pin.Title,
			Index:    -1,
			Ordinal:  pin.Ordinal,
		})
	}

	return byBundle
}

// defaultSpacesByBundle converts a display count's configured
// default-space rules into a bundle ID -> target logical ordinal lookup,
// for planFallbackMoves to consult ahead of its prevalent-Space
// heuristic. If the same bundle ID appears more than once (a config
// mistake — config.Load does not reject it), the last rule for that
// bundle ID wins, matching how a plain YAML map would behave for a
// duplicate key.
func defaultSpacesByBundle(rules []config.DefaultSpaceRule) map[string]space.Ordinal {
	byBundle := make(map[string]space.Ordinal, len(rules))

	for _, rule := range rules {
		byBundle[rule.BundleID] = rule.Ordinal
	}

	return byBundle
}

// claimedWindowIDs collects the live window IDs a set of move targets has
// already claimed, for filtering a lower-precedence phase's candidate pool.
func claimedWindowIDs(targets []moveTarget) map[uint32]bool {
	claimed := make(map[uint32]bool, len(targets))

	for _, target := range targets {
		claimed[target.windowID] = true
	}

	return claimed
}

// filterLiveByBundle returns a copy of liveByBundle with every window whose
// ID appears in claimed removed, so a lower-precedence matching phase never
// reconsiders a window a higher-precedence phase already claimed. Bundles
// left with no remaining windows are omitted entirely.
func filterLiveByBundle(
	liveByBundle map[string][]window.AcrossSpacesEntry,
	claimed map[uint32]bool,
) map[string][]window.AcrossSpacesEntry {
	if len(claimed) == 0 {
		return liveByBundle
	}

	filtered := make(map[string][]window.AcrossSpacesEntry, len(liveByBundle))

	for bundleID, entries := range liveByBundle {
		remaining := make([]window.AcrossSpacesEntry, 0, len(entries))

		for _, entry := range entries {
			if !claimed[entry.WindowID] {
				remaining = append(remaining, entry)
			}
		}

		if len(remaining) > 0 {
			filtered[bundleID] = remaining
		}
	}

	return filtered
}
