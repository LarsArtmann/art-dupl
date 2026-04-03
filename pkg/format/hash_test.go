package format

import (
	"encoding/hex"
	"testing"
)

func TestHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"zero", 0, "0000000000000000"},
		{"one", 1, "0000000000000001"},
		{"max uint64", ^uint64(0), "ffffffffffffffff"},
		{"small value", 0x1234, "0000000000001234"},
		{"medium value", 0xdeadbeef, "00000000deadbeef"},
		{"large value", 0x123456789abcdef0, "123456789abcdef0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Hash(tt.input)
			if result != tt.expected {
				t.Errorf("Hash(%#x) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHashMatchesHexEncoding(t *testing.T) {
	t.Parallel()

	testCases := []uint64{
		0,
		1,
		0x1234,
		0xdeadbeef,
		0x123456789abcdef0,
		^uint64(0),
		0x5555555555555555,
		0xaaaaaaaaaaaaaaaa,
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			result := Hash(tc)

			expected := hex.EncodeToString([]byte{
				byte(tc >> 56), byte(tc >> 48), byte(tc >> 40), byte(tc >> 32),
				byte(tc >> 24), byte(tc >> 16), byte(tc >> 8), byte(tc),
			})
			if result != expected {
				t.Errorf("Hash(%#x) = %q, want %q", tc, result, expected)
			}
		})
	}
}

func TestHashLength(t *testing.T) {
	t.Parallel()

	for i := range 100 {
		result := Hash(uint64(i))
		if len(result) != 16 {
			t.Errorf("Hash(%d) returned string of length %d, want 16", i, len(result))
		}
	}
}

func BenchmarkHash(b *testing.B) {
	benchmarks := []struct {
		name  string
		input uint64
	}{
		{"zero", 0},
		{"small", 0x1234},
		{"medium", 0xdeadbeef},
		{"large", 0x123456789abcdef0},
		{"max", ^uint64(0)},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()

			for range b.N {
				_ = Hash(bm.input)
			}
		})
	}
}
