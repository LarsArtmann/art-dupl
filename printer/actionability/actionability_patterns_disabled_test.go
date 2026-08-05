package actionability

import (
	"slices"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestAllActionabilityPatterns_Count(t *testing.T) {
	t.Parallel()

	patterns := AllActionabilityPatterns()

	want := len(actionabilityPatternTable) + len(propertyLabels)
	if len(patterns) != want {
		t.Errorf("AllActionabilityPatterns() returned %d patterns, want %d (denylist + property labels)",
			len(patterns), want)
	}
}

func TestAllActionabilityPatterns_ContainsLabels(t *testing.T) {
	t.Parallel()

	patterns := AllActionabilityPatterns()

	required := []PatternLabel{
		PatternSignatureOnly,
		PatternInterfaceImpl,
		PatternInterfaceMethod,
		PatternRAIIDefer,
		PatternErrorPropagation,
		PatternGuardClause,
		PatternAssignErrorCheck,
		PatternSingleCallExpr,
		PatternSingleSimpleStmt,
		PatternBoolAccumulatorInitializer,
		PatternSingleDeclaration,
		PatternTestHelperDelegate,
		PatternErrorWrapping,
		PatternAssertionChain,
		PatternCobraBoilerplate,
		PatternTestData,
		PatternTableDrivenTest,
		PatternTestScaffolding,
		PatternDataDominated,
		PatternDescribeTable,
		PatternBuilderCallback,
		PatternBoolGuard,
		PatternTemplRenderingIdiom,
	}

	for _, label := range required {
		if !slices.Contains(patterns, label) {
			t.Errorf("AllActionabilityPatterns() missing pattern %q", label)
		}
	}
}

func TestListActionabilityPatterns(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	ListActionabilityPatterns(&buf)

	output := buf.String()
	patterns := AllActionabilityPatterns()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != len(patterns) {
		t.Errorf("ListActionabilityPatterns wrote %d lines, want %d", len(lines), len(patterns))
	}

	for i, p := range patterns {
		if i >= len(lines) {
			break
		}

		if lines[i] != string(p) {
			t.Errorf("pattern %q: got line %q", p, lines[i])
		}
	}
}

func TestEvaluateActionabilityWithDisabled(t *testing.T) {
	t.Parallel()

	// A guard clause pattern: single IfStmt with return-only body, no else
	guardClause := mustGuardClauseSeq()
	seqs := [][]*domain.CloneNode{guardClause, guardClause}

	t.Run("without disable - suppressed", func(t *testing.T) {
		t.Parallel()

		result := EvaluateActionabilityWithDisabled(seqs, nil)
		if result != domain.NonActionable {
			t.Errorf("expected NonActionable, got %q", result)
		}
	})

	t.Run("disable guard-clause - becomes actionable", func(t *testing.T) {
		t.Parallel()

		disabled := map[PatternLabel]bool{PatternGuardClause: true}

		result := EvaluateActionabilityWithDisabled(seqs, disabled)
		if result != domain.Actionable {
			t.Errorf("expected Actionable after disabling guard-clause, got %q", result)
		}
	})

	t.Run("disable unrelated pattern - still suppressed", func(t *testing.T) {
		t.Parallel()

		disabled := map[PatternLabel]bool{PatternRAIIDefer: true}

		result := EvaluateActionabilityWithDisabled(seqs, disabled)
		if result != domain.NonActionable {
			t.Errorf("expected NonActionable (guard-clause not disabled), got %q", result)
		}
	})
}

func mustGuardClauseSeq() []*domain.CloneNode {
	return []*domain.CloneNode{
		{
			BaseType:             golang.IfStmt,
			EnclosingReturnArity: 1, // value-returning function — guard clause is extractable
			Children: []*domain.CloneNode{
				{
					BaseType: golang.BlockStmt,
					Children: []*domain.CloneNode{
						{BaseType: golang.ReturnStmt},
					},
				},
			},
		},
	}
}

func mustBoolAccumulatorSeq() []*domain.CloneNode {
	return []*domain.CloneNode{
		{
			BaseType: golang.AssignStmt,
			Children: []*domain.CloneNode{
				{BaseType: golang.Ident, Name: "false"},
				{BaseType: golang.Ident, Name: "hasX"},
			},
		},
		{
			BaseType: golang.AssignStmt,
			Children: []*domain.CloneNode{
				{BaseType: golang.Ident, Name: "false"},
				{BaseType: golang.Ident, Name: "hasY"},
			},
		},
	}
}

func TestEvaluateActionabilityWithDisabled_BoolAccumulator(t *testing.T) {
	t.Parallel()

	seqs := [][]*domain.CloneNode{mustBoolAccumulatorSeq(), mustBoolAccumulatorSeq()}

	t.Run("without disable - suppressed", func(t *testing.T) {
		t.Parallel()

		result := EvaluateActionabilityWithDisabled(seqs, nil)
		if result != domain.NonActionable {
			t.Errorf("expected NonActionable, got %q", result)
		}
	})

	t.Run("disable bool-accumulator-initializer - becomes actionable", func(t *testing.T) {
		t.Parallel()

		disabled := map[PatternLabel]bool{PatternBoolAccumulatorInitializer: true}

		result := EvaluateActionabilityWithDisabled(seqs, disabled)
		if result != domain.Actionable {
			t.Errorf("expected Actionable after disabling bool-accumulator-initializer, got %q", result)
		}
	})
}
