package layout //nolint:testpackage // tests unexported default-space helpers

import (
	"testing"

	"github.com/adonh/mumu/internal/config"
)

const defaultSpacesTestSlackBundle = "com.example.slack"

func TestDefaultSpacesByBundle_Converts(t *testing.T) {
	t.Parallel()

	rules := []config.DefaultSpaceRule{
		{BundleID: defaultSpacesTestSlackBundle, Ordinal: 1},
		{BundleID: "com.example.chrome", Ordinal: 5},
	}

	got := defaultSpacesByBundle(rules)

	want := map[string]int{
		defaultSpacesTestSlackBundle: 1,
		"com.example.chrome":         5,
	}

	if len(got) != len(want) {
		t.Fatalf("defaultSpacesByBundle() = %#v, want %#v", got, want)
	}

	for bundleID, ordinal := range want {
		if got[bundleID] != ordinal {
			t.Fatalf("defaultSpacesByBundle()[%q] = %d, want %d", bundleID, got[bundleID], ordinal)
		}
	}
}

func TestDefaultSpacesByBundle_DuplicateBundleIDLastWins(t *testing.T) {
	t.Parallel()

	rules := []config.DefaultSpaceRule{
		{BundleID: defaultSpacesTestSlackBundle, Ordinal: 1},
		{BundleID: defaultSpacesTestSlackBundle, Ordinal: 3},
	}

	got := defaultSpacesByBundle(rules)

	if got[defaultSpacesTestSlackBundle] != 3 {
		t.Fatalf(
			"defaultSpacesByBundle()[slack] = %d, want 3 (last rule wins)",
			got[defaultSpacesTestSlackBundle],
		)
	}
}

func TestDefaultSpacesByBundle_Empty(t *testing.T) {
	t.Parallel()

	got := defaultSpacesByBundle(nil)

	if len(got) != 0 {
		t.Fatalf("defaultSpacesByBundle(nil) = %#v, want empty", got)
	}
}
