package layout_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adonh/mumu/internal/layout"
	"github.com/adonh/mumu/internal/space"
)

// setDataDir points mumu's config at a fresh temp HOME so layout
// persistence tests never touch a real config.yaml or saved layouts.
func setDataDir(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
}

func layoutsDirForTest(t *testing.T) string {
	t.Helper()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	return filepath.Join(home, "Library/Application Support", "mumu", "layouts")
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	setDataDir(t)

	saved := &layout.Layout{
		DisplayCount: 2,
		SpaceCounts:  []int{2, 3},
		SavedAt:      time.Now().UTC().Truncate(time.Second),
		Entries: []layout.Entry{
			{
				BundleID: "com.example.App",
				Title:    "Window",
				Index:    0,
				Ordinal:  space.Ordinal{Display: 1, Space: 1},
			},
		},
	}

	err := layout.Save(saved)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := layout.Load(2)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	match := got.DisplayCount == 2 &&
		len(got.Entries) == 1 &&
		got.Entries[0].BundleID == "com.example.App"
	if !match {
		t.Fatalf("Load() = %#v, want match for saved layout", got)
	}
}

func TestSave_OverwritesSameDisplayCount(t *testing.T) {
	setDataDir(t)

	err := layout.Save(&layout.Layout{DisplayCount: 2, Entries: []layout.Entry{{BundleID: "old"}}})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	err = layout.Save(&layout.Layout{DisplayCount: 2, Entries: []layout.Entry{{BundleID: "new"}}})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := layout.Load(2)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(got.Entries) != 1 || got.Entries[0].BundleID != "new" {
		t.Fatalf("Load() = %#v, want overwritten entry", got)
	}
}

func TestSave_EachDisplayCountHasItsOwnFile(t *testing.T) {
	setDataDir(t)

	err := layout.Save(&layout.Layout{DisplayCount: 2})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	err = layout.Save(&layout.Layout{DisplayCount: 3})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	dir := layoutsDirForTest(t)

	for _, name := range []string{"2.json", "3.json"} {
		_, statErr := os.Stat(filepath.Join(dir, name))
		if statErr != nil {
			t.Fatalf("expected %s to exist in %s: %v", name, dir, statErr)
		}
	}

	counts, err := layout.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(counts) != 2 || counts[0] != 2 || counts[1] != 3 {
		t.Fatalf("List() = %v, want [2 3]", counts)
	}
}

func TestLoad_MissingDisplayCount(t *testing.T) {
	setDataDir(t)

	_, err := layout.Load(5)
	if err == nil {
		t.Fatal("Load() error = nil, want error for missing display count")
	}
}

func TestExists(t *testing.T) {
	setDataDir(t)

	if layout.Exists(2) {
		t.Fatal("Exists(2) = true before any save, want false")
	}

	err := layout.Save(&layout.Layout{DisplayCount: 2})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if !layout.Exists(2) {
		t.Fatal("Exists(2) = false after save, want true")
	}
}

func TestDelete_RemovesOnlyTargetedFile(t *testing.T) {
	setDataDir(t)

	err := layout.Save(&layout.Layout{DisplayCount: 2})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	err = layout.Save(&layout.Layout{DisplayCount: 3})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	err = layout.Delete(2)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if layout.Exists(2) {
		t.Fatal("Exists(2) = true after Delete(2), want false")
	}

	if !layout.Exists(3) {
		t.Fatal("Exists(3) = false after Delete(2), want true (unrelated file)")
	}
}

func TestDelete_MissingDisplayCount(t *testing.T) {
	setDataDir(t)

	err := layout.Delete(5)
	if err == nil {
		t.Fatal("Delete() error = nil, want error for missing display count")
	}
}

func TestList_EmptyWhenNoDirectoryExists(t *testing.T) {
	setDataDir(t)

	counts, err := layout.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(counts) != 0 {
		t.Fatalf("List() = %v, want empty", counts)
	}
}

func TestLoad_MalformedLayoutFile(t *testing.T) {
	setDataDir(t)

	dir := layoutsDirForTest(t)

	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	err = os.WriteFile(
		filepath.Join(dir, "2.json"),
		[]byte("{unterminated"),
		0o644,
	)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = layout.Load(2)
	if err == nil {
		t.Fatal("Load() error = nil, want error for malformed layout file")
	}

	if strings.Contains(err.Error(), "run 'mumu save' again") {
		t.Fatalf(
			"Load() error = %v, want the malformed-JSON error, not the schema-version error",
			err,
		)
	}
}

func TestLoad_OutdatedSchemaVersion(t *testing.T) {
	setDataDir(t)

	dir := layoutsDirForTest(t)

	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	// Schema version 1's Entry.Ordinal was a bare number, not a
	// {"display","space"} object — this is what an old-schema saved
	// layout file actually looked like on disk.
	oldSchemaLayout := `{
		"schemaVersion": 1,
		"displayCount": 2,
		"spaceCounts": [2, 3],
		"entries": [
			{"bundleId": "com.example.App", "title": "Window", "index": 0, "ordinal": 1}
		],
		"savedAt": "2024-01-01T00:00:00Z"
	}`

	err = os.WriteFile(filepath.Join(dir, "2.json"), []byte(oldSchemaLayout), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = layout.Load(2)
	if err == nil {
		t.Fatal("Load() error = nil, want error for outdated schema version")
	}

	if !strings.Contains(err.Error(), "run 'mumu save' again") {
		t.Fatalf(
			"Load() error = %v, want a clear \"run 'mumu save' again\" message",
			err,
		)
	}
}
