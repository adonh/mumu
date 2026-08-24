package layout

import (
	"fmt"
	"sort"

	"github.com/adonh/mumu/internal/config"
	derrors "github.com/adonh/mumu/internal/errors"
	"github.com/adonh/mumu/internal/space"
	"github.com/adonh/mumu/internal/window"
)

// SkipReason describes why a saved window entry could not be restored.
type SkipReason string

// Skip reasons reported in a RestoreSummary.
const (
	SkipAppNotRunning             SkipReason = "application is not running"
	SkipUnmatchedWindow           SkipReason = "no matching open window found"
	SkipOrdinalOutOfRange         SkipReason = "saved space no longer exists"
	SkipFallbackTargetUnavailable SkipReason = "fallback space is unavailable"
	SkipMoveFailed                SkipReason = "failed to move window"
)

// SkippedEntry records a saved entry or transient fallback target that could
// not be restored.
type SkippedEntry struct {
	Entry    Entry
	Reason   SkipReason
	Fallback bool
	// DefaultConfigured is set when this fallback placement (or its
	// failure) was chosen because of a configured default_spaces rule,
	// rather than the prevalent-Space heuristic. Always accompanied by
	// Fallback = true; never set together with Fuzzy.
	DefaultConfigured bool
	// Fuzzy is set when the entry had already been matched to a window by
	// approximate title similarity (see matchEntries) before the skip
	// occurred, e.g. the move itself then failed. Never set together with
	// Fallback — see moveTarget.
	Fuzzy bool
}

// RestoreSummary reports the outcome of a Restore call.
type RestoreSummary struct {
	Moved   int
	Skipped []SkippedEntry
}

// ArrangementDrift describes the saved vs. current per-display space-count
// sequence, for detecting whether the display arrangement has changed since
// a layout was saved.
type ArrangementDrift struct {
	Saved   []int
	Current []int
}

// Mismatched reports whether the current arrangement differs from the one
// recorded when the layout was saved.
func (d ArrangementDrift) Mismatched() bool {
	return !intSlicesEqual(d.Saved, d.Current)
}

// DetectDrift compares a saved layout's recorded per-display space-count
// sequence against the current one.
func DetectDrift(saved *Layout) ArrangementDrift {
	return ArrangementDrift{
		Saved:   saved.SpaceCounts,
		Current: space.LeftToRightSpaceCounts(),
	}
}

type moveTarget struct {
	windowID uint32
	sid      uint64
	entry    Entry // for progress reporting and skip-detail on move failure
	fallback bool
	// fuzzy is set when this target's window was chosen by approximate
	// (non-exact) title similarity rather than a perfect word-set match.
	// fallback targets come only from planFallbackMoves, which never sets
	// fuzzy, and fuzzy targets come only from planDirectMoves, which never
	// sets fallback, so a single target is never both.
	fuzzy bool
	// defaultConfigured marks a fallback target chosen because of a
	// configured default_spaces rule rather than the prevalent-Space
	// heuristic. Always accompanied by fallback = true.
	defaultConfigured bool
}

type fallbackTarget struct {
	ordinal int
	sid     uint64
}

func moveFailureSkip(target moveTarget) SkippedEntry {
	return SkippedEntry{
		Entry:             target.entry,
		Reason:            SkipMoveFailed,
		Fallback:          target.fallback,
		DefaultConfigured: target.defaultConfigured,
		Fuzzy:             target.fuzzy,
	}
}

