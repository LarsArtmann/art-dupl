package domain_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestDomainCloneValidation(t *testing.T) {
	t.Parallel()

	t.Run("should accept valid clones", func(t *testing.T) {
		clone := domain.Clone{
			ID:         "clone-1",
			Filename:   "test.go",
			StartLine:  10,
			EndLine:    20,
			StartPos:   100,
			EndPos:     200,
			Fragment:   "test code",
			Hash:       "abc123",
			Confidence: 0.95,
			Complexity: 5,
			Status:     domain.FileProcessingStateCompleted,
		}

		if err := clone.IsValid(); err != nil {
			t.Errorf("Expected valid clone, got error: %v", err)
		}
	})

	t.Run("should reject clones with empty ID", func(t *testing.T) {
		clone := domain.Clone{
			Filename:  "test.go",
			StartLine: 10,
			EndLine:   20,
		}

		if err := clone.IsValid(); err == nil {
			t.Error("Expected error for empty clone ID, got nil")
		}
	})

	t.Run("should reject clones with invalid position", func(t *testing.T) {
		clone := domain.Clone{
			ID:        "clone-1",
			Filename:  "test.go",
			StartLine: 10,
			EndLine:   20,
			StartPos:  200,
			EndPos:    100, // Invalid: end < start
		}

		if err := clone.IsValid(); err == nil {
			t.Error("Expected error for invalid position, got nil")
		}
	})

	t.Run("should reject clones with invalid confidence", func(t *testing.T) {
		clone := domain.Clone{
			ID:         "clone-1",
			Filename:   "test.go",
			Confidence: 1.5, // Invalid: > 1.0
		}

		if err := clone.IsValid(); err == nil {
			t.Error("Expected error for invalid confidence, got nil")
		}
	})
}

func TestDomainNodeToClone(t *testing.T) {
	t.Parallel()

	t.Run("should create valid clones from nodes", func(t *testing.T) {
		node := &syntax.Node{
			Filename: "test.go",
			Pos:      50,
			End:      150,
		}

		fileContent := []byte("line 1\nline 2\nline 3\nline 4\nline 5\nfunc test() {}\nline 7\nline 8\nline 9\nline 10\nline 11\nline 12\nline 13\nline 14\nline 15\nline 16")

		clone := domain.NodeToClone(node, "test.go", fileContent)
		if clone.Filename != "test.go" {
			t.Errorf("Expected filename 'test.go', got %s", clone.Filename)
		}
		if clone.StartPos != 50 {
			t.Errorf("Expected start pos 50, got %d", clone.StartPos)
		}
		if clone.EndPos != 150 {
			t.Errorf("Expected end pos 150, got %d", clone.EndPos)
		}
		if clone.Status != domain.FileProcessingStateCompleted {
			t.Errorf("Expected status completed, got %v", clone.Status)
		}
		if err := clone.IsValid(); err != nil {
			t.Errorf("Expected valid clone, got error: %v", err)
		}
	})
}

func TestDomainCloneGroupValidation(t *testing.T) {
	t.Parallel()

	t.Run("should accept valid clone groups", func(t *testing.T) {
		group := domain.CloneGroup{
			ID:   "group-1",
			Hash: "abc123",
			Size: 100,
			Clones: []domain.Clone{
				{
					ID:        "clone-1",
					Filename:  "test1.go",
					StartLine: 10,
					EndLine:   20,
					Status:    domain.FileProcessingStateCompleted,
				},
				{
					ID:        "clone-2",
					Filename:  "test2.go",
					StartLine: 15,
					EndLine:   25,
					Status:    domain.FileProcessingStateCompleted,
				},
			},
			Severity: domain.CloneSeverityMedium,
			Status:   domain.FileProcessingStateCompleted,
		}

		if err := group.IsValid(); err != nil {
			t.Errorf("Expected valid clone group, got error: %v", err)
		}
	})

	t.Run("should reject groups with empty clones", func(t *testing.T) {
		group := domain.CloneGroup{
			ID:       "group-1",
			Hash:     "abc123",
			Size:     100,
			Clones:   []domain.Clone{},
			Severity: domain.CloneSeverityMedium,
		}

		if err := group.IsValid(); err == nil {
			t.Error("Expected error for empty clones, got nil")
		}
	})

	t.Run("should reject groups with invalid severity", func(t *testing.T) {
		group := domain.CloneGroup{
			ID:   "group-1",
			Hash: "abc123",
			Size: 100,
			Clones: []domain.Clone{
				{
					ID:        "clone-1",
					Filename:  "test.go",
					StartLine: 10,
					EndLine:   20,
					Status:    domain.FileProcessingStateCompleted,
				},
			},
			Severity: domain.CloneSeverity("invalid"),
		}

		if err := group.IsValid(); err == nil {
			t.Error("Expected error for invalid severity, got nil")
		}
	})

	t.Run("should propagate clone validation errors", func(t *testing.T) {
		group := domain.CloneGroup{
			ID:   "group-1",
			Hash: "abc123",
			Size: 100,
			Clones: []domain.Clone{
				{
					// Invalid: empty ID
					Filename:  "test.go",
					StartLine: 10,
					EndLine:   20,
					Status:    domain.FileProcessingStateCompleted,
				},
			},
			Severity: domain.CloneSeverityMedium,
		}

		if err := group.IsValid(); err == nil {
			t.Error("Expected error for invalid clone, got nil")
		}
	})
}

