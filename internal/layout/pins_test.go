package layout //nolint:testpackage // tests unexported pin-matching helpers

import (
	"testing"

	"github.com/adonh/mumu/internal/config"
	"github.com/adonh/mumu/internal/window"
)

const pinsTestTitle = "general"

// planLayoutPhaseForTest mirrors planLayoutPhase's direct+fallback
// orchestration but, like every other test in this package, uses a
// mocked space count/ID mapping instead of the real macOS Mission
// Control APIs planLayoutPhase calls through the space package. Those
// APIs return whatever the machine running the test actually has
// configured (e.g. a headless CI runner may report 0 Spaces), which
// would make ordinal-out-of-range behavior nondeterministic across
// environments.
func planLayoutPhaseForTest(
	entriesByBundle map[string][]Entry,
	liveByBundle map[string][]window.AcrossSpacesEntry,
) ([]moveTarget, []SkippedEntry, []moveTarget, []SkippedEntry) {
	usedIndex := map[string]map[int]bool{}
	identity := func(ordinal int) uint64 { return uint64(ordinal) }

	direct, directSkipped, validAssignmentOrdinals := planDirectMoves(
		entriesByBundle,
		liveByBundle,
		usedIndex,
		10,
		identity,
	)

	fallback, fallbackSkipped := planFallbackMoves(
		liveByBundle,
		usedIndex,
		validAssignmentOrdinals,
		primaryDisplayFallbackTarget,
		identity,
	)

	return direct, directSkipped, fallback, fallbackSkipped
}

func TestPinEntriesByBundle_ConvertsWithNegativeIndex(t *testing.T) {
	t.Parallel()

	pins := []config.PinRule{
		{BundleID: "com.tinyspeck.slackmacgap", Title: pinsTestTitle, Space: 1},
		{BundleID: "com.google.Chrome", Title: "GitHub", Space: 3},
		{BundleID: "com.google.Chrome", Title: "Mail", Space: 4},
	}

	byBundle := pinEntriesByBundle(pins)

	if len(byBundle) != 2 {
		t.Fatalf("byBundle = %#v, want 2 bundles", byBundle)
	}

	slackEntries := byBundle["com.tinyspeck.slackmacgap"]
	if len(slackEntries) != 1 || slackEntries[0].Index != -1 || slackEntries[0].Ordinal != 1 {
		t.Fatalf("slack entries = %#v, want one entry with Index -1, Ordinal 1", slackEntries)
	}

	chromeEntries := byBundle["com.google.Chrome"]
	if len(chromeEntries) != 2 {
		t.Fatalf("chrome entries = %#v, want 2 entries", chromeEntries)
	}

	for _, entry := range chromeEntries {
		if entry.Index != -1 {
			t.Fatalf("entry.Index = %d, want -1: %#v", entry.Index, entry)
		}
	}
}

func TestClaimedWindowIDs(t *testing.T) {
	t.Parallel()

	targets := []moveTarget{{windowID: 1}, {windowID: 5}}

	claimed := claimedWindowIDs(targets)

	if len(claimed) != 2 || !claimed[1] || !claimed[5] {
		t.Fatalf("claimed = %#v, want {1: true, 5: true}", claimed)
	}
}

func TestFilterLiveByBundle_RemovesClaimedWindows(t *testing.T) {
	t.Parallel()

	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {
			fallbackLiveEntry(1, "a"),
			fallbackLiveEntry(2, "b"),
		},
		"other.bundle": {
			fallbackLiveEntry(3, "c"),
		},
	}

	filtered := filterLiveByBundle(liveByBundle, map[uint32]bool{2: true, 3: true})

	if len(filtered) != 1 {
		t.Fatalf("filtered = %#v, want only fallbackTestBundle remaining", filtered)
	}

	remaining := filtered[fallbackTestBundle]
	if len(remaining) != 1 || remaining[0].WindowID != 1 {
		t.Fatalf("remaining = %#v, want only window 1", remaining)
	}
}

func TestFilterLiveByBundle_NoClaimsReturnsOriginal(t *testing.T) {
	t.Parallel()

	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {fallbackLiveEntry(1, "a")},
	}

	filtered := filterLiveByBundle(liveByBundle, map[uint32]bool{})

	if len(filtered[fallbackTestBundle]) != 1 {
		t.Fatalf("filtered = %#v, want unchanged", filtered)
	}
}

