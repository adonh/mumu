package permissions_test

import (
	"strings"
	"testing"

	"github.com/adonh/mumu/internal/permissions"
)

func TestWarnings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result permissions.CheckResult
		want   int
	}{
		{
			name:   "both granted produces no warnings",
			result: permissions.CheckResult{Accessibility: true, ScreenRecording: true},
			want:   0,
		},
		{
			name:   "accessibility missing produces one warning",
			result: permissions.CheckResult{Accessibility: false, ScreenRecording: true},
			want:   1,
		},
		{
			name:   "screen recording missing produces one warning",
			result: permissions.CheckResult{Accessibility: true, ScreenRecording: false},
			want:   1,
		},
		{
			name:   "both missing produces two warnings",
			result: permissions.CheckResult{Accessibility: false, ScreenRecording: false},
			want:   2,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := permissions.Warnings(testCase.result)
			if len(got) != testCase.want {
				t.Fatalf("Warnings(%+v) returned %d warning(s), want %d: %v",
					testCase.result, len(got), testCase.want, got)
			}
		})
	}
}

func TestWarnings_MentionsAffectedPermission(t *testing.T) {
	t.Parallel()

	accessibility := permissions.Warnings(
		permissions.CheckResult{Accessibility: false, ScreenRecording: true},
	)
	if len(accessibility) != 1 || !strings.Contains(accessibility[0], "Accessibility") {
		t.Fatalf("expected an Accessibility warning, got %v", accessibility)
	}

	screenRecording := permissions.Warnings(
		permissions.CheckResult{Accessibility: true, ScreenRecording: false},
	)
	if len(screenRecording) != 1 || !strings.Contains(screenRecording[0], "Screen Recording") {
		t.Fatalf("expected a Screen Recording warning, got %v", screenRecording)
	}
}