// Restore applies a saved layout: for each application with saved window
// entries, it batch-matches those entries against that application's
// currently open windows by title similarity (see matchEntries) and moves
// each matched window to the space corresponding to its entry's logical
// ordinal. Applications that are not currently running are skipped and
// never launched. Matching scores every remaining (entry, window) pair by
// shared title words and commits pairs greedily from the highest score
// down, so no open window is ever claimed by more than one saved entry; an
// entry goes unmatched only once its application has no open window left
// to claim. Matches whose title didn't exactly match are reported with a
// "(fuzzy)" marker. After matching, any remaining windows for an
// application with valid assignments use that application's prevalent
// target space; tied targets use the current space on the primary display.
// Restore never creates or removes spaces; entries whose target space no
// longer exists are skipped and reported.
//
// sortKey determines both the order windows are moved in and the order
// their per-window progress lines print in (see SortKey); it has no
// effect on which windows are matched, moved, or skipped.
//
// pins and precedence apply the window-pinning capability: pins are
// matched against currently open windows the same way saved entries are,
// and precedence controls whether pin matching or saved-layout matching
// (including its own application-level fallback) runs first and claims
// windows before the other. Pins never receive an application-level
// fallback of their own — an unmatched pin is simply skipped. Pass a nil
// or empty pins slice when no pins are configured for the current display
// count.
//
// defaultSpaces configures, per application bundle identifier, a fixed
// application-level fallback target for that application's leftover
// unclaimed windows: when present for a bundle ID, it always wins over
// the prevalent-Space heuristic for that bundle's leftover windows this
// restore, and it applies even when the bundle has zero valid
// saved-entry assignments (a case that otherwise receives no fallback at
// all). Pass a nil or empty slice when no default spaces are configured
// for the current display count.
//
// progress, if non-nil, receives status updates while windows are scanned
// and moved — the latter can take a while since each move is paced to let
// WindowServer catch up; pass nil to discard them.
//
// Callers are responsible for any arrangement-drift confirmation prompt
// (see DetectDrift) before calling Restore.
func Restore(
	saved *Layout,
	pins []config.PinRule,
	precedence config.PinPrecedence,
	defaultSpaces []config.DefaultSpaceRule,
	sortKey SortKey,
	progress ProgressFunc,
) (RestoreSummary, error) {
	summary := RestoreSummary{}

	err := ensureLayoutPermissions()
	if err != nil {
		return summary, err
	}

	progress.emit("Scanning currently open windows...")

	liveEntries, err := window.AllAcrossSpaces()
	if err != nil {
		return summary, derrors.Wrapf(
			err,
			derrors.CodeActionFailed,
			"failed to enumerate live windows",
		)
	}

	liveByBundle := groupLiveByBundle(liveEntries)
	entriesByBundle := groupEntriesByBundle(saved.Entries)
	pinsByBundle := pinEntriesByBundle(pins)
	defaultSpacesMap := defaultSpacesByBundle(defaultSpaces)

	var (
		directMoves, fallbackMoves, pinMoves       []moveTarget
		directSkipped, fallbackSkipped, pinSkipped []SkippedEntry
	)

	if precedence == config.PinPrecedenceLayout {
		directMoves, directSkipped, fallbackMoves, fallbackSkipped = planLayoutPhase(
			entriesByBundle,
			liveByBundle,
			defaultSpacesMap,
		)

		claimed := claimedWindowIDs(
			append(append([]moveTarget{}, directMoves...), fallbackMoves...),
		)
		pinLiveByBundle := filterLiveByBundle(liveByBundle, claimed)

		pinMoves, pinSkipped, _ = planDirectMoves(
			pinsByBundle,
			pinLiveByBundle,
			map[string]map[int]bool{},
			space.LogicalCount(),
			space.LogicalSpaceID,
		)
	} else {
		pinMoves, pinSkipped, _ = planDirectMoves(
			pinsByBundle,
			liveByBundle,
			map[string]map[int]bool{},
			space.LogicalCount(),
			space.LogicalSpaceID,
		)

		claimed := claimedWindowIDs(pinMoves)
		layoutLiveByBundle := filterLiveByBundle(liveByBundle, claimed)

		directMoves, directSkipped, fallbackMoves, fallbackSkipped = planLayoutPhase(
			entriesByBundle,
			layoutLiveByBundle,
			defaultSpacesMap,
		)
	}

	summary.Skipped = append(summary.Skipped, directSkipped...)
	summary.Skipped = append(summary.Skipped, fallbackSkipped...)
	summary.Skipped = append(summary.Skipped, pinSkipped...)

	toMove := directMoves
	toMove = append(toMove, fallbackMoves...)
	toMove = append(toMove, pinMoves...)

	mcOrdinal := newMissionControlOrdinalLookup()
	sort.SliceStable(toMove, func(i, j int) bool {
		return entryLess(toMove[i].entry, toMove[j].entry, sortKey, mcOrdinal)
	})

	if len(toMove) > 0 {
		progress.emit(fmt.Sprintf("Moving %d window(s)...", len(toMove)))
	}

	for moveIdx, target := range toMove {
		marker := moveMarker(target)

		progress.emit(fmt.Sprintf(
			"  %s %s — %s — %q%s",
			FormatIndex(moveIdx+1, len(toMove)),
			space.DualLabel(target.entry.Ordinal),
			target.entry.BundleID,
			displayTitle(target.entry.Title),
			marker,
		))

		err := window.MoveWindowIDToSpace(target.windowID, target.sid)
		if err != nil {
			summary.Skipped = append(
				summary.Skipped,
				moveFailureSkip(target),
			)
		} else {
			summary.Moved++
		}
	}

	return summary, nil
}

// moveMarker builds the trailing annotation for a restore progress line,
// distinguishing a configured-default placement ("(default)") from a
// prevalent-Space placement ("(fallback)") and, independently, an
// approximate title match ("(fuzzy)").
func moveMarker(target moveTarget) string {
	marker := ""

	switch {
	case target.defaultConfigured:
		marker += " (default)"
	case target.fallback:
		marker += " (fallback)"
	}

	if target.fuzzy {
		marker += " (fuzzy)"
	}

	return marker
}

