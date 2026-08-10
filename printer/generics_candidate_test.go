package printer

import (
	"strings"
	"testing"

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
			{Name: "highest", VarType: "int64"},
		},
		{
			{Name: "g", VarType: "string"},
			{Name: "highest", VarType: "int64"},
		},
	}

	isCandidate, hint := ClassifyGenericsCandidate(seqs)

	if !isCandidate {
		t.Error("expected generics candidate when types differ at position 0")
	}

	if hint == "" {
		t.Error("expected non-empty hint")
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
			{Name: "highest"}, // empty VarType
		},
		{
			{Name: "g", VarType: "string"},
			{Name: "highest"}, // empty VarType
		},
	}

	isCandidate, hint := ClassifyGenericsCandidate(seqs)

	if !isCandidate {
		t.Error("expected generics candidate: position 0 has differing types")
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
				},
			},
		},
		{
			{
				Name:    "for",
				VarType: "",
				Children: []*domain.CloneNode{
					{Name: "g", VarType: "db.MemberGrowthPoint"},
				},
			},
		},
	}

	isCandidate, hint := ClassifyGenericsCandidate(seqs)

	if !isCandidate {
		t.Error("expected generics candidate: child nodes have differing types")
	}

	if hint == "" {
		t.Error("expected non-empty hint")
	}
}

func TestClassifyGenericsCandidate_HintFormat(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{{Name: "x", VarType: "[]int"}},
		{{Name: "y", VarType: "[]string"}},
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
		{{Name: "x", VarType: "github.com/larsartmann/erraudit/internal/analyzer.MainAnalyzerOption"}},
		{{Name: "y", VarType: "github.com/larsartmann/erraudit/internal/ast.FileCacheOption"}},
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
