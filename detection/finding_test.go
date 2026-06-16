package detection

import (
	"context"
	"os"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestHasNonEmptyFrag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		frags [][]*syntax.Node
		want  bool
	}{
		{"nil frags", nil, false},
		{"empty outer", [][]*syntax.Node{}, false},
		{"single empty inner", [][]*syntax.Node{{}}, false},
		{"multiple empty inner", [][]*syntax.Node{{}, {}}, false},
		{"single non-empty", [][]*syntax.Node{{&syntax.Node{}}}, true},
		{"mixed", [][]*syntax.Node{{}, {&syntax.Node{}}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := hasNonEmptyFrag(tt.frags); got != tt.want {
				t.Errorf("hasNonEmptyFrag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTodoIssueToFinding(t *testing.T) {
	t.Parallel()

	issue := TodoIssue{
		Filename: "main.go",
		Line:     42,
		Text:     "refactor this",
		Type:     "TODO",
		Tags:     []string{"tech-debt"},
	}

	finding := todoIssueToFinding(issue)

	if finding.Filename != "main.go" {
		t.Errorf("Filename = %q, want 'main.go'", finding.Filename)
	}

	if finding.Line != 42 {
		t.Errorf("Line = %d, want 42", finding.Line)
	}

	if finding.Type != domain.FindingTypeTodo {
		t.Errorf("Type = %q, want %q", finding.Type, domain.FindingTypeTodo)
	}

	if finding.Message != "refactor this" {
		t.Errorf("Message = %q, want 'refactor this'", finding.Message)
	}

	if finding.Priority != domain.PriorityLow {
		t.Errorf("Priority = %q, want %q", finding.Priority, domain.PriorityLow)
	}
}

func TestLegacyIssueToFinding(t *testing.T) {
	t.Parallel()

	issue := LegacyIssue{
		Filename: "old.go",
		Line:     10,
		Type:     "deprecated",
		Message:  "uses deprecated API",
		Severity: domain.PriorityHigh,
	}

	finding := legacyIssueToFinding(issue)

	if finding.Type != domain.FindingTypeLegacy {
		t.Errorf("Type = %q, want %q", finding.Type, domain.FindingTypeLegacy)
	}

	if finding.Priority != domain.PriorityHigh {
		t.Errorf("Priority = %q, want %q", finding.Priority, domain.PriorityHigh)
	}
}

func TestMultiDetector_FindFindings_TodosMethod(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := tmpDir + "/test.go"

	goCode := `package test

// TODO: fix this
func Foo() {}
`

	err := os.WriteFile(testFile, []byte(goCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tree := suffixtree.New()
	data := []*syntax.Node{createTestNode(testFile, 1, 100)}

	detector := NewMultiDetector(
		config.DetectionConfig{Methods: config.DetectionMethods{config.DetectionMethodTodos}},
		data,
		tree,
	)

	findingCount := 0

	for finding := range detector.FindFindings(context.Background()) {
		findingCount++

		if finding.Type != domain.FindingTypeTodo {
			t.Errorf("Type = %q, want %q", finding.Type, domain.FindingTypeTodo)
		}

		if finding.Filename == "" {
			t.Error("Filename should not be empty")
		}
	}

	if findingCount == 0 {
		t.Error("Expected at least one TODO finding")
	}
}
