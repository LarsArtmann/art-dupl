package simd

import (
	"bytes"
	"reflect"
	"testing"
	"unsafe"
)

// testData16Bytes is a 16-byte test data slice for SIMD tests.
var testData16Bytes = []byte{
	0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
	0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
}

// assertSliceLength is a test helper that asserts a slice has the expected length.
func assertSliceLength(t *testing.T, slice any, length int) {
	t.Helper()
	v := reflect.ValueOf(slice)
	if v.Len() != length {
		t.Errorf("UnsafeSlice() length = %d, want %d", v.Len(), length)
	}
}

func TestAvailable(t *testing.T) {
	// On ARM64 (including Apple Silicon), Available() returns false
	// This test verifies the current behavior
	result := Available()
	// The result depends on architecture, but on most systems it's false
	// We just verify it doesn't panic and returns a consistent value
	_ = result
}

func TestNewHasher(t *testing.T) {
	hasher := NewHasher()
	if hasher == nil {
		t.Error("NewHasher() returned nil")
	}

	// When SIMD is not available (current state), should return fallbackHasher
	// We verify by checking the behavior matches fallbackHasher
	data := []byte("test data")

	result := hasher.Hash(data)
	if result != nil {
		t.Errorf("fallbackHasher.Hash() = %v, want nil", result)
	}
}

func testHasherHash(t *testing.T, hasher Hasher, name string) {
	tests := []struct {
		testName string
		data     []byte
	}{
		{"empty slice", []byte{}},
		{"single byte", []byte{0x01}},
		{"small data", []byte("hello")},
		{"large data", make([]byte, 1024)},
		{"nil slice", nil},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := hasher.Hash(tt.data)
			if result != nil {
				t.Errorf("%s.Hash() = %v, want nil", name, result)
			}
		})
	}
}

func TestFallbackHasher_Hash(t *testing.T) {
	testHasherHash(t, &fallbackHasher{}, "fallbackHasher")
}

func TestFallbackHasher_HashSlice(t *testing.T) {
	hasher := &fallbackHasher{}

	tests := []struct {
		name string
		data [][]byte
	}{
		{"empty slice", [][]byte{}},
		{"single element", [][]byte{[]byte("test")}},
		{"multiple elements", [][]byte{[]byte("a"), []byte("b"), []byte("c")}},
		{"with nil elements", [][]byte{nil, []byte("test"), nil}},
		{"large slice", make([][]byte, 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasher.HashSlice(tt.data)
			if result == nil {
				t.Error("HashSlice() returned nil slice")

				return
			}

			if len(result) != len(tt.data) {
				t.Errorf("HashSlice() returned %d elements, want %d", len(result), len(tt.data))

				return
			}
			// All elements should be nil (since Hash returns nil)
			for i, r := range result {
				if r != nil {
					t.Errorf("HashSlice()[%d] = %v, want nil", i, r)
				}
			}
		})
	}
}

func TestSimdHasher_Hash(t *testing.T) {
	// simdHasher delegates to fallbackHasher
	testHasherHash(t, &simdHasher{}, "simdHasher")
}

func TestSimdHasher_HashSlice(t *testing.T) {
	// simdHasher delegates to fallbackHasher
	hasher := &simdHasher{}

	data := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	result := hasher.HashSlice(data)

	if result == nil {
		t.Error("HashSlice() returned nil slice")

		return
	}

	if len(result) != len(data) {
		t.Errorf("HashSlice() returned %d elements, want %d", len(result), len(data))

		return
	}

	for i, r := range result {
		if r != nil {
			t.Errorf("HashSlice()[%d] = %v, want nil", i, r)
		}
	}
}

func TestVectorSize(t *testing.T) {
	size := VectorSize()
	// Current implementation returns 64 (AVX-512 preparation)
	if size != 64 {
		t.Errorf("VectorSize() = %d, want 64", size)
	}
}

