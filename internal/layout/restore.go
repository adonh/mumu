package layout

import (
	"fmt"
	"sort"

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
}

func moveFailureSkip(target moveTarget) SkippedEntry {
	return SkippedEntry{
		Entry:    target.entry,
		Reason:   SkipMoveFailed,
		Fallback: target.fallback,
	}
}

// Restore applies a saved layout: for each saved window entry, it finds a
// matching, currently open window belonging to an already-running
// application and moves it to the space corresponding to the entry's
// logical ordinal. Applications that are not currently running are skipped
// and never launched. If an application now has exactly one window left
// unclaimed, an otherwise-unmatched entry for it is matched to that window
// regardless of title or saved position (see matchWindowIndex), since
// there's no real ambiguity about which window it refers to. After direct
// matches are established, any remaining windows for an application with
// valid assignments use that application's prevalent target space; tied
// targets use the current space on the primary display. Restore never creates
// or removes spaces; entries whose target space no longer exists are skipped
// and reported.
//
// sortKey determines both the order windows are moved in and the order
// their per-window progress lines print in (see SortKey); it has no
// effect on which windows are matched, moved, or skipped.
//
// progress, if non-nil, receives status updates while windows are scanned
// and moved — the latter can take a while since each move is paced to let
// WindowServer catch up; pass nil to discard them.
//
// Callers are responsible for any arrangement-drift confirmation prompt
// (see DetectDrift) before calling Restore.
func Restore(saved *Layout, sortKey SortKey, progress ProgressFunc) (RestoreSummary, error) {
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
	usedIndex := map[string]map[int]bool{}
	validAssignmentOrdinals := map[string][]int{}
	currentSpaceCount := space.LogicalCount()

	var toMove []moveTarget

	for _, entry := range saved.Entries {
		live := liveByBundle[entry.BundleID]
		if len(live) == 0 {
			summary.Skipped = append(
				summary.Skipped,
				SkippedEntry{Entry: entry, Reason: SkipAppNotRunning},
			)

			continue
		}

		if usedIndex[entry.BundleID] == nil {
			usedIndex[entry.BundleID] = map[int]bool{}
		}

		used := usedIndex[entry.BundleID]

		matchIdx := matchWindowIndex(entry, live, used)
		if matchIdx < 0 {
			summary.Skipped = append(
				summary.Skipped,
				SkippedEntry{Entry: entry, Reason: SkipUnmatchedWindow},
			)

			continue
		}

		used[matchIdx] = true

		if entry.Ordinal < 1 || entry.Ordinal > currentSpaceCount {
			summary.Skipped = append(
				summary.Skipped,
				SkippedEntry{Entry: entry, Reason: SkipOrdinalOutOfRange},
			)

			continue
		}

		sid := space.LogicalSpaceID(entry.Ordinal)
		if sid == 0 {
			summary.Skipped = append(
				summary.Skipped,
				SkippedEntry{Entry: entry, Reason: SkipOrdinalOutOfRange},
			)

			continue
		}

		toMove = append(
			toMove,
			moveTarget{windowID: live[matchIdx].WindowID, sid: sid, entry: entry},
		)
		validAssignmentOrdinals[entry.BundleID] = append(
			validAssignmentOrdinals[entry.BundleID],
			entry.Ordinal,
		)
	}

	fallbackMoves, fallbackSkipped := planFallbackMoves(
		liveByBundle,
		usedIndex,
		validAssignmentOrdinals,
		space.MenuBarActiveLogicalIndex,
		space.LogicalSpaceID,
	)
	toMove = append(toMove, fallbackMoves...)
	summary.Skipped = append(summary.Skipped, fallbackSkipped...)

	mcOrdinal := newMissionControlOrdinalLookup()
	sort.SliceStable(toMove, func(i, j int) bool {
		return entryLess(toMove[i].entry, toMove[j].entry, sortKey, mcOrdinal)
	})

	if len(toMove) > 0 {
		progress.emit(fmt.Sprintf("Moving %d window(s)...", len(toMove)))
	}

	for moveIdx, target := range toMove {
		fallbackMarker := ""
		if target.fallback {
			fallbackMarker = " (fallback)"
		}

		progress.emit(fmt.Sprintf(
			"  %s %s — %s — %q%s",
			FormatIndex(moveIdx+1, len(toMove)),
			space.DualLabel(target.entry.Ordinal),
			target.entry.BundleID,
			displayTitle(target.entry.Title),
			fallbackMarker,
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

// matchWindowIndex resolves a saved entry to the index of a currently open
// window within the same application's live window list. It tries, in
// order: (1) an exact, unambiguous title match; (2) the entry's saved
// positional index, if still available; (3) whether exactly one of the
// app's windows remains unclaimed, in which case there is no real
// ambiguity about which window this entry refers to regardless of title or
// saved position. The third tier matters for apps like browsers, whose
// window title reflects page content and rarely matches across save and
// restore, and for apps that had multiple windows saved but now have only
// one open. Returns -1 if no candidate is available.
func matchWindowIndex(entry Entry, live []window.AcrossSpacesEntry, used map[int]bool) int {
	if entry.Title != "" {
		matchIdx := -1
		ambiguous := false

		for candidateIdx, w := range live {
			if used[candidateIdx] || w.Title != entry.Title {
				continue
			}

			if matchIdx >= 0 {
				ambiguous = true

				break
			}

			matchIdx = candidateIdx
		}

		if matchIdx >= 0 && !ambiguous {
			return matchIdx
		}
	}

	if entry.Index >= 0 && entry.Index < len(live) && !used[entry.Index] {
		return entry.Index
	}

	return soleRemainingCandidate(live, used)
}

// soleRemainingCandidate returns the index of the only not-yet-claimed
// window in live, or -1 if zero or more than one remain unclaimed.
func soleRemainingCandidate(live []window.AcrossSpacesEntry, used map[int]bool) int {
	sole := -1

	for candidateIdx := range live {
		if used[candidateIdx] {
			continue
		}

		if sole >= 0 {
			return -1
		}

		sole = candidateIdx
	}

	return sole
}

func planFallbackMoves(
	liveByBundle map[string][]window.AcrossSpacesEntry,
	usedByBundle map[string]map[int]bool,
	assignmentOrdinals map[string][]int,
	primaryDisplayOrdinal func() (int, error),
	logicalSpaceID func(int) uint64,
) ([]moveTarget, []SkippedEntry) {
	bundleIDs := make([]string, 0, len(assignmentOrdinals))
	for bundleID := range assignmentOrdinals {
		bundleIDs = append(bundleIDs, bundleID)
	}

	sort.Strings(bundleIDs)

	var (
		targets []moveTarget
		skipped []SkippedEntry
	)

	for _, bundleID := range bundleIDs {
		ordinal, hasTarget, err := fallbackTargetOrdinal(
			assignmentOrdinals[bundleID],
			primaryDisplayOrdinal,
		)
		if !hasTarget {
			continue
		}

		live := liveByBundle[bundleID]

		used := usedByBundle[bundleID]
		if used == nil {
			used = map[int]bool{}
			usedByBundle[bundleID] = used
		}

		sid := uint64(0)
		if err == nil {
			sid = logicalSpaceID(ordinal)
			if sid == 0 {
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
				Ordinal:  ordinal,
			}
			if err != nil {
				skipped = append(
					skipped,
					SkippedEntry{
						Entry:    entry,
						Reason:   SkipFallbackTargetUnavailable,
						Fallback: true,
					},
				)

				continue
			}

			used[liveIdx] = true

			targets = append(
				targets,
				moveTarget{
					windowID: liveEntry.WindowID,
					sid:      sid,
					entry:    entry,
					fallback: true,
				},
			)
		}
	}

	return targets, skipped
}

func fallbackTargetOrdinal(
	assignmentOrdinals []int,
	primaryDisplayOrdinal func() (int, error),
) (int, bool, error) {
	if len(assignmentOrdinals) == 0 {
		return 0, false, nil
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
		return ordinal, true, nil
	}

	primaryOrdinal, err := primaryDisplayOrdinal()
	if err != nil {
		return 0, true, err
	}

	if primaryOrdinal < 1 {
		return 0, true, derrors.New(
			derrors.CodeActionFailed,
			"failed to resolve the primary display's current logical space",
		)
	}

	return primaryOrdinal, true, nil
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
