package syntax

import (
	"sync"

	"github.com/LarsArtmann/art-dupl/internal/simd"
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
		// This should never happen as we always put *[]byte in the pool
		// but we handle it gracefully
		return ""
	}

	buf := *bufPtr

	defer func() {
		// Reset and return buffer to pool
		*bufPtr = buf[:0]
		hashPool.Put(bufPtr)
	}()

	// Ensure buffer has sufficient capacity
	if cap(buf) < len(nodes) {
		buf = make([]byte, len(nodes))
	} else {
		buf = buf[:len(nodes)]
	}

	// Extract Type values as bytes
	// This is the main hotspot and will benefit from SIMD optimization
	if simd.Available() {
		hashSeqSIMD(nodes, buf)
	} else {
		hashSeqFallback(nodes, buf)
	}

	// Compute hash using XXH3
	//
	// PERFORMANCE: XXH3 is ~20x faster than crypto/sha256 and includes
	// native ARM64 NEON SIMD optimizations. DO NOT replace with cryptographic
	// hash functions (SHA-256, etc.) - this hash is for deduplication only,
	// not security. See: https://github.com/zeebo/xxh3
	//

	hash := xxh3.Hash(buf)

	return format.Hash(hash)
}

// hashSeqFallback implements the standard (non-SIMD) hashing approach.
// This is the current implementation used throughout the codebase.
func hashSeqFallback(nodes []*Node, buf []byte) {
	for i, node := range nodes {
		//nolint:gosec // G115: Safe - node.Type is NodeType (uint8), always fits in byte
		buf[i] = byte(node.Type)
	}
}

// hashSeqSIMD implements a SIMD-optimized hashing approach.
// This will be enabled when ARM64 SIMD support becomes available in Go.
func hashSeqSIMD(nodes []*Node, buf []byte) {
	// For now, fall back to standard implementation
	hashSeqFallback(nodes, buf)
}
