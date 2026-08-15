package layout //nolint:testpackage // tests unexported planDirectMoves / intSlicesEqual

import (
	"testing"

	"github.com/adonh/mumu/internal/window"
)

func liveEntry(title string) window.AcrossSpacesEntry {
	return window.AcrossSpacesEntry{Title: title}
}

func TestPlanDirectMoves_MatchesByTitleSimilarity(t *testing.T) {
	t.Parallel()

	entriesByBundle := map[string][]Entry{
		fallbackTestBundle: {
			{BundleID: fallbackTestBundle, Title: "Beta", Index: 0, Ordinal: 3},
		},
	}
	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {liveEntry("Alpha"), liveEntry("Beta"), liveEntry("Gamma")},
	}

	toMove, skipped, validOrdinals := planDirectMoves(
		entriesByBundle,
		liveByBundle,
		map[string]map[int]bool{},
		10,
		func(ordinal int) uint64 { return uint64(ordinal) },
	)

	if len(skipped) != 0 {
		t.Fatalf("skipped = %#v, want none", skipped)
	}

	if len(toMove) != 1 || toMove[0].windowID != liveByBundle[fallbackTestBundle][1].WindowID {
		t.Fatalf("toMove = %#v, want the exact-title match (live[1])", toMove)
	}

	if toMove[0].fuzzy {
		t.Fatal("exact title match marked fuzzy, want false")
	}

	if got := validOrdinals[fallbackTestBundle]; len(got) != 1 || got[0] != 3 {
		t.Fatalf("validAssignmentOrdinals = %#v, want [3]", got)
	}
}

func TestPlanDirectMoves_OneToOneWhenTwoEntriesPreferTheSameWindow(t *testing.T) {
	t.Parallel()

	// live[0] ("a b c d") is the naive top pick for both entries; a
	// sequential per-entry matcher processing entries in order would
	// double-claim it for the first entry it reaches. The batch matcher
	// must instead give it to whichever entry it's the best match for,
	// and resolve the other to its own distinct, still-valid alternative
	// (live[2]) rather than leaving it unmatched.
	entriesByBundle := map[string][]Entry{
		fallbackTestBundle: {
			{BundleID: fallbackTestBundle, Title: "a b c d", Index: 0, Ordinal: 1},
			{BundleID: fallbackTestBundle, Title: "a b c d e", Index: 1, Ordinal: 2},
		},
	}
	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {
			fallbackLiveEntry(1, "a b c d"),
			fallbackLiveEntry(2, "a b"),
			fallbackLiveEntry(3, "c d e"),
		},
	}

	toMove, skipped, _ := planDirectMoves(
		entriesByBundle,
		liveByBundle,
		map[string]map[int]bool{},
		10,
		func(ordinal int) uint64 { return uint64(ordinal) },
	)

	if len(skipped) != 0 {
		t.Fatalf("skipped = %#v, want none", skipped)
	}

	if len(toMove) != 2 {
		t.Fatalf("toMove = %#v, want 2 moves", toMove)
	}

	claimed := map[uint32]bool{}

	for _, target := range toMove {
		if claimed[target.windowID] {
			t.Fatalf("window %d claimed by more than one move: %#v", target.windowID, toMove)
		}

		claimed[target.windowID] = true
	}

	byOrdinal := map[int]moveTarget{}
	for _, target := range toMove {
		byOrdinal[target.entry.Ordinal] = target
	}

	if got := byOrdinal[1]; got.windowID != 1 || got.fuzzy {
		t.Fatalf("ordinal 1 target = %#v, want window 1, not fuzzy (exact match)", got)
	}

	if got := byOrdinal[2]; got.windowID != 3 || !got.fuzzy {
		t.Fatalf("ordinal 2 target = %#v, want window 3, fuzzy (approximate match)", got)
	}
}

func TestPlanDirectMoves_UnmatchedWhenAppNotRunning(t *testing.T) {
	t.Parallel()

	entriesByBundle := map[string][]Entry{
		fallbackTestBundle: {{BundleID: fallbackTestBundle, Title: "Report", Ordinal: 1}},
	}

	toMove, skipped, validOrdinals := planDirectMoves(
		entriesByBundle,
		map[string][]window.AcrossSpacesEntry{},
		map[string]map[int]bool{},
		10,
		func(int) uint64 { return 0 },
	)

	if len(toMove) != 0 {
		t.Fatalf("toMove = %#v, want none", toMove)
	}

	if len(skipped) != 1 || skipped[0].Reason != SkipAppNotRunning {
		t.Fatalf("skipped = %#v, want one SkipAppNotRunning", skipped)
	}

	if len(validOrdinals) != 0 {
		t.Fatalf("validAssignmentOrdinals = %#v, want none", validOrdinals)
	}
}

