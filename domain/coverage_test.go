package domain

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// stringerValidator is a type constraint for types with String() and IsValid() methods.
type stringerValidator interface {
	String() string
	IsValid() bool
}

// testEnumMethods tests String() and IsValid() methods for enum types.
func testEnumMethods[T stringerValidator](t *testing.T, val T, wantStr string, wantValid bool) {
	t.Helper()

	if got := val.String(); got != wantStr {
		t.Errorf("String() = %v, want %v", got, wantStr)
	}

	if got := val.IsValid(); got != wantValid {
		t.Errorf("IsValid() = %v, want %v", got, wantValid)
	}
}

// runEnumTests runs testEnumMethods for all test cases.
func runEnumTests[T stringerValidator](t *testing.T, testCases []struct {
	name    string
	state   T
	str     string
	isValid bool
},
) {
	t.Helper()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			testEnumMethods(t, tt.state, tt.str, tt.isValid)
		})
	}
}

// TestEnumTypes tests enum type String and IsValid methods.
func TestFileProcessingState_Methods(t *testing.T) {
	tests := []struct {
		name    string
		state   FileProcessingState
		str     string
		isValid bool
	}{
		{"pending", FileProcessingStatePending, "pending", true},
		{"processing", FileProcessingStateProcessing, "processing", true},
		{"completed", FileProcessingStateCompleted, "completed", true},
		{"failed", FileProcessingStateFailed, "failed", true},
		{"invalid", FileProcessingState("invalid"), "invalid", false},
	}

	runEnumTests(t, tests)
}

func TestDetectionState_Methods(t *testing.T) {
	tests := []struct {
		name    string
		state   DetectionState
		str     string
		isValid bool
	}{
		{"idle", DetectionStateIdle, "idle", true},
		{"running", DetectionStateRunning, "running", true},
		{"completed", DetectionStateCompleted, "completed", true},
		{"failed", DetectionStateFailed, "failed", true},
		{"invalid", DetectionState("invalid"), "invalid", false},
	}

	runEnumTests(t, tests)
}

func TestAnalysisMode_Methods(t *testing.T) {
	tests := []struct {
		name    string
		state   AnalysisMode
		str     string
		isValid bool
	}{
		{"full", AnalysisModeFull, "full", true},
		{"quick", AnalysisModeQuick, "quick", true},
		{"deep", AnalysisModeDeep, "deep", true},
		{"invalid", AnalysisMode("invalid"), "invalid", false},
	}

	runEnumTests(t, tests)
}

func TestCloneSeverity_Methods(t *testing.T) {
	tests := []struct {
		name    string
		state   CloneSeverity
		str     string
		isValid bool
	}{
		{"low", CloneSeverityLow, "low", true},
		{"medium", CloneSeverityMedium, "medium", true},
		{"high", CloneSeverityHigh, "high", true},
		{"critical", CloneSeverityCritical, "critical", true},
		{"invalid", CloneSeverity("invalid"), "invalid", false},
	}

	runEnumTests(t, tests)
}

