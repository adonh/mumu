package cmd //nolint:testpackage // Tests unexported restore summary rendering.

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/adonh/mumu/internal/config"
	"github.com/adonh/mumu/internal/layout"
)

const (
	testShellOffCommand = "echo off-command"
	testShellOnCommand  = "echo on-command"
	testFlagDefaultOff  = "false"
	testCommandFalse    = "false"
	testGlobalOff       = "global-off"
	testGlobalOn        = "global-on"
	testLayoutOff       = "layout-off"
	testLayoutOn        = "layout-on"
	testSlackBundleID   = "com.tinyspeck.slackmacgap"
)

var errTestRestoreFailed = errors.New("restore failed")

func TestLayoutRestoreCmd_NoHooksFlagRegistered(t *testing.T) {
	t.Parallel()

	flag := layoutRestoreCmd.Flags().Lookup("no-hooks")
	if flag == nil {
		t.Fatal("layoutRestoreCmd has no --no-hooks flag registered")
	}

	if flag.DefValue != testFlagDefaultOff {
		t.Fatalf("--no-hooks default = %q, want %q", flag.DefValue, testFlagDefaultOff)
	}
}

func TestRunRestoreWithHooks_OffRunsBeforeAndOnRunsAfterRestore(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	cfg := &config.Config{
		Hooks: config.Hooks{
			Off: []config.Command{{Shell: testShellOffCommand}},
			On:  []config.Command{{Shell: testShellOnCommand}},
		},
	}

	restoreCalled := false

	_, err := runRestoreWithHooks(cmd, cfg, 2, false, func() (layout.RestoreSummary, error) {
		restoreCalled = true

		cmd.Println("restore-progress")

		return layout.RestoreSummary{Moved: 1}, nil
	})
	if err != nil {
		t.Fatalf("runRestoreWithHooks() error = %v", err)
	}

	if !restoreCalled {
		t.Fatal("restore was never called")
	}

	got := output.String()

	offIdx := strings.Index(got, "off-command")
	restoreIdx := strings.Index(got, "restore-progress")
	onIdx := strings.Index(got, "on-command")

	if offIdx == -1 || restoreIdx == -1 || onIdx == -1 ||
		(offIdx >= restoreIdx || restoreIdx >= onIdx) {
		t.Fatalf("output = %q, want off-command before restore-progress before on-command", got)
	}
}

func TestRunRestoreWithHooks_NoHooksSuppressesBoth(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	cfg := &config.Config{
		Hooks: config.Hooks{
			Off: []config.Command{{Shell: testShellOffCommand}},
			On:  []config.Command{{Shell: testShellOnCommand}},
		},
	}

	_, err := runRestoreWithHooks(cmd, cfg, 2, true, func() (layout.RestoreSummary, error) {
		return layout.RestoreSummary{Moved: 1}, nil
	})
	if err != nil {
		t.Fatalf("runRestoreWithHooks() error = %v", err)
	}

	got := output.String()
	if strings.Contains(got, "off-command") || strings.Contains(got, "on-command") {
		t.Fatalf("output = %q, want no hook commands to run with noHooks=true", got)
	}
}

func TestRunRestoreWithHooks_FailingOffCommandDoesNotPreventRestore(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	cfg := &config.Config{
		Hooks: config.Hooks{
			Off: []config.Command{{Argv: []string{testCommandFalse}}},
		},
	}

	restoreCalled := false

	_, err := runRestoreWithHooks(cmd, cfg, 2, false, func() (layout.RestoreSummary, error) {
		restoreCalled = true

		return layout.RestoreSummary{Moved: 1}, nil
	})
	if err != nil {
		t.Fatalf("runRestoreWithHooks() error = %v", err)
	}

	if !restoreCalled {
		t.Fatal("restore was not called after a failing off command")
	}
}

