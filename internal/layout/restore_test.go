package layout //nolint:testpackage // tests unexported matchWindowIndex / intSlicesEqual

import (
	"testing"

	"github.com/adonh/mumu/internal/window"
)

func liveEntry(title string) window.AcrossSpacesEntry {
	return window.AcrossSpacesEntry{Title: title}
}

// noSuchTitle is a saved title used across tests to exercise the
// no-title-match path without matching any live window.
const noSuchTitle = "Missing"

func TestMatchWindowIndex_ExactTitleMatch(t *testing.T) {
	t.Parallel()

	live := []window.AcrossSpacesEntry{liveEntry("Alpha"), liveEntry("Beta"), liveEntry("Gamma")}
	entry := Entry{Title: "Beta", Index: 0}

	got := matchWindowIndex(entry, live, map[int]bool{})
	if got != 1 {
		t.Fatalf("matchWindowIndex() = %d, want 1 (exact title match)", got)
	}
}

func TestMatchWindowIndex_AmbiguousTitleFallsBackToIndex(t *testing.T) {
	t.Parallel()

	live := []window.AcrossSpacesEntry{
		liveEntry("Untitled"),
		liveEntry("Untitled"),
		liveEntry("Untitled"),
	}
	entry := Entry{Title: "Untitled", Index: 2}

	got := matchWindowIndex(entry, live, map[int]bool{})
	if got != 2 {
		t.Fatalf("matchWindowIndex() = %d, want 2 (index fallback on ambiguous title)", got)
	}
}

func TestMatchWindowIndex_NoTitleMatchFallsBackToIndex(t *testing.T) {
	t.Parallel()

	live := []window.AcrossSpacesEntry{liveEntry("One"), liveEntry("Two")}
	entry := Entry{Title: noSuchTitle, Index: 1}

	got := matchWindowIndex(entry, live, map[int]bool{})
	if got != 1 {
		t.Fatalf("matchWindowIndex() = %d, want 1 (index fallback, no title match)", got)
	}
}

func TestMatchWindowIndex_EmptyTitleUsesIndexDirectly(t *testing.T) {
	t.Parallel()

	live := []window.AcrossSpacesEntry{liveEntry("One"), liveEntry("")}
	entry := Entry{Title: "", Index: 1}

	got := matchWindowIndex(entry, live, map[int]bool{})
	if got != 1 {
		t.Fatalf("matchWindowIndex() = %d, want 1", got)
	}
}

func TestMatchWindowIndex_UsedWindowSkippedInTitleMatch(t *testing.T) {
	t.Parallel()

	live := []window.AcrossSpacesEntry{liveEntry("Alpha"), liveEntry("Beta")}
	entry := Entry{Title: "Alpha", Index: 1}

	// live[0] has the matching title but is already used, so title matching
	// must skip it, leaving no unambiguous match; this falls back to the
	// saved index (1), which is available.
	got := matchWindowIndex(entry, live, map[int]bool{0: true})
	if got != 1 {
		t.Fatalf(
			"matchWindowIndex() = %d, want 1 (fallback to index since the matching title is used)",
			got,
		)
	}
}

func TestMatchWindowIndex_NoCandidateReturnsNegativeOne(t *testing.T) {
	t.Parallel()

	// Two unclaimed candidates, neither matching by title or index: genuine
	// ambiguity, so no fallback tier should guess.
	live := []window.AcrossSpacesEntry{liveEntry("One"), liveEntry("Two")}
	entry := Entry{Title: noSuchTitle, Index: 5}

	got := matchWindowIndex(entry, live, map[int]bool{})
	if got != -1 {
		t.Fatalf("matchWindowIndex() = %d, want -1", got)
	}
}

func TestMatchWindowIndex_SoleRemainingCandidateMatchesRegardlessOfTitleOrIndex(t *testing.T) {
	t.Parallel()

	// Only one window is currently open for this app; even though its
	// title doesn't match (common for browsers, whose title reflects page
	// content) and the saved index (5) is out of range — e.g. because
	// several other windows from the same saved layout have since been
	// closed — there's no real ambiguity about which window this is.
	live := []window.AcrossSpacesEntry{liveEntry("Completely different title")}
	entry := Entry{Title: "Old title", Index: 5}

	got := matchWindowIndex(entry, live, map[int]bool{})
	if got != 0 {
		t.Fatalf("matchWindowIndex() = %d, want 0 (sole remaining candidate)", got)
	}
}

func TestMatchWindowIndex_SoleRemainingCandidateSkippedIfAlreadyUsed(t *testing.T) {
	t.Parallel()

	live := []window.AcrossSpacesEntry{liveEntry("One")}
	entry := Entry{Title: noSuchTitle, Index: 5}

	// The only window is already claimed by another entry in this restore
	// pass, so nothing remains for this one.
	got := matchWindowIndex(entry, live, map[int]bool{0: true})
	if got != -1 {
		t.Fatalf("matchWindowIndex() = %d, want -1 (sole candidate already used)", got)
	}
}

func TestMatchWindowIndex_IndexAlreadyUsed(t *testing.T) {
	t.Parallel()

	// Two windows remain unclaimed besides the one at the entry's saved
	// index, so there's still genuine ambiguity and no fallback tier
	// should guess.
	live := []window.AcrossSpacesEntry{liveEntry("One"), liveEntry("Two"), liveEntry("Three")}
	entry := Entry{Title: "", Index: 0}

	got := matchWindowIndex(entry, live, map[int]bool{0: true})
	if got != -1 {
		t.Fatalf("matchWindowIndex() = %d, want -1 (index already used, still ambiguous)", got)
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
