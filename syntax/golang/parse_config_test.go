package golang

import (
	"errors"
	"fmt"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
)

// TestErrInvalidDetectionModeAlias locks in the invariant that
// syntax/golang.ErrInvalidDetectionMode is the SAME sentinel as
// domain.ErrInvalidDetectionMode (not a separate errors.New with the same
// message). Before this was fixed, the two were distinct pointers, so
// errors.Is across packages silently returned false — a latent bug where
// callers checking errors.Is(err, domain.ErrInvalidDetectionMode) would NOT
// match an error wrapped at the syntax/golang layer.
func TestErrInvalidDetectionModeAlias(t *testing.T) {
	t.Parallel()

	if !errors.Is(ErrInvalidDetectionMode, domain.ErrInvalidDetectionMode) {
		t.Error("syntax/golang.ErrInvalidDetectionMode must be errors.Is-equal to domain.ErrInvalidDetectionMode")
	}

	if !errors.Is(domain.ErrInvalidDetectionMode, ErrInvalidDetectionMode) {
		t.Error("domain.ErrInvalidDetectionMode must be errors.Is-equal to syntax/golang.ErrInvalidDetectionMode (symmetry)")
	}

	// A wrapped instance must also match the sentinel via errors.Is.
	wrapped := fmt.Errorf("context: %w", ErrInvalidDetectionMode)

	if !errors.Is(wrapped, domain.ErrInvalidDetectionMode) {
		t.Error("wrapped syntax/golang sentinel must match domain.ErrInvalidDetectionMode via errors.Is")
	}
}
