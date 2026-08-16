package domain

import (
	"testing"
	"unsafe"
)

// cacheLineSize is the assumed CPU cache line width (64 bytes on all modern
// x86/ARM/PPC processors).
const cacheLineSize = 64

// TestCloneNodeScalarFieldsInOneCacheLine verifies that all scalar fields of
// CloneNode are grouped together within a single 64-byte cache line window.
//
// The actionability pattern-matching layer reads BaseType, InterfaceMethod,
// and IsAlias on every node during evaluation. Grouping these scalars into one
// cache line (after the pointer/string fields) avoids a cross-line access.
//
// If this test fails, a field was added in the wrong position. Fix it by
// keeping all pointer/string fields before all scalar fields in the struct
// definition. Also update syntaxToCloneNode in printer/clone_processor.go to
// include the new field.
func TestCloneNodeScalarFieldsInOneCacheLine(t *testing.T) {
	t.Parallel()

	type fieldInfo struct {
		name   string
		offset uintptr
		size   uintptr
	}

	scalarFields := []fieldInfo{
		{"BaseType", unsafe.Offsetof(CloneNode{}.BaseType), unsafe.Sizeof(CloneNode{}.BaseType)},
		{
			"EnclosingReturnArity",
			unsafe.Offsetof(CloneNode{}.EnclosingReturnArity),
			unsafe.Sizeof(CloneNode{}.EnclosingReturnArity),
		},
		{"InterfaceMethod", unsafe.Offsetof(CloneNode{}.InterfaceMethod), unsafe.Sizeof(CloneNode{}.InterfaceMethod)},
		{"IsAlias", unsafe.Offsetof(CloneNode{}.IsAlias), unsafe.Sizeof(CloneNode{}.IsAlias)},
	}

	minOffset := scalarFields[0].offset
	maxEnd := scalarFields[0].offset + scalarFields[0].size

	for _, f := range scalarFields {
		if f.offset < minOffset {
			minOffset = f.offset
		}

		end := f.offset + f.size
		if end > maxEnd {
			maxEnd = end
		}
	}

	if minOffset < cacheLineSize {
		t.Errorf(
			"scalar fields start at offset %d, expected >= %d (pointer/string fields should come first)",
			minOffset, cacheLineSize,
		)
	}

	span := maxEnd - minOffset
	if span > cacheLineSize {
		t.Errorf(
			"scalar fields span %d bytes (offset %d to %d), expected <= %d (one cache line)",
			span, minOffset, maxEnd, cacheLineSize,
		)
	}
}

// TestCloneNodeSizeConsistency verifies the CloneNode struct size hasn't grown
// unexpectedly. The current layout is 88 bytes. A change here means a field
// was added, removed, or retyped — review the layout to ensure the cache line
// invariant still holds.
func TestCloneNodeSizeConsistency(t *testing.T) {
	t.Parallel()

	const expectedSize = 88

	actualSize := unsafe.Sizeof(CloneNode{})

	if actualSize != expectedSize {
		t.Errorf(
			"CloneNode size changed: got %d bytes, expected %d — verify scalar fields still fit in one cache line",
			actualSize, expectedSize,
		)
	}
}
