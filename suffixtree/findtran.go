package suffixtree

// findTran finds a transition matching the given token value.
// Uses O(1) map lookup for constant-time performance regardless of transition count.
//
// Performance characteristics:
// - O(1) time complexity
// - O(1) additional space
//
// This is called frequently during suffix tree construction and search.
func (s *state) findTran(c TokenValue) *tran {
	return s.trans[c]
}
