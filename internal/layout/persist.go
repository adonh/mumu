package layout

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	derrors "github.com/adonh/mumu/internal/errors"
	"github.com/adonh/mumu/internal/paths"
)

// DefaultDir is the default directory where saved layouts are stored,
// following the "~/.local/share/<app>" user-data convention. This changed
// from the pre-rebrand app's equivalent path as part of the rename: there's
// no installed user base to preserve continuity for, and keeping the old
// app name here would undercut the rebrand's whole point of avoiding
// confusion with that unrelated, similarly-named app.
const DefaultDir = "~/.local/share/mumu/layouts"

func dirPath() string {
	return paths.ExpandHome(DefaultDir)
}

func filePath(displayCount int) string {
	return filepath.Join(dirPath(), strconv.Itoa(displayCount)+".json")
}

// Save persists a layout as JSON keyed by its display count, overwriting
// any previously saved layout for that same display count.
func Save(saved *Layout) error {
	dir := dirPath()

	err := os.MkdirAll(dir, 0o755) //nolint:mnd
	if err != nil {
		return derrors.Wrapf(err, derrors.CodeConfigIOFailed, "creating layouts directory")
	}

	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return derrors.Wrapf(err, derrors.CodeSerializationFailed, "marshaling layout")
	}

	err = os.WriteFile(filePath(saved.DisplayCount), data, 0o644) //nolint:mnd
	if err != nil {
		return derrors.Wrapf(err, derrors.CodeConfigIOFailed, "writing layout file")
	}

	return nil
}

// Load reads the saved layout for the given display count, returning a
// clear error if none exists.
func Load(displayCount int) (*Layout, error) {
	data, err := os.ReadFile(filePath(displayCount))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, derrors.Newf(
				derrors.CodeInvalidInput,
				"no saved layout found for %d display(s); run 'mumu save' first",
				displayCount,
			)
		}

		return nil, derrors.Wrapf(err, derrors.CodeConfigIOFailed, "reading layout file")
	}

	var loaded Layout

	err = json.Unmarshal(data, &loaded)
	if err != nil {
		return nil, derrors.Wrapf(err, derrors.CodeSerializationFailed, "parsing layout file")
	}

	return &loaded, nil
}

// Exists reports whether a saved layout exists for the given display count.
func Exists(displayCount int) bool {
	_, err := os.Stat(filePath(displayCount))

	return err == nil
}

// Delete removes the saved layout for the given display count.
func Delete(displayCount int) error {
	err := os.Remove(filePath(displayCount))
	if err != nil {
		if os.IsNotExist(err) {
			return derrors.Newf(
				derrors.CodeInvalidInput,
				"no saved layout found for %d display(s)",
				displayCount,
			)
		}

		return derrors.Wrapf(err, derrors.CodeConfigIOFailed, "deleting layout file")
	}

	return nil
}

// List returns the display counts for which a layout has been saved,
// sorted ascending.
func List() ([]int, error) {
	entries, err := os.ReadDir(dirPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []int{}, nil
		}

		return nil, derrors.Wrapf(err, derrors.CodeConfigIOFailed, "reading layouts directory")
	}

	counts := make([]int, 0, len(entries))

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		if filepath.Ext(name) != ".json" {
			continue
		}

		n, err := strconv.Atoi(strings.TrimSuffix(name, ".json"))
		if err != nil {
			continue
		}

		counts = append(counts, n)
	}

	sort.Ints(counts)

	return counts, nil
}
