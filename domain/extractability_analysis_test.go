package domain

import (
	"testing"
)

func TestExtractabilityModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		analysis ExtractabilityAnalysis
		harmful  bool
	}{
		{
			name: "all properties true + high confidence = harmful",
			analysis: ExtractabilityAnalysis{
				MechanicallyExtractable: true,
				ControlFlowExtractable:  true,
				ROIPositive:             true,
				Parameterizable:         true,
				Confidence:              0.9,
			},
			harmful: true,
		},
		{
			name: "control-flow not extractable = not harmful",
			analysis: ExtractabilityAnalysis{
				MechanicallyExtractable: true,
				ControlFlowExtractable:  false,
				ROIPositive:             true,
				Parameterizable:         true,
				Confidence:              0.9,
			},
			harmful: false,
		},
		{
			name: "ROI negative = not harmful",
			analysis: ExtractabilityAnalysis{
				MechanicallyExtractable: true,
				ControlFlowExtractable:  true,
				ROIPositive:             false,
				Parameterizable:         true,
				Confidence:              0.9,
			},
			harmful: false,
		},
		{
			name: "not parameterizable = not harmful",
			analysis: ExtractabilityAnalysis{
				MechanicallyExtractable: true,
				ControlFlowExtractable:  true,
				ROIPositive:             true,
				Parameterizable:         false,
				Confidence:              0.9,
			},
			harmful: false,
		},
		{
			name: "low confidence = not harmful (prevent suppression without evidence)",
			analysis: ExtractabilityAnalysis{
				MechanicallyExtractable: true,
				ControlFlowExtractable:  true,
				ROIPositive:             true,
				Parameterizable:         true,
				Confidence:              0.3,
			},
			harmful: false,
		},
		{
			name: "default analysis = not harmful (conservative)",
			analysis: ExtractabilityAnalysis{
				MechanicallyExtractable: true,
				ControlFlowExtractable:  true,
				ROIPositive:             true,
				Parameterizable:         true,
				Confidence:              0.5,
			},
			harmful: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsHarmful(tc.analysis)
			if got != tc.harmful {
				t.Errorf("IsHarmful() = %v, want %v", got, tc.harmful)
			}
		})
	}
}

func TestDefaultExtractabilityAnalysis(t *testing.T) {
	t.Parallel()

	d := DefaultExtractabilityAnalysis()
	if !d.MechanicallyExtractable || !d.ControlFlowExtractable || !d.ROIPositive || !d.Parameterizable {
		t.Error("default analysis should have all properties true (conservative)")
	}

	if d.Confidence != 0.5 {
		t.Errorf("default confidence = %v, want 0.5", d.Confidence)
	}
}

func TestActionabilityTier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		confidence float64
		want       CloneActionability
	}{
		{0.9, Actionable},
		{0.8, Actionable},
		{0.79, LowConfidence},
		{0.5, LowConfidence},
		{0.49, NonActionable},
		{0.0, NonActionable},
	}

	for _, tc := range tests {
		got := ActionabilityTier(tc.confidence)
		if got != tc.want {
			t.Errorf("ActionabilityTier(%.2f) = %s, want %s", tc.confidence, got, tc.want)
		}
	}
}
