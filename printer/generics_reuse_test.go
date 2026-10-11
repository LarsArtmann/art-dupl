package printer

import (
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
)

// reuseIndex mirrors a run that declared pkg.Box[T any] (one type param) and
// pkg.Pair[K, V] (two type params).
var reuseIndex = map[string]int{
	"github.com/x/pkg.Box":  1,
	"github.com/x/pkg.Pair": 2,
}

func TestClassifyReuseExistingGeneric_AllInstantiationsOfSameGeneric(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{
			{Name: "b", VarType: "github.com/x/pkg.Box[int]"},
			{Name: "v", VarType: "int"},
		},
		{
			{Name: "b", VarType: "github.com/x/pkg.Box[string]"},
			{Name: "v", VarType: "string"},
		},
	}

	declKey, count, reuse := ClassifyReuseExistingGeneric(seqs, reuseIndex)

	if !reuse {
		t.Fatal("expected reuse verdict when every divergence instantiates the same indexed generic")
	}

	if declKey != "github.com/x/pkg.Box" {
		t.Errorf("declKey = %q, want github.com/x/pkg.Box", declKey)
	}

	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestClassifyReuseExistingGeneric_CompositePrefixes(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{
			{Name: "b", VarType: "[]github.com/x/pkg.Box[int]"},
			{Name: "m", VarType: "map[string]*github.com/x/pkg.Pair[int,string]"},
		},
		{
			{Name: "b", VarType: "[]github.com/x/pkg.Box[float64]"},
			{Name: "m", VarType: "map[string]*github.com/x/pkg.Pair[string,bool]"},
		},
	}

	// Two DIFFERENT generics diverge (Box and Pair across positions): the
	// reuse verdict requires every divergence to target the same generic, so
	// this group keeps the extraction suggestion.
	_, _, reuse := ClassifyReuseExistingGeneric(seqs, reuseIndex)

	if reuse {
		t.Error("two distinct generics must not fire the single-generic reuse verdict")
	}
}

func TestClassifyReuseExistingGeneric_PlainTypesNoReuse(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{
			{Name: "k", VarType: "int64"},
			{Name: "h", VarType: "float64"},
		},
		{
			{Name: "g", VarType: "string"},
			{Name: "h", VarType: "time.Duration"},
		},
	}

	if _, _, reuse := ClassifyReuseExistingGeneric(seqs, reuseIndex); reuse {
		t.Error("plain-type divergence must not fire the reuse verdict")
	}
}

func TestClassifyReuseExistingGeneric_EmptyIndexDisabled(t *testing.T) {
	seqs := [][]*domain.CloneNode{
		{
			{Name: "b", VarType: "github.com/x/pkg.Box[int]"},
			{Name: "v", VarType: "int"},
		},
		{
			{Name: "b", VarType: "github.com/x/pkg.Box[string]"},
			{Name: "v", VarType: "string"},
		},
	}

	if _, _, reuse := ClassifyReuseExistingGeneric(seqs, nil); reuse {
		t.Error("nil index must disable the reuse verdict")
	}

	if _, _, reuse := ClassifyReuseExistingGeneric(seqs, map[string]int{}); reuse {
		t.Error("empty index must disable the reuse verdict")
	}
}

func TestClassifyGenericsCandidate_InstantiationOnlyDivergence(t *testing.T) {
	// T8 pin: divergence positions that are pure generic instantiations
	// (Box[int] vs Box[string] at two positions, no other type difference)
	// count toward MinDivergentPositions — the wrapper-function family
	// (same algorithm over different instantiations) is detectable since the
	// instantiation marker puts the instantiated type on the IndexExpr node.
	seqs := [][]*domain.CloneNode{
		{
			{Name: "b", VarType: "pkg.Box[int]"},
			{Name: "out", VarType: "[]pkg.Box[int]"},
		},
		{
			{Name: "b", VarType: "pkg.Box[string]"},
			{Name: "out", VarType: "[]pkg.Box[string]"},
		},
	}

	isCandidate, hint := ClassifyGenericsCandidate(seqs)

	if !isCandidate {
		t.Fatal("instantiation-only divergence at 2 positions must be a candidate")
	}

	// T9: instantiation-only divergence gets the actionable hint phrasing.
	if !strings.Contains(hint, "differs only in generic instantiation") {
		t.Errorf("hint should use the instantiation-only phrasing, got: %s", hint)
	}
}

func TestGenericsSuggestion_ReuseVsExtract(t *testing.T) {
	reuseSeqs := [][]*domain.CloneNode{
		{
			{Name: "b", VarType: "github.com/x/pkg.Box[int]"},
			{Name: "v", VarType: "int"},
		},
		{
			{Name: "b", VarType: "github.com/x/pkg.Box[string]"},
			{Name: "v", VarType: "string"},
		},
	}

	got := genericsSuggestion(processConfig{genericDecls: reuseIndex}, reuseSeqs)
	want := "Reuse existing generic pkg.Box[T1]"

	if got != want {
		t.Errorf("suggestion = %q, want %q", got, want)
	}

	extractSeqs := [][]*domain.CloneNode{
		{
			{Name: "k", VarType: "int64"},
			{Name: "h", VarType: "float64"},
		},
		{
			{Name: "g", VarType: "string"},
			{Name: "h", VarType: "time.Duration"},
		},
	}

	got = genericsSuggestion(processConfig{genericDecls: reuseIndex}, extractSeqs)
	if got != "Extract to generic function" {
		t.Errorf("non-instantiation divergence must keep the extraction suggestion, got %q", got)
	}

	got = genericsSuggestion(processConfig{}, reuseSeqs)
	if got != "Extract to generic function" {
		t.Errorf("missing index must keep the extraction suggestion, got %q", got)
	}
}
