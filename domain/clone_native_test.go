package domain_test

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
)

func TestDomainCloneValidation(t *testing.T) {
	t.Parallel()

	t.Run("should accept valid clones", func(t *testing.T) {
		clone := domain.Clone{
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
}

func TestDomainCloneGroupValidation(t *testing.T) {
	t.Parallel()

	t.Run("should accept valid clone groups", func(t *testing.T) {
		group := domain.CloneGroup{
			Hash: "abc123",
			Size: 100,
			Clones: []domain.Clone{
				{
					Filename:  "test1.go",
					StartLine: 10,
					EndLine:   20,
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
}

func TestDomainCalculateSeverity(t *testing.T) {
	t.Parallel()

	t.Run("should calculate low severity", func(t *testing.T) {
		severity := domain.CalculateSeverity(20, 5)
		if severity != domain.CloneSeverityLow {
			t.Errorf("Expected low severity, got %v", severity)
		}
	})
}
