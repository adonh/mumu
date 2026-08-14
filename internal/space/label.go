package space

import (
	"fmt"
)

// DualLabel renders a logical Space ordinal (see LogicalSpaceID) alongside
// the macOS Mission Control Space number it currently corresponds to — the
// same ordinal macOS's own "Switch to Desktop <n>" keyboard shortcut uses —
// e.g. "#03 (space 21)". Both numbers are zero-padded to 2 digits so entries
// line up in a list. The two numbers can differ whenever the primary
// display isn't the leftmost one. Resolved fresh against the current
// display arrangement; falls back to "(space unknown)" if the logical
// ordinal no longer maps to an existing Space.
func DualLabel(logicalOrdinal int) string {
	missionControl := "unknown"

	if sid := LogicalSpaceID(logicalOrdinal); sid != 0 {
		if idx := MissionControlIndexForSpace(sid); idx != 0 {
			missionControl = fmt.Sprintf("%02d", idx)
		}
	}

	return fmt.Sprintf("#%02d (space %s)", logicalOrdinal, missionControl)
}
