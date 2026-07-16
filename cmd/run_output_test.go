package cmd

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
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
