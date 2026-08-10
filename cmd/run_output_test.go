package cmd

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestShouldSuppressGroup_MinLines(t *testing.T) {
	t.Parallel()

	cloneWithLines := func(start, end int) domain.ProcessedClone {
		return domain.ProcessedClone{
			CloneRef: domain.CloneRef{
				LineStart: start,
				LineEnd:   end,
			},
		}
	}

	tests := []struct {
		name     string
		group    domain.ProcessedCloneGroup
		minLines int
		expected bool
	}{
		{
			name:     "minLines=0 disables filter",
			group:    domain.ProcessedCloneGroup{Clones: []domain.ProcessedClone{cloneWithLines(1, 2)}},
			minLines: 0,
			expected: false,
		},
		{
			name:     "clone shorter than minLines is suppressed",
			group:    domain.ProcessedCloneGroup{Clones: []domain.ProcessedClone{cloneWithLines(1, 2)}},
			minLines: 5,
			expected: true,
		},
		{
			name:     "clone exactly at minLines is not suppressed",
			group:    domain.ProcessedCloneGroup{Clones: []domain.ProcessedClone{cloneWithLines(1, 5)}},
			minLines: 5,
			expected: false,
		},
		{
			name:     "clone longer than minLines is not suppressed",
			group:    domain.ProcessedCloneGroup{Clones: []domain.ProcessedClone{cloneWithLines(1, 10)}},
			minLines: 5,
			expected: false,
		},
		{
			name: "multiple clones: minimum determines suppression",
			group: domain.ProcessedCloneGroup{Clones: []domain.ProcessedClone{
				cloneWithLines(1, 10),
				cloneWithLines(1, 2),
				cloneWithLines(1, 8),
			}},
			minLines: 5,
			expected: true,
		},
		{
			name: "multiple clones: all above minLines not suppressed",
			group: domain.ProcessedCloneGroup{Clones: []domain.ProcessedClone{
				cloneWithLines(1, 10),
				cloneWithLines(1, 8),
				cloneWithLines(1, 6),
			}},
			minLines: 5,
			expected: false,
		},
		{
			name:     "empty clones not suppressed",
			group:    domain.ProcessedCloneGroup{Clones: nil},
			minLines: 5,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := shouldSuppressGroup(tc.group, SuppressionConfig{MinLines: tc.minLines})
			if result != tc.expected {
				t.Errorf("shouldSuppressGroup() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestMinCloneLineCount(t *testing.T) {
	t.Parallel()

	cloneWithLines := func(start, end int) domain.ProcessedClone {
		return domain.ProcessedClone{
			CloneRef: domain.CloneRef{LineStart: start, LineEnd: end},
		}
	}

	tests := []struct {
		name     string
		group    domain.ProcessedCloneGroup
		expected int
	}{
		{name: "empty group returns 0", group: domain.ProcessedCloneGroup{}, expected: 0},
		{
			name:     "single clone",
			group:    domain.ProcessedCloneGroup{Clones: []domain.ProcessedClone{cloneWithLines(1, 5)}},
			expected: 5,
		},
		{
			name: "multiple clones returns minimum",
			group: domain.ProcessedCloneGroup{Clones: []domain.ProcessedClone{
				cloneWithLines(1, 10),
				cloneWithLines(1, 3),
				cloneWithLines(1, 7),
			}},
			expected: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := minCloneLineCount(tc.group)
			if result != tc.expected {
				t.Errorf("minCloneLineCount() = %d, want %d", result, tc.expected)
			}
		})
	}
}

// countingPrinter tracks how many times PrintClones is called and captures
// suppression stats. Used to verify that --suggest-generics does NOT filter
// clone groups (enhancer, not filter).
type countingPrinter struct {
	printedGroups int
	stats         printer.SuppressionStats
}

func (c *countingPrinter) PrintHeader() error { return nil }

func (c *countingPrinter) PrintClones(_ domain.ProcessedCloneGroup, _ ...config.SortCriteria) error {
	c.printedGroups++
	return nil
}

func (c *countingPrinter) PrintFooter() error { return nil }

func (c *countingPrinter) SetSuppressionStats(stats printer.SuppressionStats) {
	c.stats = stats
}

var (
	_ printer.Printer                = (*countingPrinter)(nil)
	_ printer.SuppressionStatsSetter = (*countingPrinter)(nil)
)

func noopReadFileForOutputTest(filename string) ([]byte, error) {
	return []byte("package p\n"), nil
}

// TestPrintCloneGroups_ShowsAllClonesRegardlessOfGenericsCandidate verifies
// that printCloneGroups shows ALL clone groups — both generics candidates and
// non-candidates. This is the core enhancer behavior: --suggest-generics
// annotates rather than filters.
func TestPrintCloneGroups_ShowsAllClonesRegardlessOfGenericsCandidate(t *testing.T) {
	t.Parallel()

	// Group A: generics candidate (different VarType across instances)
	groupA := [][]*syntax.Node{
		{
			{Type: golang.Ident, Name: "k", VarType: "int64", Filename: "a.go", Pos: 1, End: 2},
		},
		{
			{Type: golang.Ident, Name: "g", VarType: "string", Filename: "b.go", Pos: 1, End: 2},
		},
	}

	// Group B: non-generics candidate (same VarType)
	groupB := [][]*syntax.Node{
		{
			{Type: golang.Ident, Name: "x", VarType: "int", Filename: "c.go", Pos: 1, End: 2},
		},
		{
			{Type: golang.Ident, Name: "y", VarType: "int", Filename: "d.go", Pos: 1, End: 2},
		},
	}

	groups := map[string][][]*syntax.Node{
		"hash-a": groupA,
		"hash-b": groupB,
	}
	keys := []string{"hash-a", "hash-b"}

	cp := &countingPrinter{}

	err := printCloneGroups(
		cp,
		noopReadFileForOutputTest,
		groups,
		keys,
		config.SortBySize,
		true,                // semantic
		SuppressionConfig{}, // no suppression at all
	)
	if err != nil {
		t.Fatalf("printCloneGroups failed: %v", err)
	}

	if cp.printedGroups != 2 {
		t.Errorf("expected 2 groups printed (enhancer mode shows all), got %d", cp.printedGroups)
	}

	if cp.stats.Shown != 2 {
		t.Errorf("expected stats.Shown=2, got %d", cp.stats.Shown)
	}
}
