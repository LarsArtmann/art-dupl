package printer

import (
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
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

// TestProcessClones_ReuseSuggestionNotDoubled (plan T10) pins the no-double-
// suggest contract at the pipeline level: a group whose divergences are all
// instantiations of one indexed generic carries the REUSE suggestion only —
// never the extraction suggestion alongside it, and never an extractability
// verdict that contradicts reuse. Precedence lives in genericsSuggestion:
// reuse wins, extraction is suppressed.
func TestProcessClones_ReuseSuggestionNotDoubled(t *testing.T) {
	t.Parallel()

	instantiationNodes := func(filename, inst string) []*syntax.Node {
		return []*syntax.Node{
			{
				Type:     golang.Ident,
				Name:     "b",
				VarType:  inst,
				Filename: filename,
				Pos:      3,
				End:      13,
			},
			{
				Type:     golang.Ident,
				Name:     "v",
				VarType:  strings.TrimSuffix(strings.TrimSuffix(inst, "int]"), "int") + "x",
				Filename: filename,
				Pos:      6,
				End:      13,
			},
		}
	}

	dups := [][]*syntax.Node{
		instantiationNodes("test_a.go", "github.com/x/pkg.Box[int]"),
		instantiationNodes("test_b.go", "github.com/x/pkg.Box[string]"),
	}

	clones, err := ProcessClones(
		multiLineReadFile, dups,
		WithGenericsMinLines(0), WithGenericDecls(reuseIndex),
	)
	if err != nil {
		t.Fatalf("ProcessClones failed: %v", err)
	}

	for i, c := range clones {
		suggestion := c.Classification.Suggestion

		if !strings.HasPrefix(suggestion, "Reuse existing generic ") {
			t.Errorf("clone %d: expected reuse suggestion, got %q", i, suggestion)
		}

		if strings.Contains(suggestion, "Extract to generic") {
			t.Errorf("clone %d: double suggestion — reuse and extract both present: %q", i, suggestion)
		}

		if strings.Contains(c.Classification.GenericsHint, "Extract to generic") {
			t.Errorf("clone %d: hint carries extraction advice on a reuse group: %q", i, c.Classification.GenericsHint)
		}
	}
}

// TestProcessClones_AlreadyGenericAnnotation (plan T6) pins the
// over-engineering note: candidates that instantiate an indexed shared
// generic but diverge at plain-type positions too (the per-exporter Option
// wrapper class) carry the note in their hint; pure reuse groups and
// plain-type groups do not.
func TestProcessClones_AlreadyGenericAnnotation(t *testing.T) {
	t.Parallel()

	nodes := func(filename, boxInst, exporterType string) []*syntax.Node {
		return []*syntax.Node{
			{
				Type:     golang.Ident,
				Name:     "opt",
				VarType:  boxInst,
				Filename: filename,
				Pos:      3,
				End:      13,
			},
			{
				Type:     golang.Ident,
				Name:     "e",
				VarType:  exporterType,
				Filename: filename,
				Pos:      6,
				End:      13,
			},
		}
	}

	dups := [][]*syntax.Node{
		nodes("test_a.go", "github.com/x/pkg.Box[int]", "*d2.Exporter"),
		nodes("test_b.go", "github.com/x/pkg.Box[string]", "*openapi.Exporter"),
	}

	clones, err := ProcessClones(
		multiLineReadFile, dups,
		WithGenericsMinLines(0), WithGenericDecls(reuseIndex),
	)
	if err != nil {
		t.Fatalf("ProcessClones failed: %v", err)
	}

	for i, c := range clones {
		if !strings.Contains(c.Classification.GenericsHint, "over-engineering") {
			t.Errorf("clone %d: expected already-generic note in hint, got %q", i, c.Classification.GenericsHint)
		}

		if strings.HasPrefix(c.Classification.Suggestion, "Reuse existing generic ") {
			t.Errorf("clone %d: mixed divergence must not get the reuse suggestion", i)
		}
	}

	// Pure reuse groups keep the clean instantiation hint without the note.
	reuseDups := [][]*syntax.Node{
		nodes("test_a.go", "github.com/x/pkg.Box[int]", "int"),
		nodes("test_b.go", "github.com/x/pkg.Box[string]", "string"),
	}

	reuseClones, err := ProcessClones(
		multiLineReadFile, reuseDups,
		WithGenericsMinLines(0), WithGenericDecls(reuseIndex),
	)
	if err != nil {
		t.Fatalf("ProcessClones (reuse) failed: %v", err)
	}

	for i, c := range reuseClones {
		if strings.Contains(c.Classification.GenericsHint, "over-engineering") {
			t.Errorf("reuse clone %d: hint must not carry the note, got %q", i, c.Classification.GenericsHint)
		}
	}
}
