package layout

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/adonh/mumu/internal/config"
	derrors "github.com/adonh/mumu/internal/errors"
)

// layoutsDirName is the name of the subdirectory, inside the configured
// data directory, that holds one JSON file per saved layout. Saved layouts
// are internal state, not a user-facing editable file: JSON, rather than
// YAML, signals that they're not meant for hand-editing.
const layoutsDirName = "layouts"

const dirMode = 0o755

const fileMode = 0o644

const jsonIndent = "  "

// layoutsDir resolves the full path to the layouts subdirectory, using the
// data directory from mumu's configuration.
func layoutsDir() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}

	return filepath.Join(cfg.DataDir, layoutsDirName), nil
}

func filePath(dir string, displayCount int) string {
	return filepath.Join(dir, strconv.Itoa(displayCount)+".json")
}

// Save persists a layout as its own JSON file, named for its display
// count, overwriting any previously saved layout for that same display
// count.
func Save(saved *Layout) error {
	dir, err := layoutsDir()
	if err != nil {
		return err
	}

	err = os.MkdirAll(dir, dirMode)
	if err != nil {
		return derrors.Wrapf(err, derrors.CodeConfigIOFailed, "creating layouts directory")
	}

	entry := *saved
	entry.SchemaVersion = SchemaVersion

	data, err := json.MarshalIndent(&entry, "", jsonIndent)
	if err != nil {
		return derrors.Wrapf(err, derrors.CodeSerializationFailed, "marshaling layout")
	}

	err = os.WriteFile(filePath(dir, saved.DisplayCount), data, fileMode)
	if err != nil {
		return derrors.Wrapf(err, derrors.CodeConfigIOFailed, "writing layout file")
	}

	return nil
}

// Load reads the saved layout for the given display count, returning a
// clear error if none exists.
func Load(displayCount int) (*Layout, error) {
	dir, err := layoutsDir()
	if err != nil {
		return nil, err
	}

	path := filePath(dir, displayCount)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, derrors.Newf(
				derrors.CodeInvalidInput,
				"no saved layout found for %d display(s); run 'mumu save' first",
				displayCount,
			)
		}

		return nil, derrors.Wrapf(err, derrors.CodeConfigIOFailed, "reading layout file %s", path)
	}

	err = checkSchemaVersion(data, path)
	if err != nil {
		return nil, err
	}

	var loaded Layout

	err = json.Unmarshal(data, &loaded)
	if err != nil {
		return nil, derrors.Wrapf(
			err,
			derrors.CodeSerializationFailed,
			"parsing layout file %s",
			path,
		)
	}

	return &loaded, nil
}

// checkSchemaVersion pre-parses just a saved layout file's schemaVersion
// field, ahead of the full Layout unmarshal, so a file written by an
// older mumu version (whose Entry.Ordinal shape no longer matches — see
// SchemaVersion's doc comment) fails with a clear, actionable error
// instead of a raw encoding/json type-mismatch error from the full
// unmarshal.
func checkSchemaVersion(data []byte, path string) error {
	var versioned struct {
		SchemaVersion int `json:"schemaVersion"`
	}

	err := json.Unmarshal(data, &versioned)
	if err != nil {
		// Malformed JSON: let the full unmarshal below produce today's
		// existing "malformed saved-layout file" error.
		return nil //nolint:nilerr // intentional: defer to the full unmarshal's error.
	}

	if versioned.SchemaVersion != SchemaVersion {
		return derrors.Newf(
			derrors.CodeSerializationFailed,
			"saved layout file %s was written by an incompatible mumu version "+
				"(schema %d, expected %d); run 'mumu save' again to recreate it",
			path,
			versioned.SchemaVersion,
			SchemaVersion,
		)
	}

	return nil
}

// Exists reports whether a saved layout exists for the given display count.
func Exists(displayCount int) bool {
	dir, err := layoutsDir()
	if err != nil {
		return false
	}

	_, err = os.Stat(filePath(dir, displayCount))

	return err == nil
}

// Delete removes the saved layout for the given display count.
func Delete(displayCount int) error {
	dir, err := layoutsDir()
	if err != nil {
		return err
	}

	err = os.Remove(filePath(dir, displayCount))
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
	dir, err := layoutsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
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
