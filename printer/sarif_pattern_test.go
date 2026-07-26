package printer

import (
	"bytes"
	"encoding/json/v2"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
)

func TestSARIFOutput_NonActionablePattern(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	printer := NewSARIF(&buf, mockSARIFReadFile, 15).(*sarifPrinter)
	printer.SetHash("test-hash-abc")

	group := domain.ProcessedCloneGroup{
		Hash: "test-hash-abc",
		Clones: []domain.ProcessedClone{
			{
				CloneRef: domain.CloneRef{
					Filename:  "a.go",
					LineStart: 10,
					LineEnd:   20,
					Fragment:  "test code",
				},
				Classification: domain.CloneClassification{
					CloneType:            domain.CloneType2,
					Category:             domain.CategoryFunction,
					NonActionablePattern: "guard-clause",
				},
			},
		},
	}

	if err := printer.PrintClones(group); err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	if err := printer.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	var output SARIFOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to unmarshal SARIF: %v", err)
	}

	if len(output.Runs) == 0 || len(output.Runs[0].Results) == 0 {
		t.Fatal("No SARIF results")
	}

	result := output.Runs[0].Results[0]
	pattern, ok := result.Properties["non_actionable_pattern"]
	if !ok {
		t.Error("SARIF result missing 'non_actionable_pattern' property")
	}

	if pattern != "guard-clause" {
		t.Errorf("non_actionable_pattern = %q, want %q", pattern, "guard-clause")
	}

	category, ok := result.Properties["category"]
	if !ok {
		t.Error("SARIF result missing 'category' property")
	}

	if category != "function" {
		t.Errorf("category = %q, want %q", category, "function")
	}
}

func TestSARIFOutput_NoPatternWhenActionable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	printer := NewSARIF(&buf, mockSARIFReadFile, 15).(*sarifPrinter)
	printer.SetHash("test-hash-xyz")

	group := domain.ProcessedCloneGroup{
		Hash: "test-hash-xyz",
		Clones: []domain.ProcessedClone{
			{
				CloneRef: domain.CloneRef{
					Filename:  "b.go",
					LineStart: 1,
					LineEnd:   5,
					Fragment:  "test code",
				},
				Classification: domain.CloneClassification{
					CloneType:            domain.CloneType1,
					Category:             domain.CategoryFunction,
					NonActionablePattern: "",
				},
			},
		},
	}

	if err := printer.PrintClones(group); err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	if err := printer.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	var output SARIFOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to unmarshal SARIF: %v", err)
	}

	result := output.Runs[0].Results[0]
	if _, ok := result.Properties["non_actionable_pattern"]; ok {
		t.Error("SARIF result should NOT have 'non_actionable_pattern' when actionable")
	}
}
