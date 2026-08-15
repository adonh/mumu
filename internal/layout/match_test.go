package layout //nolint:testpackage // tests unexported matchEntries

import (
	"testing"

	"github.com/adonh/mumu/internal/window"
)

func matchLiveEntry(title string) window.AcrossSpacesEntry {
	return window.AcrossSpacesEntry{Title: title}
}

// noOverlapTitle is a saved title used across tests to exercise the
// zero-word-overlap path without matching any live window by title.
const noOverlapTitle = "zulu"

func TestMatchEntries_HighestScoringPairWins(t *testing.T) {
	t.Parallel()

	entries := []Entry{{Title: "hello world", Index: 0}}
	live := []window.AcrossSpacesEntry{
		matchLiveEntry("hello"),
		matchLiveEntry("hello world"),
		matchLiveEntry("world"),
	}

	got := matchEntries(entries, live)
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("matchEntries() = %v, want [1] (exact match beats partial overlaps)", got)
	}
}

func TestMatchEntries_OneToOneWhenTwoEntriesPreferTheSameWindow(t *testing.T) {
	t.Parallel()

	// Both entries score highest against live[0]; independently matching
	// each entry to its own top pick would double-claim it. Each entry
	// also has its own distinct, lower-scoring fallback: entries[0]'s is
	// live[1] (score 0.5), entries[1]'s is live[2] (score 0.6).
	entries := []Entry{
		{Title: "a b c d", Index: 0},
		{Title: "a b c d e", Index: 1},
	}
	live := []window.AcrossSpacesEntry{
		matchLiveEntry("a b c d"), // exact match for entries[0] (score 1.0)
		matchLiveEntry("a b"),     // entries[0]'s distinct fallback (score 0.5)
		matchLiveEntry("c d e"),   // entries[1]'s distinct fallback (score 0.6)
	}

	got := matchEntries(entries, live)
	if len(got) != 2 {
		t.Fatalf("matchEntries() = %v, want 2 results", got)
	}

	if got[0] != 0 {
		t.Fatalf("entries[0] matched live[%d], want live[0] (its exact match)", got[0])
	}

	if got[1] != 2 {
		t.Fatalf("entries[1] matched live[%d], want live[2] (its own distinct fallback)", got[1])
	}

	if got[0] == got[1] {
		t.Fatalf("matchEntries() = %v, both entries claimed the same window", got)
	}
}

func TestMatchEntries_ZeroOverlapStillMatchesSoleCandidate(t *testing.T) {
	t.Parallel()

	entries := []Entry{{Title: noOverlapTitle, Index: 0}}
	live := []window.AcrossSpacesEntry{matchLiveEntry("foo")}

	got := matchEntries(entries, live)
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("matchEntries() = %v, want [0] (only candidate, despite zero word overlap)", got)
	}
}

func TestMatchEntries_TieBrokenBySavedIndex(t *testing.T) {
	t.Parallel()

	entries := []Entry{{Title: noOverlapTitle, Index: 1}}
	live := []window.AcrossSpacesEntry{matchLiveEntry("foo"), matchLiveEntry("bar")}

	got := matchEntries(entries, live)
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("matchEntries() = %v, want [1] (tie broken by saved Index)", got)
	}
}

func TestMatchEntries_TieWithNoIndexMatchIsDeterministic(t *testing.T) {
	t.Parallel()

	entries := []Entry{{Title: noOverlapTitle, Index: 5}}
	live := []window.AcrossSpacesEntry{matchLiveEntry("foo"), matchLiveEntry("bar")}

	first := matchEntries(entries, live)
	second := matchEntries(entries, live)

	if len(first) != 1 || first[0] != 0 {
		t.Fatalf("matchEntries() = %v, want [0] (fixed tie-break order)", first)
	}

	if second[0] != first[0] {
		t.Fatalf("matchEntries() = %v on repeat call, want %v (repeatable)", second, first)
	}
}

func TestMatchEntries_UnmatchedOnlyWhenWindowsRunOut(t *testing.T) {
	t.Parallel()

	entries := []Entry{{Title: "foo", Index: 0}, {Title: "bar", Index: 1}}
	live := []window.AcrossSpacesEntry{matchLiveEntry("foo")}

	got := matchEntries(entries, live)
	if len(got) != 2 {
		t.Fatalf("matchEntries() = %v, want 2 results", got)
	}

	if got[0] != 0 {
		t.Fatalf("entries[0] matched live[%d], want live[0]", got[0])
	}

	if got[1] != -1 {
		t.Fatalf("entries[1] matched live[%d], want -1 (no window left to claim)", got[1])
	}
}

func TestGroupEntriesByBundle(t *testing.T) {
	t.Parallel()

	entries := []Entry{
		{BundleID: "com.example.a", Title: "one"},
		{BundleID: "com.example.b", Title: "two"},
		{BundleID: "com.example.a", Title: "three"},
	}

	got := groupEntriesByBundle(entries)

	if len(got["com.example.a"]) != 2 {
		t.Fatalf("com.example.a entries = %d, want 2", len(got["com.example.a"]))
	}

	if len(got["com.example.b"]) != 1 {
		t.Fatalf("com.example.b entries = %d, want 1", len(got["com.example.b"]))
	}
}
