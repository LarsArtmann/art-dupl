package detection

import "testing"

// TestMethodStringContract verifies that detection's method string constants
// match the canonical protocol values used by config and SDK packages.
// This catches drift if anyone changes a method string in one package
// without updating the others.
func TestMethodStringContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		detectionConst string
		expectedValue  string
	}{
		{MethodArtDupl, "art-dupl"},
		{MethodHash, "hash"},
	}

	for _, tc := range cases {
		if tc.detectionConst != tc.expectedValue {
			t.Errorf("detection.%q = %q, want %q (must match config/pkg.artdupl constants)",
				tc.expectedValue, tc.detectionConst, tc.expectedValue)
		}
	}
}
