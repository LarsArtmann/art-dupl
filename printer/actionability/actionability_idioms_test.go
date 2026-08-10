package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestIsDeferCallPattern(t *testing.T) {
	t.Run("defer bare function call is non-actionable", func(t *testing.T) {
		seqs := makeDeferBareCallSeqs("unsubscribe", 2)
		assertPatternMatch(t, "isDeferCallPattern", seqs, true)
	})

	t.Run("defer method call is NOT matched (business logic)", func(t *testing.T) {
		seqs := makeDeferSelectorCallSeqs("svc", "processOrder", 2)
		assertPatternMatch(t, "isDeferCallPattern", seqs, false)
	})

	t.Run("two-node sequence is NOT matched", func(t *testing.T) {
		seqs := [][]*domain.CloneNode{
			{mustDeferBareCall("cancel"), mustDeferBareCall("cleanup")},
			{mustDeferBareCall("cancel"), mustDeferBareCall("cleanup")},
		}
		assertPatternMatch(t, "isDeferCallPattern", seqs, false)
	})
}

func TestIsTestFrameworkCallPattern(t *testing.T) {
	t.Run("t.Parallel() is non-actionable", func(t *testing.T) {
		seqs := makeTestFrameworkCallSeqs("t", "Parallel", 2)
		assertPatternMatch(t, "isTestFrameworkCallPattern", seqs, true)
	})

	t.Run("b.Helper() is non-actionable", func(t *testing.T) {
		seqs := makeTestFrameworkCallSeqs("b", "Helper", 2)
		assertPatternMatch(t, "isTestFrameworkCallPattern", seqs, true)
	})

	t.Run("non-test-framework call is NOT matched", func(t *testing.T) {
		seqs := makeTestFrameworkCallSeqs("svc", "Process", 2)
		assertPatternMatch(t, "isTestFrameworkCallPattern", seqs, false)
	})
}

func TestIsStateFlagMutation(t *testing.T) {
	t.Run("struct field assignment is non-actionable", func(t *testing.T) {
		seqs := makeStateFlagMutationSeqs(2)
		assertPatternMatch(t, "isStateFlagMutation", seqs, true)
	})

	t.Run("non-field assignment is NOT matched", func(t *testing.T) {
		node := &domain.CloneNode{
			BaseType: golang.AssignStmt,
			Children: []*domain.CloneNode{
				{BaseType: golang.Ident, Name: "x"},
				{BaseType: golang.BasicLit, Name: "1"},
			},
		}
		seqs := [][]*domain.CloneNode{{node}, {node}}
		assertPatternMatch(t, "isStateFlagMutation", seqs, false)
	})
}

func TestIsEmptyDefault(t *testing.T) {
	t.Run("if x == empty string { x = default } is non-actionable", func(t *testing.T) {
		seqs := makeEmptyDefaultSeqs(2)
		assertPatternMatch(t, "isEmptyDefault", seqs, true)
	})

	t.Run("non-empty-string IfStmt is NOT matched", func(t *testing.T) {
		seqs := [][]*domain.CloneNode{
			{mustIfErrReturnNil()},
			{mustIfErrReturnNil()},
		}
		assertPatternMatch(t, "isEmptyDefault", seqs, false)
	})
}

// --- Helpers ---

func mustDeferBareCall(funcName string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.DeferStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.CallExpr,
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: funcName},
				},
			},
		},
	}
}

func makeDeferBareCallSeqs(funcName string, count int) [][]*domain.CloneNode {
	node := mustDeferBareCall(funcName)

	seqs := make([][]*domain.CloneNode, count)
	for i := range seqs {
		seqs[i] = []*domain.CloneNode{node}
	}

	return seqs
}

func makeDeferSelectorCallSeqs(receiver, method string, count int) [][]*domain.CloneNode {
	node := mustDeferSelectorCall(receiver, method)

	seqs := make([][]*domain.CloneNode, count)
	for i := range seqs {
		seqs[i] = []*domain.CloneNode{node}
	}

	return seqs
}

func makeTestFrameworkCallSeqs(receiver, method string, count int) [][]*domain.CloneNode {
	node := &domain.CloneNode{
		BaseType: golang.ExprStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.CallExpr,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.SelectorExpr,
						Name:     method,
						Children: []*domain.CloneNode{
							{BaseType: golang.Ident, Name: receiver},
						},
					},
				},
			},
		},
	}

	seqs := make([][]*domain.CloneNode, count)
	for i := range seqs {
		seqs[i] = []*domain.CloneNode{node}
	}

	return seqs
}

func makeStateFlagMutationSeqs(count int) [][]*domain.CloneNode {
	node := &domain.CloneNode{
		BaseType: golang.AssignStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.SelectorExpr,
				Name:     "flag",
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "w"},
				},
			},
			{BaseType: golang.Ident, Name: "true"},
		},
	}

	seqs := make([][]*domain.CloneNode, count)
	for i := range seqs {
		seqs[i] = []*domain.CloneNode{node}
	}

	return seqs
}

func makeEmptyDefaultSeqs(count int) [][]*domain.CloneNode {
	node := &domain.CloneNode{
		BaseType: golang.IfStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.BinaryExpr,
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "x"},
					{BaseType: golang.BasicLit, Name: `""`},
				},
			},
			{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.AssignStmt,
						Children: []*domain.CloneNode{
							{BaseType: golang.Ident, Name: "x"},
							{BaseType: golang.Ident, Name: "default"},
						},
					},
				},
			},
		},
	}

	seqs := make([][]*domain.CloneNode, count)
	for i := range seqs {
		seqs[i] = []*domain.CloneNode{node}
	}

	return seqs
}

func assertPatternMatch(t *testing.T, pattern string, seqs [][]*domain.CloneNode, expected bool) {
	t.Helper()

	var result bool

	switch pattern {
	case "isDeferCallPattern":
		result = isDeferCallPattern(seqs)
	case "isTestFrameworkCallPattern":
		result = isTestFrameworkCallPattern(seqs)
	case "isStateFlagMutation":
		result = isStateFlagMutation(seqs)
	case "isEmptyDefault":
		result = isEmptyDefault(seqs)
	default:
		t.Fatalf("unknown pattern: %s", pattern)
	}

	if result != expected {
		t.Errorf("%s: expected %v, got %v", pattern, expected, result)
	}
}
