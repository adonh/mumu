package space

import (
	"fmt"
)

// DualLabel renders a logical Space Ordinal (see IDForOrdinal)
// alongside the macOS Mission Control Space number it currently
// corresponds to — the same ordinal macOS's own "Switch to Desktop <n>"
// keyboard shortcut uses — e.g. "#2:01 (space 21)". The Mission Control
// number is zero-padded to 2 digits so entries line up in a list. The two
// numbers can differ whenever the primary display isn't the leftmost one.
// Resolved fresh against the current display arrangement; falls back to
// "(space unknown)" if the Ordinal no longer maps to an existing Space.
func DualLabel(ordinal Ordinal) string {
	missionControl := "unknown"

	if sid := IDForOrdinal(ordinal); sid != 0 {
		if idx := MissionControlIndexForSpace(sid); idx != 0 {
			missionControl = fmt.Sprintf("%02d", idx)
		}
	}

	return fmt.Sprintf("#%s (space %s)", ordinal, missionControl)
}
