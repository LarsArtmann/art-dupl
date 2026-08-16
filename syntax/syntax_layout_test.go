package syntax

import (
	"testing"
	"unsafe"
)

// cacheLineSize is the assumed CPU cache line width (64 bytes on all modern
// x86/ARM/PPC processors). The layout invariant tested below relies on this.
const cacheLineSize = 64

// TestNodeScalarFieldsInOneCacheLine verifies that all scalar fields of Node
// are grouped together within a single 64-byte cache line window.
//
// This is a regression guard for the field-ordering optimization: pointer and
// string fields (Children, Filename, Name, VarType) are placed first, followed
// by all scalar fields (Type, Pos, End, Owns, Fingerprint, EnclosingReturnArity,
// Statement, InterfaceMethod, IsAlias). The hot-path reads in Val() touch Type,
// Fingerprint, and Statement; the actionability layer reads BaseType (decoded
// from Type), InterfaceMethod, and IsAlias. Grouping these scalars into one
// cache line avoids a cross-line access on every node evaluation.
//
// If this test fails, a new field was likely added in the wrong position (e.g.,
// a string field inserted between scalar fields, or a scalar field moved before
// the pointer/string block). Fix it by keeping all pointer/string fields before
// all scalar fields in the struct definition. Then update serial() and Clone()
// to include the new field — TestSerializePreservesAllFields will catch a
// missing field in serial(), but Clone() has no such guard.
func TestNodeScalarFieldsInOneCacheLine(t *testing.T) {
	t.Parallel()

	type fieldInfo struct {
		name   string
		offset uintptr
		size   uintptr
	}

	scalarFields := []fieldInfo{
		{"Type", unsafe.Offsetof(Node{}.Type), unsafe.Sizeof(Node{}.Type)},
		{"Pos", unsafe.Offsetof(Node{}.Pos), unsafe.Sizeof(Node{}.Pos)},
		{"End", unsafe.Offsetof(Node{}.End), unsafe.Sizeof(Node{}.End)},
		{"Owns", unsafe.Offsetof(Node{}.Owns), unsafe.Sizeof(Node{}.Owns)},
		{"Fingerprint", unsafe.Offsetof(Node{}.Fingerprint), unsafe.Sizeof(Node{}.Fingerprint)},
		{
			"EnclosingReturnArity",
			unsafe.Offsetof(Node{}.EnclosingReturnArity),
			unsafe.Sizeof(Node{}.EnclosingReturnArity),
		},
		{"Statement", unsafe.Offsetof(Node{}.Statement), unsafe.Sizeof(Node{}.Statement)},
		{"InterfaceMethod", unsafe.Offsetof(Node{}.InterfaceMethod), unsafe.Sizeof(Node{}.InterfaceMethod)},
		{"IsAlias", unsafe.Offsetof(Node{}.IsAlias), unsafe.Sizeof(Node{}.IsAlias)},
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

	// All scalar fields must start at or after the first cache line boundary
	// (i.e., after the pointer/string block that occupies the first 64+ bytes).
	if minOffset < cacheLineSize {
		t.Errorf(
			"scalar fields start at offset %d, expected >= %d (pointer/string fields should come first)",
			minOffset, cacheLineSize,
		)
	}

	// All scalar fields must fit within one cache line window.
	span := maxEnd - minOffset
	if span > cacheLineSize {
		t.Errorf(
			"scalar fields span %d bytes (offset %d to %d), expected <= %d (one cache line)",
			span, minOffset, maxEnd, cacheLineSize,
		)
	}

	// Verify Val() hot-path fields are within one cache line.
	valFields := []fieldInfo{
		{"Type", unsafe.Offsetof(Node{}.Type), unsafe.Sizeof(Node{}.Type)},
		{"Fingerprint", unsafe.Offsetof(Node{}.Fingerprint), unsafe.Sizeof(Node{}.Fingerprint)},
		{"Statement", unsafe.Offsetof(Node{}.Statement), unsafe.Sizeof(Node{}.Statement)},
	}

	valMin := valFields[0].offset
	valMax := valFields[0].offset + valFields[0].size

	for _, f := range valFields {
		if f.offset < valMin {
			valMin = f.offset
		}

		end := f.offset + f.size
		if end > valMax {
			valMax = end
		}
	}

	valSpan := valMax - valMin
	if valSpan > cacheLineSize {
		t.Errorf(
			"Val() hot-path fields span %d bytes, expected <= %d",
			valSpan, cacheLineSize,
		)
	}
}

// TestNodeSizeConsistency verifies the Node struct size hasn't grown
// unexpectedly. The current layout is 104 bytes (72 bytes of pointer/string
// fields + 32 bytes of scalar fields including padding). A change here means
// a field was added, removed, or retyped — review the layout to ensure the
// cache line invariant still holds.
func TestNodeSizeConsistency(t *testing.T) {
	t.Parallel()

	const expectedSize = 104

	actualSize := unsafe.Sizeof(Node{})

	if actualSize != expectedSize {
		t.Errorf(
			"Node size changed: got %d bytes, expected %d — verify scalar fields still fit in one cache line",
			actualSize, expectedSize,
		)
	}
}
