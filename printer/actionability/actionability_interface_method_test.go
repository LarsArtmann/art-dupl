package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestIsInterfaceMethodBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		seqs     [][]*domain.CloneNode
		expected bool
	}{
		{
			name: "String method with small body",
			seqs: [][]*domain.CloneNode{
				mustInterfaceMethodNode("String", 2),
				mustInterfaceMethodNode("String", 2),
			},
			expected: true,
		},
		{
			name: "Read method with body at limit (4)",
			seqs: [][]*domain.CloneNode{
				mustInterfaceMethodNode("Read", maxInterfaceMethodBodyNodes),
			},
			expected: true,
		},
		{
			name: "non-interface method name",
			seqs: [][]*domain.CloneNode{
				mustInterfaceMethodNode("ProcessData", 2),
				mustInterfaceMethodNode("ProcessData", 2),
			},
			expected: false,
		},
		{
			name: "type-aware: custom interface method suppressed via flag",
			seqs: [][]*domain.CloneNode{
				mustTypeAwareInterfaceMethodNode("Validate", 2),
				mustTypeAwareInterfaceMethodNode("Validate", 2),
			},
			expected: true,
		},
		{
			name: "type-aware: custom interface method with large body not suppressed",
			seqs: [][]*domain.CloneNode{
				mustTypeAwareInterfaceMethodNode("Validate", maxInterfaceMethodBodyNodes+1),
			},
			expected: false,
		},
		{
			name: "type-aware: flag takes priority over non-matching static name",
			seqs: [][]*domain.CloneNode{
				mustTypeAwareInterfaceMethodNode("DoCustomThing", 1),
				mustTypeAwareInterfaceMethodNode("DoCustomThing", 1),
			},
			expected: true,
		},
		{
			name: "mixed: one type-aware flagged, one static name (both suppress)",
			seqs: [][]*domain.CloneNode{
				mustTypeAwareInterfaceMethodNode("Validate", 2),
				mustInterfaceMethodNode("String", 2),
			},
			expected: true,
		},
		{
			name: "body too large (5 nodes)",
			seqs: [][]*domain.CloneNode{
				mustInterfaceMethodNode("String", maxInterfaceMethodBodyNodes+1),
			},
			expected: false,
		},
		{
			name: "non-FuncDecl node",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.AssignStmt}},
			},
			expected: false,
		},
		{
			name: "statement-level: interface method body suppressed via propagated flag",
			seqs: [][]*domain.CloneNode{
				mustStatementInterfaceMethodSeq(1),
				mustStatementInterfaceMethodSeq(1),
			},
			expected: true,
		},
		{
			name: "statement-level: 3-statement interface method body suppressed",
			seqs: [][]*domain.CloneNode{
				mustStatementInterfaceMethodSeq(3),
				mustStatementInterfaceMethodSeq(3),
			},
			expected: true,
		},
		{
			name: "statement-level: at limit (4 statements) suppressed",
			seqs: [][]*domain.CloneNode{
				mustStatementInterfaceMethodSeq(maxInterfaceMethodBodyNodes),
			},
			expected: true,
		},
		{
			name: "statement-level: over limit (5 statements) not suppressed",
			seqs: [][]*domain.CloneNode{
				mustStatementInterfaceMethodSeq(maxInterfaceMethodBodyNodes + 1),
			},
			expected: false,
		},
		{
			name: "statement-level: without InterfaceMethod flag not suppressed",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.ReturnStmt}},
				{{BaseType: golang.ReturnStmt}},
			},
			expected: false,
		},
		{
			name:     "empty seqs",
			seqs:     [][]*domain.CloneNode{},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isInterfaceMethodBody(tc.seqs)
			if result != tc.expected {
				t.Errorf("isInterfaceMethodBody() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func mustInterfaceMethodNode(methodName string, bodyChildCount int) []*domain.CloneNode {
	children := make([]*domain.CloneNode, 0, bodyChildCount)
	for range bodyChildCount {
		children = append(children, &domain.CloneNode{BaseType: golang.ReturnStmt})
	}

	return []*domain.CloneNode{
		{
			BaseType: golang.FuncDecl,
			Name:     methodName,
			Children: []*domain.CloneNode{
				{BaseType: golang.BlockStmt, Children: children},
			},
		},
	}
}

func mustTypeAwareInterfaceMethodNode(methodName string, bodyChildCount int) []*domain.CloneNode {
	seq := mustInterfaceMethodNode(methodName, bodyChildCount)
	seq[0].InterfaceMethod = true

	return seq
}

func mustStatementInterfaceMethodSeq(stmtCount int) []*domain.CloneNode {
	seq := make([]*domain.CloneNode, stmtCount)
	for i := range seq {
		seq[i] = &domain.CloneNode{
			BaseType:        golang.ReturnStmt,
			InterfaceMethod: true,
		}
	}

	return seq
}
