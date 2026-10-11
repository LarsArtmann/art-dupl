package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// Collector-tail and option-apply-loop pattern tests. Shapes mirror the
// measured false-positive families from the 2026-10-11 precision re-measure
// (docs/research/2026-10-11_suggest-generics-precision-baseline.md ledger).
//
// Hand-built builders serialize assignments in the REAL [rhs, lhs] child
// order (syntax/golang/transform.go) — the first version of these tests
// built target-first trees and masked a matcher that read positions
// backwards (see the child-order contract in actionability_collector.go).
// The TestCollectorTailRealSource cases parse actual Go statements as
// further insurance against hand-built/serialized divergence.

func node(baseType int32, children ...*domain.CloneNode) *domain.CloneNode {
	return &domain.CloneNode{BaseType: baseType, Children: children}
}

func collectorIdent(name string) *domain.CloneNode {
	return &domain.CloneNode{BaseType: golang.Ident, Name: name}
}

func appendAssign(target, value string) *domain.CloneNode {
	return node(golang.AssignStmt,
		node(golang.CallExpr, collectorIdent("append"), collectorIdent(target), collectorIdent(value)),
		collectorIdent(target),
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

func cloneNodeFrom(n *syntax.Node) *domain.CloneNode {
	cn := &domain.CloneNode{
		Name:     n.Name,
		BaseType: golang.DecodeBaseType(n.Type),
	}

	if len(n.Children) > 0 {
		cn.Children = make([]*domain.CloneNode, len(n.Children))
		for i, c := range n.Children {
			cn.Children[i] = cloneNodeFrom(c)
		}
	}

	return cn
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
			name: "append loop with make prealloc and return matches",
			seqs: [][]*domain.CloneNode{
				{
					node(golang.AssignStmt,
						node(golang.CallExpr,
							collectorIdent("make"),
							collectorIdent("v4"),
							node(golang.BasicLit),
							node(golang.CallExpr, collectorIdent("len"), collectorIdent("v0")),
						),
						collectorIdent("v2"),
					),
					collectorLoop(appendAssign("v2", "v1")),
					node(golang.ReturnStmt, collectorIdent("v2")),
				},
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
			name: "deep-copy loop matches (deref binding + clone field copies)",
			seqs: [][]*domain.CloneNode{
				{collectorLoop(
					node(golang.AssignStmt,
						node(golang.StarExpr, collectorIdent("v1")),
						collectorIdent("v3"),
					),
					node(golang.AssignStmt,
						node(golang.CallExpr,
							node(golang.SelectorExpr, collectorIdent("slices"), collectorIdent("Clone")),
							node(golang.SelectorExpr, collectorIdent("v1"), collectorIdent("Content")),
						),
						node(golang.SelectorExpr, collectorIdent("v3"), collectorIdent("Content")),
					),
					appendAssign("v2", "v3"),
				)},
			},
			want: true,
		},
		{
			name: "plain field copy without clone call matches",
			seqs: [][]*domain.CloneNode{
				{collectorLoop(
					node(golang.AssignStmt,
						node(golang.SelectorExpr, collectorIdent("v1"), collectorIdent("Name")),
						node(golang.SelectorExpr, collectorIdent("v3"), collectorIdent("Name")),
					),
					appendAssign("v2", "v3"),
				)},
			},
			want: true,
		},
		{
			name: "append of constructed composite does NOT match",
			seqs: [][]*domain.CloneNode{
				{collectorLoop(node(golang.AssignStmt,
					node(golang.CallExpr, collectorIdent("append"), collectorIdent("v2"), node(golang.CompositeLit)),
					collectorIdent("v2"),
				))},
			},
			want: false,
		},
		{
			name: "append to different slice does NOT match",
			seqs: [][]*domain.CloneNode{
				{collectorLoop(node(golang.AssignStmt,
					node(golang.CallExpr, collectorIdent("append"), collectorIdent("v5"), collectorIdent("v1")),
					collectorIdent("v2"),
				))},
			},
			want: false,
		},
		{
			name: "loop with real processing call does NOT match",
			seqs: [][]*domain.CloneNode{
				{collectorLoop(node(golang.AssignStmt,
					node(golang.CallExpr, collectorIdent("normalize"), collectorIdent("v1")),
					collectorIdent("v2"),
				))},
				{collectorLoop(node(golang.AssignStmt,
					node(golang.CallExpr, collectorIdent("normalize"), collectorIdent("v1")),
					collectorIdent("v2"),
				))},
			},
			want: false,
		},
		{
			name: "clone call on non-stdlib receiver does NOT match",
			seqs: [][]*domain.CloneNode{
				{collectorLoop(
					node(golang.AssignStmt,
						node(golang.CallExpr,
							node(golang.SelectorExpr, collectorIdent("v1"), collectorIdent("Clone")),
						),
						collectorIdent("v3"),
					),
					appendAssign("v2", "v3"),
				)},
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
			name: "guard-only sequence without accumulation does NOT match",
			seqs: [][]*domain.CloneNode{
				{node(golang.IfStmt,
					node(golang.BinaryExpr),
					node(golang.ReturnStmt),
				)},
				{node(golang.IfStmt,
					node(golang.BinaryExpr),
					node(golang.ReturnStmt),
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

// TestCollectorTailRealSource pins the pattern against REAL serialized
// trees: each case parses actual Go statements through the golang
// transformer (the same path detection uses), so child ordering, node
// shapes, and normalization all match production.
func TestCollectorTailRealSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want bool
	}{
		{
			name: "deep-copy loop with deref matches (go-paperless documents shape)",
			src: `for _, doc := range s.documents {
	copied := *doc
	copied.Content = slices.Clone(doc.Content)
	documents = append(documents, copied)
}`,
			want: true,
		},
		{
			name: "deep-copy loop with value binding matches (go-paperless patches shape)",
			src: `for _, patch := range s.patches {
	copied := patch
	copied.Body = slices.Clone(patch.Body)
	patches = append(patches, copied)
}`,
			want: true,
		},
		{
			name: "plain filter loop matches",
			src: `for _, x := range xs {
	if x.Active {
		out = append(out, x)
	}
}`,
			want: true,
		},
		{
			name: "append loop with make prealloc matches",
			src: `kept := make([]T, 0, len(items))
for _, item := range items {
	kept = append(kept, item)
}
return kept`,
			want: true,
		},
		{
			name: "predicate filter with skip-guards does NOT match (deliberate: mining verdict)",
			src: `for _, item := range items {
	if rule, hasRule := ruleID(item); hasRule {
		if _, off := disabled[strings.ToLower(rule)]; off {
			continue
		}
	}

	kept = append(kept, item)
}`,
			want: false,
		},
		{
			name: "transform loop does NOT match",
			src: `for _, u := range users {
	out = append(out, normalize(u))
}`,
			want: false,
		},
		{
			name: "clone method on domain receiver does NOT match",
			src: `for _, doc := range docs {
	copied := doc.Clone()
	out = append(out, copied)
}`,
			want: false,
		},
		{
			name: "option-apply loop matches",
			src: `for _, opt := range opts {
	opt(s)
}
return s`,
			want: false, // matches option-apply-loop, NOT append-collector-tail
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			nodes, err := golang.ParseSnippet(
				golang.SnippetStatements, tc.src, "fixture.go", 0, golang.DetectionModeSemantic,
			)
			if err != nil {
				t.Fatalf("ParseSnippet error: %v", err)
			}

			if len(nodes) == 0 {
				t.Fatal("ParseSnippet returned no nodes")
			}

			seq := make([]*domain.CloneNode, len(nodes))
			for i, n := range nodes {
				seq[i] = cloneNodeFrom(n)
			}

			label, action := evaluateActionabilityWithDisabled([][]*domain.CloneNode{seq}, disabledExcept(
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

// TestOptionApplyLoopRealSource pins the option-apply pattern against real
// serialized statements (the corpus-dominant false class).
func TestOptionApplyLoopRealSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want bool
	}{
		{
			name: "bare apply loop with return matches",
			src: `for _, opt := range opts {
	opt(s)
}
return s`,
			want: true,
		},
		{
			name: "apply loop without return matches",
			src: `for _, option := range options {
	option(cfg)
}`,
			want: true,
		},
		{
			name: "constructor with defaulting after loop does NOT match",
			src: `for _, opt := range opts {
	opt(s)
}
if s.Name == "" {
	s.Name = defaultName
}
return s`,
			want: false,
		},
		{
			name: "apply loop on selector receiver does NOT match",
			src: `for _, opt := range opts {
	opt.Apply(s)
}`,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			nodes, err := golang.ParseSnippet(
				golang.SnippetStatements, tc.src, "fixture.go", 0, golang.DetectionModeSemantic,
			)
			if err != nil {
				t.Fatalf("ParseSnippet error: %v", err)
			}

			if len(nodes) == 0 {
				t.Fatal("ParseSnippet returned no nodes")
			}

			seq := make([]*domain.CloneNode, len(nodes))
			for i, n := range nodes {
				seq[i] = cloneNodeFrom(n)
			}

			label, action := evaluateActionabilityWithDisabled([][]*domain.CloneNode{seq}, disabledExcept(
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
