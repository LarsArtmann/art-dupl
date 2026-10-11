package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// Collector-tail and option-apply-loop pattern tests. Shapes mirror the
// measured false-positive families from the 2026-10-11 precision re-measure
// (docs/research/2026-10-11_suggest-generics-precision-baseline.md ledger).

func node(baseType int32, children ...*domain.CloneNode) *domain.CloneNode {
	return &domain.CloneNode{BaseType: baseType, Children: children}
}

func collectorIdent(name string) *domain.CloneNode {
	return &domain.CloneNode{BaseType: golang.Ident, Name: name}
}

func appendAssign(target, value string) *domain.CloneNode {
	return node(golang.AssignStmt,
		collectorIdent(target),
		node(golang.CallExpr, collectorIdent("append"), collectorIdent(target), collectorIdent(value)),
	)
}

func collectorLoop(body ...*domain.CloneNode) *domain.CloneNode {
	return node(golang.RangeStmt,
		collectorIdent("v1"),
		collectorIdent("v0"),
		node(golang.BlockStmt, body...),
	)
}

func optionApplyLoop() *domain.CloneNode {
	return node(golang.RangeStmt,
		collectorIdent("v1"),
		collectorIdent("v0"),
		node(golang.BlockStmt,
			node(golang.ExprStmt,
				node(golang.CallExpr, collectorIdent("v1"), collectorIdent("v2")),
			),
		),
	)
}

func TestAppendCollectorTail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		seqs [][]*domain.CloneNode
		want bool
	}{
		{
			name: "bare append loop matches",
			seqs: [][]*domain.CloneNode{
				{collectorLoop(appendAssign("v2", "v1"))},
				{collectorLoop(appendAssign("v2", "v1"))},
			},
			want: true,
		},
		{
			name: "filtered append loop matches (if with collector body)",
			seqs: [][]*domain.CloneNode{
				{collectorLoop(node(golang.IfStmt,
					node(golang.BinaryExpr, collectorIdent("v1"), collectorIdent("v3")),
					node(golang.BlockStmt, appendAssign("v2", "v1")),
				))},
				{collectorLoop(node(golang.IfStmt,
					node(golang.BinaryExpr, collectorIdent("v1"), collectorIdent("v3")),
					node(golang.BlockStmt, appendAssign("v2", "v1")),
				))},
			},
			want: true,
		},
		{
			name: "append loop with make prealloc and return matches",
			seqs: [][]*domain.CloneNode{
				{
					node(golang.AssignStmt,
						collectorIdent("v2"),
						node(golang.CallExpr,
							collectorIdent("make"),
							collectorIdent("v4"),
							node(golang.BasicLit),
							node(golang.CallExpr, collectorIdent("len"), collectorIdent("v0")),
						),
					),
					collectorLoop(appendAssign("v2", "v1")),
					node(golang.ReturnStmt, collectorIdent("v2")),
				},
				{collectorLoop(appendAssign("v2", "v1"))},
			},
			want: true,
		},
		{
			name: "loop with real processing call does NOT match",
			seqs: [][]*domain.CloneNode{
				{collectorLoop(node(golang.AssignStmt,
					collectorIdent("v2"),
					node(golang.CallExpr, collectorIdent("normalize"), collectorIdent("v1")),
				))},
				{collectorLoop(node(golang.AssignStmt,
					collectorIdent("v2"),
					node(golang.CallExpr, collectorIdent("normalize"), collectorIdent("v1")),
				))},
			},
			want: false,
		},
		{
			name: "append of constructed composite does NOT match",
			seqs: [][]*domain.CloneNode{
				{collectorLoop(node(golang.AssignStmt,
					collectorIdent("v2"),
					node(golang.CallExpr, collectorIdent("append"), collectorIdent("v2"), node(golang.CompositeLit)),
				))},
			},
			want: false,
		},
		{
			name: "append to different slice does NOT match",
			seqs: [][]*domain.CloneNode{
				{collectorLoop(node(golang.AssignStmt,
					collectorIdent("v2"),
					node(golang.CallExpr, collectorIdent("append"), collectorIdent("v5"), collectorIdent("v1")),
				))},
			},
			want: false,
		},
		{
			name: "if with else does NOT match",
			seqs: [][]*domain.CloneNode{
				{node(golang.IfStmt,
					node(golang.BinaryExpr),
					node(golang.BlockStmt, appendAssign("v2", "v1")),
					node(golang.BlockStmt, appendAssign("v2", "v1")),
				)},
			},
			want: false,
		},
		{
			name: "guard body with real call does NOT match",
			seqs: [][]*domain.CloneNode{
				{node(golang.IfStmt,
					node(golang.BinaryExpr),
					node(golang.BlockStmt,
						node(golang.ExprStmt,
							node(golang.CallExpr, collectorIdent("strings"), collectorIdent("v1")),
						),
					),
				)},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			label, action := evaluateActionabilityWithDisabled(tc.seqs, disabledExcept(
				PatternAppendCollectorTail,
			))
			if tc.want && action != domain.NonActionable {
				t.Errorf("expected NonActionable, got %v (label %q)", action, label)
			}

			if !tc.want && label == PatternAppendCollectorTail {
				t.Errorf("pattern matched but should not have")
			}
		})
	}
}

