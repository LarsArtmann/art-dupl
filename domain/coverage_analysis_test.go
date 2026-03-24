package domain

import (
	"testing"
)

// validatable is a type constraint for types with IsValid() error method.
type validatable interface {
	IsValid() error
}

// newDetectionOptions creates a DetectionOptions with default values for testing.
// The threshold parameter allows customizing the threshold value (including 0).
func newDetectionOptions(threshold Threshold) DetectionOptions {
	return DetectionOptions{
		Threshold:    threshold,
		Mode:         AnalysisModeFull,
		Paths:        []Filepath{"./src"},
		OutputFormat: "text",
	}
}

func testValidMethods[T validatable](t *testing.T, testCases []struct {
	name    string
	value   T
	wantErr bool
},
) {
	t.Helper()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAnalysis_IsValid tests Analysis.IsValid method.
func TestAnalysis_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		value   Analysis
		wantErr bool
	}{
		{
			name: "valid analysis",
			value: Analysis{
				ID:        "test-1",
				State:     DetectionStateCompleted,
				Mode:      AnalysisModeFull,
				Threshold: 15,
				CreatedAt: "2024-01-01T00:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "invalid state",
			value: Analysis{
				ID:        "test-2",
				State:     DetectionState("invalid"),
				Mode:      AnalysisModeFull,
				Threshold: 15,
				CreatedAt: "2024-01-01T00:00:00Z",
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			value: Analysis{
				ID:        "test-3",
				State:     DetectionStateIdle,
				Mode:      AnalysisMode("invalid"),
				Threshold: 15,
				CreatedAt: "2024-01-01T00:00:00Z",
			},
			wantErr: true,
		},
		{
			name: "zero threshold",
			value: Analysis{
				ID:        "test-4",
				State:     DetectionStateIdle,
				Mode:      AnalysisModeFull,
				Threshold: 0,
				CreatedAt: "2024-01-01T00:00:00Z",
			},
			wantErr: true,
		},
		{
			name: "empty created at",
			value: Analysis{
				ID:        "test-5",
				State:     DetectionStateIdle,
				Mode:      AnalysisModeFull,
				Threshold: 15,
				CreatedAt: "",
			},
			wantErr: true,
		},
		{
			name: "invalid clone group",
			value: Analysis{
				ID:        "test-6",
				State:     DetectionStateCompleted,
				Mode:      AnalysisModeFull,
				Threshold: 15,
				CreatedAt: "2024-01-01T00:00:00Z",
				CloneGroups: []CloneGroup{
					{ID: "group-1", Clones: nil},
				},
			},
			wantErr: true,
		},
	}

	testValidMethods(t, tests)
}

// TestAnalysisStats_IsValid tests AnalysisStats.IsValid method.
func TestAnalysisStats_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		value   AnalysisStats
		wantErr bool
	}{
		{
			name: "valid stats",
			value: AnalysisStats{
				FilesAnalyzed:    10,
				TotalClones:      5,
				TotalTokenSize:   100,
				ComplexityScore:  2.5,
				DuplicationRatio: 0.3,
				ProcessingTime:   1000,
			},
			wantErr: false,
		},
		{
			name: "zero files analyzed",
			value: AnalysisStats{
				FilesAnalyzed:    0,
				ProcessingTime:   1000,
				ComplexityScore:  1.0,
				DuplicationRatio: 0.5,
			},
			wantErr: true,
		},
		{
			name: "zero processing time",
			value: AnalysisStats{
				FilesAnalyzed:    10,
				ProcessingTime:   0,
				ComplexityScore:  1.0,
				DuplicationRatio: 0.5,
			},
			wantErr: true,
		},
		{
			name: "negative complexity score",
			value: AnalysisStats{
				FilesAnalyzed:    10,
				ProcessingTime:   1000,
				ComplexityScore:  -1.0,
				DuplicationRatio: 0.5,
			},
			wantErr: true,
		},
		{
			name: "negative duplication ratio",
			value: AnalysisStats{
				FilesAnalyzed:    10,
				ProcessingTime:   1000,
				ComplexityScore:  1.0,
				DuplicationRatio: -0.1,
			},
			wantErr: true,
		},
	}

	testValidMethods(t, tests)
}

// TestDetectionOptions_IsValid tests DetectionOptions.IsValid method.
func TestDetectionOptions_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		options DetectionOptions
		wantErr bool
	}{
		{
			name:    "valid options",
			options: newDetectionOptions(15),
			wantErr: false,
		},
		{
			name:    "zero threshold",
			options: newDetectionOptions(0),
			wantErr: true,
		},
		{
			name: "invalid mode",
			options: DetectionOptions{
				Threshold:    15,
				Mode:         AnalysisMode("invalid"),
				Paths:        []Filepath{"./src"},
				OutputFormat: "text",
			},
			wantErr: true,
		},
		{
			name: "empty paths",
			options: DetectionOptions{
				Threshold:    15,
				Mode:         AnalysisModeFull,
				Paths:        []Filepath{},
				OutputFormat: "text",
			},
			wantErr: true,
		},
		{
			name: "nil paths",
			options: DetectionOptions{
				Threshold:    15,
				Mode:         AnalysisModeFull,
				Paths:        nil,
				OutputFormat: "text",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectionOptions.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