func TestRunRestoreWithHooks_FailingOnCommandDoesNotChangeRestoreResult(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	cfg := &config.Config{
		Hooks: config.Hooks{
			On: []config.Command{{Argv: []string{testCommandFalse}}},
		},
	}

	summary, err := runRestoreWithHooks(cmd, cfg, 2, false, func() (layout.RestoreSummary, error) {
		return layout.RestoreSummary{Moved: 3}, nil
	})
	if err != nil {
		t.Fatalf("runRestoreWithHooks() error = %v", err)
	}

	if summary.Moved != 3 {
		t.Fatalf("summary.Moved = %d, want 3 (unaffected by failing on command)", summary.Moved)
	}
}

func TestRunRestoreWithHooks_RestoreErrorSkipsOnCommands(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	cfg := &config.Config{
		Hooks: config.Hooks{
			On: []config.Command{{Shell: testShellOnCommand}},
		},
	}

	wantErr := errTestRestoreFailed

	_, err := runRestoreWithHooks(cmd, cfg, 2, false, func() (layout.RestoreSummary, error) {
		return layout.RestoreSummary{}, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("runRestoreWithHooks() error = %v, want %v", err, wantErr)
	}

	if strings.Contains(output.String(), "on-command") {
		t.Fatalf("output = %q, want on commands skipped when restore fails", output.String())
	}
}

func TestPrintRestoreSummaryMarksFallbackFailures(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)

	printRestoreSummary(
		cmd,
		layout.RestoreSummary{
			Skipped: []layout.SkippedEntry{
				{
					Entry: layout.Entry{
						BundleID: "com.example.chrome",
						Title:    "window",
						Ordinal:  1,
					},
					Reason:   layout.SkipMoveFailed,
					Fallback: true,
				},
			},
		},
		layout.SortByDisplay,
	)

	if !strings.Contains(output.String(), "(fallback)") {
		t.Fatalf("restore summary = %q, want fallback marker", output.String())
	}
}

func TestPrintRestoreSummaryMarksFuzzyMatches(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)

	printRestoreSummary(
		cmd,
		layout.RestoreSummary{
			Skipped: []layout.SkippedEntry{
				{
					Entry: layout.Entry{
						BundleID: "com.example.chrome",
						Title:    "window",
						Ordinal:  1,
					},
					Reason: layout.SkipMoveFailed,
					Fuzzy:  true,
				},
			},
		},
		layout.SortByDisplay,
	)

	if !strings.Contains(output.String(), "(fuzzy)") {
		t.Fatalf("restore summary = %q, want fuzzy marker", output.String())
	}
}

func TestPrintConfiguredPins_ListsEachRule(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)

	printConfiguredPins(cmd, []config.PinRule{
		{BundleID: testSlackBundleID, Title: "general", Space: 1},
	})

	got := output.String()
	for _, want := range []string{"1 configured pin(s)", testSlackBundleID, "general"} {
		if !strings.Contains(got, want) {
			t.Fatalf("printConfiguredPins output = %q, want it to contain %q", got, want)
		}
	}
}

func TestPrintConfiguredDefaultSpaces_ListsEachRule(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)

	printConfiguredDefaultSpaces(cmd, []config.DefaultSpaceRule{
		{BundleID: testSlackBundleID, Space: 1},
	})

	got := output.String()
	for _, want := range []string{"1 configured default space(s)", testSlackBundleID} {
		if !strings.Contains(got, want) {
			t.Fatalf("printConfiguredDefaultSpaces output = %q, want it to contain %q", got, want)
		}
	}
}

func TestPrintConfiguredDefaultSpaces_NoneConfiguredPrintsNothing(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)

	printConfiguredDefaultSpaces(cmd, nil)

	if output.Len() != 0 {
		t.Fatalf(
			"printConfiguredDefaultSpaces output = %q, want empty output when none are configured",
			output.String(),
		)
	}
}

