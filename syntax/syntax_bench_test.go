package syntax

import "testing"

func BenchmarkSerialize_Small(b *testing.B) {
	root := &Node{
		Type: 1,
		Children: []*Node{
			{Type: 2, Name: "x"},
			{Type: 3, Name: "y", Children: []*Node{{Type: 4}}},
		},
	}

	b.ResetTimer()

	for range b.N {
		_ = Serialize(root)
	}
}

func BenchmarkSerialize_Large(b *testing.B) {
	root := genDeepTree(100)

	b.ResetTimer()

	for range b.N {
		_ = Serialize(root)
	}
}

func BenchmarkSerialize_Statements(b *testing.B) {
	root := &Node{
		Type: 1,
		Children: []*Node{
			{Type: 2, Statement: true, Children: []*Node{{Type: 3, Name: "x"}}},
			{Type: 4, Statement: true, Children: []*Node{{Type: 5, Name: "y"}}},
			{Type: 6, Statement: true, Children: []*Node{{Type: 7, Name: "z"}}},
		},
	}

	b.ResetTimer()

	for range b.N {
		_ = Serialize(root)
	}
}

func BenchmarkSerialize_Idempotent(b *testing.B) {
	root := genDeepTree(50)

	b.ResetTimer()

	for range b.N {
		// Two serializations to verify idempotency overhead
		_ = Serialize(root)
		_ = Serialize(root)
	}
}

func genDeepTree(depth int) *Node {
	root := &Node{Type: 1}
	current := root

	for i := 1; i < depth; i++ {
		child := &Node{Type: int32(i + 1), Name: "var"}
		current.Children = []*Node{child}
		current = child
	}

	return root
}
