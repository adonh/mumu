package space //nolint:testpackage // Tests unexported primary-display resolution helper.

import "testing"

func TestPrimaryDisplayCurrentSpace(t *testing.T) {
	t.Parallel()

	t.Run("returns logical index for primary display space", func(t *testing.T) {
		t.Parallel()

		spaceID, logicalIndex, err := primaryDisplayCurrentSpace(42, func(spaceID uint64) int {
			if spaceID != 42 {
				t.Fatalf("lookup space ID = %d, want 42", spaceID)
			}

			return 7
		})
		if err != nil {
			t.Fatalf("primaryDisplayCurrentSpace() error = %v, want nil", err)
		}

		if spaceID != 42 || logicalIndex != 7 {
			t.Fatalf(
				"primaryDisplayCurrentSpace() = (%d, %d), want (42, 7)",
				spaceID,
				logicalIndex,
			)
		}
	})

	t.Run("rejects unresolved primary display space", func(t *testing.T) {
		t.Parallel()

		_, _, err := primaryDisplayCurrentSpace(0, func(uint64) int {
			t.Fatal("logical index lookup must not run for an unresolved primary display space")

			return 0
		})
		if err == nil {
			t.Fatal("primaryDisplayCurrentSpace() error = nil, want error")
		}
	})

	t.Run("rejects primary display space outside logical ordering", func(t *testing.T) {
		t.Parallel()

		_, _, err := primaryDisplayCurrentSpace(42, func(uint64) int { return 0 })
		if err == nil {
			t.Fatal("primaryDisplayCurrentSpace() error = nil, want error")
		}
	})
}

func TestPrimaryDisplayCurrentLogicalIndexIsResolvable(t *testing.T) {
	t.Parallel()

	if len(LeftToRightSpaceCounts()) == 0 {
		t.Skip("no displays reported; skipping on headless environment")
	}

	logicalIndex, err := PrimaryDisplayCurrentLogicalIndex()
	if err != nil {
		t.Fatalf("PrimaryDisplayCurrentLogicalIndex() error = %v, want nil", err)
	}

	if sid := LogicalSpaceID(logicalIndex); sid == 0 {
		t.Fatalf(
			"LogicalSpaceID(PrimaryDisplayCurrentLogicalIndex()=%d) = 0, want a current space",
			logicalIndex,
		)
	}
}
