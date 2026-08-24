package layout //nolint:testpackage // tests unexported planDirectMoves / intSlicesEqual

import (
	"testing"

	"github.com/adonh/mumu/internal/space"
	"github.com/adonh/mumu/internal/window"
)

// testSpaceCounts is a single display with 10 Spaces — big enough to
// exercise in-range ordinals while still letting a deliberately large
// Space number (e.g. 99) trigger the out-of-range path.
var testSpaceCounts = []int{10}

const testTitleReport = "Report"

func liveEntry(title string) window.AcrossSpacesEntry {
	return window.AcrossSpacesEntry{Title: title}
}

func TestPlanDirectMoves_MatchesByTitleSimilarity(t *testing.T) {
	t.Parallel()

	entriesByBundle := map[string][]Entry{
		fallbackTestBundle: {
			{BundleID: fallbackTestBundle, Title: "Beta", Index: 0, Ordinal: ord(3)},
		},
	}
	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {liveEntry("Alpha"), liveEntry("Beta"), liveEntry("Gamma")},
	}

	toMove, skipped, validOrdinals := planDirectMoves(
		entriesByBundle,
		liveByBundle,
		map[string]map[int]bool{},
		testSpaceCounts,
		identityIDForOrdinal,
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

	if got := validOrdinals[fallbackTestBundle]; len(got) != 1 || got[0] != ord(3) {
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
			{BundleID: fallbackTestBundle, Title: "a b c d", Index: 0, Ordinal: ord(1)},
			{BundleID: fallbackTestBundle, Title: "a b c d e", Index: 1, Ordinal: ord(2)},
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
		testSpaceCounts,
		identityIDForOrdinal,
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

	byOrdinal := map[space.Ordinal]moveTarget{}
	for _, target := range toMove {
		byOrdinal[target.entry.Ordinal] = target
	}

	if got := byOrdinal[ord(1)]; got.windowID != 1 || got.fuzzy {
		t.Fatalf("ordinal 1 target = %#v, want window 1, not fuzzy (exact match)", got)
	}

	if got := byOrdinal[ord(2)]; got.windowID != 3 || !got.fuzzy {
		t.Fatalf("ordinal 2 target = %#v, want window 3, fuzzy (approximate match)", got)
	}
}

func TestPlanDirectMoves_UnmatchedWhenAppNotRunning(t *testing.T) {
	t.Parallel()

	entriesByBundle := map[string][]Entry{
		fallbackTestBundle: {
			{BundleID: fallbackTestBundle, Title: testTitleReport, Ordinal: ord(1)},
		},
	}

	toMove, skipped, validOrdinals := planDirectMoves(
		entriesByBundle,
		map[string][]window.AcrossSpacesEntry{},
		map[string]map[int]bool{},
		testSpaceCounts,
		func(space.Ordinal) uint64 { return 0 },
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
			{BundleID: fallbackTestBundle, Title: "foo", Index: 0, Ordinal: ord(1)},
			{BundleID: fallbackTestBundle, Title: "bar", Index: 1, Ordinal: ord(2)},
		},
	}
	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {liveEntry("foo")},
	}

	toMove, skipped, _ := planDirectMoves(
		entriesByBundle,
		liveByBundle,
		map[string]map[int]bool{},
		testSpaceCounts,
		identityIDForOrdinal,
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
		fallbackTestBundle: {
			{BundleID: fallbackTestBundle, Title: testTitleReport, Ordinal: ord(99)},
		},
	}
	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {liveEntry(testTitleReport)},
	}

	toMove, skipped, validOrdinals := planDirectMoves(
		entriesByBundle,
		liveByBundle,
		map[string]map[int]bool{},
		testSpaceCounts,
		identityIDForOrdinal,
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

func TestPlanDirectMoves_SkipsOutOfRangeDisplay(t *testing.T) {
	t.Parallel()

	// Display 2 doesn't exist in a single-display test arrangement, even
	// though Space 1 would be in range on display 1 — proving the
	// display part is bounds-checked independently of the Space part.
	entriesByBundle := map[string][]Entry{
		fallbackTestBundle: {
			{
				BundleID: fallbackTestBundle,
				Title:    testTitleReport,
				Ordinal:  space.Ordinal{Display: 2, Space: 1},
			},
		},
	}
	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {liveEntry(testTitleReport)},
	}

	toMove, skipped, _ := planDirectMoves(
		entriesByBundle,
		liveByBundle,
		map[string]map[int]bool{},
		testSpaceCounts,
		identityIDForOrdinal,
	)

	if len(toMove) != 0 {
		t.Fatalf("toMove = %#v, want none", toMove)
	}

	if len(skipped) != 1 || skipped[0].Reason != SkipOrdinalOutOfRange {
		t.Fatalf("skipped = %#v, want one SkipOrdinalOutOfRange", skipped)
	}
}

