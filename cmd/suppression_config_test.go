package cmd

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
)

// TestBuildSuppressionConfigPopulatesAllFields is a regression guard for the
// bug where 3 of 6 SuppressionConfig construction sites used truncated struct
// literals, silently omitting AcceptDirectives, NoActionability, and
// DisabledPatterns. A nil AcceptDirectives means "skip the check entirely"
// (never suppress), which caused //art-dupl:accept directives to be ignored.
func TestBuildSuppressionConfigPopulatesAllFields(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		SuppressTestLow:  true,
		TestThreshold:    15,
		MinLines:         7,
		NoActionability:  true,
		DisabledPatterns: []string{"single-call-expression"},
	}

	s := buildSuppressionConfig(cfg)

	if s.AcceptDirectives == nil {
		t.Fatal("AcceptDirectives must not be nil when NoAcceptDirectives is false; " +
			"nil causes shouldSuppressGroup to skip the check entirely")
	}

	if !s.SuppressTestLow {
		t.Error("SuppressTestLow = false, want true")
	}

	if s.TestThreshold != 15 {
		t.Errorf("TestThreshold = %d, want 15", s.TestThreshold)
	}

	if s.MinLines != 7 {
		t.Errorf("MinLines = %d, want 7", s.MinLines)
	}

	if !s.NoActionability {
		t.Error("NoActionability = false, want true")
	}

	if len(s.DisabledPatterns) == 0 {
		t.Error("DisabledPatterns is empty, expected entries from config")
	}
}

// TestBuildSuppressionConfigNilAcceptDirectivesWhenDisabled verifies that
// AcceptDirectives is nil when NoAcceptDirectives is true — this is the
// intentional "disable" path, distinct from the bug's accidental nil.
func TestBuildSuppressionConfigNilAcceptDirectivesWhenDisabled(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		NoAcceptDirectives: true,
	}

	s := buildSuppressionConfig(cfg)

	if s.AcceptDirectives != nil {
		t.Error("AcceptDirectives should be nil when NoAcceptDirectives is true")
	}
}
