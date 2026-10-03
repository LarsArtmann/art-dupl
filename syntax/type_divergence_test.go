package syntax

import (
	"slices"
	"testing"
)

func TestCountTypeDivergencePositions(t *testing.T) {
	node := func(name, varType string, children ...*Node) *Node {
		return &Node{Name: name, VarType: varType, Children: children}
	}

	tests := []struct {
		name  string
		frags [][]*Node
		want  int
	}{
		{
			name:  "fewer than two fragments",
			frags: [][]*Node{{node("v0", "int")}},
			want:  0,
		},
		{
			name: "zero divergence, identical types",
			frags: [][]*Node{
				{node("v0", "int", node("v1", "string"))},
				{node("v0", "int", node("v1", "string"))},
			},
			want: 0,
		},
		{
			name: "one divergent position",
			frags: [][]*Node{
				{node("v0", "int", node("v1", "string"))},
				{node("v0", "int64", node("v1", "string"))},
			},
			want: 1,
		},
		{
			name: "two divergent positions",
			frags: [][]*Node{
				{node("v0", "int", node("v1", "string"))},
				{node("v0", "int64", node("v1", "[]byte"))},
			},
			want: 2,
		},
		{
			name: "empty VarType on either side is ignored",
			frags: [][]*Node{
				{node("v0", "int", node("v1", ""))},
				{node("v0", "", node("v1", "string"))},
			},
			want: 0,
		},
		{
			name: "three fragments, union of positions across all pairs",
			frags: [][]*Node{
				{node("v0", "int", node("v1", "string"), node("v2", "bool"))},
				{node("v0", "int64", node("v1", "string"), node("v2", "bool"))},
				{node("v0", "int64", node("v1", "[]byte"), node("v2", "bool"))},
			},
			want: 2,
		},
		{
			name: "length mismatch compares shared prefix only",
			frags: [][]*Node{
				{node("v0", "int", node("v1", "string"))},
				{node("v0", "int64")},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountTypeDivergencePositions(tt.frags); got != tt.want {
				t.Errorf("CountTypeDivergencePositions() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIsGenericsCandidateStructure(t *testing.T) {
	node := func(name, varType string, children ...*Node) *Node {
		return &Node{Name: name, VarType: varType, Children: children}
	}

	tests := []struct {
		name  string
		frags [][]*Node
		want  bool
	}{
		{
			name: "zero divergence is not candidate structure",
			frags: [][]*Node{
				{node("v0", "int", node("v1", "string"))},
				{node("v0", "int", node("v1", "string"))},
			},
			want: false,
		},
		{
			name: "single receiver-only divergence is not candidate structure",
			frags: [][]*Node{
				{node("v0", "time.Time")},
				{node("v0", "*big.Int")},
			},
			want: false,
		},
		{
			name: "two divergences is candidate structure",
			frags: [][]*Node{
				{node("v0", "int", node("v1", "string"))},
				{node("v0", "int64", node("v1", "[]byte"))},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGenericsCandidateStructure(tt.frags); got != tt.want {
				t.Errorf("IsGenericsCandidateStructure() = %t, want %t", got, tt.want)
			}
		})
	}
}

// TestMinDivergentPositionsValue pins the canonical constant: the printer
// re-export and the combined-mode gate both derive from it.
func TestMinDivergentPositionsValue(t *testing.T) {
	if MinDivergentPositions != 2 {
		t.Errorf("MinDivergentPositions = %d, want 2", MinDivergentPositions)
	}
}

// TestFlattenNodesPreOrder guards the traversal order that must mirror
// printer's ClassifyGenericsCandidate flattening for position alignment.
func TestFlattenNodesPreOrder(t *testing.T) {
	tree := &Node{Name: "a", Children: []*Node{
		{Name: "b", Children: []*Node{{Name: "d"}}},
		{Name: "c"},
	}}

	got := flattenNodes([]*Node{tree})

	gotNames := make([]string, 0, len(got))
	for _, n := range got {
		gotNames = append(gotNames, n.Name)
	}

	if !slices.Equal(gotNames, []string{"a", "b", "d", "c"}) {
		t.Errorf("flattenNodes() = %v, want [a b d c]", gotNames)
	}
}
