package layout //nolint:testpackage // tests unexported entryLess

import (
	"testing"

	"github.com/adonh/mumu/internal/space"
)

const (
	testBundleSafari = "com.apple.Safari"
	testBundleChrome = "com.google.Chrome"
	testBundleSame   = "com.same"
)

// stubMCOrdinal returns a lookup function usable as entryLess's mcOrdinal
// parameter without any native calls, backed by a fixed logical-ordinal ->
// Mission Control-ordinal mapping.
func stubMCOrdinal(mapping map[int]int) func(int) int {
	return func(logicalOrdinal int) int {
		return mapping[logicalOrdinal]
	}
}

func TestEntryLess_SortByDisplay(t *testing.T) {
	t.Parallel()

	entryA := Entry{Ordinal: 1, BundleID: "com.b", Title: "Z"}
	entryB := Entry{Ordinal: 2, BundleID: "com.a", Title: "A"}

	mcOrdinal := stubMCOrdinal(nil)

	if !entryLess(entryA, entryB, SortByDisplay, mcOrdinal) {
		t.Fatalf(
			"entryLess(entryA, entryB, SortByDisplay) = false, want true (entryA.Ordinal=1 < entryB.Ordinal=2)",
		)
	}

	if entryLess(entryB, entryA, SortByDisplay, mcOrdinal) {
		t.Fatalf("entryLess(entryB, entryA, SortByDisplay) = true, want false")
	}
}

func TestEntryLess_SortByApp(t *testing.T) {
	t.Parallel()

	entryA := Entry{Ordinal: 5, BundleID: testBundleSafari, Title: "Z"}
	entryB := Entry{Ordinal: 1, BundleID: testBundleChrome, Title: "A"}

	mcOrdinal := stubMCOrdinal(nil)

	if !entryLess(entryA, entryB, SortByApp, mcOrdinal) {
		t.Fatalf(
			"entryLess(entryA, entryB, SortByApp) = false, want true " +
				"(bundle ID entryA < entryB despite entryA.Ordinal > entryB.Ordinal)",
		)
	}
}

func TestEntryLess_SortByMacOS(t *testing.T) {
	t.Parallel()

	// Logical ordinal 1 maps to Mission Control ordinal 9 (e.g. primary
	// display isn't leftmost), so SortByMacOS should order entryB before
	// entryA even though entryA's logical ordinal is smaller.
	entryA := Entry{Ordinal: 1, BundleID: "com.a"}
	entryB := Entry{Ordinal: 2, BundleID: "com.b"}

	mcOrdinal := stubMCOrdinal(map[int]int{1: 9, 2: 3})

	if entryLess(entryA, entryB, SortByMacOS, mcOrdinal) {
		t.Fatalf(
			"entryLess(entryA, entryB, SortByMacOS) = true, want false (mcOrdinal(entryA)=9 > mcOrdinal(entryB)=3)",
		)
	}

	if !entryLess(entryB, entryA, SortByMacOS, mcOrdinal) {
		t.Fatalf("entryLess(entryB, entryA, SortByMacOS) = false, want true")
	}
}

func TestEntryLess_TieBreakCascade(t *testing.T) {
	t.Parallel()

	mcOrdinal := stubMCOrdinal(map[int]int{4: 4})

	t.Run("same primary key falls back to Ordinal then BundleID then Title", func(t *testing.T) {
		t.Parallel()

		// Same Ordinal and BundleID: Title breaks the tie.
		entryA := Entry{Ordinal: 4, BundleID: testBundleSame, Title: "Alpha"}
		entryB := Entry{Ordinal: 4, BundleID: testBundleSame, Title: "Beta"}

		for _, key := range []SortKey{SortByDisplay, SortByMacOS, SortByApp} {
			if !entryLess(entryA, entryB, key, mcOrdinal) {
				t.Fatalf("entryLess(entryA, entryB, %s) = false, want true (Title tiebreak)", key)
			}
		}
	})

	t.Run("app sort ties on BundleID fall back to Ordinal then Title", func(t *testing.T) {
		t.Parallel()

		// Equal BundleID for SortByApp: Ordinal breaks the tie next.
		entryA := Entry{Ordinal: 1, BundleID: testBundleSame, Title: "Z"}
		entryB := Entry{Ordinal: 2, BundleID: testBundleSame, Title: "A"}

		if !entryLess(entryA, entryB, SortByApp, mcOrdinal) {
			t.Fatalf(
				"entryLess(entryA, entryB, SortByApp) = false, want true (Ordinal tiebreak: 1 < 2)",
			)
		}
	})
}

