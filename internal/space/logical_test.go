package space_test

import (
	"testing"

	"github.com/y3owk1n/mimi/internal/space"
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
