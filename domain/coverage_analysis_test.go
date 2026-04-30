package domain

import (
	"testing"
	"time"
)

// validatable is a type constraint for types with IsValid() error method.
type validatable interface {
	IsValid() error
}

// baseAnalysis returns a valid Analysis with customizable overrides.
func baseAnalysis(overrides ...func(*Analysis)) Analysis {
	a := Analysis{
		ID:        "test-1",
		State:     DetectionStateIdle,
		Mode:      AnalysisModeFull,
		Threshold: 15,
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	for _, override := range overrides {
		override(&a)
	}

	return a
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

// analysisStatsCase creates a test case for AnalysisStats validation.
func analysisStatsCase(
	name string,
	files FileCount,
	time ProcessingTime,
	complexity, ratio float64,
	wantErr bool,
) struct {
	name    string
	value   AnalysisStats
	wantErr bool
} {
	return struct {
		name    string
		value   AnalysisStats
		wantErr bool
	}{
		name: name,
		value: AnalysisStats{
			FilesAnalyzed:    files,
			ProcessingTime:   time,
			ComplexityScore:  complexity,
			DuplicationRatio: ratio,
		},
		wantErr: wantErr,
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
			validationErr := tt.value.IsValid()
			if (validationErr != nil) != tt.wantErr {
				t.Errorf("IsValid() validation error = %v, wantErr %v", validationErr, tt.wantErr)
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
			value: baseAnalysis(
				func(a *Analysis) { a.ID = "test-1"; a.State = DetectionStateCompleted },
			),
			wantErr: false,
		},
		{
			name: "invalid state",
			value: baseAnalysis(
				func(a *Analysis) { a.ID = "test-2"; a.State = DetectionState("invalid") },
			),
			wantErr: true,
		},
		{
			name: "invalid mode",
			value: baseAnalysis(
				func(a *Analysis) { a.ID = "test-3"; a.Mode = AnalysisMode("invalid") },
			),
			wantErr: true,
		},
		{
			name:    "zero threshold",
			value:   baseAnalysis(func(a *Analysis) { a.ID = "test-4"; a.Threshold = 0 }),
			wantErr: true,
		},
		{
			name:    "empty created at",
			value:   baseAnalysis(func(a *Analysis) { a.ID = "test-5"; a.CreatedAt = time.Time{} }),
			wantErr: true,
		},
		{
			name: "invalid clone group",
			value: baseAnalysis(func(a *Analysis) {
				a.ID = "test-6"
				a.State = DetectionStateCompleted
				a.CloneGroups = []CloneGroup{{ID: "group-1", Clones: nil}}
			}),
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
		analysisStatsCase("zero files analyzed", 0, 1000, 1.0, 0.5, true),
		analysisStatsCase("zero processing time", 10, 0, 1.0, 0.5, true),
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
			chkErr := tt.options.IsValid()

			wantErr := tt.wantErr
			if (chkErr != nil) != wantErr {
				t.Errorf("DetectionOptions.IsValid() error = %v, wantErr %v", chkErr, wantErr)
			}
		})
	}
}
