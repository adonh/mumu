package config_test

import (
	"os"
	"path/filepath"
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
