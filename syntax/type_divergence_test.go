package syntax

import (
	"slices"
	"testing"
)

func TestCountTypeDivergencePositions(t *testing.T) {
	tests := []struct {
		name  string
		frags [][]*Node
		want  int
	}{
		{
			name:  "fewer than two fragments",
			frags: [][]*Node{{{VarType: "int"}}},
			want:  0,
		},
		{
			name: "zero divergence, identical types",
			frags: [][]*Node{
				{Name: "v0", VarType: "int", Children: []*Node{{Name: "v1", VarType: "string"}}},
				{Name: "v0", VarType: "int", Children: []*Node{{Name: "v1", VarType: "string"}}},
			},
			want: 0,
		},
		{
			name: "one divergent position",
			frags: [][]*Node{
				{Name: "v0", VarType: "int", Children: []*Node{{Name: "v1", VarType: "string"}}},
				{Name: "v0", VarType: "int64", Children: []*Node{{Name: "v1", VarType: "string"}}},
			},
			want: 1,
		},
		{
			name: "two divergent positions",
			frags: [][]*Node{
				{Name: "v0", VarType: "int", Children: []*Node{{Name: "v1", VarType: "string"}}},
				{Name: "v0", VarType: "int64", Children: []*Node{{Name: "v1", VarType: "[]byte"}}},
			},
			want: 2,
		},
		{
			name: "empty VarType on either side is ignored",
			frags: [][]*Node{
				{Name: "v0", VarType: "int", Children: []*Node{{Name: "v1"}}},
				{Name: "v0", Children: []*Node{{Name: "v1", VarType: "string"}}},
			},
			want: 0,
		},
		{
			name: "three fragments, union of positions across all pairs",
			frags: [][]*Node{
				{Name: "v0", VarType: "int", Children: []*Node{{Name: "v1", VarType: "string"}, {Name: "v2", VarType: "bool"}}},
				{Name: "v0", VarType: "int64", Children: []*Node{{Name: "v1", VarType: "string"}, {Name: "v2", VarType: "bool"}}},
				{Name: "v0", VarType: "int64", Children: []*Node{{Name: "v1", VarType: "[]byte"}, {Name: "v2", VarType: "bool"}}},
			},
			want: 2,
		},
		{
			name: "length mismatch compares shared prefix only",
			frags: [][]*Node{
				{Name: "v0", VarType: "int", Children: []*Node{{Name: "v1", VarType: "string"}}},
				{Name: "v0", VarType: "int64"},
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
	makeFrags := func(typeA, typeB string) [][]*Node {
		return [][]*Node{
			{Name: "v0", VarType: typeA, Children: []*Node{{Name: "v1", VarType: "string"}}},
			{Name: "v0", VarType: typeB, Children: []*Node{{Name: "v1", VarType: "[]byte"}}},
		}
	}

	tests := []struct {
		name  string
		frags [][]*Node
		want  bool
	}{
		{name: "zero divergence is not candidate structure", frags: makeFrags("int", "int"), want: false},
		{name: "one divergence is not candidate structure", frags: [][]*Node{
			{Name: "v0", VarType: "time.Time"}, {Name: "v0", VarType: "*big.Int"},
		}, want: false},
		{name: "two divergences is candidate structure", frags: makeFrags("int", "int64"), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGenericsCandidateStructure(tt.frags); got != tt.want {
				t.Errorf("IsGenericsCandidateStructure() = %t, want %t", got, tt.want)
			}
		})
	}
}

// TestMinDivergentPositionsMatchesPrinter pins the canonical constant against
// the printer re-export so the two can never drift.
func TestMinDivergentPositionsMatchesPrinter(t *testing.T) {
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
	want := []string{"a", "b", "d", "c"}

	var gotNames []string
	for _, n := range got {
		gotNames = append(gotNames, n.Name)
	}

	if !slices.Equal(gotNames, want) {
		t.Errorf("flattenNodes() = %v, want %v", gotNames, want)
	}
}
