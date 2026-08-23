package hooks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/adonh/mumu/internal/config"
)

// ProgressFunc receives human-readable status updates as each command
// starts, succeeds, or fails, in the same shape layout.ProgressFunc
// already uses. A nil ProgressFunc is valid and silently discards
// updates.
type ProgressFunc func(message string)

func (f ProgressFunc) emit(message string) {
	if f != nil {
		f(message)
	}
}

// Run executes commands sequentially, one at a time, in the order given.
// A Shell-form command runs through "sh -c"; an Argv-form command runs
// directly, with no shell involved. Each command's standard output and
// standard error stream live to stdout/stderr as it runs. A command that
// exits non-zero or fails to start is reported via progress and does not
// stop the remaining commands.
func Run(commands []config.Command, stdout, stderr io.Writer, progress ProgressFunc) {
	for _, command := range commands {
		runOne(command, stdout, stderr, progress)
	}
}

func runOne(command config.Command, stdout, stderr io.Writer, progress ProgressFunc) {
	label := Describe(command)
	progress.emit("Running: " + label)

	cmd := buildCmd(command)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			progress.emit(fmt.Sprintf("✗ command failed (exit %d): %s", exitErr.ExitCode(), label))
		} else {
			progress.emit(fmt.Sprintf("✗ command failed to start (%s): %s", err, label))
		}

		return
	}

	progress.emit("✓ " + label)
}

// buildCmd constructs the *exec.Cmd for command, using "sh -c" for a
// Shell-form command and direct argv execution otherwise. There's no
// per-command timeout by design (see design.md's Non-Goals) — hooks run
// to completion, so context.Background() is used rather than a
// cancellable context.
func buildCmd(command config.Command) *exec.Cmd {
	if command.Shell != "" {
		return exec.CommandContext(context.Background(), "sh", "-c", command.Shell)
	}

	return exec.CommandContext(context.Background(), command.Argv[0], command.Argv[1:]...)
}

// Describe renders command as written, for progress messages and preview
// output: the shell string, or the argv list space-joined.
func Describe(command config.Command) string {
	if command.Shell != "" {
		return command.Shell
	}

	return strings.Join(command.Argv, " ")
}
