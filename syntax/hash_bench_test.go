package syntax

import (
	"fmt"
	"testing"
)

// GenerateNodes creates test nodes for benchmarking
func GenerateNodes(count int) []*Node {
	nodes := make([]*Node, count)
	for i := range nodes {
		nodes[i] = &Node{
			Type:     i % 256, // Test various types
			Filename: fmt.Sprintf("file_%d.go", i%10),
			Pos:      i,
			End:      i + 1,
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
