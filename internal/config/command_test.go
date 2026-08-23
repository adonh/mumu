package config_test

import (
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/adonh/mumu/internal/config"
)

func TestCommand_UnmarshalYAML_ShellString(t *testing.T) {
	t.Parallel()

	var cmds []config.Command

	err := yaml.Unmarshal([]byte("- echo hi\n"), &cmds)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(cmds) != 1 || cmds[0].Shell != "echo hi" || cmds[0].Argv != nil {
		t.Fatalf("cmds = %+v, want single shell command %q", cmds, "echo hi")
	}
}

func TestCommand_UnmarshalYAML_ArgvList(t *testing.T) {
	t.Parallel()

	var cmds []config.Command

	err := yaml.Unmarshal([]byte("- [echo, hi]\n"), &cmds)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := []string{"echo", "hi"}
	if len(cmds) != 1 || cmds[0].Shell != "" || len(cmds[0].Argv) != len(want) {
		t.Fatalf("cmds = %+v, want single argv command %v", cmds, want)
	}

	for i, arg := range want {
		if cmds[0].Argv[i] != arg {
			t.Fatalf("cmds[0].Argv = %v, want %v", cmds[0].Argv, want)
		}
	}
}

func TestCommand_UnmarshalYAML_MixedArray(t *testing.T) {
	t.Parallel()

	var cmds []config.Command

	err := yaml.Unmarshal([]byte("- echo hi\n- [echo, bye]\n"), &cmds)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(cmds) != 2 {
		t.Fatalf("cmds = %+v, want 2 entries", cmds)
	}

	if cmds[0].Shell != "echo hi" {
		t.Fatalf("cmds[0].Shell = %q, want %q", cmds[0].Shell, "echo hi")
	}

	if len(cmds[1].Argv) != 2 || cmds[1].Argv[0] != "echo" || cmds[1].Argv[1] != "bye" {
		t.Fatalf("cmds[1].Argv = %v, want [echo bye]", cmds[1].Argv)
	}
}

func TestCommand_UnmarshalYAML_EmptyString(t *testing.T) {
	t.Parallel()

	var cmds []config.Command

	err := yaml.Unmarshal([]byte("- \"\"\n"), &cmds)
	if err == nil {
		t.Fatal("Unmarshal() error = nil, want error for empty string command")
	}
}

func TestCommand_UnmarshalYAML_EmptyList(t *testing.T) {
	t.Parallel()

	var cmds []config.Command

	err := yaml.Unmarshal([]byte("- []\n"), &cmds)
	if err == nil {
		t.Fatal("Unmarshal() error = nil, want error for empty list command")
	}
}

func TestCommand_UnmarshalYAML_ListWithEmptyString(t *testing.T) {
	t.Parallel()

	var cmds []config.Command

	err := yaml.Unmarshal([]byte("- [echo, \"\"]\n"), &cmds)
	if err == nil {
		t.Fatal("Unmarshal() error = nil, want error for list containing an empty string")
	}
}

func TestCommand_UnmarshalYAML_UnsupportedNodeKind(t *testing.T) {
	t.Parallel()

	var cmds []config.Command

	err := yaml.Unmarshal([]byte("- key: value\n"), &cmds)
	if err == nil {
		t.Fatal("Unmarshal() error = nil, want error for mapping node")
	}
}