func groupLiveByBundle(entries []window.AcrossSpacesEntry) map[string][]window.AcrossSpacesEntry {
	byBundle := map[string][]window.AcrossSpacesEntry{}

	for _, e := range entries {
		if e.Fullscreen {
			continue
		}

		byBundle[e.BundleID] = append(byBundle[e.BundleID], e)
	}

	return byBundle
}

// planLayoutPhase runs saved-layout direct matching and its
// application-level fallback as a self-contained unit against the given
// live-window pool, with its own fresh usedIndex map so the two calls
// stay internally consistent regardless of whether liveByBundle is the
// full live-window pool or one already filtered to exclude windows a
// higher-precedence pin phase claimed.
func planLayoutPhase(
	entriesByBundle map[string][]Entry,
	liveByBundle map[string][]window.AcrossSpacesEntry,
	defaultSpaces map[string]int,
) ([]moveTarget, []SkippedEntry, []moveTarget, []SkippedEntry) {
	usedIndex := map[string]map[int]bool{}

	direct, directSkipped, validAssignmentOrdinals := planDirectMoves(
		entriesByBundle,
		liveByBundle,
		usedIndex,
		space.LogicalCount(),
		space.LogicalSpaceID,
	)

	fallback, fallbackSkipped := planFallbackMoves(
		liveByBundle,
		usedIndex,
		validAssignmentOrdinals,
		defaultSpaces,
		primaryDisplayFallbackTarget,
		space.LogicalSpaceID,
	)

	return direct, directSkipped, fallback, fallbackSkipped
}

// planDirectMoves batch-matches each application's saved entries against
// its currently open windows (see matchEntries) and resolves every match
// to either a moveTarget or a skip reason. usedIndex is mutated in place
// to record every live window index claimed by a match, keyed by bundle
// ID, so planFallbackMoves can later see exactly which windows remain
// unclaimed. The returned map records, per bundle ID, the logical
// ordinals of every valid assignment, for planFallbackMoves to pick a
// prevalent fallback target from.
func planDirectMoves(
	entriesByBundle map[string][]Entry,
	liveByBundle map[string][]window.AcrossSpacesEntry,
	usedIndex map[string]map[int]bool,
	currentSpaceCount int,
	logicalSpaceID func(int) uint64,
) ([]moveTarget, []SkippedEntry, map[string][]int) {
	bundleIDs := make([]string, 0, len(entriesByBundle))
	for bundleID := range entriesByBundle {
		bundleIDs = append(bundleIDs, bundleID)
	}

	sort.Strings(bundleIDs)

	var (
		toMove  []moveTarget
		skipped []SkippedEntry
	)

	validAssignmentOrdinals := map[string][]int{}

	for _, bundleID := range bundleIDs {
		entries := entriesByBundle[bundleID]
		live := liveByBundle[bundleID]

		if len(live) == 0 {
			for _, entry := range entries {
				skipped = append(
					skipped,
					SkippedEntry{Entry: entry, Reason: SkipAppNotRunning},
				)
			}

			continue
		}

		used := usedIndex[bundleID]
		if used == nil {
			used = map[int]bool{}
			usedIndex[bundleID] = used
		}

		matches := matchEntries(entries, live)

		for i, entry := range entries {
			matchIdx := matches[i]
			if matchIdx < 0 {
				skipped = append(
					skipped,
					SkippedEntry{Entry: entry, Reason: SkipUnmatchedWindow},
				)

				continue
			}

			used[matchIdx] = true

			if entry.Ordinal < 1 || entry.Ordinal > currentSpaceCount {
				skipped = append(
					skipped,
					SkippedEntry{Entry: entry, Reason: SkipOrdinalOutOfRange},
				)

				continue
			}

			sid := logicalSpaceID(entry.Ordinal)
			if sid == 0 {
				skipped = append(
					skipped,
					SkippedEntry{Entry: entry, Reason: SkipOrdinalOutOfRange},
				)

				continue
			}

			toMove = append(
				toMove,
				moveTarget{
					windowID: live[matchIdx].WindowID,
					sid:      sid,
					entry:    entry,
					fuzzy:    titleSimilarity(entry.Title, live[matchIdx].Title) < 1,
				},
			)
			validAssignmentOrdinals[bundleID] = append(
				validAssignmentOrdinals[bundleID],
				entry.Ordinal,
			)
		}
	}

	return toMove, skipped, validAssignmentOrdinals
}

