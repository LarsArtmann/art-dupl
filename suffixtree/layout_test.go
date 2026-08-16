package suffixtree

import (
	"testing"
	"unsafe"
)

// TestStateLayout verifies that the state struct stays compact after the
// transition map was replaced by a sorted []tran slice of values. The struct
// is 32 bytes: a 24-byte slice header plus the linkState pointer. The trans
// field (most accessed during search) must stay at offset 0. Transitions
// live BY VALUE in the slice's backing array — one allocation per internal
// state instead of a map hash table, and none at all for leaves (nil slice).
func TestStateLayout(t *testing.T) {
	t.Parallel()

	var s state

	size := unsafe.Sizeof(s)
	if size != 32 {
		t.Errorf("state struct size: got %d, want 32 (slice header + linkState)", size)
	}

	transOffset := unsafe.Offsetof(s.trans)
	if transOffset != 0 {
		t.Errorf("trans field offset: got %d, want 0 (most accessed field should be first)", transOffset)
	}

	linkOffset := unsafe.Offsetof(s.linkState)
	if linkOffset != 24 {
		t.Errorf("linkState field offset: got %d, want 24", linkOffset)
	}
}

// TestTranLayout verifies the tran struct is compact: two int32 Pos fields
// plus one *state pointer pack into 16 bytes with no padding. Four trans fit
// per 64-byte cache line as slice elements.
func TestTranLayout(t *testing.T) {
	t.Parallel()

	var tr tran

	size := unsafe.Sizeof(tr)
	if size != 16 {
		t.Errorf("tran struct size: got %d, want 16 (two int32 + pointer, no padding)", size)
	}
}
