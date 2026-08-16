package suffixtree

// linearScanMax is the transition count below which findTran uses a linear
// scan of the sorted slice instead of binary search. Instrumentation on
// 10k-token trees shows 90%+ of states are leaves (0 transitions) and most
// internal states hold 2-10 transitions; for such small n a linear scan of
// cache-adjacent tran values beats hashing.
const linearScanMax = 8

// findTran finds the transition matching the given token value. The slice is
// sorted by key, where a transition's key is derivable as data[tr.start] —
// no separate key field is stored. Linear scan covers the common case
// (<= linearScanMax transitions); binary search covers high-fanout states
// such as the root.
//
// Performance characteristics:
//   - O(linearScanMax) for the common case, O(log n) for large fanout
//   - O(1) additional space
//
// This is called frequently during suffix tree construction and search.
// The returned pointer is only valid until the next addTran on s.
func (s *state) findTran(data []TokenValue, c TokenValue) *tran {
	if len(s.trans) <= linearScanMax {
		for i := range s.trans {
			if data[s.trans[i].start] == c {
				return &s.trans[i]
			}
		}

		return nil
	}

	low, high := 0, len(s.trans)
	for low < high {
		mid := int(uint(low+high) >> 1)
		if data[s.trans[mid].start] < c {
			low = mid + 1
		} else {
			high = mid
		}
	}

	if low < len(s.trans) && data[s.trans[low].start] == c {
		return &s.trans[low]
	}

	return nil
}