func TestCloneSeverity_MarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		severity  CloneSeverity
		want      string
		wantError bool
	}{
		{"low", CloneSeverityLow, `"low"`, false},
		{"medium", CloneSeverityMedium, `"medium"`, false},
		{"high", CloneSeverityHigh, `"high"`, false},
		{"critical", CloneSeverityCritical, `"critical"`, false},
		{"invalid", CloneSeverity("invalid"), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.severity.MarshalJSON()
			if (err != nil) != tt.wantError {
				t.Errorf("MarshalJSON() error = %v, wantError %v", err, tt.wantError)

				return
			}

			if !tt.wantError && string(got) != tt.want {
				t.Errorf("MarshalJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestCloneSeverity_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      CloneSeverity
		wantError bool
	}{
		{"low", `"low"`, CloneSeverityLow, false},
		{"medium", `"medium"`, CloneSeverityMedium, false},
		{"high", `"high"`, CloneSeverityHigh, false},
		{"critical", `"critical"`, CloneSeverityCritical, false},
		{"invalid", `"invalid"`, CloneSeverity(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got CloneSeverity

			err := got.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantError {
				t.Errorf("UnmarshalJSON() error = %v, wantError %v", err, tt.wantError)

				return
			}

			if !tt.wantError && got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCalculateSeverity tests the CalculateSeverity function.
func TestCalculateSeverity(t *testing.T) {
	tests := []struct {
		name       string
		size       uint
		complexity uint
		want       CloneSeverity
	}{
		{"low severity", 10, 5, CloneSeverityLow},
		{"medium by size", 60, 5, CloneSeverityMedium},
		{"medium by complexity", 10, 15, CloneSeverityMedium},
		{"high by size", 120, 5, CloneSeverityHigh},
		{"high by complexity", 10, 25, CloneSeverityHigh},
		{"critical by size", 250, 5, CloneSeverityCritical},
		{"critical by complexity", 10, 60, CloneSeverityCritical},
		{"boundary medium size", 51, 5, CloneSeverityMedium},
		{"boundary medium complexity", 10, 11, CloneSeverityMedium},
		{"boundary high size", 101, 5, CloneSeverityHigh},
		{"boundary high complexity", 10, 21, CloneSeverityHigh},
		{"boundary critical size", 201, 5, CloneSeverityCritical},
		{"boundary critical complexity", 10, 51, CloneSeverityCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateSeverity(tt.size, tt.complexity)
			if got != tt.want {
				t.Errorf("CalculateSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNodeToClone tests the NodeToClone function.
func TestNodeToClone(t *testing.T) {
	t.Run("with valid file content", func(t *testing.T) {
		node := &syntax.Node{
			Type: 1,
			Pos:  0,
			End:  12,
		}
		filename := TestFilename
		fileContent := []byte("func main() {}")

		clone := NodeToClone(node, filename, fileContent)

		if clone.Status != FileProcessingStateCompleted {
			t.Errorf("Status = %v, want %v", clone.Status, FileProcessingStateCompleted)
		}

		if clone.StartLine.Uint() == 0 {
			t.Error("StartLine should not be zero")
		}

		if clone.EndLine.Uint() == 0 {
			t.Error("EndLine should not be zero")
		}
	})

	t.Run("with nil file content", func(t *testing.T) {
		node := &syntax.Node{
			Type: 1,
			Pos:  0,
			End:  12,
		}
		filename := TestFilename

		clone := NodeToClone(node, filename, nil)

		if clone.Status != FileProcessingStateCompleted {
			t.Errorf("Status = %v, want %v", clone.Status, FileProcessingStateCompleted)
		}

		if clone.StartLine.Uint() != 1 {
			t.Errorf("StartLine = %v, want 1", clone.StartLine.Uint())
		}
	})

	t.Run("with nested children", func(t *testing.T) {
		node := &syntax.Node{
			Type: 1,
			Pos:  0,
			End:  12,
			Children: []*syntax.Node{
				{Type: 2, Pos: 5, End: 10},
				{Type: 3, Pos: 5, End: 10, Children: []*syntax.Node{{Type: 4, Pos: 6, End: 9}}},
			},
		}
		filename := TestFilename
		fileContent := []byte("func main() {}")

		clone := NodeToClone(node, filename, fileContent)

		if clone.Complexity.Uint() == 0 {
			t.Error("Complexity should be calculated from children")
		}
	})
}

// TestCloneStringMethods tests Clone string accessor methods.
func TestCloneStringMethods(t *testing.T) {
	clone := Clone{}
	clone.SetFilename(TestFilename)
	clone.SetFragment("func main() {}")
	clone.SetHash("abc123")

	if got := clone.FilenameString(); got != TestFilename {
		t.Errorf("FilenameString() = %v, want 'test.go'", got)
	}

	if got := clone.FragmentString(); got != "func main() {}" {
		t.Errorf("FragmentString() = %v, want 'func main() {}'", got)
	}

	if got := clone.HashString(); got != "abc123" {
		t.Errorf("HashString() = %v, want 'abc123'", got)
	}
}
