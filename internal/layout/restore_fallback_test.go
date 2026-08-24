package layout //nolint:testpackage // tests unexported restore planning helpers

import (
	"errors"
	"testing"

	"github.com/adonh/mumu/internal/window"
)

const fallbackTestBundle = "com.example.chrome"

var errPrimarySpaceUnavailable = errors.New("primary space unavailable")

func fallbackLiveEntry(windowID uint32, title string) window.AcrossSpacesEntry {
	return window.AcrossSpacesEntry{WindowID: windowID, Title: title}
}

func TestPlanFallbackMoves(t *testing.T) {
	t.Parallel()

	t.Run("uses the uniquely most prevalent assignment", func(t *testing.T) {
		t.Parallel()

		liveByBundle := map[string][]window.AcrossSpacesEntry{
			fallbackTestBundle: {
				fallbackLiveEntry(1, "matched"),
				fallbackLiveEntry(2, "unmatched one"),
				fallbackLiveEntry(3, "unmatched two"),
			},
		}
		usedByBundle := map[string]map[int]bool{
			fallbackTestBundle: {0: true},
		}
		assignmentOrdinals := map[string][]int{
			fallbackTestBundle: {4, 4, 7},
		}

		targets, skipped := planFallbackMoves(
			liveByBundle,
			usedByBundle,
			assignmentOrdinals,
			map[string]int{},
			func() (fallbackTarget, error) {
				t.Fatal("primary display resolver must not run for a unique target")

				return fallbackTarget{}, nil
			},
			func(ordinal int) uint64 {
				if ordinal != 4 {
					t.Fatalf("fallback ordinal = %d, want 4", ordinal)
				}

				return 104
			},
		)

		if len(skipped) != 0 {
			t.Fatalf("skipped = %#v, want none", skipped)
		}

		if len(targets) != 2 {
			t.Fatalf("fallback targets = %d, want 2", len(targets))
		}

		for _, target := range targets {
			if !target.fallback {
				t.Fatal("fallback target marker = false, want true")
			}

			if target.sid != 104 {
				t.Fatalf("fallback space ID = %d, want 104", target.sid)
			}

			if target.entry.Ordinal != 4 {
				t.Fatalf("fallback ordinal = %d, want 4", target.entry.Ordinal)
			}
		}

		if !usedByBundle[fallbackTestBundle][1] || !usedByBundle[fallbackTestBundle][2] {
			t.Fatalf(
				"unmatched live windows were not claimed: %#v",
				usedByBundle[fallbackTestBundle],
			)
		}
	})

	t.Run("uses the exact primary display space when assignments tie", func(t *testing.T) {
		t.Parallel()

		liveByBundle := map[string][]window.AcrossSpacesEntry{
			fallbackTestBundle: {
				fallbackLiveEntry(1, "matched"),
				fallbackLiveEntry(2, "unmatched"),
			},
		}
		usedByBundle := map[string]map[int]bool{
			fallbackTestBundle: {0: true},
		}
		assignmentOrdinals := map[string][]int{
			fallbackTestBundle: {2, 5},
		}

		targets, skipped := planFallbackMoves(
			liveByBundle,
			usedByBundle,
			assignmentOrdinals,
			map[string]int{},
			func() (fallbackTarget, error) {
				return fallbackTarget{ordinal: 7, sid: 107}, nil
			},
			func(int) uint64 {
				t.Fatal("tied fallback must use the primary display's resolved space ID")

				return 0
			},
		)

		if len(skipped) != 0 {
			t.Fatalf("skipped = %#v, want none", skipped)
		}

		if len(targets) != 1 {
			t.Fatalf("fallback targets = %d, want 1", len(targets))
		}

		if targets[0].entry.Ordinal != 7 || targets[0].sid != 107 {
			t.Fatalf("fallback target = %#v, want primary ordinal 7 / space ID 107", targets[0])
		}
	})

	t.Run("reports an unavailable primary target without claiming windows", func(t *testing.T) {
		t.Parallel()

		liveByBundle := map[string][]window.AcrossSpacesEntry{
			fallbackTestBundle: {
				fallbackLiveEntry(1, "matched"),
				fallbackLiveEntry(2, "unmatched"),
			},
		}
		usedByBundle := map[string]map[int]bool{
			fallbackTestBundle: {0: true},
		}
		assignmentOrdinals := map[string][]int{
			fallbackTestBundle: {2, 5},
		}

		targets, skipped := planFallbackMoves(
			liveByBundle,
			usedByBundle,
			assignmentOrdinals,
			map[string]int{},
			func() (fallbackTarget, error) {
				return fallbackTarget{}, errPrimarySpaceUnavailable
			},
			func(int) uint64 {
				t.Fatal("logical space lookup must not run after primary target resolution fails")

				return 0
			},
		)

		if len(targets) != 0 {
			t.Fatalf("fallback targets = %#v, want none", targets)
		}

		if len(skipped) != 1 || skipped[0].Reason != SkipFallbackTargetUnavailable {
			t.Fatalf("skipped = %#v, want one unavailable-fallback skip", skipped)
		}

		if !skipped[0].Fallback {
			t.Fatal("unavailable fallback skip is not marked as a fallback")
		}

		if usedByBundle[fallbackTestBundle][1] {
			t.Fatalf(
				"unavailable fallback claimed live window: %#v",
				usedByBundle[fallbackTestBundle],
			)
		}
	})

	t.Run("does not move windows without valid assignments", func(t *testing.T) {
		t.Parallel()

		liveByBundle := map[string][]window.AcrossSpacesEntry{
			fallbackTestBundle: {fallbackLiveEntry(1, "unmatched")},
		}
		usedByBundle := map[string]map[int]bool{
			fallbackTestBundle: {},
		}

		targets, skipped := planFallbackMoves(
			liveByBundle,
			usedByBundle,
			map[string][]int{},
			map[string]int{},
			func() (fallbackTarget, error) {
				t.Fatal("primary display resolver must not run without valid assignments")

				return fallbackTarget{}, nil
			},
			func(int) uint64 {
				t.Fatal("logical space lookup must not run without valid assignments")

				return 0
			},
		)

		if len(targets) != 0 || len(skipped) != 0 {
			t.Fatalf("targets/skipped = %#v/%#v, want none", targets, skipped)
		}

		if usedByBundle[fallbackTestBundle][0] {
			t.Fatalf(
				"window without valid assignments was claimed: %#v",
				usedByBundle[fallbackTestBundle],
			)
		}
	})

	t.Run("configured default overrides an unambiguous prevalent Space", func(t *testing.T) {
		t.Parallel()

		liveByBundle := map[string][]window.AcrossSpacesEntry{
			fallbackTestBundle: {
				fallbackLiveEntry(1, "matched"),
				fallbackLiveEntry(2, "unmatched"),
			},
		}
		usedByBundle := map[string]map[int]bool{
			fallbackTestBundle: {0: true},
		}
		assignmentOrdinals := map[string][]int{
			fallbackTestBundle: {2, 2, 2},
		}
		defaultSpaces := map[string]int{
			fallbackTestBundle: 6,
		}

		targets, skipped := planFallbackMoves(
			liveByBundle,
			usedByBundle,
			assignmentOrdinals,
			defaultSpaces,
			func() (fallbackTarget, error) {
				t.Fatal("primary display resolver must not run when a default is configured")

				return fallbackTarget{}, nil
			},
			func(ordinal int) uint64 {
				if ordinal != 6 {
					t.Fatalf("fallback ordinal = %d, want configured default 6", ordinal)
				}

				return 106
			},
		)

		if len(skipped) != 0 {
			t.Fatalf("skipped = %#v, want none", skipped)
		}

		if len(targets) != 1 {
			t.Fatalf("fallback targets = %d, want 1", len(targets))
		}

		if targets[0].entry.Ordinal != 6 || targets[0].sid != 106 {
			t.Fatalf("fallback target = %#v, want configured ordinal 6 / space ID 106", targets[0])
		}

		if !targets[0].defaultConfigured {
			t.Fatal("defaultConfigured marker = false, want true")
		}
	})

	t.Run(
		"configured default overrides a tied prevalent Space without a primary-display lookup",
		func(t *testing.T) {
			t.Parallel()

			liveByBundle := map[string][]window.AcrossSpacesEntry{
				fallbackTestBundle: {
					fallbackLiveEntry(1, "matched"),
					fallbackLiveEntry(2, "unmatched"),
				},
			}
			usedByBundle := map[string]map[int]bool{
				fallbackTestBundle: {0: true},
			}
			assignmentOrdinals := map[string][]int{
				fallbackTestBundle: {2, 5},
			}
			defaultSpaces := map[string]int{
				fallbackTestBundle: 8,
			}

			targets, skipped := planFallbackMoves(
				liveByBundle,
				usedByBundle,
				assignmentOrdinals,
				defaultSpaces,
				func() (fallbackTarget, error) {
					t.Fatal("primary display resolver must not run when a default is configured")

					return fallbackTarget{}, nil
				},
				func(ordinal int) uint64 { return uint64(ordinal) },
			)

			if len(skipped) != 0 {
				t.Fatalf("skipped = %#v, want none", skipped)
			}

			if len(targets) != 1 || targets[0].entry.Ordinal != 8 {
				t.Fatalf("targets = %#v, want configured ordinal 8", targets)
			}
		},
	)

	t.Run(
		"configured default activates a fallback with zero valid assignments",
		func(t *testing.T) {
			t.Parallel()

			liveByBundle := map[string][]window.AcrossSpacesEntry{
				fallbackTestBundle: {
					fallbackLiveEntry(1, "one"),
					fallbackLiveEntry(2, "two"),
				},
			}
			usedByBundle := map[string]map[int]bool{}
			defaultSpaces := map[string]int{
				fallbackTestBundle: 9,
			}

			targets, skipped := planFallbackMoves(
				liveByBundle,
				usedByBundle,
				map[string][]int{},
				defaultSpaces,
				func() (fallbackTarget, error) {
					t.Fatal("primary display resolver must not run when a default is configured")

					return fallbackTarget{}, nil
				},
				func(ordinal int) uint64 { return uint64(ordinal) },
			)

			if len(skipped) != 0 {
				t.Fatalf("skipped = %#v, want none", skipped)
			}

			if len(targets) != 2 {
				t.Fatalf(
					"fallback targets = %d, want 2 (both windows, zero valid assignments)",
					len(targets),
				)
			}

			for _, target := range targets {
				if target.entry.Ordinal != 9 || !target.defaultConfigured {
					t.Fatalf("target = %#v, want ordinal 9 and defaultConfigured = true", target)
				}
			}
		},
	)
}

func TestMoveFailureSkipPreservesFallbackMarker(t *testing.T) {
	t.Parallel()

	entry := Entry{BundleID: fallbackTestBundle, Title: "window", Ordinal: 4}

	t.Run("direct match", func(t *testing.T) {
		t.Parallel()

		skipped := moveFailureSkip(moveTarget{entry: entry})
		if skipped.Fallback {
			t.Fatal("direct-match failure is marked as a fallback")
		}
	})

	t.Run("fallback placement", func(t *testing.T) {
		t.Parallel()

		skipped := moveFailureSkip(moveTarget{entry: entry, fallback: true})
		if !skipped.Fallback {
			t.Fatal("fallback move failure is not marked as a fallback")
		}
	})
}
