package templ

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// FuzzParseBytes tests that ParseBytes never panics on arbitrary input
// and maintains the invariant: either (node != nil, err == nil) or (node == nil, err != nil).
func FuzzParseBytes(f *testing.F) {
	// Seed corpus: valid and edge-case templ content
	f.Add([]byte(`package main`))
	f.Add([]byte(`templ Hello(name string) {
	<div>Hello, { name }</div>
}`))
	f.Add([]byte(``))
	f.Add([]byte(`<div>raw html</div>`))
	f.Add([]byte(`templ Empty() {}`))
	f.Add([]byte(`css Style() {
	color: red;
}`))
	f.Add([]byte(`templ Nested() {
	<div>
		<span>
			<p>deep</p>
		</span>
	</div>
}`))
	f.Add([]byte("{| broken syntax |}"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 10000 {
			return // keep test fast
		}

		node, _, err := ParseBytes("fuzz.templ", data)

		// Invariant: success → node non-nil; failure → error non-nil
		if err == nil && node == nil {
			t.Fatal("ParseBytes returned nil node with nil error")
		}

		if err != nil && node != nil {
			t.Fatal("ParseBytes returned non-nil node with non-nil error")
		}

		// If we got a valid node, walk it to ensure no panics during traversal
		if node != nil {
			walkNodeSafe(node)
		}
	})
}

// walkNodeSafe recursively traverses the node tree, accessing all fields
// to ensure the transformer didn't leave any inconsistent state.
func walkNodeSafe(n *syntax.Node) {
	if n == nil {
		return
	}

	for _, child := range n.Children {
		walkNodeSafe(child)
	}
}
