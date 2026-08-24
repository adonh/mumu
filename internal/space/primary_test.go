package space //nolint:testpackage // Tests unexported primary-display resolution helper.

import "testing"

func TestPrimaryDisplayCurrentSpace(t *testing.T) {
	t.Parallel()

	t.Run("returns logical ordinal for primary display space", func(t *testing.T) {
		t.Parallel()

		want := Ordinal{Display: 1, Space: 7}

		spaceID, ordinal, err := primaryDisplayCurrentSpace(42, func(spaceID uint64) Ordinal {
			if spaceID != 42 {
				t.Fatalf("lookup space ID = %d, want 42", spaceID)
			}

			return want
		})
		if err != nil {
			t.Fatalf("primaryDisplayCurrentSpace() error = %v, want nil", err)
		}

		if spaceID != 42 || ordinal != want {
			t.Fatalf(
				"primaryDisplayCurrentSpace() = (%d, %v), want (42, %v)",
				spaceID,
				ordinal,
				want,
			)
		}
	})

	t.Run("rejects unresolved primary display space", func(t *testing.T) {
		t.Parallel()

		_, _, err := primaryDisplayCurrentSpace(0, func(uint64) Ordinal {
			t.Fatal("ordinal lookup must not run for an unresolved primary display space")

			return Ordinal{}
		})
		if err == nil {
			t.Fatal("primaryDisplayCurrentSpace() error = nil, want error")
		}
	})

	t.Run("rejects primary display space outside logical ordering", func(t *testing.T) {
		t.Parallel()

		_, _, err := primaryDisplayCurrentSpace(42, func(uint64) Ordinal { return Ordinal{} })
		if err == nil {
			t.Fatal("primaryDisplayCurrentSpace() error = nil, want error")
		}
	})
}

func TestPrimaryDisplayCurrentOrdinalIsResolvable(t *testing.T) {
	t.Parallel()

	if len(LeftToRightSpaceCounts()) == 0 {
		t.Skip("no displays reported; skipping on headless environment")
	}

	ordinal, err := PrimaryDisplayCurrentOrdinal()
	if err != nil {
		t.Fatalf("PrimaryDisplayCurrentOrdinal() error = %v, want nil", err)
	}

	if sid := IDForOrdinal(ordinal); sid == 0 {
		t.Fatalf(
			"IDForOrdinal(PrimaryDisplayCurrentOrdinal()=%v) = 0, want a current space",
			ordinal,
		)
	}
}
