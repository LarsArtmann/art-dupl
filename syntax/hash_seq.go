package syntax

import (
	"encoding/binary"
	"sync"

	"github.com/LarsArtmann/art-dupl/pkg/format"
	"github.com/zeebo/xxh3"
)

// hashPool is a sync.Pool for reusing byte slices in hashSeq operations.
// This reduces memory allocations for hash operations.

var hashPool = sync.Pool{
	New: func() any {
		// Pre-allocate for common sizes (up to 10,000 nodes * 4 bytes = 40KB)
		buf := make([]byte, 0, 40_000)

		return &buf
	},
}

// hashSeq computes an XXH3 hash of a sequence of nodes.
// This is the main hashing operation in the duplicate detection pipeline.
//
// Each node contributes 4 bytes (full int32 Type) to the hash, preserving
// semantic encoding (identifier hashes, operator hashes) that would be
// lost if only the lower 8 bits were used.
//
// Performance characteristics:
//   - O(n) time complexity where n is the number of nodes
//   - O(1) additional space (reuses pooled buffers)
//   - Uses SIMD-optimized XXH3 hashing (~20x faster than SHA-256)
func hashSeq(nodes []*Node) string {
	if len(nodes) == 0 {
		return ""
	}

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

	// Each node contributes 4 bytes (full int32 Type)
	needed := len(nodes) * 4

	// Ensure buffer has sufficient capacity
	if cap(buf) < needed {
		buf = make([]byte, needed)
	} else {
		buf = buf[:needed]
	}

	for i, node := range nodes {
		//nolint:gosec // G115: node.Type bounded by int32
		binary.LittleEndian.PutUint32(buf[i*4:], uint32(node.Type))
	}

	hash := xxh3.Hash(buf)

	return format.Hash(hash)
}
