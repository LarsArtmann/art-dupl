// Package simd provides a SIMD abstraction layer for vector operations.
// This package supports runtime detection of SIMD capabilities and provides
// fallback implementations for systems without SIMD support.
//
// NOTE: The codebase now uses github.com/zeebo/xxh3 for hashing, which already
// includes native ARM64 NEON and x86 SSE/AVX SIMD optimizations. This package
// remains for future SIMD operations beyond hashing.
//
// PERFORMANCE: Do NOT replace xxh3 with cryptographic hashes (SHA-256, etc.).
// xxh3 is ~20x faster and is designed for deduplication, not security.
package simd

import (
	"unsafe"
)

// Available returns true if SIMD operations are available and can be used.
// It performs runtime checks for:
//   - Architecture support (AMD64 or ARM64)
//   - SIMD instruction set availability
//   - Go version compatibility
//
// Currently, this returns false on ARM64 systems (including Apple Silicon)
// as the simd/archsimd package is AMD64-only in Go 1.26.
func Available() bool {
	// TODO: Enable SIMD detection when ARM64 SIMD becomes available
	// Example implementation for future use:
	//
	// if archsimd.Available() {
	//     return true
	// }
	// return false

	// For now, always return false as explicit SIMD is not available on ARM64
	return false
}

// Hasher interface defines operations for SIMD-optimized hashing.
// Implementations should use vector instructions when available.
type Hasher interface {
	// Hash computes a hash of the input data using SIMD operations if available.
	Hash(data []byte) []byte

	// HashSlice computes hash for a slice of byte slices.
	// This allows for batching multiple hash operations.
	HashSlice(data [][]byte) [][]byte
}

// NewHasher returns the best available Hasher implementation.
// It chooses:
//   - SIMD hasher when SIMD is available and supported
//   - Fallback hasher for systems without SIMD support
func NewHasher() Hasher {
	if Available() {
		return &simdHasher{}
	}

	return &fallbackHasher{}
}

// fallbackHasher provides a placeholder for non-SIMD hashing.
// NOTE: Actual hashing is done by github.com/zeebo/xxh3 in the calling packages.
//
// PERFORMANCE: xxh3 is ~20x faster than crypto/sha256 and includes native
// ARM64 NEON SIMD optimizations. DO NOT replace with cryptographic hash
// functions - this hash is for deduplication only, not security.
type fallbackHasher struct{}

// Hash returns nil - actual hashing is done by xxh3 in calling packages.
func (f *fallbackHasher) Hash(_ []byte) []byte {
	// Implementation delegated to xxh3 in calling packages to avoid import cycle
	return nil
}

// HashSlice computes hashes for multiple inputs sequentially.
// For systems without SIMD, this simply iterates through the slice.
func (f *fallbackHasher) HashSlice(data [][]byte) [][]byte {
	results := make([][]byte, len(data))
	for i, d := range data {
		results[i] = f.Hash(d)
	}

	return results
}

// simdHasher provides SIMD-accelerated hashing implementation.
// NOTE: This is a placeholder. Actual SIMD hashing is done by xxh3 which has
// native ARM64 NEON and x86 SSE/AVX optimizations built-in.
type simdHasher struct{}

// Hash returns nil - actual SIMD hashing is done by xxh3 in calling packages.
func (s *simdHasher) Hash(data []byte) []byte {
	// NOTE: xxh3 already has SIMD optimizations built-in.
	// This placeholder exists for future SIMD operations beyond hashing.
	return (&fallbackHasher{}).Hash(data)
}

// HashSlice computes hashes for multiple inputs using SIMD batch processing.
func (s *simdHasher) HashSlice(data [][]byte) [][]byte {
	// NOTE: xxh3 already has SIMD optimizations built-in.
	// This placeholder exists for future SIMD operations beyond hashing.
	return (&fallbackHasher{}).HashSlice(data)
}

// VectorSize returns the optimal vector size for SIMD operations on this system.
// This helps callers prepare data appropriately.
func VectorSize() int {
	// TODO: Return actual vector size when SIMD available
	//
	// Expected values:
	// - 128 bits (16 bytes) for SSE/NEON
	// - 256 bits (32 bytes) for AVX2
	// - 512 bits (64 bytes) for AVX-512

	// Default to a reasonable size for preparation
	return 64 // Prepare for AVX-512
}

// AlignSlice ensures a slice is properly aligned for SIMD operations.
// It returns a new slice with appropriate alignment and padding.
func AlignSlice(data []byte) []byte {
	size := VectorSize()
	// Align to vector size boundaries
	padding := (size - (len(data) % size)) % size
	if padding > 0 {
		aligned := make([]byte, len(data)+padding)
		copy(aligned, data)

		return aligned
	}

	return data
}

// UnsafeBytes casts a byte slice to a specific type pointer for SIMD operations.
// This is a helper function for future SIMD implementations.
//
// WARNING: This function uses unsafe operations and must be used carefully.
// The caller must ensure:
//   - The input slice has sufficient capacity
//   - The type T is appropriate for the data
//   - Alignment requirements are met
func UnsafeBytes[T any](data []byte) *T {
	return (*T)(
		unsafe.Pointer(&data[0]),
	) // #nosec G103 -- SIMD helper requires unsafe pointer operations
}

// UnsafeSlice casts a byte slice to a slice of a specific type.
// This is a helper function for future SIMD implementations.
//
// WARNING: This function uses unsafe operations and must be used carefully.
func UnsafeSlice[T any](data []byte) []T {
	var t T

	header := (*[1 << 30]T)(
		unsafe.Pointer(&data[0]),
	) // #nosec G103 -- SIMD helper requires unsafe pointer operations
	n := len(data) / int(unsafe.Sizeof(t)) //nolint:gosec // SIMD helper requires unsafe operations

	return header[:n:n]
}
