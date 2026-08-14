package space //nolint:testpackage // Tests unexported primary-display resolution helper.

import "testing"

func TestMenuBarDisplayLogicalIndex(t *testing.T) {
	t.Parallel()

	t.Run("returns logical index for primary display space", func(t *testing.T) {
		t.Parallel()

		got, err := menuBarDisplayLogicalIndex(42, func(spaceID uint64) int {
			if spaceID != 42 {
				t.Fatalf("lookup space ID = %d, want 42", spaceID)
			}

			return 7
		})
		if err != nil {
			t.Fatalf("menuBarDisplayLogicalIndex() error = %v, want nil", err)
		}

		if got != 7 {
			t.Fatalf("menuBarDisplayLogicalIndex() = %d, want 7", got)
		}
	})

	t.Run("rejects unresolved primary display space", func(t *testing.T) {
		t.Parallel()

		_, err := menuBarDisplayLogicalIndex(0, func(uint64) int {
			t.Fatal("logical index lookup must not run for an unresolved menu-bar display space")

			return 0
		})
		if err == nil {
			t.Fatal("menuBarDisplayLogicalIndex() error = nil, want error")
		}
	})

	t.Run("rejects primary display space outside logical ordering", func(t *testing.T) {
		t.Parallel()

		_, err := menuBarDisplayLogicalIndex(42, func(uint64) int { return 0 })
		if err == nil {
			t.Fatal("menuBarDisplayLogicalIndex() error = nil, want error")
		}
	})
}

func TestMenuBarActiveLogicalIndexIsResolvable(t *testing.T) {
	t.Parallel()

	if len(LeftToRightSpaceCounts()) == 0 {
		t.Skip("no displays reported; skipping on headless environment")
	}

	logicalIndex, err := MenuBarActiveLogicalIndex()
	if err != nil {
		t.Fatalf("MenuBarActiveLogicalIndex() error = %v, want nil", err)
	}

	if sid := LogicalSpaceID(logicalIndex); sid == 0 {
		t.Fatalf(
			"LogicalSpaceID(MenuBarActiveLogicalIndex()=%d) = 0, want a current space",
			logicalIndex,
		)
	}
}
