package space

/*
#cgo CFLAGS: -x objective-c
#include "../native/mumu.h"
#include <stdlib.h>
*/
import "C"

import "unsafe"

// LogicalDisplayCount returns the number of currently connected displays,
// in left-to-right order (i.e. the valid range for an Ordinal's Display
// field is 1..this value). This numbering is scoped to the layout
// save/restore capability and is independent of which display is
// primary — see internal/native/mumu.h for details.
func LogicalDisplayCount() int {
	return int(C.MumuLogicalDisplayCount())
}

// IDForOrdinal returns the macOS space ID at the given logical
// Ordinal, or 0 if either the display or the space-within-display part is
// out of range.
func IDForOrdinal(o Ordinal) uint64 {
	return uint64(C.MumuOrdinalSpaceID(C.int(o.Display), C.int(o.Space)))
}

// OrdinalForSpace returns the logical Ordinal for a given macOS space ID,
// or the zero Ordinal if the space ID is not currently known.
func OrdinalForSpace(sid uint64) Ordinal {
	var display, spaceIdx C.int

	if C.MumuSpaceOrdinal(C.uint64_t(sid), &display, &spaceIdx) == 0 {
		return Ordinal{}
	}

	return Ordinal{Display: int(display), Space: int(spaceIdx)}
}

// LeftToRightSpaceCounts returns the per-display space-count sequence in
// left-to-right order (e.g. [3, 5, 2] for three displays with 3, 5, and 2
// spaces respectively, sorted left to right). Used to detect display
// arrangement drift between a layout save and a later restore.
func LeftToRightSpaceCounts() []int {
	var cCount C.int

	ptr := C.MumuLeftToRightSpaceCounts(&cCount)
	if ptr == nil || cCount == 0 {
		return []int{}
	}
	defer C.free(unsafe.Pointer(ptr)) //nolint:nlreturn

	count := int(cCount)
	cSlice := (*[1 << 20]C.int)(unsafe.Pointer(ptr))[:count:count]

	result := make([]int, count)
	for i, v := range cSlice {
		result[i] = int(v)
	}

	return result
}
