package space_test

import (
	"testing"

	"github.com/adonh/mumu/internal/space"
)

// TestLogicalNumbering_SelfConsistent validates the logical
// per-display Space numbering's internal invariants against whatever
// display/Space arrangement is live on the machine running the test. It
// intentionally does not assert a fixed expected mapping — SLSCopyManagedDisplaySpaces
// has no test seam for synthetic multi-display arrangements — but round-trip
// and count invariants alone are enough to catch regressions in the
// per-display indexing logic. Skips (rather than fails) when no displays
// are reported, since that's expected in headless CI environments.
func TestLogicalNumbering_SelfConsistent(t *testing.T) {
	t.Parallel()

	counts := space.LeftToRightSpaceCounts()
	if len(counts) == 0 {
		t.Skip("no displays reported; skipping on headless environment")
	}

	if got := space.LogicalDisplayCount(); got != len(counts) {
		t.Fatalf(
			"LogicalDisplayCount() = %d, want %d (len(LeftToRightSpaceCounts()) %v)",
			got,
			len(counts),
			counts,
		)
	}

	seen := make(map[uint64]bool)

	for displayOrdinal, spaceCount := range counts {
		for spaceOrdinal := 1; spaceOrdinal <= spaceCount; spaceOrdinal++ {
			ordinal := space.Ordinal{Display: displayOrdinal + 1, Space: spaceOrdinal}

			sid := space.IDForOrdinal(ordinal)
			if sid == 0 {
				t.Fatalf(
					"IDForOrdinal(%v) = 0, want a non-zero Space ID (counts=%v)",
					ordinal,
					counts,
				)
			}

			if seen[sid] {
				t.Fatalf(
					"IDForOrdinal(%v) = %d, which was already returned for a different ordinal",
					ordinal,
					sid,
				)
			}

			seen[sid] = true

			if back := space.OrdinalForSpace(sid); back != ordinal {
				t.Fatalf(
					"OrdinalForSpace(IDForOrdinal(%v)=%d) = %v, want %v (round trip)",
					ordinal,
					sid,
					back,
					ordinal,
				)
			}
		}

		if got := space.IDForOrdinal(
			space.Ordinal{Display: displayOrdinal + 1, Space: spaceCount + 1},
		); got != 0 {
			t.Fatalf(
				"IDForOrdinal({%d, %d}) = %d, want 0 (out of range)",
				displayOrdinal+1,
				spaceCount+1,
				got,
			)
		}
	}

	if got := space.IDForOrdinal(space.Ordinal{Display: len(counts) + 1, Space: 1}); got != 0 {
		t.Fatalf("IDForOrdinal({%d, 1}) = %d, want 0 (out of range)", len(counts)+1, got)
	}

	if got := space.IDForOrdinal(space.Ordinal{}); got != 0 {
		t.Fatalf("IDForOrdinal({}) = %d, want 0 (out of range)", got)
	}
}

// TestMissionControlIndexForSpace_SelfConsistent validates the reverse
// Mission Control Space lookup's internal invariants against whatever
// display/Space arrangement is live on the machine running the test,
// mirroring TestLogicalNumbering_SelfConsistent above. It doesn't assert a
// fixed expected mapping for the same reason that test doesn't, but
// confirms every currently valid Space ID round-trips to a distinct index
// within space.Focus's valid 1..Count() range — i.e. the reverse lookup is
// a bijection onto that range, not just "some number that happens to be
// non-zero". Skips (rather than fails) when no displays are reported,
// since that's expected in headless CI environments.
func TestMissionControlIndexForSpace_SelfConsistent(t *testing.T) {
	t.Parallel()

	total := space.Count()
	if total == 0 {
		t.Skip("no displays reported; skipping on headless environment")
	}

	counts := space.LeftToRightSpaceCounts()

	seenIndex := make(map[int]bool, total)

	for displayOrdinal, spaceCount := range counts {
		for spaceOrdinal := 1; spaceOrdinal <= spaceCount; spaceOrdinal++ {
			ordinal := space.Ordinal{Display: displayOrdinal + 1, Space: spaceOrdinal}

			sid := space.IDForOrdinal(ordinal)
			if sid == 0 {
				t.Fatalf(
					"IDForOrdinal(%v) = 0, want a non-zero Space ID (total=%d)",
					ordinal,
					total,
				)
			}

			mcIdx := space.MissionControlIndexForSpace(sid)
			if mcIdx < 1 || mcIdx > total {
				t.Fatalf(
					"MissionControlIndexForSpace(%d) = %d, want a value in 1..%d (Focus's valid index range)",
					sid,
					mcIdx,
					total,
				)
			}

			if seenIndex[mcIdx] {
				t.Fatalf(
					"MissionControlIndexForSpace(%d) = %d, which was already returned for a different Space ID",
					sid,
					mcIdx,
				)
			}

			seenIndex[mcIdx] = true
		}
	}

	if got := space.MissionControlIndexForSpace(0); got != 0 {
		t.Fatalf("MissionControlIndexForSpace(0) = %d, want 0 (not found)", got)
	}
}