func TestSortEntries_SortByApp(t *testing.T) {
	t.Parallel()

	entries := []Entry{
		{Ordinal: 1, BundleID: testBundleChrome, Title: "New Tab"},
		{Ordinal: 3, BundleID: testBundleSafari, Title: "Page 2"},
		{Ordinal: 2, BundleID: testBundleSafari, Title: "Page 1"},
	}

	SortEntries(entries, SortByApp)

	want := []string{testBundleSafari, testBundleSafari, testBundleChrome}
	for i, w := range want {
		if entries[i].BundleID != w {
			t.Fatalf("SortEntries(SortByApp)[%d].BundleID = %q, want %q", i, entries[i].BundleID, w)
		}
	}

	// Within the same bundle ID, Ordinal breaks the tie (2 before 3).
	if entries[0].Ordinal != 2 || entries[1].Ordinal != 3 {
		t.Fatalf(
			"SortEntries(SortByApp) Safari order = [%d, %d], want [2, 3]",
			entries[0].Ordinal,
			entries[1].Ordinal,
		)
	}
}

func TestParseSortKey(t *testing.T) {
	t.Parallel()

	valid := map[string]SortKey{
		"display": SortByDisplay,
		"macos":   SortByMacOS,
		"app":     SortByApp,
	}

	for raw, want := range valid {
		got, err := ParseSortKey(raw)
		if err != nil {
			t.Fatalf("ParseSortKey(%q) returned error: %v", raw, err)
		}

		if got != want {
			t.Fatalf("ParseSortKey(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestParseSortKey_Invalid(t *testing.T) {
	t.Parallel()

	_, err := ParseSortKey("bogus")
	if err == nil {
		t.Fatal("ParseSortKey(\"bogus\") returned nil error, want a validation error")
	}
}

// TestSortEntries_SortByMacOS_LiveSelfConsistent validates SortByMacOS
// against whatever display/Space arrangement is live on the machine
// running the test, mirroring internal/space's own self-consistency tests
// (e.g. TestMissionControlIndexForSpace_SelfConsistent). It builds one
// synthetic Entry per currently valid logical ordinal, sorts them by
// SortByMacOS, and checks the result is non-decreasing by each entry's
// live-resolved Mission Control ordinal — i.e. that SortEntries actually
// used the real native lookup rather than, say, silently falling back to
// logical order. Skips (rather than fails) when no displays are
// reported, since that's expected in headless CI environments.
func TestSortEntries_SortByMacOS_LiveSelfConsistent(t *testing.T) {
	t.Parallel()

	total := space.LogicalCount()
	if total == 0 {
		t.Skip("no displays reported; skipping on headless environment")
	}

	entries := make([]Entry, total)
	for i := range total {
		logicalOrdinal := i + 1
		entries[i] = Entry{
			Ordinal:  logicalOrdinal,
			BundleID: "com.example.synthetic",
			Title:    "synthetic",
		}
	}

	SortEntries(entries, SortByMacOS)

	prevMCIndex := -1

	for _, entry := range entries {
		sid := space.LogicalSpaceID(entry.Ordinal)
		if sid == 0 {
			t.Fatalf(
				"LogicalSpaceID(%d) = 0, want a non-zero Space ID (total=%d)",
				entry.Ordinal,
				total,
			)
		}

		mcIndex := space.MissionControlIndexForSpace(sid)
		if mcIndex == 0 {
			t.Fatalf(
				"MissionControlIndexForSpace(%d) = 0 for logical ordinal %d, want non-zero",
				sid,
				entry.Ordinal,
			)
		}

		if mcIndex < prevMCIndex {
			t.Fatalf(
				"SortEntries(SortByMacOS) produced non-monotonic Mission Control ordinals: %d after %d",
				mcIndex,
				prevMCIndex,
			)
		}

		prevMCIndex = mcIndex
	}
}
