package layout_test

import (
	"testing"

	"github.com/adonh/mumu/internal/layout"
)

func TestFormatIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current int
		total   int
		want    string
	}{
		{name: "single digit total", current: 3, total: 9, want: "[3/9]"},
		{
			name:    "double digit total pads single digit current",
			current: 3,
			total:   12,
			want:    "[03/12]",
		},
		{name: "double digit current and total", current: 12, total: 12, want: "[12/12]"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := layout.FormatIndex(testCase.current, testCase.total); got != testCase.want {
				t.Fatalf(
					"FormatIndex(%d, %d) = %q, want %q",
					testCase.current,
					testCase.total,
					got,
					testCase.want,
				)
			}
		})
	}
}

func TestFormatIndexWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current int
		total   int
		width   int
		want    string
	}{
		{
			name:    "width matches total's own digit count",
			current: 3,
			total:   9,
			width:   1,
			want:    "[3/9]",
		},
		{
			name:    "width wider than total's own digit count, e.g. to align with a sibling list",
			current: 1,
			total:   1,
			width:   2,
			want:    "[01/01]",
		},
		{
			name:    "width narrower than total's digits doesn't truncate the total",
			current: 3,
			total:   12,
			width:   1,
			want:    "[3/12]",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := layout.FormatIndexWidth(testCase.current, testCase.total, testCase.width)
			if got != testCase.want {
				t.Fatalf(
					"FormatIndexWidth(%d, %d, %d) = %q, want %q",
					testCase.current, testCase.total, testCase.width, got, testCase.want,
				)
			}
		})
	}
}