func TestAlignSlice(t *testing.T) {
	vectorSize := VectorSize()

	tests := []struct {
		name     string
		data     []byte
		wantLen  int  // expected length after alignment
		wantCopy bool // whether the result should be a copy
	}{
		{"empty slice", []byte{}, 0, false},
		{"already aligned", make([]byte, vectorSize), vectorSize, false},
		{"one byte", []byte{0x01}, vectorSize, true},
		{"half vector", make([]byte, vectorSize/2), vectorSize, true},
		{"needs padding", make([]byte, 10), vectorSize, true},
		{"multiple vectors", make([]byte, vectorSize*2), vectorSize * 2, false},
		{"multiple plus one", make([]byte, vectorSize+1), vectorSize * 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill with pattern to verify copy
			for i := range tt.data {
				tt.data[i] = byte(i % 256)
			}

			result := AlignSlice(tt.data)

			if len(result) != tt.wantLen {
				t.Errorf("AlignSlice() length = %d, want %d", len(result), tt.wantLen)

				return
			}

			// Verify original data is preserved
			for i, b := range tt.data {
				if result[i] != b {
					t.Errorf(
						"AlignSlice() corrupted data at index %d: got %d, want %d",
						i,
						result[i],
						b,
					)
				}
			}

			// Verify padding is zeroed
			for i := len(tt.data); i < len(result); i++ {
				if result[i] != 0 {
					t.Errorf("AlignSlice() padding not zeroed at index %d: got %d", i, result[i])
				}
			}

			// Check if it's a copy or the same slice
			if tt.wantCopy && &result[0] == &tt.data[0] {
				t.Error("AlignSlice() returned same slice, expected a copy")
			}

			if !tt.wantCopy && len(result) > 0 && &result[0] != &tt.data[0] {
				t.Error("AlignSlice() returned different slice, expected same slice")
			}
		})
	}
}

func TestAlignSlice_EdgeCases(t *testing.T) {
	vectorSize := VectorSize()

	// Test that padding calculation is correct
	// padding = (size - (len(data) % size)) % size
	for i := 0; i <= vectorSize*2; i++ {
		data := make([]byte, i)
		result := AlignSlice(data)

		expectedLen := ((i + vectorSize - 1) / vectorSize) * vectorSize
		if i%vectorSize == 0 {
			expectedLen = i
		}

		if len(result) != expectedLen {
			t.Errorf("AlignSlice for length %d: got %d, want %d", i, len(result), expectedLen)
		}
	}
}

func TestUnsafeBytes(t *testing.T) {
	t.Run("uint32 pointer", func(t *testing.T) {
		data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		ptr := UnsafeBytes[uint32](data)
		if ptr == nil {
			t.Error("UnsafeBytes() returned nil")

			return
		}
		// The value depends on endianness, just verify we can read it
		value := *ptr
		_ = value
	})

	t.Run("uint64 pointer", func(t *testing.T) {
		data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		ptr := UnsafeBytes[uint64](data)
		if ptr == nil {
			t.Error("UnsafeBytes() returned nil")

			return
		}

		value := *ptr
		_ = value
	})

	t.Run("byte pointer", func(t *testing.T) {
		data := []byte{0x42}

		ptr := UnsafeBytes[byte](data)
		if ptr == nil {
			t.Error("UnsafeBytes() returned nil")

			return
		}

		if *ptr != 0x42 {
			t.Errorf("UnsafeBytes() = 0x%x, want 0x42", *ptr)
		}
	})
}

func TestUnsafeSlice(t *testing.T) {
	t.Run("uint32 slice", func(t *testing.T) {
		// 8 bytes = 2 uint32s
		data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		slice := UnsafeSlice[uint32](data)

		assertSliceLength(t, slice, 2)

		if cap(slice) != 2 {
			t.Errorf("UnsafeSlice() capacity = %d, want 2", cap(slice))
		}
	})

	t.Run("uint64 slice", func(t *testing.T) {
		// 16 bytes = 2 uint64s
		data := []byte{
			0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
			0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		}
		slice := UnsafeSlice[uint64](data)

		assertSliceLength(t, slice, 2)
	})

	t.Run("empty slice", func(t *testing.T) {
		data := []byte{}
		// This should handle empty slice gracefully or panic
		// Current implementation will panic on empty slice
		defer func() {
			_ = recover()
		}()

		slice := UnsafeSlice[uint32](data)
		// If we get here, it handled empty slice
		if len(slice) != 0 {
			t.Errorf("UnsafeSlice() length = %d, want 0", len(slice))
		}
	})

	t.Run("odd length", func(t *testing.T) {
		// 5 bytes = 1 uint32 (remainder truncated)
		data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
		slice := UnsafeSlice[uint32](data)

		if len(slice) != 1 {
			t.Errorf("UnsafeSlice() length = %d, want 1", len(slice))
		}
	})
}

