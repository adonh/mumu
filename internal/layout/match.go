package layout

import (
	"sort"

	"github.com/adonh/mumu/internal/window"
)

// groupEntriesByBundle groups saved entries by their owning application's
// bundle identifier, mirroring groupLiveByBundle's grouping of live
// windows, so both sides of a restore can be matched application by
// application.
func groupEntriesByBundle(entries []Entry) map[string][]Entry {
	byBundle := map[string][]Entry{}

	for _, entry := range entries {
		byBundle[entry.BundleID] = append(byBundle[entry.BundleID], entry)
	}

	return byBundle
}

// matchEntries computes a one-to-one assignment from one application's
// saved entries to its currently open windows: it scores every (entry,
// window) pair by titleSimilarity, then commits pairs greedily from the
// highest score down, skipping any entry or window already committed,
// until no further pair can be assigned. This guarantees no live window is
// ever committed to more than one entry, regardless of how many entries
// would independently have preferred it.
//
// When multiple pairs tie for the top remaining score, the pair whose
// window's live position equals the entry's saved Index is preferred; if
// that doesn't uniquely resolve the tie, ties fall back to a fixed,
// repeatable order (ascending live window index, then ascending entry
// index), so the same input always produces the same assignment.
//
// The result is parallel to entries: result[i] is the matched index into
// live, or -1 if entries[i] had no live window left to claim (more saved
// entries than open windows).
func matchEntries(entries []Entry, live []window.AcrossSpacesEntry) []int {
	result := make([]int, len(entries))
	for i := range result {
		result[i] = -1
	}

	candidates := scoreCandidates(entries, live)

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidateLess(candidates[i], candidates[j], entries)
	})

	entryClaimed := make([]bool, len(entries))
	liveClaimed := make([]bool, len(live))
	unmatched := len(entries)

	for _, candidate := range candidates {
		if unmatched == 0 {
			break
		}

		if entryClaimed[candidate.entryIdx] || liveClaimed[candidate.liveIdx] {
			continue
		}

		entryClaimed[candidate.entryIdx] = true
		liveClaimed[candidate.liveIdx] = true
		result[candidate.entryIdx] = candidate.liveIdx
		unmatched--
	}

	return result
}

// matchCandidate is one (saved entry, live window) pair under
// consideration by matchEntries, along with its title similarity score.
type matchCandidate struct {
	entryIdx int
	liveIdx  int
	score    float64
}

func scoreCandidates(entries []Entry, live []window.AcrossSpacesEntry) []matchCandidate {
	candidates := make([]matchCandidate, 0, len(entries)*len(live))

	for entryIdx, entry := range entries {
		for liveIdx, w := range live {
			candidates = append(candidates, matchCandidate{
				entryIdx: entryIdx,
				liveIdx:  liveIdx,
				score:    titleSimilarity(entry.Title, w.Title),
			})
		}
	}

	return candidates
}

// candidateLess orders candidates for matchEntries's greedy assignment:
// highest score first, then a candidate whose live position matches the
// entry's saved Index, then a fixed deterministic order.
func candidateLess(left, right matchCandidate, entries []Entry) bool {
	if left.score != right.score {
		return left.score > right.score
	}

	leftAtSavedIndex := left.liveIdx == entries[left.entryIdx].Index
	rightAtSavedIndex := right.liveIdx == entries[right.entryIdx].Index

	if leftAtSavedIndex != rightAtSavedIndex {
		return leftAtSavedIndex
	}

	if left.liveIdx != right.liveIdx {
		return left.liveIdx < right.liveIdx
	}

	return left.entryIdx < right.entryIdx
}
