package syntax

import (
	"fmt"
	"testing"

	"github.com/LarsArtmann/art-dupl/suffixtree"
)

// GenerateNodes creates test nodes for benchmarking.
func GenerateNodes(count int) []*Node {
	nodes := make([]*Node, count)
	for i := range nodes {
		nodes[i] = &Node{
			Type:     int32(i % 256), // Test various types
			Filename: fmt.Sprintf("file_%d.go", i%10),
			Pos:      int32(i),
			End:      int32(i + 1),
		}
	}

	return nodes
}

func BenchmarkHashSeqSmall(b *testing.B) {
	b.ReportAllocs()

	nodes := GenerateNodes(10)

	b.ResetTimer()

	for b.Loop() {
		hashSeq(nodes)
	}
}

func BenchmarkHashSeqMedium(b *testing.B) {
	b.ReportAllocs()

	nodes := GenerateNodes(1000)

	b.ResetTimer()

	for b.Loop() {
		hashSeq(nodes)
	}
}

func BenchmarkHashSeqLarge(b *testing.B) {
	b.ReportAllocs()

	nodes := GenerateNodes(10000)

	b.ResetTimer()

	for b.Loop() {
		hashSeq(nodes)
	}
}

func BenchmarkHashSeqVeryLarge(b *testing.B) {
	b.ReportAllocs()

	nodes := GenerateNodes(100000)

	b.ResetTimer()

	for b.Loop() {
		hashSeq(nodes)
	}
}

// BenchmarkHashSeqFallback benchmarks the fallback (non-SIMD) implementation.
func BenchmarkHashSeqFallback(b *testing.B) {
	b.ReportAllocs()

	nodes := GenerateNodes(10000)
	buf := make([]byte, len(nodes))

	b.ResetTimer()

	for b.Loop() {
		hashSeqFallback(nodes, buf)
	}
}

// BenchmarkBatchHash benchmarks batch hashing of multiple sequences.
func BenchmarkBatchHash(b *testing.B) {
	b.ReportAllocs()

	batchSizes := []int{10, 100, 1000}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("Batch%d", batchSize), func(b *testing.B) {
			sequences := make([][]*Node, batchSize)
			for i := range sequences {
				sequences[i] = GenerateNodes(100)
			}

			b.ResetTimer()

			for b.Loop() {
				BatchHash(sequences)
			}
		})
	}
}

// BenchmarkHashSeqWithConfig benchmarks hashSeq with custom configuration.
func BenchmarkHashSeqWithConfig(b *testing.B) {
	b.ReportAllocs()

	nodes := GenerateNodes(10000)
	config := HashConfig{
		UseSIMD:   false,
		BatchSize: 64,
	}

	b.ResetTimer()

	for b.Loop() {
		HashSeqWithConfig(nodes, config)
	}
}

// BenchmarkHashSeqParallel benchmarks parallel hashing.
func BenchmarkHashSeqParallel(b *testing.B) {
	b.ReportAllocs()

	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			nodes := GenerateNodes(size)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					hashSeq(nodes)
				}
			})
		})
	}
}

// BenchmarkSerialize benchmarks the Serialize function.
func BenchmarkSerialize(b *testing.B) {
	b.ReportAllocs()

	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			// Build a simple tree structure
			root := NewNode()
			nodes := GenerateNodes(size - 1)
			root.AddChildren(nodes...)

			b.ResetTimer()

			for b.Loop() {
				Serialize(root)
			}
		})
	}
}

// BenchmarkFindSyntaxUnits benchmarks the FindSyntaxUnits function.
func BenchmarkFindSyntaxUnits(b *testing.B) {
	b.ReportAllocs()

	// Create test data
	data := GenerateNodes(10000)

	// Create a mock suffix tree match
	match := Match{
		Frags: make([][]*Node, 2),
	}
	match.Frags[0] = data[100:200]
	match.Frags[1] = data[500:600]

	b.ResetTimer()

	for b.Loop() {
		// This will be different in actual usage, but benchmarks the pattern
		FindSyntaxUnits(data, suffixtree.Match{
			Ps:  []suffixtree.Pos{100, 500},
			Len: 100,
		}, 50)
	}
}

// BenchmarkMemoryPool benchmarks the memory pool efficiency.
func BenchmarkMemoryPool(b *testing.B) {
	b.ReportAllocs()

	nodes := GenerateNodes(10000)

	b.ResetTimer()

	for b.Loop() {
		hashSeq(nodes)
		hashSeq(nodes)
		hashSeq(nodes)
		hashSeq(nodes)
		hashSeq(nodes)
	}
}