func TestPlanDirectMoves_UnmatchedWhenMoreEntriesThanWindows(t *testing.T) {
	t.Parallel()

	entriesByBundle := map[string][]Entry{
		fallbackTestBundle: {
			{BundleID: fallbackTestBundle, Title: "foo", Index: 0, Ordinal: 1},
			{BundleID: fallbackTestBundle, Title: "bar", Index: 1, Ordinal: 2},
		},
	}
	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {liveEntry("foo")},
	}

	toMove, skipped, _ := planDirectMoves(
		entriesByBundle,
		liveByBundle,
		map[string]map[int]bool{},
		10,
		func(ordinal int) uint64 { return uint64(ordinal) },
	)

	if len(toMove) != 1 {
		t.Fatalf("toMove = %#v, want 1 match", toMove)
	}

	if len(skipped) != 1 || skipped[0].Reason != SkipUnmatchedWindow {
		t.Fatalf("skipped = %#v, want one SkipUnmatchedWindow", skipped)
	}
}

func TestPlanDirectMoves_SkipsOutOfRangeOrdinal(t *testing.T) {
	t.Parallel()

	entriesByBundle := map[string][]Entry{
		fallbackTestBundle: {{BundleID: fallbackTestBundle, Title: "Report", Ordinal: 99}},
	}
	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {liveEntry("Report")},
	}

	toMove, skipped, validOrdinals := planDirectMoves(
		entriesByBundle,
		liveByBundle,
		map[string]map[int]bool{},
		10,
		func(ordinal int) uint64 { return uint64(ordinal) },
	)

	if len(toMove) != 0 {
		t.Fatalf("toMove = %#v, want none", toMove)
	}

	if len(skipped) != 1 || skipped[0].Reason != SkipOrdinalOutOfRange {
		t.Fatalf("skipped = %#v, want one SkipOrdinalOutOfRange", skipped)
	}

	if len(validOrdinals) != 0 {
		t.Fatalf("validAssignmentOrdinals = %#v, want none", validOrdinals)
	}
}

// TestFallbackAndFuzzyMarkersAreDisjoint guards the invariant documented on
// moveTarget: fallback placements (from planFallbackMoves) and fuzzy
// matches (from planDirectMoves) come from disjoint code paths and must
// never both be set on the same target.
func TestFallbackAndFuzzyMarkersAreDisjoint(t *testing.T) {
	t.Parallel()

	entriesByBundle := map[string][]Entry{
		fallbackTestBundle: {
			{BundleID: fallbackTestBundle, Title: "totally different", Index: 0, Ordinal: 4},
		},
	}
	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {
			fallbackLiveEntry(1, "totally different"),
			fallbackLiveEntry(2, "unmatched"),
		},
	}
	usedIndex := map[string]map[int]bool{}

	directMoves, _, validOrdinals := planDirectMoves(
		entriesByBundle,
		liveByBundle,
		usedIndex,
		10,
		func(ordinal int) uint64 { return uint64(ordinal) },
	)

	fallbackMoves, _ := planFallbackMoves(
		liveByBundle,
		usedIndex,
		validOrdinals,
		func() (fallbackTarget, error) { return fallbackTarget{}, nil },
		func(ordinal int) uint64 { return uint64(ordinal) },
	)

	for _, target := range append(directMoves, fallbackMoves...) {
		if target.fallback && target.fuzzy {
			t.Fatalf("target = %#v, want fallback and fuzzy never both true", target)
		}
	}

	if len(directMoves) != 1 || directMoves[0].fallback {
		t.Fatalf("directMoves = %#v, want one non-fallback match", directMoves)
	}

	if len(fallbackMoves) != 1 || fallbackMoves[0].fuzzy {
		t.Fatalf("fallbackMoves = %#v, want one non-fuzzy fallback", fallbackMoves)
	}
}

func TestArrangementDrift_Mismatched(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		saved   []int
		current []int
		want    bool
	}{
		{name: "identical", saved: []int{4, 13, 7, 1}, current: []int{4, 13, 7, 1}, want: false},
		{name: "different counts", saved: []int{4, 13, 7, 1}, current: []int{4, 13, 7}, want: true},
		{
			name:    "different values",
			saved:   []int{4, 13, 7, 1},
			current: []int{4, 13, 8, 1},
			want:    true,
		},
		{name: "reordered", saved: []int{4, 13}, current: []int{13, 4}, want: true},
		{name: "both empty", saved: []int{}, current: []int{}, want: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			drift := ArrangementDrift{Saved: testCase.saved, Current: testCase.current}
			if got := drift.Mismatched(); got != testCase.want {
				t.Fatalf("Mismatched() = %v, want %v", got, testCase.want)
			}
		})
	}
}
