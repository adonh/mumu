package layout

import (
	"fmt"
	"strconv"
)

// ProgressFunc receives human-readable status updates during long-running
// operations (Capture, Restore) so callers like the CLI can print
// something while native window enumeration and window moves are in
// flight, rather than appearing to hang. A nil ProgressFunc is valid and
// silently discards updates.
type ProgressFunc func(message string)

// emit calls f with message if f is non-nil.
func (f ProgressFunc) emit(message string) {
	if f != nil {
		f(message)
	}
}

// displayTitle returns title, or a placeholder if it's empty (windows with
// no title, e.g. some utility/panel windows, are captured with "").
func displayTitle(title string) string {
	if title == "" {
		return "(untitled)"
	}

	return title
}

// FormatIndex renders a 1-based progress index like "[03/12]", with both
// numbers zero-padded to the digit width of total so entries line up
// vertically in a list regardless of how many digits total itself has.
func FormatIndex(current, total int) string {
	return FormatIndexWidth(current, total, len(strconv.Itoa(total)))
}

// FormatIndexWidth renders a 1-based progress index like "[03/12]", with
// both numbers zero-padded to width. Unlike FormatIndex, width is supplied
// by the caller rather than derived from total, so multiple lists printed
// one after another (e.g. "mumu show"'s entries, pins, and default-space
// sections) can share a single width and stay vertically aligned with each
// other even though each list has its own total.
func FormatIndexWidth(current, total, width int) string {
	return fmt.Sprintf("[%0*d/%0*d]", width, current, width, total)
}