func TestOptionApplyLoop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		seqs [][]*domain.CloneNode
		want bool
	}{
		{
			name: "bare apply loop + return matches",
			seqs: [][]*domain.CloneNode{
				{optionApplyLoop(), node(golang.ReturnStmt, collectorIdent("v2"))},
				{optionApplyLoop(), node(golang.ReturnStmt, collectorIdent("v2"))},
			},
			want: true,
		},
		{
			name: "apply loop without return matches",
			seqs: [][]*domain.CloneNode{
				{optionApplyLoop()},
				{optionApplyLoop()},
			},
			want: true,
		},
		{
			name: "constructor with logic after loop does NOT match",
			seqs: [][]*domain.CloneNode{
				{optionApplyLoop(), node(golang.IfStmt,
					node(golang.BinaryExpr, collectorIdent("v3"), node(golang.BasicLit)),
					node(golang.BlockStmt, node(golang.AssignStmt, collectorIdent("v4"))),
				)},
			},
			want: false,
		},
		{
			name: "loop calling non-range function does NOT match",
			seqs: [][]*domain.CloneNode{
				{node(golang.RangeStmt,
					collectorIdent("v1"),
					collectorIdent("v0"),
					node(golang.BlockStmt,
						node(golang.ExprStmt,
							node(golang.CallExpr, collectorIdent("v9"), collectorIdent("v2")),
						),
					),
				)},
			},
			want: false,
		},
		{
			name: "selector call body (t.Run-style) does NOT match",
			seqs: [][]*domain.CloneNode{
				{node(golang.RangeStmt,
					collectorIdent("v1"),
					collectorIdent("v0"),
					node(golang.BlockStmt,
						node(golang.ExprStmt,
							node(golang.CallExpr,
								node(golang.SelectorExpr, collectorIdent("v2"), collectorIdent("Run")),
							),
						),
					),
				)},
			},
			want: false,
		},
		{
			name: "two-argument opt call does NOT match",
			seqs: [][]*domain.CloneNode{
				{node(golang.RangeStmt,
					collectorIdent("v1"),
					collectorIdent("v0"),
					node(golang.BlockStmt,
						node(golang.ExprStmt,
							node(golang.CallExpr, collectorIdent("v1"), collectorIdent("v2"), collectorIdent("v3")),
						),
					),
				)},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			label, action := evaluateActionabilityWithDisabled(tc.seqs, disabledExcept(
				PatternOptionApplyLoop,
			))
			if tc.want && action != domain.NonActionable {
				t.Errorf("expected NonActionable, got %v (label %q)", action, label)
			}

			if !tc.want && label == PatternOptionApplyLoop {
				t.Errorf("pattern matched but should not have")
			}
		})
	}
}

// disabledExcept builds a disabled set covering every pattern except want,
// so the test isolates one pattern even inside the full table.
func disabledExcept(want PatternLabel) map[PatternLabel]bool {
	disabled := make(map[PatternLabel]bool)

	for _, p := range AllActionabilityPatterns() {
		if p != want {
			disabled[p] = true
		}
	}

	return disabled
}
