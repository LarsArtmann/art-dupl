package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
)

func TestPriorityHigher(t *testing.T) {
	tests := []struct {
		a, b   ClonePriority
		result bool
	}{
		{domain.PriorityCritical, domain.PriorityHigh, true},
		{domain.PriorityCritical, domain.PriorityCritical, false},
		{domain.PriorityHigh, domain.PriorityMedium, true},
		{domain.PriorityMedium, domain.PriorityLow, true},
		{domain.PriorityLow, domain.PriorityLow, false},
		{domain.PriorityLow, domain.PriorityCritical, false},
	}

	for _, tt := range tests {
		name := string(tt.a) + "_vs_" + string(tt.b)
		t.Run(name, func(t *testing.T) {
			if got := priorityHigher(tt.a, tt.b); got != tt.result {
				t.Errorf("priorityHigher(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.result)
			}
		})
	}
}
