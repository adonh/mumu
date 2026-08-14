package space

/*
#cgo CFLAGS: -x objective-c
#include "../native/mumu.h"
#include <stdlib.h>
*/
import "C"

import "unsafe"

// LogicalCount returns the total number of Mission Control spaces, counted
// in logical left-to-right order. Numerically identical to Count(); only the
// ordering used to reach that total differs. This numbering is scoped to the
// layout save/restore capability and is independent of which display is
// primary — see internal/native/mumu.h for details.
func LogicalCount() int {
	return int(C.MumuLogicalSpaceCount())
}

// LogicalSpaceID returns the macOS space ID at the given 1-based logical
// left-to-right index, or 0 if the index is out of range.
func LogicalSpaceID(logicalIndex int) uint64 {
	return uint64(C.MumuLogicalSpaceID(C.int(logicalIndex)))
}

// LogicalIndexForSpace returns the 1-based logical left-to-right index for a
// given macOS space ID, or 0 if the space ID is not currently known.
func LogicalIndexForSpace(sid uint64) int {
	return int(C.MumuLogicalIndexForSpace(C.uint64_t(sid)))
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
