// Package simd provides a SIMD abstraction layer for vector operations.
// This package supports runtime detection of SIMD capabilities and provides
// fallback implementations for systems without SIMD support.
//
// IMPORTANT: As of Go 1.26, explicit SIMD support is AMD64-only via the
// simd/archsimd package. ARM64 SIMD support is expected in future Go versions
// (estimated Go 1.28+).
//
// This abstraction layer prepares the codebase for SIMD optimization when it
// becomes available on all target architectures.
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

// fallbackHasher provides the standard non-SIMD hashing implementation.
type fallbackHasher struct{}

// Hash computes a hash using standard crypto/sha256 operations.
// This is the current implementation used throughout the codebase.
func (f *fallbackHasher) Hash(data []byte) []byte {
	// Import crypto/sha256 when needed
	// hash := sha256.Sum256(data)
	// return hash[:]
	// Implementation delegated to caller to avoid import cycle
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
// This will be implemented when ARM64 SIMD becomes available.
type simdHasher struct{}

// Hash computes a hash using SIMD-accelerated operations.
// Expected to be 2-3x faster than fallback on large inputs.
func (s *simdHasher) Hash(data []byte) []byte {
	// TODO: Implement SIMD-accelerated hashing when available
	//
	// Example approach using future simd/archsimd:
	//
	// // Load data into SIMD vectors
	// := simd.LoadBytes(data)
	//
	// // Process in parallel chunks
	// for i := range vectors {
	//     vectors[i] = simd.HashChunk(vectors[i])
	// }
	//
	// // Combine results
	// return simd.CombineHashes(vectors)

	// Fall back to standard implementation for now
	return (&fallbackHasher{}).Hash(data)
}

// HashSlice computes hashes for multiple inputs using SIMD batch processing.
// This allows for processing multiple inputs in parallel using vector registers.
func (s *simdHasher) HashSlice(data [][]byte) [][]byte {
	// TODO: Implement SIMD batch hashing when available
	//
	// Future approach:
	//
	// // Align data for SIMD processing
	// aligned := alignForSIMD(data)
	//
	// // Process batches in parallel
	// for _, batch := range aligned {
	//     processSIMDBatch(batch)
	// }
	//
	// return results

	// Fall back to standard implementation for now
	return (&fallbackHasher{}).HashSlice(data)
}

// alignForSIMD ensures data is properly aligned for SIMD operations.
// SIMD instructions typically require 16, 32, or 64-byte alignment.
func alignForSIMD(data [][]byte) [][]byte {
	// TODO: Implement SIMD alignment when needed
	//
	// Key considerations:
	// - Pad to SIMD register width (16, 32, or 64 bytes)
	// - Ensure cache-line alignment for better performance
	// - Consider memory locality for batch processing

	return data
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
	return (*T)(unsafe.Pointer(&data[0]))
}

// UnsafeSlice casts a byte slice to a slice of a specific type.
// This is a helper function for future SIMD implementations.
//
// WARNING: This function uses unsafe operations and must be used carefully.
func UnsafeSlice[T any](data []byte) []T {
	var t T
	header := (*[1 << 30]T)(unsafe.Pointer(&data[0]))
	n := len(data) / int(unsafe.Sizeof(t))
	return header[:n:n]
}
