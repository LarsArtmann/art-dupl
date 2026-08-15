package printer

import (
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
)

func TestClassifyGenericsCandidate_IdenticalTypes_NotCandidate(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{
			{Name: "k", VarType: "int64"},
			{Name: "highest", VarType: "int64"},
		},
		{
			{Name: "g", VarType: "int64"},
			{Name: "highest", VarType: "int64"},
		},
	}

	isCandidate, hint := ClassifyGenericsCandidate(seqs)

	if isCandidate {
		t.Errorf("expected NOT a generics candidate when all types match, got candidate with hint: %s", hint)
	}

	if hint != "" {
		t.Errorf("expected empty hint, got: %s", hint)
	}
}

func TestClassifyGenericsCandidate_DifferentTypes_IsCandidate(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{
			{Name: "k", VarType: "int64"},
			{Name: "highest", VarType: "float64"},
		},
		{
			{Name: "g", VarType: "string"},
			{Name: "highest", VarType: "time.Duration"},
		},
	}

	isCandidate, hint := ClassifyGenericsCandidate(seqs)

	if !isCandidate {
		t.Error("expected generics candidate when types differ at 2 positions")
	}

	if hint == "" {
		t.Error("expected non-empty hint")
	}
}

func TestClassifyGenericsCandidate_SingleDivergentPosition_NotCandidate(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{
			{Name: "k", VarType: "int64"},
			{Name: "highest", VarType: "int64"},
		},
		{
			{Name: "g", VarType: "string"},
			{Name: "highest", VarType: "int64"},
		},
	}

	isCandidate, hint := ClassifyGenericsCandidate(seqs)

	if isCandidate {
		t.Errorf(
			"expected NOT a candidate with a single divergent position (below MinDivergentPositions), got hint: %s",
			hint,
		)
	}

	if hint != "" {
		t.Errorf("expected empty hint, got: %s", hint)
	}
}

func TestClassifyGenericsCandidate_SingleSequence_NotCandidate(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{
			{Name: "k", VarType: "int64"},
		},
	}

	isCandidate, _ := ClassifyGenericsCandidate(seqs)

	if isCandidate {
		t.Error("expected NOT a generics candidate with only 1 sequence")
	}
}

func TestClassifyGenericsCandidate_NoVarType_NotCandidate(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{
			{Name: "k"},
			{Name: "highest"},
		},
		{
			{Name: "g"},
			{Name: "highest"},
		},
	}

	isCandidate, _ := ClassifyGenericsCandidate(seqs)

	if isCandidate {
		t.Error("expected NOT a generics candidate when no VarType data (non-type-aware mode)")
	}
}

func TestClassifyGenericsCandidate_EmptyVarTypeSkipped(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{
			{Name: "k", VarType: "int64"},
			{Name: "skipped"}, // empty VarType
			{Name: "highest", VarType: "float64"},
		},
		{
			{Name: "g", VarType: "string"},
			{Name: "skipped"}, // empty VarType
			{Name: "highest", VarType: "time.Duration"},
		},
	}

	isCandidate, hint := ClassifyGenericsCandidate(seqs)

	if !isCandidate {
		t.Error("expected generics candidate: positions 0 and 2 have differing types")
	}

	if hint == "" {
		t.Error("expected non-empty hint")
	}
}

func TestClassifyGenericsCandidate_WithChildren(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{
			{
				Name:    "for",
				VarType: "",
				Children: []*domain.CloneNode{
					{Name: "k", VarType: "db.AuthorKindActivity"},
					{Name: "v", VarType: "db.AuthorLabel"},
				},
			},
		},
		{
			{
				Name:    "for",
				VarType: "",
				Children: []*domain.CloneNode{
					{Name: "g", VarType: "db.MemberGrowthPoint"},
					{Name: "v", VarType: "db.MemberAlias"},
				},
			},
		},
	}

	isCandidate, hint := ClassifyGenericsCandidate(seqs)

	if !isCandidate {
		t.Error("expected generics candidate: child nodes have differing types at 2 positions")
	}

	if hint == "" {
		t.Error("expected non-empty hint")
	}
}

