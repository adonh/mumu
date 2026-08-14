package space_test

import (
	"testing"

	"github.com/adonh/mumu/internal/space"
)

// TestLogicalNumbering_SelfConsistent validates the logical left-to-right
// Space numbering's internal invariants against whatever display/Space
// arrangement is live on the machine running the test. It intentionally
// does not assert a fixed expected mapping — SLSCopyManagedDisplaySpaces has
// no test seam for synthetic multi-display arrangements — but round-trip
// and count invariants alone are enough to catch regressions in the
// sorting/concatenation logic. Skips (rather than fails) when no displays
// are reported, since that's expected in headless CI environments.
func TestLogicalNumbering_SelfConsistent(t *testing.T) {
	t.Parallel()

	counts := space.LeftToRightSpaceCounts()
	if len(counts) == 0 {
		t.Skip("no displays reported; skipping on headless environment")
	}

	total := 0
	for _, c := range counts {
		total += c
	}

	if got := space.LogicalCount(); got != total {
		t.Fatalf(
			"LogicalCount() = %d, want %d (sum of LeftToRightSpaceCounts %v)",
			got,
			total,
			counts,
		)
	}

	seen := make(map[uint64]bool, total)

	for logicalIdx := 1; logicalIdx <= total; logicalIdx++ {
		sid := space.LogicalSpaceID(logicalIdx)
		if sid == 0 {
			t.Fatalf(
				"LogicalSpaceID(%d) = 0, want a non-zero Space ID (total=%d)",
				logicalIdx,
				total,
			)
		}

		if seen[sid] {
			t.Fatalf(
				"LogicalSpaceID(%d) = %d, which was already returned for a different index",
				logicalIdx,
				sid,
			)
		}

		seen[sid] = true

		if back := space.LogicalIndexForSpace(sid); back != logicalIdx {
			t.Fatalf(
				"LogicalIndexForSpace(LogicalSpaceID(%d)=%d) = %d, want %d (round trip)",
				logicalIdx,
				sid,
				back,
				logicalIdx,
			)
		}
	}

	if got := space.LogicalSpaceID(total + 1); got != 0 {
		t.Fatalf("LogicalSpaceID(%d) = %d, want 0 (out of range)", total+1, got)
	}

	if got := space.LogicalSpaceID(0); got != 0 {
		t.Fatalf("LogicalSpaceID(0) = %d, want 0 (out of range)", got)
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

	if got := space.LogicalCount(); got != total {
		t.Fatalf(
			"LogicalCount() = %d, want %d (Count(), numerically equal by contract)",
			got,
			total,
		)
	}

	seenIndex := make(map[int]bool, total)

	// LogicalSpaceID(1..total) enumerates every currently valid Space ID
	// (just in a different order than Mission Control's), so it's a
	// convenient source of real Space IDs to round-trip through the
	// function under test.
	for logicalIdx := 1; logicalIdx <= total; logicalIdx++ {
		sid := space.LogicalSpaceID(logicalIdx)
		if sid == 0 {
			t.Fatalf(
				"LogicalSpaceID(%d) = 0, want a non-zero Space ID (total=%d)",
				logicalIdx,
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

	if got := space.MissionControlIndexForSpace(0); got != 0 {
		t.Fatalf("MissionControlIndexForSpace(0) = %d, want 0 (not found)", got)
	}
}
