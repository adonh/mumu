package config

import (
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"

	derrors "github.com/adonh/mumu/internal/errors"
	"github.com/adonh/mumu/internal/paths"
)

// Config holds mumu's user-editable settings, loaded from its config.yaml.
type Config struct {
	// DataDir is the directory mumu uses to store its data (currently
	// just a "layouts" subdirectory holding the saved window-to-Space
	// layouts, one internal JSON file per display count), with any
	// leading "~" already expanded to an absolute path.
	DataDir string
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

// fileFormat is the on-disk shape of config.yaml.
type fileFormat struct {
	DataDir string `yaml:"data_dir"` //nolint:tagliatelle // Stable user-facing config key name.
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

	return &Config{DataDir: paths.ExpandHome(parsed.DataDir)}, nil
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

	return &Config{DataDir: paths.ExpandHome(dataDirRaw)}, nil
}

func defaultConfigYAML(dataDir string) string {
	return "# mumu configuration.\n" +
		"#\n" +
		"# data_dir: the directory mumu uses to store its data (currently just\n" +
		"# layouts/, the saved window-to-Space layouts, one internal JSON file\n" +
		"# per display count). Supports a leading \"~\" for your home directory.\n" +
		"# Defaults to $XDG_DATA_HOME/mumu if XDG_DATA_HOME is set, otherwise\n" +
		"# ~/Library/Application Support/mumu.\n" +
		"data_dir: " + dataDir + "\n"
}
