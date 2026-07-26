package printer

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
				mustFuncDeclWithBody("String", 2),
				mustFuncDeclWithBody("String", 2),
			},
			expected: true,
		},
		{
			name: "Read method with body at limit (4)",
			seqs: [][]*domain.CloneNode{
				mustFuncDeclWithBody("Read", maxInterfaceMethodBodyNodes),
			},
			expected: true,
		},
		{
			name: "non-interface method name",
			seqs: [][]*domain.CloneNode{
				mustFuncDeclWithBody("ProcessData", 2),
				mustFuncDeclWithBody("ProcessData", 2),
			},
			expected: false,
		},
		{
			name: "body too large (5 nodes)",
			seqs: [][]*domain.CloneNode{
				mustFuncDeclWithBody("String", maxInterfaceMethodBodyNodes+1),
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

func mustFuncDeclWithBody(methodName string, bodyChildCount int) []*domain.CloneNode {
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
