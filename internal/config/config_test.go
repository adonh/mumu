package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adonh/mumu/internal/config"
)

func TestFilePath_XDGConfigHomeSet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")

	got := config.FilePath()
	want := filepath.Join("/custom/config", "mumu", "config.yaml")

	if got != want {
		t.Fatalf("FilePath() = %q, want %q", got, want)
	}
}

func TestFilePath_XDGConfigHomeUnset(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	home := t.TempDir()
	t.Setenv("HOME", home)

	got := config.FilePath()
	want := filepath.Join(home, "Library/Application Support", "mumu", "config.yaml")

	if got != want {
		t.Fatalf("FilePath() = %q, want %q", got, want)
	}
}

func TestLoad_CreatesDefaultWhenMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	wantDataDir := filepath.Join(home, "Library/Application Support", "mumu")
	if cfg.DataDir != wantDataDir {
		t.Fatalf("cfg.DataDir = %q, want %q", cfg.DataDir, wantDataDir)
	}

	path := config.FilePath()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected config file to be created at %s: %v", path, err)
	}

	if len(data) == 0 {
		t.Fatalf("expected non-empty default config file content")
	}
}

func TestLoad_DoesNotOverwriteExistingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")

	path := config.FilePath()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	err = os.WriteFile(path, []byte("data_dir: /custom/data\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DataDir != "/custom/data" {
		t.Fatalf("cfg.DataDir = %q, want %q", cfg.DataDir, "/custom/data")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(data) != "data_dir: /custom/data\n" {
		t.Fatalf("existing config file was modified: %q", string(data))
	}
}

func TestLoad_ExpandsTildeInDataDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path := config.FilePath()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	err = os.WriteFile(path, []byte("data_dir: ~/mumu-data\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := filepath.Join(home, "mumu-data")
	if cfg.DataDir != want {
		t.Fatalf("cfg.DataDir = %q, want %q", cfg.DataDir, want)
	}
}

func TestLoad_MalformedYAML(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path := config.FilePath()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	err = os.WriteFile(path, []byte("data_dir: [unterminated\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for malformed YAML")
	}
}

func TestLoad_EmptyDataDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path := config.FilePath()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	err = os.WriteFile(path, []byte("data_dir: \"\"\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for empty data_dir")
	}
}

func TestLoad_ParsesPinsAndPrecedence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path := config.FilePath()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	contents := "data_dir: /custom/data\n" +
		"pin_precedence: layout\n" +
		"pins:\n" +
		"  2:\n" +
		"    - bundle_id: com.tinyspeck.slackmacgap\n" +
		"      title: \"Slack\"\n" +
		"      space: 1\n" +
		"  4:\n" +
		"    - bundle_id: com.google.Chrome\n" +
		"      title: \"GitHub\"\n" +
		"      space: 5\n"

	err = os.WriteFile(path, []byte(contents), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.PinPrecedence != config.PinPrecedenceLayout {
		t.Fatalf("cfg.PinPrecedence = %q, want %q", cfg.PinPrecedence, config.PinPrecedenceLayout)
	}

	want := map[int][]config.PinRule{
		2: {{BundleID: "com.tinyspeck.slackmacgap", Title: "Slack", Space: 1}},
		4: {{BundleID: "com.google.Chrome", Title: "GitHub", Space: 5}},
	}

	if len(cfg.Pins) != len(want) {
		t.Fatalf("cfg.Pins = %+v, want %+v", cfg.Pins, want)
	}

	for displayCount, wantRules := range want {
		gotRules := cfg.Pins[displayCount]
		if len(gotRules) != len(wantRules) || gotRules[0] != wantRules[0] {
			t.Fatalf("cfg.Pins[%d] = %+v, want %+v", displayCount, gotRules, wantRules)
		}
	}
}

func TestLoad_PinsAbsentDefaultsToNoPinsAndPinPrecedence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path := config.FilePath()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	err = os.WriteFile(path, []byte("data_dir: /custom/data\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Pins) != 0 {
		t.Fatalf("cfg.Pins = %+v, want empty", cfg.Pins)
	}

	if cfg.PinPrecedence != config.PinPrecedencePin {
		t.Fatalf("cfg.PinPrecedence = %q, want %q", cfg.PinPrecedence, config.PinPrecedencePin)
	}
}

func TestLoad_InvalidPinMissingBundleID(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path := config.FilePath()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	contents := "data_dir: /custom/data\n" +
		"pins:\n" +
		"  2:\n" +
		"    - title: \"Slack\"\n" +
		"      space: 1\n"

	err = os.WriteFile(path, []byte(contents), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for missing bundle_id")
	}

	if cfg != nil {
		t.Fatalf("Load() returned non-nil Config on error: %+v", cfg)
	}
}

func TestLoad_InvalidPinMissingTitle(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path := config.FilePath()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	contents := "data_dir: /custom/data\n" +
		"pins:\n" +
		"  2:\n" +
		"    - bundle_id: com.tinyspeck.slackmacgap\n" +
		"      space: 1\n"

	err = os.WriteFile(path, []byte(contents), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for missing title")
	}
}

func TestLoad_InvalidPinNonPositiveSpace(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path := config.FilePath()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	contents := "data_dir: /custom/data\n" +
		"pins:\n" +
		"  2:\n" +
		"    - bundle_id: com.tinyspeck.slackmacgap\n" +
		"      title: \"Slack\"\n" +
		"      space: 0\n"

	err = os.WriteFile(path, []byte(contents), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for non-positive space")
	}
}

func TestLoad_InvalidPinPrecedence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path := config.FilePath()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	contents := "data_dir: /custom/data\n" +
		"pin_precedence: bogus\n"

	err = os.WriteFile(path, []byte(contents), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for invalid pin_precedence")
	}
}

func TestLoad_DefaultFileDocumentsPinsAndPrecedence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")

	_, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	data, err := os.ReadFile(config.FilePath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	for _, want := range []string{"pins:", "pin_precedence:", "bundle_id:", "space:"} {
		if !strings.Contains(content, want) {
			t.Fatalf("default config file missing %q in content:\n%s", want, content)
		}
	}
}

func TestLoad_NonStringDataDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path := config.FilePath()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	err = os.WriteFile(path, []byte("data_dir: [1, 2, 3]\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for non-string data_dir")
	}
}
