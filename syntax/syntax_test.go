package syntax

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testhelpers"
)

const testFilename = "test.go"

func TestSerialization(t *testing.T) {
	t.Parallel()

	n := genNodes(7)
	n[0].AddChildren(n[1], n[2], n[3])
	n[1].AddChildren(n[4], n[5])
	n[2].AddChildren(n[6])

	m := genNodes(6)
	m[0].AddChildren(m[1], m[2], m[3], m[4], m[5])
	testCases := []struct {
		t        *Node
		expected []int
	}{
		{n[0], []int{6, 2, 0, 0, 1, 0, 0}},
		{m[0], []int{5, 0, 0, 0, 0, 0}},
	}

	for _, tc := range testCases {
		compareSeries(t, Serialize(tc.t), tc.expected)
	}
}

func genNodes(cnt int) []*Node {
	nodes := make([]*Node, cnt)
	for i := range nodes {
		nodes[i] = NewNode()
	}

	return nodes
}

func compareSeries(t *testing.T, stream []*Node, owns []int) {
	t.Helper()

	if len(stream) != len(owns) {
		t.Errorf("series aren't the same length; got %d, want %d", len(stream), len(owns))

		return
	}

	for i, item := range stream {
		if item.Owns != int32(owns[i]) {
			t.Errorf("got %d, want %d", item.Owns, owns[i])
		}
	}
}

func TestGetUnitsIndexes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		seq       string
		threshold int
		expected  []int
	}{
		{"a8 a0 a2 a0", 3, []int{2}},
		{"a0 a8 a2 a0", 1, []int{2}},
		{"a3 a0 a1", 3, []int{0}},
		{"a3 a0 ", 1, []int{1}},
		{"a1 a0 a1 a0", 2, []int{0, 2}},
	}

Loop:
	for _, tc := range testCases {
		nodes := str2nodes(tc.seq)

		indexes := getUnitsIndexes(nodes, tc.threshold)
		for i := range tc.expected {
			if i > len(indexes)-1 || tc.expected[i] != indexes[i] {
				t.Errorf("for seq '%s', got %v, want %v", tc.seq, indexes, tc.expected)
			}

			continue Loop
		}
	}
}

func TestCyclicDupl(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		seq      string
		indexes  []int
		expected bool
	}{
		{"a1 b0 a2 b0", []int{0, 2}, false},
		{"a1 b0 a1 b0", []int{0, 2}, true},
		{"a0 a0", []int{0, 1}, true}, //nolint:dupword // Intentional duplicate for testing
		{"a1 b0 c1 b0 a1 b0 c1 b0", []int{0, 2, 4, 6}, true},
		{"a1 b0 c1 b0 a1 b0", []int{0, 2, 4}, false},
		{"a0 b0 a0 c0", []int{0, 1, 2, 3}, false},
		{"a0 b0 a0 b0 a0", []int{0, 1, 2}, false},
		{"a1 b0 a1 b0 c1 b0", []int{0, 2, 4}, false},
		{
			"a1 ",
			[]int{0, 4},
			false,
		},
		{
			"a2 b0 a2 b0 a2 b0 a2 b0 a2 b0",
			[]int{0, 3, 6, 9, 12},
			false,
		},
	}

	for _, tc := range testCases {
		nodes := str2nodes(tc.seq)
		if tc.expected != isCyclic(tc.indexes, nodes) {
			t.Errorf(
				"for seq '%s', indexes %v, got %t, want %t",
				tc.seq,
				tc.indexes,
				!tc.expected,
				tc.expected,
			)
		}
	}
}

// str2nodes converts strint to a sequence of *Node by following principle:
//   - node is represented by 2 characters
//   - first character is node type
//   - second character is the number for Node.Owns.
func str2nodes(str string) []*Node {
	chars := []rune(str)

	nodes := make([]*Node, (len(chars)+1)/3)
	for i := 0; i < len(chars)-1; i += 3 {
		nodes[i/3] = &Node{Type: chars[i], Owns: chars[i+1] - '0'}
	}

	return nodes
}

func FuzzSerialize(f *testing.F) {
	// Add seed corpus with valid Go code patterns
	f.Add("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}")
	f.Add("func test() int {\n\treturn 42\n}")
	f.Add("type Foo struct {\n\tX int\n}")
	f.Add("if x > 0 {\n\treturn true\n}")
	f.Add("for i := 0; i < 10; i++ {\n\tfmt.Println(i)\n}")
	f.Add("var x int = 5")
	f.Add("func (f *Foo) Method() string {\n\treturn \"test\"\n}")
	f.Add("switch v {\ncase 1:\n\treturn \"one\"\ndefault:\n\treturn \"unknown\"\n}")
	f.Add("defer func() {\n\tlog.Println(\"done\")\n}()")
	f.Add("package main")

	f.Fuzz(func(t *testing.T, input string) {
		// Parse input to create AST nodes
		// Since we can't reliably parse all fuzz inputs as Go code,
		// we'll create synthetic nodes based on input characteristics
		defer testhelpers.PanicRecovery(t, input)()

		// Create a simple node tree structure
		root := createTestNodeTree(input)

		// Serialize the node tree
		stream := Serialize(root)

		// Verify invariants
		if stream == nil {
			t.Error("Serialize returned nil stream")
		}

		// Verify stream is not empty for non-empty input
		if len(input) > 0 && len(stream) == 0 {
			t.Error("Serialize returned empty stream for non-empty input")
		}

		// Verify each node in stream is non-nil
		for i, node := range stream {
			if node == nil {
				t.Errorf("Stream contains nil node at index %d", i)
			}
		}

		// Verify root owns the correct number of descendants
		if len(stream) > 0 && root.Owns != int32(len(stream)-1) {
			t.Errorf("Root Owns mismatch: got %d, want %d", root.Owns, len(stream)-1)
		}
	})
}

// createTestNodeTree creates a synthetic node tree for fuzz testing.
func createTestNodeTree(input string) *Node {
	if len(input) == 0 {
		return NewNode()
	}

	// Create a tree based on input length and content
	root := NewNode()
	root.Type = int32(len(input) % 100)
	root.Filename = testFilename
	root.Pos = 0
	root.End = int32(len(input))

	// Add children based on input characteristics
	childCount := len(input) % 20
	for i := range childCount {
		child := NewNode()
		child.Type = int32(int(input[i%len(input)]) % 50)
		child.Filename = testFilename
		child.Pos = int32(i)
		child.End = int32(i + 1)
		root.AddChildren(child)

		// Add grandchildren
		if i%2 == 0 && i+1 < childCount {
			grandchild := NewNode()
			grandchild.Type = int32(int(input[(i+1)%len(input)]) % 30)
			grandchild.Filename = testFilename
			grandchild.Pos = int32(i + 1)
			grandchild.End = int32(i + 2)
			child.AddChildren(grandchild)
		}
	}

	return root
}
