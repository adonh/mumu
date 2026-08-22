package config

import (
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"

	derrors "github.com/adonh/mumu/internal/errors"
	"github.com/adonh/mumu/internal/paths"
)

// PinPrecedence controls whether pin rules or saved-layout entries win
// when both would otherwise claim the same currently open window during
// restore (see the window-pinning capability).
type PinPrecedence string

// Valid PinPrecedence values.
const (
	PinPrecedencePin    PinPrecedence = "pin"
	PinPrecedenceLayout PinPrecedence = "layout"
)

// PinRule is a single user-configured app+title-pattern-to-Space pin, matched
// during restore using the same title-similarity heuristic saved-layout
// entries use.
type PinRule struct {
	// BundleID is the pinned application's bundle identifier.
	BundleID string
	// Title is the approximate title pattern matched against that
	// application's currently open window titles.
	Title string
	// Space is the target logical left-to-right Space ordinal (see
	// internal/space's logical numbering).
	Space int
}

// Config holds mumu's user-editable settings, loaded from its config.yaml.
type Config struct {
	// DataDir is the directory mumu uses to store its data (currently
	// just a "layouts" subdirectory holding the saved window-to-Space
	// layouts, one internal JSON file per display count), with any
	// leading "~" already expanded to an absolute path.
	DataDir string
	// Pins maps a connected-display-count to the pin rules configured
	// for it. A display count with no configured pins has no entry.
	Pins map[int][]PinRule
	// PinPrecedence controls pin-vs-saved-layout precedence during
	// restore. Defaults to PinPrecedencePin.
	PinPrecedence PinPrecedence
}

const (
	configFileName = "config.yaml"
	appDirName     = "mumu"
	// appSupportRelPath is the native macOS per-user data location,
	// relative to the home directory, used as the fallback when the
	// corresponding XDG environment variable isn't set.
	appSupportRelPath = "Library/Application Support"

	fileMode = 0o644
	dirMode  = 0o755
)

// pinRuleFileFormat is the on-disk shape of one entry under a config.yaml
// "pins" display-count list.
type pinRuleFileFormat struct {
	BundleID string `yaml:"bundle_id"` //nolint:tagliatelle // Stable user-facing config key name.
	Title    string `yaml:"title"`
	Space    int    `yaml:"space"`
}

// fileFormat is the on-disk shape of config.yaml.
type fileFormat struct {
	DataDir string `yaml:"data_dir"` //nolint:tagliatelle // Stable user-facing config key name.
	// Pins maps a display count to its list of pin rules. Absent or empty
	// means no pins are configured for that display count.
	Pins map[int][]pinRuleFileFormat `yaml:"pins"`
	// PinPrecedence is "pin" or "layout"; empty means the default ("pin").
	PinPrecedence string `yaml:"pin_precedence"` //nolint:tagliatelle // Stable user-facing config key name.
}

// FilePath resolves the path to mumu's configuration file: it follows
// $XDG_CONFIG_HOME if set to a non-empty value, and otherwise falls back to
// ~/Library/Application Support/mumu/config.yaml.
func FilePath() string {
	return filepath.Join(configDir(), configFileName)
}

func configDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appDirName)
	}

	return filepath.Join(paths.ExpandHome("~"), appSupportRelPath, appDirName)
}

// defaultDataDirRaw returns the default data_dir value written into a
// freshly created config file. It uses a literal "~" (left unexpanded) for
// the macOS-fallback case so the generated file stays portable and matches
// what a user would naturally type, but an absolute path when derived from
// an XDG environment variable (which is already absolute by convention).
func defaultDataDirRaw() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, appDirName)
	}

	return filepath.Join("~", appSupportRelPath, appDirName)
}

// Load reads mumu's configuration file, creating it with commented
// defaults if it doesn't yet exist, and returns the resolved settings.
func Load() (*Config, error) {
	path := FilePath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return createDefault(path)
		}

		return nil, derrors.Wrapf(err, derrors.CodeConfigIOFailed, "reading config file %s", path)
	}

	var parsed fileFormat

	err = yaml.Unmarshal(data, &parsed)
	if err != nil {
		return nil, derrors.Wrapf(err, derrors.CodeInvalidConfig, "parsing config file %s", path)
	}

	if parsed.DataDir == "" {
		return nil, derrors.Newf(
			derrors.CodeInvalidConfig,
			"config file %s: data_dir must be a non-empty string",
			path,
		)
	}

	pins, err := validatePins(path, parsed.Pins)
	if err != nil {
		return nil, err
	}

	precedence, err := validatePinPrecedence(path, parsed.PinPrecedence)
	if err != nil {
		return nil, err
	}

	return &Config{
		DataDir:       paths.ExpandHome(parsed.DataDir),
		Pins:          pins,
		PinPrecedence: precedence,
	}, nil
}

