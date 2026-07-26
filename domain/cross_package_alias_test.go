package domain_test

import (
	"errors"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/pkg/artdupl"
	syntaxgolang "github.com/LarsArtmann/art-dupl/syntax/golang"
)

// TestAliasedSentinelsAreIdentical verifies that every sentinel re-exported
// from domain/ into config/, pkg/artdupl/, or syntax/golang/ points to the
// exact same error value. This prevents the class of bug where two
// errors.New("same message") in different packages cause errors.Is to
// silently return false across package boundaries.
func TestAliasedSentinelsAreIdentical(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		canon   error
		aliased error
	}{
		{"config.ErrInvalidDetectionMethod", domain.ErrInvalidDetectionMethod, config.ErrInvalidDetectionMethod},
		{"config.ErrInvalidDetectionMode", domain.ErrInvalidDetectionMode, config.ErrInvalidDetectionMode},
		{"config.ErrInvalidDiffMode", domain.ErrInvalidDiffMode, config.ErrInvalidDiffMode},
		{"config.ErrInvalidThreshold", domain.ErrInvalidThreshold, config.ErrInvalidThreshold},
		{"config.ErrThresholdTooLarge", domain.ErrThresholdTooLarge, config.ErrThresholdTooLarge},
		{"config.ErrInvalidOutputFormat", domain.ErrInvalidOutputFormat, config.ErrInvalidOutputFormat},
		{"config.ErrInvalidSortCriteria", domain.ErrInvalidSortCriteria, config.ErrInvalidSortCriteria},
		{"pkg/artdupl.ErrInvalidThreshold", domain.ErrInvalidThreshold, artdupl.ErrInvalidThreshold},
		{"pkg/artdupl.ErrThresholdTooLarge", domain.ErrThresholdTooLarge, artdupl.ErrThresholdTooLarge},
		{"syntax/golang.ErrInvalidDetectionMode", domain.ErrInvalidDetectionMode, syntaxgolang.ErrInvalidDetectionMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if !errors.Is(tt.aliased, tt.canon) {
				t.Errorf("aliased sentinel %s does not match canonical via errors.Is", tt.name)
			}

			if !errors.Is(tt.canon, tt.aliased) {
				t.Errorf("canonical does not match aliased %s via errors.Is (reverse)", tt.name)
			}

			//nolint:errorlint // intentional pointer-identity check: aliases must be the same value
			if tt.canon != tt.aliased {
				t.Errorf("%s is a distinct error value, not an alias", tt.name)
			}
		})
	}
}

// TestNoDuplicateMessageStrings verifies that no two independent
// errors.New() calls share a message string. Duplicate messages with
// different pointers cause silent errors.Is failures across packages.
func TestNoDuplicateMessageStrings(t *testing.T) {
	t.Parallel()

	allSentinels := []struct {
		msg     string
		varName string
	}{
		{"threshold must be >= 1", "domain.ErrInvalidThreshold"},
		{"threshold too large (max 1000)", "domain.ErrThresholdTooLarge"},
		{"invalid detection method", "domain.ErrInvalidDetectionMethod"},
		{"invalid detection mode", "domain.ErrInvalidDetectionMode"},
		{"invalid diff mode", "domain.ErrInvalidDiffMode"},
		{"invalid output format", "domain.ErrInvalidOutputFormat"},
		{"invalid sort criteria", "domain.ErrInvalidSortCriteria"},
	}

	seen := make(map[string]string)

	for _, s := range allSentinels {
		if prev, exists := seen[s.msg]; exists {
			t.Errorf(
				"duplicate sentinel message %q in both %s and %s",
				s.msg, prev, s.varName,
			)
		}

		seen[s.msg] = s.varName
	}
}