func TestDomainAnalysisValidation(t *testing.T) {
	t.Parallel()

	t.Run("should accept valid analysis", func(t *testing.T) {
		analysis := domain.Analysis{
			ID:        "analysis-1",
			State:     domain.DetectionStateCompleted,
			Mode:      domain.AnalysisModeFull,
			Threshold: 10,
			CreatedAt: time.Now().Format(time.RFC3339),
			CloneGroups: []domain.CloneGroup{
				{
					ID:   "group-1",
					Hash: "abc123",
					Size: 100,
					Clones: []domain.Clone{
						{
							ID:        "clone-1",
							Filename:  "test.go",
							StartLine: 10,
							EndLine:   20,
							Status:    domain.FileProcessingStateCompleted,
						},
					},
					Severity: domain.CloneSeverityMedium,
					Status:   domain.FileProcessingStateCompleted,
				},
			},
			Stats: domain.AnalysisStats{
				FilesAnalyzed:   5,
				TotalClones:     1,
				TotalTokenSize:  100,
				ComplexityScore: 0.5,
				DuplicationRatio: 0.2,
				ProcessingTime:   1000,
			},
		}

		if err := analysis.IsValid(); err != nil {
			t.Errorf("Expected valid analysis, got error: %v", err)
		}
	})

	t.Run("should reject analysis with empty ID", func(t *testing.T) {
		analysis := domain.Analysis{
			State:     domain.DetectionStateCompleted,
			Mode:      domain.AnalysisModeFull,
			Threshold: 10,
			CreatedAt: time.Now().Format(time.RFC3339),
		}

		if err := analysis.IsValid(); err == nil {
			t.Error("Expected error for empty analysis ID, got nil")
		}
	})

	t.Run("should reject analysis with zero threshold", func(t *testing.T) {
		analysis := domain.Analysis{
			ID:        "analysis-1",
			State:     domain.DetectionStateCompleted,
			Mode:      domain.AnalysisModeFull,
			Threshold: 0, // Invalid
			CreatedAt: time.Now().Format(time.RFC3339),
		}

		if err := analysis.IsValid(); err == nil {
			t.Error("Expected error for zero threshold, got nil")
		}
	})
}

func TestDomainCloneSeverity(t *testing.T) {
	t.Parallel()

	t.Run("should validate all severity levels", func(t *testing.T) {
		validSeverities := []domain.CloneSeverity{
			domain.CloneSeverityLow,
			domain.CloneSeverityMedium,
			domain.CloneSeverityHigh,
			domain.CloneSeverityCritical,
		}

		for _, severity := range validSeverities {
			if !severity.IsValid() {
				t.Errorf("Severity %s should be valid", severity)
			}
		}
	})

	t.Run("should reject invalid severity", func(t *testing.T) {
		invalidSeverity := domain.CloneSeverity("invalid")
		if invalidSeverity.IsValid() {
			t.Error("Invalid severity should not be valid")
		}
	})

	t.Run("should marshal and unmarshal JSON correctly", func(t *testing.T) {
		original := domain.CloneSeverityHigh
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		if string(data) != `"high"` {
			t.Errorf("Expected \"high\", got %s", string(data))
		}

		var unmarshaled domain.CloneSeverity
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if unmarshaled != original {
			t.Errorf("Expected %v, got %v", original, unmarshaled)
		}
	})
}

func TestDomainDetectionOptions(t *testing.T) {
	t.Parallel()

	t.Run("should accept valid options", func(t *testing.T) {
		options := domain.DetectionOptions{
			Threshold:     10,
			Mode:          domain.AnalysisModeFull,
			IncludeVendor: false,
			Verbose:       false,
			Paths:         []string{"./src"},
			OutputFormat:  "json",
		}

		if err := options.IsValid(); err != nil {
			t.Errorf("Expected valid options, got error: %v", err)
		}
	})

	t.Run("should reject options with zero threshold", func(t *testing.T) {
		options := domain.DetectionOptions{
			Threshold:    0, // Invalid
			Mode:         domain.AnalysisModeFull,
			Paths:        []string{"./src"},
			OutputFormat: "json",
		}

		if err := options.IsValid(); err == nil {
			t.Error("Expected error for zero threshold, got nil")
		}
	})

	t.Run("should reject options with empty paths", func(t *testing.T) {
		options := domain.DetectionOptions{
			Threshold:    10,
			Mode:         domain.AnalysisModeFull,
			Paths:        []string{}, // Invalid
			OutputFormat: "json",
		}

		if err := options.IsValid(); err == nil {
			t.Error("Expected error for empty paths, got nil")
		}
	})
}