// validatePins checks every configured pin rule has a non-empty bundle_id
// and title and a positive space ordinal, returning a clear error naming
// the config file path and the offending entry otherwise.
func validatePins(path string, raw map[int][]pinRuleFileFormat) (map[int][]PinRule, error) {
	if len(raw) == 0 {
		return map[int][]PinRule{}, nil
	}

	pins := make(map[int][]PinRule, len(raw))

	for displayCount, rules := range raw {
		converted := make([]PinRule, 0, len(rules))

		for _, rule := range rules {
			if rule.BundleID == "" {
				return nil, derrors.Newf(
					derrors.CodeInvalidConfig,
					"config file %s: pins for %d display(s): bundle_id must be a non-empty string",
					path,
					displayCount,
				)
			}

			if rule.Title == "" {
				return nil, derrors.Newf(
					derrors.CodeInvalidConfig,
					"config file %s: pins for %d display(s), app %s: title must be a non-empty string",
					path,
					displayCount,
					rule.BundleID,
				)
			}

			if rule.Space <= 0 {
				return nil, derrors.Newf(
					derrors.CodeInvalidConfig,
					"config file %s: pins for %d display(s), app %s: space must be a positive integer, got %d",
					path,
					displayCount,
					rule.BundleID,
					rule.Space,
				)
			}

			converted = append(converted, PinRule(rule))
		}

		pins[displayCount] = converted
	}

	return pins, nil
}

// validatePinPrecedence resolves the raw pin_precedence string to a
// PinPrecedence, defaulting to PinPrecedencePin when absent and reporting a
// clear error for any other value.
func validatePinPrecedence(path, raw string) (PinPrecedence, error) {
	switch raw {
	case "":
		return PinPrecedencePin, nil
	case string(PinPrecedencePin):
		return PinPrecedencePin, nil
	case string(PinPrecedenceLayout):
		return PinPrecedenceLayout, nil
	default:
		return "", derrors.Newf(
			derrors.CodeInvalidConfig,
			"config file %s: pin_precedence must be %q or %q, got %q",
			path,
			PinPrecedencePin,
			PinPrecedenceLayout,
			raw,
		)
	}
}

func createDefault(path string) (*Config, error) {
	dataDirRaw := defaultDataDirRaw()

	err := os.MkdirAll(filepath.Dir(path), dirMode)
	if err != nil {
		return nil, derrors.Wrapf(err, derrors.CodeConfigIOFailed, "creating config directory")
	}

	err = os.WriteFile(path, []byte(defaultConfigYAML(dataDirRaw)), fileMode)
	if err != nil {
		return nil, derrors.Wrapf(err, derrors.CodeConfigIOFailed, "writing default config file")
	}

	return &Config{DataDir: paths.ExpandHome(dataDirRaw), PinPrecedence: PinPrecedencePin}, nil
}

func defaultConfigYAML(dataDir string) string {
	return "# mumu configuration.\n" +
		"#\n" +
		"# data_dir: the directory mumu uses to store its data (currently just\n" +
		"# layouts/, the saved window-to-Space layouts, one internal JSON file\n" +
		"# per display count). Supports a leading \"~\" for your home directory.\n" +
		"# Defaults to $XDG_DATA_HOME/mumu if XDG_DATA_HOME is set, otherwise\n" +
		"# ~/Library/Application Support/mumu.\n" +
		"data_dir: " + dataDir + "\n" +
		"\n" +
		"# pins: fixed application-window-to-Space assignments, applied by\n" +
		"# \"mumu restore\", keyed by the number of connected displays (different\n" +
		"# display counts can declare entirely different pins). Each rule needs\n" +
		"# an app's bundle_id, an approximate title pattern (matched the same way\n" +
		"# restore matches saved layouts), and a target space (mumu's logical\n" +
		"# left-to-right Space number). Absent or empty means no pins.\n" +
		"#\n" +
		"# pins:\n" +
		"#   2:\n" +
		"#     - bundle_id: com.tinyspeck.slackmacgap\n" +
		"#       title: \"Slack\"\n" +
		"#       space: 1\n" +
		"\n" +
		"# pin_precedence: whether pins (\"pin\", the default) or the saved layout\n" +
		"# (\"layout\") wins when both would claim the same open window during\n" +
		"# \"mumu restore\".\n" +
		"#\n" +
		"# pin_precedence: pin\n"
}