func TestOrdinalInBounds_IndependentPerDisplay(t *testing.T) {
	t.Parallel()

	// Two displays: 2 Spaces on display 1, 5 Spaces on display 2.
	counts := []int{2, 5}

	cases := []struct {
		name    string
		ordinal space.Ordinal
		want    bool
	}{
		{name: "in range on display 1", ordinal: space.Ordinal{Display: 1, Space: 2}, want: true},
		{
			name:    "out of range on display 1 even though in range on display 2",
			ordinal: space.Ordinal{Display: 1, Space: 5},
			want:    false,
		},
		{name: "in range on display 2", ordinal: space.Ordinal{Display: 2, Space: 5}, want: true},
		{
			name:    "out of range on display 2 even though in range on display 1",
			ordinal: space.Ordinal{Display: 2, Space: 2},
			want:    true,
		},
		{name: "display out of range", ordinal: space.Ordinal{Display: 3, Space: 1}, want: false},
		{name: "zero-value ordinal is out of range", ordinal: space.Ordinal{}, want: false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ordinalInBounds(tt.ordinal, counts); got != tt.want {
				t.Fatalf("ordinalInBounds(%v, %v) = %v, want %v", tt.ordinal, counts, got, tt.want)
			}
		})
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
			{BundleID: fallbackTestBundle, Title: "totally different", Index: 0, Ordinal: ord(4)},
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
		testSpaceCounts,
		identityIDForOrdinal,
	)

	fallbackMoves, _ := planFallbackMoves(
		liveByBundle,
		usedIndex,
		validOrdinals,
		map[string]space.Ordinal{},
		func() (fallbackTarget, error) { return fallbackTarget{}, nil },
		identityIDForOrdinal,
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

func TestMoveMarker_DistinguishesDefaultConfiguredFromPrevalentFallback(t *testing.T) {
	t.Parallel()

	defaultTarget := moveTarget{fallback: true, defaultConfigured: true}
	if got := moveMarker(defaultTarget); got != " (default)" {
		t.Fatalf("moveMarker(configured default) = %q, want %q", got, " (default)")
	}

	prevalentTarget := moveTarget{fallback: true}
	if got := moveMarker(prevalentTarget); got != " (fallback)" {
		t.Fatalf("moveMarker(prevalent fallback) = %q, want %q", got, " (fallback)")
	}

	if got := moveMarker(defaultTarget); got == moveMarker(prevalentTarget) {
		t.Fatalf(
			"moveMarker() text does not differ between configured-default and prevalent fallback: %q",
			got,
		)
	}

	fuzzyDirect := moveTarget{fuzzy: true}
	if got := moveMarker(fuzzyDirect); got != " (fuzzy)" {
		t.Fatalf("moveMarker(fuzzy direct match) = %q, want %q", got, " (fuzzy)")
	}

	plain := moveTarget{}
	if got := moveMarker(plain); got != "" {
		t.Fatalf("moveMarker(plain direct match) = %q, want empty", got)
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