func TestDomainCalculateSeverity(t *testing.T) {
	t.Parallel()

	t.Run("should calculate low severity for simple clones", func(t *testing.T) {
		severity := domain.CalculateSeverity(20, 5)
		if severity != domain.CloneSeverityLow {
			t.Errorf("Expected low severity, got %v", severity)
		}
	})

	t.Run("should calculate medium severity for moderate clones", func(t *testing.T) {
		severity := domain.CalculateSeverity(60, 15)
		if severity != domain.CloneSeverityMedium {
			t.Errorf("Expected medium severity, got %v", severity)
		}
	})

	t.Run("should calculate high severity for complex clones", func(t *testing.T) {
		severity := domain.CalculateSeverity(120, 25)
		if severity != domain.CloneSeverityHigh {
			t.Errorf("Expected high severity, got %v", severity)
		}
	})

	t.Run("should calculate critical severity for very complex clones", func(t *testing.T) {
		severity := domain.CalculateSeverity(250, 60)
		if severity != domain.CloneSeverityCritical {
			t.Errorf("Expected critical severity, got %v", severity)
		}
	})
}

func TestDomainDetectionMethods(t *testing.T) {
	t.Parallel()

	// Test String method
	dm := domain.DetectionMethodHash
	if dm.String() != "hash" {
		t.Errorf("Expected 'hash', got %s", dm.String())
	}

	// Test IsValid
	if !domain.DetectionMethodHash.IsValid() {
		t.Error("Expected hash to be valid")
	}
	if domain.DetectionMethod("invalid").IsValid() {
		t.Error("Expected invalid method to be invalid")
	}

	// Test AllDetectionMethods
	methods := domain.AllDetectionMethods()
	if len(methods) != 4 {
		t.Errorf("Expected 4 methods, got %d", len(methods))
	}

	// Test ParseDetectionMethods
	parsed, err := domain.ParseDetectionMethods("hash,art-dupl")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(parsed) != 2 {
		t.Errorf("Expected 2 methods, got %d", len(parsed))
	}

	// Test empty string defaults
	parsed, err = domain.ParseDetectionMethods("")
	if err != nil {
		t.Errorf("Expected no error for empty string, got %v", err)
	}
	if len(parsed) != 1 || parsed[0] != domain.DetectionMethodArtDupl {
		t.Error("Expected default art-dupl method for empty string")
	}

	// Test Contains
	dms := domain.DetectionMethods{domain.DetectionMethodHash, domain.DetectionMethodArtDupl}
	if !dms.Contains(domain.DetectionMethodHash) {
		t.Error("Expected Contains to find hash method")
	}
	if dms.Contains(domain.DetectionMethod("invalid")) {
		t.Error("Expected Contains to not find invalid method")
	}

	// Test IsDefault
	defaultDms := domain.DetectionMethods{domain.DetectionMethodArtDupl}
	if !defaultDms.IsDefault() {
		t.Error("Expected IsDefault to return true for art-dupl only")
	}

	mixedDms := domain.DetectionMethods{domain.DetectionMethodHash, domain.DetectionMethodArtDupl}
	if mixedDms.IsDefault() {
		t.Error("Expected IsDefault to return false for mixed methods")
	}
}

func TestDomainOutputFormats(t *testing.T) {
	t.Parallel()

	// Test AllOutputFormats
	formats := domain.AllOutputFormats()
	if len(formats) != 5 {
		t.Errorf("Expected 5 formats, got %d", len(formats))
	}

	// Test AllSortCriteria
	criteria := domain.AllSortCriteria()
	if len(criteria) != 4 {
		t.Errorf("Expected 4 sort criteria, got %d", len(criteria))
	}
}

func TestDomainJSONMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	// Test DetectionMethod JSON marshal/unmarshal
	dm := domain.DetectionMethodHash
	data, err := json.Marshal(dm)
	if err != nil {
		t.Errorf("Expected no error marshaling, got %v", err)
	}

	var parsedDM domain.DetectionMethod
	err = json.Unmarshal(data, &parsedDM)
	if err != nil {
		t.Errorf("Expected no error unmarshaling, got %v", err)
	}
	if parsedDM != dm {
		t.Error("Expected parsed detection method to match original")
	}

	// Test OutputFormat JSON marshal/unmarshal
	of := domain.OutputFormatJSON
	data, err = of.MarshalJSON()
	if err != nil {
		t.Errorf("Expected no error marshaling output format, got %v", err)
	}

	var parsedOF domain.OutputFormat
	err = json.Unmarshal(data, &parsedOF)
	if err != nil {
		t.Errorf("Expected no error unmarshaling output format, got %v", err)
	}
	if parsedOF != of {
		t.Error("Expected parsed output format to match original")
	}
}
