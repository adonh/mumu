package space_test

import (
	"testing"

	"github.com/adonh/mumu/internal/space"
)

func TestOrdinal_Less(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		a, b space.Ordinal
		want bool
	}{
		{
			name: "lower display sorts first regardless of space",
			a:    space.Ordinal{Display: 1, Space: 9},
			b:    space.Ordinal{Display: 2, Space: 1},
			want: true,
		},
		{
			name: "higher display does not sort first",
			a:    space.Ordinal{Display: 2, Space: 1},
			b:    space.Ordinal{Display: 1, Space: 9},
			want: false,
		},
		{
			name: "same display, lower space sorts first",
			a:    space.Ordinal{Display: 1, Space: 1},
			b:    space.Ordinal{Display: 1, Space: 2},
			want: true,
		},
		{
			name: "equal ordinals are not less than each other",
			a:    space.Ordinal{Display: 1, Space: 1},
			b:    space.Ordinal{Display: 1, Space: 1},
			want: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.a.Less(tt.b); got != tt.want {
				t.Fatalf("(%v).Less(%v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestOrdinal_String(t *testing.T) {
	t.Parallel()

	cases := []struct {
		ordinal space.Ordinal
		want    string
	}{
		{ordinal: space.Ordinal{Display: 2, Space: 1}, want: "2:01"},
		{ordinal: space.Ordinal{Display: 1, Space: 21}, want: "1:21"},
		{ordinal: space.Ordinal{Display: 10, Space: 3}, want: "10:03"},
	}

	for _, tt := range cases {
		if got := tt.ordinal.String(); got != tt.want {
			t.Fatalf("(%v).String() = %q, want %q", tt.ordinal, got, tt.want)
		}
	}
}

func TestParseOrdinal_Valid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw  string
		want space.Ordinal
	}{
		{raw: "2:1", want: space.Ordinal{Display: 2, Space: 1}},
		{raw: "02:01", want: space.Ordinal{Display: 2, Space: 1}},
		{raw: "2:01", want: space.Ordinal{Display: 2, Space: 1}},
		{raw: "1:21", want: space.Ordinal{Display: 1, Space: 21}},
	}

	for _, testCase := range cases {
		got, err := space.ParseOrdinal(testCase.raw)
		if err != nil {
			t.Fatalf("ParseOrdinal(%q) error = %v, want nil", testCase.raw, err)
		}

		if got != testCase.want {
			t.Fatalf("ParseOrdinal(%q) = %v, want %v", testCase.raw, got, testCase.want)
		}
	}
}

func TestParseOrdinal_Invalid(t *testing.T) {
	t.Parallel()

	invalid := []string{
		"1",
		"5",
		"",
		"2",
		"2:",
		":1",
		"a:1",
		"2:b",
		"0:1",
		"2:0",
		"-1:1",
		"2:-1",
	}

	for _, raw := range invalid {
		_, err := space.ParseOrdinal(raw)
		if err == nil {
			t.Fatalf("ParseOrdinal(%q) error = nil, want an error", raw)
		}
	}
}