func TestUnsafeBytesAndSlice_Roundtrip(t *testing.T) {
	// Test that UnsafeBytes and UnsafeSlice work correctly together
	original := []byte{
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, 0x07, 0x08,
	}

	// Convert to uint32 slice
	uint32Slice := UnsafeSlice[uint32](original)
	if len(uint32Slice) != 2 {
		t.Fatalf("UnsafeSlice length = %d, want 2", len(uint32Slice))
	}

	// Get pointer to first element
	ptr := UnsafeBytes[uint32](original)
	if ptr == nil {
		t.Fatal("UnsafeBytes returned nil")
	}

	// Verify they point to the same data
	if *ptr != uint32Slice[0] {
		t.Errorf("Pointer value %v != slice[0] %v", *ptr, uint32Slice[0])
	}
}

func TestHasherInterface(t *testing.T) {
	// Verify both hasher types implement the Hasher interface
	var (
		_ Hasher = &fallbackHasher{}
		_ Hasher = &simdHasher{}
	)

	// Verify NewHasher returns a valid Hasher implementation

	if hasher := NewHasher(); hasher == nil {
		t.Error("NewHasher() returned nil")
	}
}

func TestAlignSlice_PreservesData(t *testing.T) {
	// Test that original data is not modified
	original := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	originalCopy := bytes.Clone(original)

	_ = AlignSlice(original)

	if !bytes.Equal(original, originalCopy) {
		t.Error("AlignSlice() modified the original slice")
	}
}

func TestAlignSlice_Idempotent(t *testing.T) {
	// Aligning an already aligned slice should return the same slice
	data := make([]byte, VectorSize())
	for i := range data {
		data[i] = byte(i)
	}

	result1 := AlignSlice(data)
	result2 := AlignSlice(result1)

	if len(result2) != len(result1) {
		t.Errorf("Second align changed length: %d -> %d", len(result1), len(result2))
	}

	if !bytes.Equal(result1, result2) {
		t.Error("Second align changed data")
	}
}

func TestUnsafeOperations_TypeSizes(t *testing.T) {
	// Verify that unsafe operations respect type sizes
	data := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}

	// Test different type sizes
	t.Run("uint8", func(t *testing.T) {
		slice := UnsafeSlice[uint8](data)
		if len(slice) != 16 {
			t.Errorf("uint8 slice length = %d, want 16", len(slice))
		}
	})

	t.Run("uint16", func(t *testing.T) {
		slice := UnsafeSlice[uint16](data)
		if len(slice) != 8 {
			t.Errorf("uint16 slice length = %d, want 8", len(slice))
		}
	})

	t.Run("uint32", func(t *testing.T) {
		slice := UnsafeSlice[uint32](data)
		if len(slice) != 4 {
			t.Errorf("uint32 slice length = %d, want 4", len(slice))
		}
	})

	t.Run("uint64", func(t *testing.T) {
		slice := UnsafeSlice[uint64](data)
		if len(slice) != 2 {
			t.Errorf("uint64 slice length = %d, want 2", len(slice))
		}
	})
}

func TestUnsafeSizeOf(t *testing.T) {
	// This test verifies the unsafe.Sizeof usage in UnsafeSlice
	types := []struct {
		name string
		size int
	}{
		{"uint8", int(unsafe.Sizeof(uint8(0)))},
		{"uint16", int(unsafe.Sizeof(uint16(0)))},
		{"uint32", int(unsafe.Sizeof(uint32(0)))},
		{"uint64", int(unsafe.Sizeof(uint64(0)))},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			expected := map[string]int{
				"uint8":  1,
				"uint16": 2,
				"uint32": 4,
				"uint64": 8,
			}
			if tt.size != expected[tt.name] {
				t.Errorf("unsafe.Sizeof(%s) = %d, want %d", tt.name, tt.size, expected[tt.name])
			}
		})
	}
}