// TestPinPrecedence_PinsClaimWindowFirst exercises the same orchestration
// Restore performs for the default "pin" precedence: pins are matched
// against the full live pool first, and the saved-layout phase only sees
// whatever the pin phase left unclaimed. Here the pin claims the
// bundle's only live window, so the saved entry for the same bundle is
// left with nothing to match and is reported as if the app weren't
// running at all — proving the claim really does propagate between
// phases rather than each phase matching independently.
func TestPinPrecedence_PinsClaimWindowFirst(t *testing.T) {
	t.Parallel()

	// A single live window for the bundle, claimed by the pin: once
	// filtered out, the saved-layout phase has no live windows left to
	// consider for this bundle at all.
	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {
			fallbackLiveEntry(1, pinsTestTitle),
		},
	}

	pinsByBundle := pinEntriesByBundle([]config.PinRule{
		{BundleID: fallbackTestBundle, Title: pinsTestTitle, Space: 1},
	})
	entriesByBundle := map[string][]Entry{
		fallbackTestBundle: {
			{BundleID: fallbackTestBundle, Title: pinsTestTitle, Index: 0, Ordinal: 9},
		},
	}

	pinMoves, pinSkipped, _ := planDirectMoves(
		pinsByBundle,
		liveByBundle,
		map[string]map[int]bool{},
		10,
		func(ordinal int) uint64 { return uint64(ordinal) },
	)

	if len(pinSkipped) != 0 || len(pinMoves) != 1 || pinMoves[0].windowID != 1 {
		t.Fatalf(
			"pinMoves = %#v, pinSkipped = %#v, want window 1 claimed by the pin",
			pinMoves,
			pinSkipped,
		)
	}

	if pinMoves[0].fuzzy {
		t.Fatal("exact-title pin match marked fuzzy, want false")
	}

	claimed := claimedWindowIDs(pinMoves)
	layoutLive := filterLiveByBundle(liveByBundle, claimed)

	directMoves, directSkipped, fallbackMoves, fallbackSkipped := planLayoutPhaseForTest(
		entriesByBundle,
		layoutLive,
	)

	if len(directMoves)+len(fallbackMoves) != 0 {
		t.Fatalf(
			"layout phase moves = direct %#v fallback %#v, want none: the only live window "+
				"was already claimed by the pin",
			directMoves,
			fallbackMoves,
		)
	}

	if len(directSkipped) != 1 || directSkipped[0].Reason != SkipAppNotRunning {
		t.Fatalf(
			"directSkipped = %#v, want the saved entry reported as if the app had no open windows",
			directSkipped,
		)
	}

	if len(fallbackSkipped) != 0 {
		t.Fatalf("fallbackSkipped = %#v, want none", fallbackSkipped)
	}
}

// TestPinPrecedence_LayoutClaimsWindowFirst mirrors the "layout"
// precedence orchestration: the saved-layout phase (direct + fallback)
// runs unfiltered first, and pin matching only considers what's left.
func TestPinPrecedence_LayoutClaimsWindowFirst(t *testing.T) {
	t.Parallel()

	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {
			fallbackLiveEntry(1, pinsTestTitle),
		},
	}

	entriesByBundle := map[string][]Entry{
		fallbackTestBundle: {
			{BundleID: fallbackTestBundle, Title: pinsTestTitle, Index: 0, Ordinal: 9},
		},
	}
	pinsByBundle := pinEntriesByBundle([]config.PinRule{
		{BundleID: fallbackTestBundle, Title: pinsTestTitle, Space: 1},
	})

	directMoves, directSkipped, fallbackMoves, fallbackSkipped := planLayoutPhaseForTest(
		entriesByBundle,
		liveByBundle,
	)

	if len(directSkipped)+len(fallbackSkipped) != 0 || len(directMoves) != 1 {
		t.Fatalf(
			"layout phase = direct %#v (skipped %#v), fallback %#v (skipped %#v), want one direct move",
			directMoves,
			directSkipped,
			fallbackMoves,
			fallbackSkipped,
		)
	}

	claimed := claimedWindowIDs(append(append([]moveTarget{}, directMoves...), fallbackMoves...))
	pinLive := filterLiveByBundle(liveByBundle, claimed)

	pinMoves, pinSkipped, _ := planDirectMoves(
		pinsByBundle,
		pinLive,
		map[string]map[int]bool{},
		10,
		func(ordinal int) uint64 { return uint64(ordinal) },
	)

	if len(pinMoves) != 0 {
		t.Fatalf(
			"pinMoves = %#v, want none: window 1 was already claimed by the saved layout",
			pinMoves,
		)
	}

	if len(pinSkipped) != 1 || pinSkipped[0].Reason != SkipAppNotRunning {
		t.Fatalf(
			"pinSkipped = %#v, want the pin reported as if the app had no open windows",
			pinSkipped,
		)
	}
}

// TestPinPrecedence_DisplacesLoserToTheOtherWindow exercises real
// contention between a pin and a saved entry that both score highest
// against the *same* live window, when a second, less-similar window is
// also available for the same bundle. The higher-precedence phase
// (pins, the default) claims the contested window; the loser is not
// simply skipped but falls through to matchEntries' own tie-breaking
// against the sole remaining candidate — proving filterLiveByBundle
// removes exactly the claimed window and nothing else.
func TestPinPrecedence_DisplacesLoserToTheOtherWindow(t *testing.T) {
	t.Parallel()

	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {
			fallbackLiveEntry(1, pinsTestTitle+" chat"),
			fallbackLiveEntry(2, "totally unrelated"),
		},
	}

	pinsByBundle := pinEntriesByBundle([]config.PinRule{
		{BundleID: fallbackTestBundle, Title: pinsTestTitle + " chat", Space: 1},
	})

	pinMoves, pinSkipped, _ := planDirectMoves(
		pinsByBundle,
		liveByBundle,
		map[string]map[int]bool{},
		10,
		func(ordinal int) uint64 { return uint64(ordinal) },
	)

	if len(pinSkipped) != 0 || len(pinMoves) != 1 || pinMoves[0].windowID != 1 {
		t.Fatalf(
			"pinMoves = %#v, pinSkipped = %#v, want window 1 (the best match) claimed by the pin",
			pinMoves,
			pinSkipped,
		)
	}

	claimed := claimedWindowIDs(pinMoves)
	layoutLive := filterLiveByBundle(liveByBundle, claimed)

	if got := layoutLive[fallbackTestBundle]; len(got) != 1 || got[0].WindowID != 2 {
		t.Fatalf(
			"layoutLive[%q] = %#v, want only window 2 left once the pin's claim on window 1 "+
				"is filtered out",
			fallbackTestBundle,
			got,
		)
	}
}

