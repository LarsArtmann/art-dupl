package syntax

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testhelpers"
	"github.com/LarsArtmann/art-dupl/suffixtree"
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
	nodes := make([]*Node, 0, cnt)
	for range cnt {
		nodes = append(nodes, NewNode())
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

	nodes := make([]*Node, 0, (len(chars)+1)/3)
	for i := 0; i < len(chars)-1; i += 3 {
		nodes = append(nodes, &Node{Type: chars[i], Owns: chars[i+1] - '0'})
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
		// (stream[0] is the serialized copy of root with correct Owns)
		if len(stream) > 0 && stream[0].Owns != int32(len(stream)-1) {
			t.Errorf("Root Owns mismatch: got %d, want %d", stream[0].Owns, len(stream)-1)
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

func TestVal(t *testing.T) {
	t.Parallel()

	n := NewNode()
	n.Type = 42

	if got := n.Val(); got != suffixtree.TokenValue(42) {
		t.Errorf("Val() = %v, want %v", got, suffixtree.TokenValue(42))
	}
}

func TestNewSyntheticFileNode(t *testing.T) {
	t.Parallel()

	n := NewSyntheticFileNode("main.go", 500)

	if n.Filename != "main.go" {
		t.Errorf("Filename = %q, want %q", n.Filename, "main.go")
	}

	if n.Pos != 0 {
		t.Errorf("Pos = %d, want 0", n.Pos)
	}

	if n.End != 500 {
		t.Errorf("End = %d, want 500", n.End)
	}

	if n.Type != 1 {
		t.Errorf("Type = %d, want 1", n.Type)
	}

	if len(n.Children) != 0 {
		t.Errorf("Children should be empty, got %d", len(n.Children))
	}
}

func TestUnique(t *testing.T) {
	t.Parallel()

	group := [][]*Node{
		{{Filename: "a.go", Pos: 1, End: 10}},
		{{Filename: "a.go", Pos: 1, End: 10}},
		{{Filename: "b.go", Pos: 1, End: 10}},
		{{Filename: "a.go", Pos: 5, End: 15}},
	}

	result := Unique(group)

	if len(result) != 3 {
		t.Errorf("Unique() returned %d groups, want 3", len(result))
	}
}

func TestUniqueEmptyGroup(t *testing.T) {
	t.Parallel()

	group := [][]*Node{
		{},
		{{Filename: "a.go", Pos: 1, End: 10}},
	}

	result := Unique(group)

	if len(result) != 1 {
		t.Errorf("Unique() with empty group returned %d, want 1", len(result))
	}
}

func TestCountUniqueFiles(t *testing.T) {
	t.Parallel()

	testGroup := [][]*Node{
		{{Filename: "a.go"}},
		{{Filename: "a.go"}},
		{{Filename: "b.go"}},
		{{Filename: "c.go"}},
	}

	if got := CountUniqueFiles(testGroup); got != 3 {
		t.Errorf("CountUniqueFiles() = %d, want 3", got)
	}
}

func TestCountUniqueFilesEmptySequences(t *testing.T) {
	t.Parallel()

	group := [][]*Node{
		{},
		{{Filename: "a.go"}},
		{},
	}

	if got := CountUniqueFiles(group); got != 1 {
		t.Errorf("CountUniqueFiles() = %d, want 1", got)
	}
}

func TestSerializeIdempotent(t *testing.T) {
	t.Parallel()

	// Build a tree with statement nodes (exercises fingerprintSubtree path)
	root := &Node{
		Type:      100,
		Statement: true,
		Children: []*Node{
			{Type: 200, Name: "x"},
			{Type: 300, Name: "y"},
		},
	}

	originalType := root.Type

	first := Serialize(root)
	second := Serialize(root)

	// Original tree must NOT be mutated
	if root.Type != originalType {
		t.Errorf("original tree mutated: Type was %d, now %d", originalType, root.Type)
	}

	// Both serializations must produce identical results
	if len(first) != len(second) {
		t.Fatalf("length mismatch: first=%d, second=%d", len(first), len(second))
	}

	for i := range first {
		if first[i].Type != second[i].Type {
			t.Errorf(
				"Type mismatch at %d: first=%d, second=%d",
				i, first[i].Type, second[i].Type,
			)
		}

		if first[i].Owns != second[i].Owns {
			t.Errorf(
				"Owns mismatch at %d: first=%d, second=%d",
				i, first[i].Owns, second[i].Owns,
			)
		}
	}
}

func TestSerializeDoesNotMutateOriginal(t *testing.T) {
	t.Parallel()

	// Non-statement tree: serial sets Owns on each node
	root := &Node{
		Type: 10,
		Children: []*Node{
			{Type: 20},
			{Type: 30, Children: []*Node{{Type: 40}}},
		},
	}

	originalOwns := []int32{root.Owns, root.Children[0].Owns, root.Children[1].Owns}

	_ = Serialize(root)

	// Original Owns must be unchanged
	if root.Owns != originalOwns[0] {
		t.Errorf("root.Owns mutated: was %d, now %d", originalOwns[0], root.Owns)
	}

	if root.Children[0].Owns != originalOwns[1] {
		t.Errorf("child[0].Owns mutated: was %d, now %d", originalOwns[1], root.Children[0].Owns)
	}

	if root.Children[1].Owns != originalOwns[2] {
		t.Errorf("child[1].Owns mutated: was %d, now %d", originalOwns[2], root.Children[1].Owns)
	}
}

func TestSerializePreservesIsAlias(t *testing.T) {
	t.Parallel()

	root := &Node{
		Type: 10,
		Children: []*Node{
			{Type: 50, IsAlias: true, Statement: true, Children: []*Node{{Type: 60}}},
		},
	}

	stream := Serialize(root)

	// The TypeSpec child is a statement node: serial emits it as a single
	// fingerprinted token. The shallow copy must preserve IsAlias so the
	// actionability layer can distinguish type aliases from type definitions.
	for _, n := range stream {
		if n.Statement && n.Type == 50 && !n.IsAlias {
			t.Errorf("statement node Type=%d: IsAlias lost during serialization (got false, want true)", n.Type)
		}
	}
}