func TestResolveHooks_OnlyGlobalConfigured(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Hooks: config.Hooks{
			Off: []config.Command{{Shell: testGlobalOff}},
			On:  []config.Command{{Shell: testGlobalOn}},
		},
	}

	offCmds, onCmds := resolveHooks(cfg, 2)

	if len(offCmds) != 1 || offCmds[0].Shell != testGlobalOff {
		t.Fatalf("off = %+v, want [global-off]", offCmds)
	}

	if len(onCmds) != 1 || onCmds[0].Shell != testGlobalOn {
		t.Fatalf("on = %+v, want [global-on]", onCmds)
	}
}

func TestResolveHooks_OnlyPerDisplayCountConfigured(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		LayoutHooks: map[int]config.Hooks{
			2: {
				Off: []config.Command{{Shell: testLayoutOff}},
				On:  []config.Command{{Shell: testLayoutOn}},
			},
		},
	}

	offCmds, onCmds := resolveHooks(cfg, 2)

	if len(offCmds) != 1 || offCmds[0].Shell != testLayoutOff {
		t.Fatalf("off = %+v, want [layout-off]", offCmds)
	}

	if len(onCmds) != 1 || onCmds[0].Shell != testLayoutOn {
		t.Fatalf("on = %+v, want [layout-on]", onCmds)
	}

	otherOff, otherOn := resolveHooks(cfg, 3)
	if len(otherOff) != 0 || len(otherOn) != 0 {
		t.Fatalf(
			"resolveHooks(cfg, 3) = %v, %v, want empty for unconfigured display count",
			otherOff,
			otherOn,
		)
	}
}

func TestResolveHooks_BothConfigured(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Hooks: config.Hooks{
			Off: []config.Command{{Shell: testGlobalOff}},
			On:  []config.Command{{Shell: testGlobalOn}},
		},
		LayoutHooks: map[int]config.Hooks{
			2: {
				Off: []config.Command{{Shell: testLayoutOff}},
				On:  []config.Command{{Shell: testLayoutOn}},
			},
		},
	}

	offCmds, onCmds := resolveHooks(cfg, 2)

	if len(offCmds) != 2 || offCmds[0].Shell != testGlobalOff || offCmds[1].Shell != testLayoutOff {
		t.Fatalf("off = %+v, want [global-off layout-off]", offCmds)
	}

	if len(onCmds) != 2 || onCmds[0].Shell != testLayoutOn || onCmds[1].Shell != testGlobalOn {
		t.Fatalf("on = %+v, want [layout-on global-on]", onCmds)
	}
}

func TestResolveHooks_NeitherConfigured(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	offCmds, onCmds := resolveHooks(cfg, 2)

	if len(offCmds) != 0 || len(onCmds) != 0 {
		t.Fatalf("off, on = %v, %v, want both empty", offCmds, onCmds)
	}
}

func TestPrintConfiguredHooks_ListsOffAndOnCommands(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)

	printConfiguredHooks(
		cmd,
		[]config.Command{{Shell: testGlobalOff}, {Argv: []string{testLayoutOff, "arg"}}},
		[]config.Command{{Shell: testLayoutOn}, {Shell: testGlobalOn}},
	)

	got := output.String()
	for _, want := range []string{testGlobalOff, "layout-off arg", testLayoutOn, testGlobalOn} {
		if !strings.Contains(got, want) {
			t.Fatalf("printConfiguredHooks output = %q, want it to contain %q", got, want)
		}
	}
}

func TestPrintConfiguredHooks_NoneConfiguredPrintsNothing(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)

	printConfiguredHooks(cmd, nil, nil)

	if output.Len() != 0 {
		t.Fatalf(
			"printConfiguredHooks output = %q, want empty output when no hooks are configured",
			output.String(),
		)
	}
}

func TestPrintConfiguredPins_NoPinsPrintsNothing(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var output bytes.Buffer
	cmd.SetOut(&output)

	printConfiguredPins(cmd, nil)

	if output.Len() != 0 {
		t.Fatalf(
			"printConfiguredPins output = %q, want empty output when no pins are configured",
			output.String(),
		)
	}
}