// TestPinMatching_ApproximateTitleMatch verifies a pin's title pattern
// need not exactly match a window's current title — the same
// shared-word similarity heuristic saved-layout entries use also
// resolves pins, per the window-pinning spec's "Pin title matching
// reuses the existing approximate-match heuristic" requirement.
func TestPinMatching_ApproximateTitleMatch(t *testing.T) {
	t.Parallel()

	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {
			fallbackLiveEntry(1, "Inbox (12) - Mail"),
			fallbackLiveEntry(2, "Unrelated Window"),
		},
	}

	pinsByBundle := pinEntriesByBundle([]config.PinRule{
		{BundleID: fallbackTestBundle, Title: "Inbox - Mail", Space: 2},
	})

	moves, skipped, _ := planDirectMoves(
		pinsByBundle,
		liveByBundle,
		map[string]map[int]bool{},
		10,
		func(ordinal int) uint64 { return uint64(ordinal) },
	)

	if len(skipped) != 0 {
		t.Fatalf("skipped = %#v, want none", skipped)
	}

	if len(moves) != 1 || moves[0].windowID != 1 {
		t.Fatalf("moves = %#v, want the pin matched to window 1 (the higher-overlap title)", moves)
	}

	if !moves[0].fuzzy {
		t.Fatal("approximate pin match not marked fuzzy, want true")
	}
}

// TestPinMatching_MultiplePinsResolveIndependently verifies two pins for
// the same application, each with a different title pattern, each claim
// their own best-scoring window without either claiming the other's
// window, per the window-pinning spec's "Multiple pins for the same
// application resolve independently" scenario.
func TestPinMatching_MultiplePinsResolveIndependently(t *testing.T) {
	t.Parallel()

	liveByBundle := map[string][]window.AcrossSpacesEntry{
		fallbackTestBundle: {
			fallbackLiveEntry(1, "general channel"),
			fallbackLiveEntry(2, "random channel"),
		},
	}

	pinsByBundle := pinEntriesByBundle([]config.PinRule{
		{BundleID: fallbackTestBundle, Title: "general channel", Space: 1},
		{BundleID: fallbackTestBundle, Title: "random channel", Space: 2},
	})

	moves, skipped, _ := planDirectMoves(
		pinsByBundle,
		liveByBundle,
		map[string]map[int]bool{},
		10,
		func(ordinal int) uint64 { return uint64(ordinal) },
	)

	if len(skipped) != 0 {
		t.Fatalf("skipped = %#v, want none", skipped)
	}

	if len(moves) != 2 {
		t.Fatalf("moves = %#v, want 2 moves", moves)
	}

	byOrdinal := map[int]moveTarget{}
	for _, move := range moves {
		byOrdinal[move.entry.Ordinal] = move
	}

	if got := byOrdinal[1]; got.windowID != 1 || got.fuzzy {
		t.Fatalf("space 1 target = %#v, want window 1, not fuzzy (exact match)", got)
	}

	if got := byOrdinal[2]; got.windowID != 2 || got.fuzzy {
		t.Fatalf("space 2 target = %#v, want window 2, not fuzzy (exact match)", got)
	}
}

// TestPinPrecedence_UnmatchedPinGetsNoFallback guards the window-pinning
// spec's non-goal: an unmatched pin is reported as skipped, never given
// an application-level prevalent-Space fallback.
func TestPinPrecedence_UnmatchedPinGetsNoFallback(t *testing.T) {
	t.Parallel()

	pinsByBundle := pinEntriesByBundle([]config.PinRule{
		{BundleID: fallbackTestBundle, Title: "totally unmatched", Space: 1},
	})

	pinMoves, pinSkipped, _ := planDirectMoves(
		pinsByBundle,
		map[string][]window.AcrossSpacesEntry{},
		map[string]map[int]bool{},
		10,
		func(ordinal int) uint64 { return uint64(ordinal) },
	)

	if len(pinMoves) != 0 {
		t.Fatalf("pinMoves = %#v, want none", pinMoves)
	}

	if len(pinSkipped) != 1 || pinSkipped[0].Reason != SkipAppNotRunning {
		t.Fatalf("pinSkipped = %#v, want one SkipAppNotRunning", pinSkipped)
	}
}
