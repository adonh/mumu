package hooks_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/adonh/mumu/internal/config"
	"github.com/adonh/mumu/internal/hooks"
)

func TestRun_SuccessfulCommandsRunInOrder(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	commands := []config.Command{
		{Shell: "echo first"},
		{Shell: "echo second"},
	}

	hooks.Run(commands, &stdout, &stdout, nil)

	got := stdout.String()

	firstIdx := strings.Index(got, "first")
	secondIdx := strings.Index(got, "second")

	if firstIdx == -1 || secondIdx == -1 || firstIdx > secondIdx {
		t.Fatalf("stdout = %q, want %q before %q", got, "first", "second")
	}
}

func TestRun_FailingCommandDoesNotBlockSubsequent(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	var progressed []string

	commands := []config.Command{
		{Argv: []string{"false"}},
		{Shell: "echo after-failure"},
	}

	hooks.Run(commands, &stdout, &stdout, func(msg string) { progressed = append(progressed, msg) })

	if !strings.Contains(stdout.String(), "after-failure") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout.String(), "after-failure")
	}

	foundFailure := false

	for _, msg := range progressed {
		if strings.Contains(msg, "failed") {
			foundFailure = true
		}
	}

	if !foundFailure {
		t.Fatalf("progress messages = %v, want a failure to be reported", progressed)
	}
}

func TestRun_ShellFormExecutesThroughShell(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	hooks.Run([]config.Command{{Shell: "echo shell-form"}}, &stdout, &stdout, nil)

	if !strings.Contains(stdout.String(), "shell-form") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout.String(), "shell-form")
	}
}

func TestRun_ArgvFormExecutesDirectly(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	hooks.Run([]config.Command{{Argv: []string{"echo", "argv-form"}}}, &stdout, &stdout, nil)

	if !strings.Contains(stdout.String(), "argv-form") {
		t.Fatalf("stdout = %q, want it to contain %q", stdout.String(), "argv-form")
	}
}

func TestRun_TrueCommandReportsSuccess(t *testing.T) {
	t.Parallel()

	var progressed []string

	hooks.Run(
		[]config.Command{{Argv: []string{"true"}}},
		&bytes.Buffer{},
		&bytes.Buffer{},
		func(msg string) { progressed = append(progressed, msg) },
	)

	foundSuccess := false

	for _, msg := range progressed {
		if strings.HasPrefix(msg, "✓") {
			foundSuccess = true
		}
	}

	if !foundSuccess {
		t.Fatalf("progress messages = %v, want a success marker", progressed)
	}
}

const testDescribeWant = "echo hi"

func TestDescribe_ShellCommand(t *testing.T) {
	t.Parallel()

	got := hooks.Describe(config.Command{Shell: testDescribeWant})
	if got != testDescribeWant {
		t.Fatalf("Describe() = %q, want %q", got, testDescribeWant)
	}
}

func TestDescribe_ArgvCommand(t *testing.T) {
	t.Parallel()

	got := hooks.Describe(config.Command{Argv: []string{"echo", "hi"}})
	if got != testDescribeWant {
		t.Fatalf("Describe() = %q, want %q", got, testDescribeWant)
	}
}