func planFallbackMoves(
	liveByBundle map[string][]window.AcrossSpacesEntry,
	usedByBundle map[string]map[int]bool,
	assignmentOrdinals map[string][]int,
	defaultSpaces map[string]int,
	primaryDisplayTarget func() (fallbackTarget, error),
	logicalSpaceID func(int) uint64,
) ([]moveTarget, []SkippedEntry) {
	bundleIDs := fallbackBundleIDs(assignmentOrdinals, defaultSpaces)

	var (
		targets []moveTarget
		skipped []SkippedEntry
	)

	for _, bundleID := range bundleIDs {
		var (
			target     fallbackTarget
			hasTarget  bool
			err        error
			configured bool
		)

		if ordinal, ok := defaultSpaces[bundleID]; ok {
			target = fallbackTarget{ordinal: ordinal}
			hasTarget = true
			configured = true
		} else {
			target, hasTarget, err = fallbackTargetForAssignments(
				assignmentOrdinals[bundleID],
				primaryDisplayTarget,
			)
		}

		if !hasTarget {
			continue
		}

		live := liveByBundle[bundleID]

		used := usedByBundle[bundleID]
		if used == nil {
			used = map[int]bool{}
			usedByBundle[bundleID] = used
		}

		if err == nil && target.sid == 0 {
			target.sid = logicalSpaceID(target.ordinal)
			if target.sid == 0 {
				err = derrors.New(
					derrors.CodeActionFailed,
					"failed to resolve fallback space",
				)
			}
		}

		for liveIdx, liveEntry := range live {
			if used[liveIdx] {
				continue
			}

			entry := Entry{
				BundleID: bundleID,
				Title:    liveEntry.Title,
				Index:    -1,
				Ordinal:  target.ordinal,
			}
			if err != nil {
				skipped = append(
					skipped,
					SkippedEntry{
						Entry:             entry,
						Reason:            SkipFallbackTargetUnavailable,
						Fallback:          true,
						DefaultConfigured: configured,
					},
				)

				continue
			}

			used[liveIdx] = true

			targets = append(
				targets,
				moveTarget{
					windowID:          liveEntry.WindowID,
					sid:               target.sid,
					entry:             entry,
					fallback:          true,
					defaultConfigured: configured,
				},
			)
		}
	}

	return targets, skipped
}

// fallbackBundleIDs returns the sorted union of bundle IDs with either a
// valid saved-entry assignment this restore or a configured default_spaces
// rule, so planFallbackMoves considers a bundle ID with a configured
// default even when it has zero valid assignments.
func fallbackBundleIDs(assignmentOrdinals map[string][]int, defaultSpaces map[string]int) []string {
	seen := make(map[string]bool, len(assignmentOrdinals)+len(defaultSpaces))
	bundleIDs := make([]string, 0, len(assignmentOrdinals)+len(defaultSpaces))

	for bundleID := range assignmentOrdinals {
		if !seen[bundleID] {
			seen[bundleID] = true

			bundleIDs = append(bundleIDs, bundleID)
		}
	}

	for bundleID := range defaultSpaces {
		if !seen[bundleID] {
			seen[bundleID] = true

			bundleIDs = append(bundleIDs, bundleID)
		}
	}

	sort.Strings(bundleIDs)

	return bundleIDs
}

func primaryDisplayFallbackTarget() (fallbackTarget, error) {
	spaceID, ordinal, err := space.PrimaryDisplayCurrentSpace()
	if err != nil {
		return fallbackTarget{}, err
	}

	return fallbackTarget{ordinal: ordinal, sid: spaceID}, nil
}

func fallbackTargetForAssignments(
	assignmentOrdinals []int,
	primaryDisplayTarget func() (fallbackTarget, error),
) (fallbackTarget, bool, error) {
	if len(assignmentOrdinals) == 0 {
		return fallbackTarget{}, false, nil
	}

	counts := map[int]int{}
	maxCount := 0
	ordinal := 0
	tied := false

	for _, candidate := range assignmentOrdinals {
		counts[candidate]++

		count := counts[candidate]
		if count > maxCount {
			ordinal = candidate
			maxCount = count
			tied = false
		} else if count == maxCount && candidate != ordinal {
			tied = true
		}
	}

	if !tied {
		return fallbackTarget{ordinal: ordinal}, true, nil
	}

	target, err := primaryDisplayTarget()
	if err != nil {
		return fallbackTarget{}, true, err
	}

	if target.ordinal < 1 || target.sid == 0 {
		return fallbackTarget{}, true, derrors.New(
			derrors.CodeActionFailed,
			"failed to resolve the primary display's current space",
		)
	}

	return target, true, nil
}

func intSlicesEqual(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}

	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}

	return true
}
