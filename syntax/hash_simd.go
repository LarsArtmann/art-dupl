package syntax

import (
	"sync"

	"github.com/LarsArtmann/art-dupl/pkg/format"
	"github.com/zeebo/xxh3"
)

// hashPool is a sync.Pool for reusing byte slices in hashSeq operations.
// This reduces memory allocations for hash operations.
//

var hashPool = sync.Pool{
	New: func() any {
		// Pre-allocate for common sizes (up to 10,000 nodes)
		buf := make([]byte, 0, 10_000)

		return &buf
	},
}

// hashSeq computes an XXH3 hash of a sequence of nodes.
// This is the main hashing operation in the duplicate detection pipeline.
//
// Performance characteristics:
// - O(n) time complexity where n is the number of nodes
// - O(1) additional space (reuses pooled buffers)
// - Uses SIMD-optimized XXH3 hashing (~20x faster than SHA-256)
//
// This function is called frequently during clone detection and has been
// identified as a primary performance bottleneck (96.33% of allocations).
func hashSeq(nodes []*Node) string {
	if len(nodes) == 0 {
		return ""
	}

	// Prepare byte array for hashing
	// Extract byte slice from pool to reduce allocations
	rawBuf := hashPool.Get()

	bufPtr, ok := rawBuf.(*[]byte)
	if !ok {
		return ""
	}

	buf := *bufPtr

	defer func() {
		*bufPtr = buf[:0]
		hashPool.Put(bufPtr)
	}()

	// Ensure buffer has sufficient capacity
	if cap(buf) < len(nodes) {
		buf = make([]byte, len(nodes))
	} else {
		buf = buf[:len(nodes)]
	}

	for i, node := range nodes {
		//nolint:gosec // G115: Safe - node.Type is NodeType (uint8), always fits in byte
		buf[i] = byte(node.Type)
	}

	hash := xxh3.Hash(buf)

	return format.Hash(hash)
}
