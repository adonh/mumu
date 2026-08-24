package space

import (
	"fmt"
	"strconv"
	"strings"

	derrors "github.com/adonh/mumu/internal/errors"
)

// Ordinal is mumu's own logical Space ordinal: a display's 1-based
// left-to-right position among connected displays, paired with a Space's
// 1-based position within that display's own native ordering (see
// internal/native/mumu.h's "Logical (Two-Part Display + Space) Numbering"
// section). Scoping the Space part to its own display means adding or
// removing a Space on one display never changes the Ordinal of a Space on
// a different display.
//
// The zero value, Ordinal{}, means "unresolved" everywhere this package
// used a bare 0 for that purpose — both fields are always positive for
// any Space that currently exists, so it's a safe sentinel.
type Ordinal struct {
	// Display is the 1-based left-to-right position of the display this
	// Space belongs to, among all currently connected displays.
	Display int `json:"display"`
	// Space is the 1-based position of this Space within its display's
	// own native Space ordering, independent of any other display's
	// Space count.
	Space int `json:"space"`
}

// Less reports whether o should sort before other: by Display first, then
// by Space. This reproduces the left-to-right ordering the previous flat
// ordinal already had, since a flat ordinal was already every Display
// 1 Space, then every Display 2 Space, and so on.
func (o Ordinal) Less(other Ordinal) bool {
	if o.Display != other.Display {
		return o.Display < other.Display
	}

	return o.Space < other.Space
}

// String renders o as "D:SS" — display unpadded, Space zero-padded to 2
// digits — e.g. "2:01" for the 1st Space on the 2nd display from the left.
func (o Ordinal) String() string {
	return fmt.Sprintf("%d:%02d", o.Display, o.Space)
}

// ParseOrdinal parses a "<display>:<space>" string (e.g. "2:1" or,
// equivalently, "02:01") into an Ordinal, requiring both parts to be
// positive integers. A bare integer (the pre-two-part-ordinal format) is
// rejected with an error explicitly naming the expected format, since a
// flat ordinal can't be safely reinterpreted as a display+space pair
// without knowing the arrangement it was originally captured against.
func ParseOrdinal(raw string) (Ordinal, error) {
	parts := strings.SplitN(raw, ":", 2) //nolint:mnd // Exactly two parts: display and space.

	if len(parts) != 2 { //nolint:mnd // Exactly two parts: display and space.
		return Ordinal{}, derrors.Newf(
			derrors.CodeInvalidInput,
			`ordinal %q must be in "<display>:<space>" format (e.g. "2:1")`,
			raw,
		)
	}

	display, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || display <= 0 {
		return Ordinal{}, derrors.Newf(
			derrors.CodeInvalidInput,
			`ordinal %q must be in "<display>:<space>" format with a positive display number`,
			raw,
		)
	}

	space, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || space <= 0 {
		return Ordinal{}, derrors.Newf(
			derrors.CodeInvalidInput,
			`ordinal %q must be in "<display>:<space>" format with a positive space number`,
			raw,
		)
	}

	return Ordinal{Display: display, Space: space}, nil
}
