package layout

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
