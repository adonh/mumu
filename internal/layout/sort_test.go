package layout //nolint:testpackage // tests unexported entryLess

import (
	"testing"

	"github.com/adonh/mumu/internal/config"
	"github.com/adonh/mumu/internal/space"
)

const (
	testBundleSafari = "com.apple.Safari"
	testBundleChrome = "com.google.Chrome"
	testBundleSame   = "com.same"
	testBundleA      = "com.a"
	testBundleB      = "com.b"
	testTitleAlpha   = "Alpha"
	testTitleBeta    = "Beta"
)

// stubMCOrdinal returns a lookup function usable as entryLess's mcOrdinal
// parameter without any native calls, backed by a fixed logical-ordinal ->
// Mission Control-ordinal mapping.
func stubMCOrdinal(mapping map[int]int) func(int) int {
	return func(logicalOrdinal int) int {
		return mapping[logicalOrdinal]
	}
}

func TestEntryLess_SortByLogical(t *testing.T) {
	t.Parallel()

	entryA := Entry{Ordinal: 1, BundleID: testBundleB, Title: "Z"}
	entryB := Entry{Ordinal: 2, BundleID: testBundleA, Title: "A"}

	mcOrdinal := stubMCOrdinal(nil)

	if !entryLess(entryA, entryB, SortByLogical, mcOrdinal) {
		t.Fatalf(
			"entryLess(entryA, entryB, SortByLogical) = false, want true (entryA.Ordinal=1 < entryB.Ordinal=2)",
		)
	}

	if entryLess(entryB, entryA, SortByLogical, mcOrdinal) {
		t.Fatalf("entryLess(entryB, entryA, SortByLogical) = true, want false")
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
	entryA := Entry{Ordinal: 1, BundleID: testBundleA}
	entryB := Entry{Ordinal: 2, BundleID: testBundleB}

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
		entryA := Entry{Ordinal: 4, BundleID: testBundleSame, Title: testTitleAlpha}
		entryB := Entry{Ordinal: 4, BundleID: testBundleSame, Title: testTitleBeta}

		for _, key := range []SortKey{SortByLogical, SortByMacOS, SortByApp} {
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

func TestSortPinRules_AllKeys(t *testing.T) {
	t.Parallel()

	t.Run("logical orders by Space ascending", func(t *testing.T) {
		t.Parallel()

		pins := []config.PinRule{
			{BundleID: testBundleSafari, Title: "B", Ordinal: 3},
			{BundleID: testBundleChrome, Title: "A", Ordinal: 1},
		}

		SortPinRules(pins, SortByLogical)

		if pins[0].Ordinal != 1 || pins[1].Ordinal != 3 {
			t.Fatalf("SortPinRules(SortByLogical) order = %+v, want Space 1 before Space 3", pins)
		}
	})

	t.Run("app orders by BundleID ascending regardless of Space", func(t *testing.T) {
		t.Parallel()

		pins := []config.PinRule{
			{BundleID: testBundleChrome, Title: "B", Ordinal: 1},
			{BundleID: testBundleSafari, Title: "A", Ordinal: 9},
		}

		SortPinRules(pins, SortByApp)

		if pins[0].BundleID != testBundleSafari || pins[1].BundleID != testBundleChrome {
			t.Fatalf(
				"SortPinRules(SortByApp) order = %+v, want %q before %q",
				pins,
				testBundleSafari,
				testBundleChrome,
			)
		}
	})

	t.Run("macos orders by resolved Mission Control ordinal", func(t *testing.T) {
		t.Parallel()

		ruleA := config.PinRule{BundleID: testBundleA, Ordinal: 1}
		ruleB := config.PinRule{BundleID: testBundleB, Ordinal: 2}

		mcOrdinal := stubMCOrdinal(map[int]int{1: 9, 2: 3})

		if pinRuleLess(ruleA, ruleB, SortByMacOS, mcOrdinal) {
			t.Fatalf(
				"pinRuleLess(ruleA, ruleB, SortByMacOS) = true, want false (mcOrdinal(1)=9 > mcOrdinal(2)=3)",
			)
		}

		if !pinRuleLess(ruleB, ruleA, SortByMacOS, mcOrdinal) {
			t.Fatalf(
				"pinRuleLess(ruleB, ruleA, SortByMacOS) = false, want true (mcOrdinal(2)=3 < mcOrdinal(1)=9)",
			)
		}
	})

	t.Run("tie on Space and BundleID falls back to Title", func(t *testing.T) {
		t.Parallel()

		pins := []config.PinRule{
			{BundleID: testBundleSame, Title: testTitleBeta, Ordinal: 4},
			{BundleID: testBundleSame, Title: testTitleAlpha, Ordinal: 4},
		}

		SortPinRules(pins, SortByLogical)

		if pins[0].Title != testTitleAlpha || pins[1].Title != testTitleBeta {
			t.Fatalf("SortPinRules tie-break order = %+v, want Alpha before Beta", pins)
		}
	})
}

func TestSortDefaultSpaceRules_AllKeys(t *testing.T) {
	t.Parallel()

	t.Run("logical orders by Space ascending", func(t *testing.T) {
		t.Parallel()

		rules := []config.DefaultSpaceRule{
			{BundleID: testBundleSafari, Ordinal: 3},
			{BundleID: testBundleChrome, Ordinal: 1},
		}

		SortDefaultSpaceRules(rules, SortByLogical)

		if rules[0].Ordinal != 1 || rules[1].Ordinal != 3 {
			t.Fatalf(
				"SortDefaultSpaceRules(SortByLogical) order = %+v, want Space 1 before Space 3",
				rules,
			)
		}
	})

	t.Run("app orders by BundleID ascending regardless of Space", func(t *testing.T) {
		t.Parallel()

		rules := []config.DefaultSpaceRule{
			{BundleID: testBundleChrome, Ordinal: 1},
			{BundleID: testBundleSafari, Ordinal: 9},
		}

		SortDefaultSpaceRules(rules, SortByApp)

		if rules[0].BundleID != testBundleSafari || rules[1].BundleID != testBundleChrome {
			t.Fatalf(
				"SortDefaultSpaceRules(SortByApp) order = %+v, want %q before %q",
				rules,
				testBundleSafari,
				testBundleChrome,
			)
		}
	})

	t.Run("macos orders by resolved Mission Control ordinal", func(t *testing.T) {
		t.Parallel()

		ruleA := config.DefaultSpaceRule{BundleID: testBundleA, Ordinal: 1}
		ruleB := config.DefaultSpaceRule{BundleID: testBundleB, Ordinal: 2}

		mcOrdinal := stubMCOrdinal(map[int]int{1: 9, 2: 3})

		if defaultSpaceRuleLess(ruleA, ruleB, SortByMacOS, mcOrdinal) {
			t.Fatalf(
				"defaultSpaceRuleLess(ruleA, ruleB, SortByMacOS) = true, want false (mcOrdinal(1)=9 > mcOrdinal(2)=3)",
			)
		}

		if !defaultSpaceRuleLess(ruleB, ruleA, SortByMacOS, mcOrdinal) {
			t.Fatalf(
				"defaultSpaceRuleLess(ruleB, ruleA, SortByMacOS) = false, want true (mcOrdinal(2)=3 < mcOrdinal(1)=9)",
			)
		}
	})

	t.Run("tie on Space falls back to BundleID", func(t *testing.T) {
		t.Parallel()

		rules := []config.DefaultSpaceRule{
			{BundleID: testBundleChrome, Ordinal: 4},
			{BundleID: testBundleSafari, Ordinal: 4},
		}

		SortDefaultSpaceRules(rules, SortByLogical)

		if rules[0].BundleID != testBundleSafari || rules[1].BundleID != testBundleChrome {
			t.Fatalf(
				"SortDefaultSpaceRules tie-break order = %+v, want %q before %q",
				rules,
				testBundleSafari,
				testBundleChrome,
			)
		}
	})
}

func TestParseSortKey(t *testing.T) {
	t.Parallel()

	valid := map[string]SortKey{
		"logical": SortByLogical,
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

func TestParseSortKey_RejectsOldDisplayValue(t *testing.T) {
	t.Parallel()

	_, err := ParseSortKey("display")
	if err == nil {
		t.Fatal(
			"ParseSortKey(\"display\") returned nil error, want a validation error since \"display\" was renamed to \"logical\"",
		)
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