func TestClassifyGenericsCandidate_HintFormat(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{{Name: "x", VarType: "[]int"}, {Name: "x", VarType: "int"}},
		{{Name: "y", VarType: "[]string"}, {Name: "y", VarType: "string"}},
	}

	_, hint := ClassifyGenericsCandidate(seqs)

	if hint == "" {
		t.Fatal("expected non-empty hint")
	}

	if !strings.Contains(hint, "[]int") || !strings.Contains(hint, "[]string") {
		t.Errorf("hint should contain both types, got: %s", hint)
	}
}

func TestShortenTypeString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"int", "int"},
		{"[]int", "[]int"},
		{"*int", "*int"},
		{"github.com/pkg.Type", "pkg.Type"},
		{"[]github.com/pkg.Type", "[]pkg.Type"},
		{"*github.com/pkg.Type", "*pkg.Type"},
		{"github.com/larsartmann/erraudit/internal/analyzer.MainAnalyzerOption", "analyzer.MainAnalyzerOption"},
		{"[]github.com/larsartmann/erraudit/internal/analyzer.MainAnalyzerOption", "[]analyzer.MainAnalyzerOption"},
		{"*github.com/larsartmann/erraudit/internal/analyzer.Main", "*analyzer.Main"},
		{"map[string]github.com/pkg.Type", "map[string]pkg.Type"},
	}

	for _, tc := range tests {
		got := shortenTypeString(tc.input)
		if got != tc.want {
			t.Errorf("shortenTypeString(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestClassifyGenericsCandidate_HintShortensPackagePaths(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{
			{Name: "x", VarType: "github.com/larsartmann/erraudit/internal/analyzer.MainAnalyzerOption"},
			{Name: "x2", VarType: "github.com/larsartmann/erraudit/internal/analyzer.SecondOption"},
		},
		{
			{Name: "y", VarType: "github.com/larsartmann/erraudit/internal/ast.FileCacheOption"},
			{Name: "y2", VarType: "github.com/larsartmann/erraudit/internal/ast.SecondOption"},
		},
	}

	_, hint := ClassifyGenericsCandidate(seqs)

	if hint == "" {
		t.Fatal("expected non-empty hint")
	}

	if strings.Contains(hint, "github.com/") {
		t.Errorf("hint should not contain full package paths, got: %s", hint)
	}

	if !strings.Contains(hint, "analyzer.MainAnalyzerOption") || !strings.Contains(hint, "ast.FileCacheOption") {
		t.Errorf("hint should contain shortened type names, got: %s", hint)
	}
}

// TestFormatGenericsHint_CanonicalizesReversedPairs verifies that A-vs-B and
// B-vs-A divergences (observed from different instance pairings) collapse to a
// single hint entry.
func TestFormatGenericsHint_CanonicalizesReversedPairs(t *testing.T) {
	divs := []TypeDivergence{
		{Position: 0, TypeA: "int64", TypeB: "string"},
		{Position: 1, TypeA: "string", TypeB: "int64"},
	}

	hint := formatGenericsHint(divs)

	if strings.Contains(hint, ";") {
		t.Errorf("reversed pairs should dedup to a single entry, got: %s", hint)
	}
}

// TestGenericsMinLinesDefaultsInSync guards the mirrored default constants:
// config owns the CLI default, printer owns the no-option default. They must
// agree so JSON-config users and library callers see the same gate.
func TestGenericsMinLinesDefaultsInSync(t *testing.T) {
	if config.DefaultSuggestGenericsMinLines != DefaultGenericsMinLines {
		t.Errorf(
			"config.DefaultSuggestGenericsMinLines (%d) != printer.DefaultGenericsMinLines (%d) — keep the mirrored constants in sync",
			config.DefaultSuggestGenericsMinLines,
			DefaultGenericsMinLines,
		)
	}
}
