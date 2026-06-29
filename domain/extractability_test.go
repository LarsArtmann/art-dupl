package domain

import "testing"

func TestAssessExtractability(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		lines       int
		instances   int
		complete    bool
		wantSave    int
		wantExtract bool
	}{
		{"single instance", 10, 1, true, 0, false},
		{"complete function, 3 sites", 10, 3, true, 16, true}, // 10*2-4=16
		{"complete function, 2 sites", 20, 2, true, 16, true}, // 20*1-4=16
		{"too small to save", 3, 2, true, -1, false},          // 3*1-4=-1
		{"partial fragment", 10, 3, false, 20, false},         // not complete
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := AssessExtractability(tc.lines, tc.instances, tc.complete)
			if got.EstimatedLinesSaved != tc.wantSave {
				t.Errorf("EstimatedLinesSaved = %d, want %d", got.EstimatedLinesSaved, tc.wantSave)
			}

			if got.CanExtract != tc.wantExtract {
				t.Errorf("CanExtract = %v, want %v", got.CanExtract, tc.wantExtract)
			}

			if got.Reason == "" {
				t.Error("Reason should not be empty")
			}
		})
	}
}

func TestCloneCategory_IsCompleteUnit(t *testing.T) {
	t.Parallel()

	complete := []CloneCategory{CategoryFunction, CategoryMethod, CategoryLoop, CategoryConditional}
	for _, c := range complete {
		if !c.IsCompleteUnit() {
			t.Errorf("%q should be a complete unit", c)
		}
	}

	incomplete := []CloneCategory{CategoryStruct, CategoryUnknown, CategoryAssignment}
	for _, c := range incomplete {
		if c.IsCompleteUnit() {
			t.Errorf("%q should NOT be a complete unit", c)
		}
	}
}
