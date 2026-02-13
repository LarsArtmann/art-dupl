package domain

import (
	"encoding/json"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// testFilename is a reusable test filename constant.
const testFilename = "test.go"

// mustNewLineNumber creates a LineNumber for tests, panicking on error.
func mustNewLineNumber(n uint16) LineNumber {
	ln, err := NewLineNumber(n)
	if err != nil {
		panic(err)
	}
	return ln
}

// TestAnalysis_IsValid tests Analysis.IsValid method.
func TestAnalysis_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		analysis Analysis
		wantErr  bool
	}{
		{
			name: "valid analysis",
			analysis: Analysis{
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
			analysis: Analysis{
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
			analysis: Analysis{
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
			analysis: Analysis{
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
			analysis: Analysis{
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
			analysis: Analysis{
				ID:        "test-6",
				State:     DetectionStateCompleted,
				Mode:      AnalysisModeFull,
				Threshold: 15,
				CreatedAt: "2024-01-01T00:00:00Z",
				CloneGroups: []CloneGroup{
					{ID: "group-1", Clones: nil}, // Invalid: empty clones
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.analysis.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Analysis.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAnalysisStats_IsValid tests AnalysisStats.IsValid method.
func TestAnalysisStats_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		stats   AnalysisStats
		wantErr bool
	}{
		{
			name: "valid stats",
			stats: AnalysisStats{
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
			stats: AnalysisStats{
				FilesAnalyzed:    0,
				ProcessingTime:   1000,
				ComplexityScore:  1.0,
				DuplicationRatio: 0.5,
			},
			wantErr: true,
		},
		{
			name: "zero processing time",
			stats: AnalysisStats{
				FilesAnalyzed:    10,
				ProcessingTime:   0,
				ComplexityScore:  1.0,
				DuplicationRatio: 0.5,
			},
			wantErr: true,
		},
		{
			name: "negative complexity score",
			stats: AnalysisStats{
				FilesAnalyzed:    10,
				ProcessingTime:   1000,
				ComplexityScore:  -1.0,
				DuplicationRatio: 0.5,
			},
			wantErr: true,
		},
		{
			name: "negative duplication ratio",
			stats: AnalysisStats{
				FilesAnalyzed:    10,
				ProcessingTime:   1000,
				ComplexityScore:  1.0,
				DuplicationRatio: -0.1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stats.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("AnalysisStats.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestClone_IsValid tests Clone.IsValid method.
func TestClone_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		clone   Clone
		wantErr bool
	}{
		{
			name: "valid clone",
			clone: Clone{
				StartLine: mustNewLineNumber(10),
				EndLine:   mustNewLineNumber(20),
				StartPos:  NewBytePosition(100),
				EndPos:    NewBytePosition(200),
				Status:    FileProcessingStateCompleted,
			},
			wantErr: false,
		},
		{
			name: "end line before start line",
			clone: Clone{
				StartLine: mustNewLineNumber(20),
				EndLine:   mustNewLineNumber(10),
				Status:    FileProcessingStateCompleted,
			},
			wantErr: true,
		},
		{
			name: "end pos before start pos",
			clone: Clone{
				StartLine: mustNewLineNumber(10),
				EndLine:   mustNewLineNumber(20),
				StartPos:  NewBytePosition(200),
				EndPos:    NewBytePosition(100),
				Status:    FileProcessingStateCompleted,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			clone: Clone{
				StartLine: mustNewLineNumber(10),
				EndLine:   mustNewLineNumber(20),
				Status:    FileProcessingState("invalid"),
			},
			wantErr: true,
		},
		{
			name: "zero positions are valid",
			clone: Clone{
				StartLine: mustNewLineNumber(10),
				EndLine:   mustNewLineNumber(20),
				StartPos:  0,
				EndPos:    0,
				Status:    FileProcessingStateCompleted,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.clone.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Clone.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCloneGroup_IsValid tests CloneGroup.IsValid method.
func TestCloneGroup_IsValid(t *testing.T) {
	validClone := Clone{
		StartLine: mustNewLineNumber(10),
		EndLine:   mustNewLineNumber(20),
		Status:    FileProcessingStateCompleted,
	}

	tests := []struct {
		name       string
		cloneGroup CloneGroup
		wantErr    bool
	}{
		{
			name: "valid clone group",
			cloneGroup: CloneGroup{
				ID:       "group-1",
				Clones:   []Clone{validClone},
				Hash:     "abc123",
				Size:     10,
				Severity: CloneSeverityMedium,
				Status:   FileProcessingStateCompleted,
			},
			wantErr: false,
		},
		{
			name: "empty clones",
			cloneGroup: CloneGroup{
				ID:       "group-2",
				Clones:   []Clone{},
				Severity: CloneSeverityMedium,
				Status:   FileProcessingStateCompleted,
			},
			wantErr: true,
		},
		{
			name: "nil clones",
			cloneGroup: CloneGroup{
				ID:       "group-3",
				Clones:   nil,
				Severity: CloneSeverityMedium,
				Status:   FileProcessingStateCompleted,
			},
			wantErr: true,
		},
		{
			name: "invalid severity",
			cloneGroup: CloneGroup{
				ID:       "group-4",
				Clones:   []Clone{validClone},
				Severity: CloneSeverity("invalid"),
				Status:   FileProcessingStateCompleted,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			cloneGroup: CloneGroup{
				ID:       "group-5",
				Clones:   []Clone{validClone},
				Severity: CloneSeverityMedium,
				Status:   FileProcessingState("invalid"),
			},
			wantErr: true,
		},
		{
			name: "invalid clone in group",
			cloneGroup: CloneGroup{
				ID: "group-6",
				Clones: []Clone{
					{
						StartLine: mustNewLineNumber(20),
						EndLine:   mustNewLineNumber(10), // Invalid: end before start
						Status:    FileProcessingStateCompleted,
					},
				},
				Severity: CloneSeverityMedium,
				Status:   FileProcessingStateCompleted,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cloneGroup.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("CloneGroup.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDetectionOptions_IsValid tests DetectionOptions.IsValid method.
func TestDetectionOptions_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		options DetectionOptions
		wantErr bool
	}{
		{
			name: "valid options",
			options: DetectionOptions{
				Threshold:    15,
				Mode:         AnalysisModeFull,
				Paths:        []string{"./src"},
				OutputFormat: "text",
			},
			wantErr: false,
		},
		{
			name: "zero threshold",
			options: DetectionOptions{
				Threshold:    0,
				Mode:         AnalysisModeFull,
				Paths:        []string{"./src"},
				OutputFormat: "text",
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			options: DetectionOptions{
				Threshold:    15,
				Mode:         AnalysisMode("invalid"),
				Paths:        []string{"./src"},
				OutputFormat: "text",
			},
			wantErr: true,
		},
		{
			name: "empty paths",
			options: DetectionOptions{
				Threshold:    15,
				Mode:         AnalysisModeFull,
				Paths:        []string{},
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

// TestRepository_IsValid tests Repository.IsValid method.
func TestRepository_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		repo    Repository
		wantErr bool
	}{
		{
			name: "valid repository",
			repo: Repository{
				Path:     "/path/to/repo",
				Name:     "my-repo",
				Language: "go",
			},
			wantErr: false,
		},
		{
			name: "empty path",
			repo: Repository{
				Path:     "",
				Name:     "my-repo",
				Language: "go",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			repo: Repository{
				Path:     "/path/to/repo",
				Name:     "",
				Language: "go",
			},
			wantErr: true,
		},
		{
			name: "empty language",
			repo: Repository{
				Path:     "/path/to/repo",
				Name:     "my-repo",
				Language: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.repo.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSourceFile_IsValid tests SourceFile.IsValid method.
func TestSourceFile_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		file    SourceFile
		wantErr bool
	}{
		{
			name: "valid source file",
			file: SourceFile{
				Path: "/path/to/file.go",
				Name: "file.go",
				Size: 1024,
				Hash: "abc123",
			},
			wantErr: false,
		},
		{
			name: "empty path",
			file: SourceFile{
				Path: "",
				Name: "file.go",
				Size: 1024,
				Hash: "abc123",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			file: SourceFile{
				Path: "/path/to/file.go",
				Name: "",
				Size: 1024,
				Hash: "abc123",
			},
			wantErr: true,
		},
		{
			name: "zero size",
			file: SourceFile{
				Path: "/path/to/file.go",
				Name: "file.go",
				Size: 0,
				Hash: "abc123",
			},
			wantErr: true,
		},
		{
			name: "empty hash",
			file: SourceFile{
				Path: "/path/to/file.go",
				Name: "file.go",
				Size: 1024,
				Hash: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.file.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("SourceFile.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStringID tests StringID type methods.
func TestStringID_NewStringID(t *testing.T) {
	tests := []struct {
		name  string
		input uint32
		want  StringID
	}{
		{"zero", 0, StringID(0)},
		{"positive", 42, StringID(42)},
		{"max", ^uint32(0), StringID(^uint32(0))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStringID(tt.input)
			if got != tt.want {
				t.Errorf("NewStringID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringID_Uint32(t *testing.T) {
	sid := StringID(42)
	if got := sid.Uint32(); got != 42 {
		t.Errorf("StringID.Uint32() = %v, want %v", got, 42)
	}
}

func TestStringID_MarshalJSON(t *testing.T) {
	pool := NewStringInternPool(10)
	id := pool.Intern("test")

	tests := []struct {
		name    string
		sid     StringID
		want    string
		wantErr bool
	}{
		{"valid id", id, `"test"`, false},
		{"zero id (invalid)", StringID(0), "null", false},
		{"out of range id", StringID(9999), "null", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Temporarily use test pool for MarshalJSON
			cleanup := SetGlobalPoolForTesting(pool)
			defer cleanup()

			got, err := tt.sid.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("StringID.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("StringID.MarshalJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestStringID_UnmarshalJSON(t *testing.T) {
	pool := NewStringInternPool(10)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid string", `"test"`, false},
		{"null", "null", false},
		{"empty string", `""`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use test pool for UnmarshalJSON
			cleanup := SetGlobalPoolForTesting(pool)
			defer cleanup()

			var sid StringID
			err := sid.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("StringID.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStringInternPool tests StringInternPool methods.
func TestStringInternPool_Intern(t *testing.T) {
	pool := NewStringInternPool(10)

	// Test interning a new string
	id1 := pool.Intern("hello")
	if id1 == 0 {
		t.Error("Intern() returned 0 for new string")
	}

	// Test interning the same string again
	id2 := pool.Intern("hello")
	if id2 != id1 {
		t.Errorf("Intern() returned different IDs for same string: %v vs %v", id1, id2)
	}

	// Test interning empty string
	id3 := pool.Intern("")
	if id3 != 0 {
		t.Errorf("Intern('') = %v, want 0", id3)
	}

	// Test interning another new string
	id4 := pool.Intern("world")
	if id4 == 0 || id4 == id1 {
		t.Error("Intern() returned invalid ID for different string")
	}
}

func TestStringInternPool_Lookup(t *testing.T) {
	pool := NewStringInternPool(10)

	// Test lookup for non-existent ID
	if got := pool.Lookup(StringID(999)); got != "" {
		t.Errorf("Lookup(non-existent) = %v, want empty string", got)
	}

	// Test lookup for zero ID
	if got := pool.Lookup(StringID(0)); got != "" {
		t.Errorf("Lookup(0) = %v, want empty string", got)
	}

	// Test lookup for valid ID
	id := pool.Intern("hello")
	if got := pool.Lookup(id); got != "hello" {
		t.Errorf("Lookup() = %v, want 'hello'", got)
	}
}

func TestStringInternPool_Len(t *testing.T) {
	pool := NewStringInternPool(10)

	if got := pool.Len(); got != 0 {
		t.Errorf("Len() = %v, want 0", got)
	}

	pool.Intern("hello")
	if got := pool.Len(); got != 1 {
		t.Errorf("Len() = %v, want 1", got)
	}

	pool.Intern("world")
	if got := pool.Len(); got != 2 {
		t.Errorf("Len() = %v, want 2", got)
	}

	// Same string should not increase count
	pool.Intern("hello")
	if got := pool.Len(); got != 2 {
		t.Errorf("Len() = %v, want 2 (duplicate)", got)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.String(); got != tt.str {
				t.Errorf("String() = %v, want %v", got, tt.str)
			}
			if got := tt.state.IsValid(); got != tt.isValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.isValid)
			}
		})
	}
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.String(); got != tt.str {
				t.Errorf("String() = %v, want %v", got, tt.str)
			}
			if got := tt.state.IsValid(); got != tt.isValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.isValid)
			}
		})
	}
}

func TestAnalysisMode_Methods(t *testing.T) {
	tests := []struct {
		name    string
		mode    AnalysisMode
		str     string
		isValid bool
	}{
		{"full", AnalysisModeFull, "full", true},
		{"quick", AnalysisModeQuick, "quick", true},
		{"deep", AnalysisModeDeep, "deep", true},
		{"invalid", AnalysisMode("invalid"), "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.str {
				t.Errorf("String() = %v, want %v", got, tt.str)
			}
			if got := tt.mode.IsValid(); got != tt.isValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.isValid)
			}
		})
	}
}

func TestCloneSeverity_Methods(t *testing.T) {
	tests := []struct {
		name     string
		severity CloneSeverity
		str      string
		isValid  bool
	}{
		{"low", CloneSeverityLow, "low", true},
		{"medium", CloneSeverityMedium, "medium", true},
		{"high", CloneSeverityHigh, "high", true},
		{"critical", CloneSeverityCritical, "critical", true},
		{"invalid", CloneSeverity("invalid"), "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.str {
				t.Errorf("String() = %v, want %v", got, tt.str)
			}
			if got := tt.severity.IsValid(); got != tt.isValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.isValid)
			}
		})
	}
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
		filename := testFilename
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
		filename := testFilename

		clone := NodeToClone(node, filename, nil)

		if clone.Status != FileProcessingStateCompleted {
			t.Errorf("Status = %v, want %v", clone.Status, FileProcessingStateCompleted)
		}
		// With nil fileContent, lineStart and lineEnd default to 1
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
		filename := testFilename
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
	clone.SetFilename(testFilename)
	clone.SetFragment("func main() {}")
	clone.SetHash("abc123")

	if got := clone.FilenameString(); got != testFilename {
		t.Errorf("FilenameString() = %v, want 'test.go'", got)
	}
	if got := clone.FragmentString(); got != "func main() {}" {
		t.Errorf("FragmentString() = %v, want 'func main() {}'", got)
	}
	if got := clone.HashString(); got != "abc123" {
		t.Errorf("HashString() = %v, want 'abc123'", got)
	}
}

// TestValidationHelpers tests validation helper functions.
func TestValidateRules(t *testing.T) {
	t.Run("all rules pass", func(t *testing.T) {
		rules := []validationRule{
			{valid: true, msg: "rule1"},
			{valid: true, msg: "rule2"},
		}
		if err := validateRules(rules); err != nil {
			t.Errorf("validateRules() error = %v, want nil", err)
		}
	})

	t.Run("first rule fails", func(t *testing.T) {
		rules := []validationRule{
			{valid: false, msg: "rule1 failed"},
			{valid: true, msg: "rule2"},
		}
		err := validateRules(rules)
		if err == nil {
			t.Error("validateRules() error = nil, want error")
		}
		if err.Error() != "rule1 failed" {
			t.Errorf("validateRules() error = %v, want 'rule1 failed'", err)
		}
	})

	t.Run("middle rule fails", func(t *testing.T) {
		rules := []validationRule{
			{valid: true, msg: "rule1"},
			{valid: false, msg: "rule2 failed"},
			{valid: true, msg: "rule3"},
		}
		err := validateRules(rules)
		if err == nil {
			t.Error("validateRules() error = nil, want error")
		}
		if err.Error() != "rule2 failed" {
			t.Errorf("validateRules() error = %v, want 'rule2 failed'", err)
		}
	})
}

func TestValidateFields(t *testing.T) {
	t.Run("all fields valid", func(t *testing.T) {
		err := validateFields(
			validationRule{true, "field1"},
			validationRule{true, "field2"},
		)
		if err != nil {
			t.Errorf("validateFields() error = %v, want nil", err)
		}
	})

	t.Run("one field invalid", func(t *testing.T) {
		err := validateFields(
			validationRule{true, "field1"},
			validationRule{false, "field2 is invalid"},
		)
		if err == nil {
			t.Error("validateFields() error = nil, want error")
		}
	})
}

// TestHelpers tests the helper functions in helpers.go.
func TestMarshalStringID(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		got, err := marshalStringID("test-id", "ID cannot be empty")
		if err != nil {
			t.Errorf("marshalStringID() error = %v", err)
		}
		if string(got) != `"test-id"` {
			t.Errorf("marshalStringID() = %v, want `\"test-id\"`", string(got))
		}
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := marshalStringID("", "ID cannot be empty")
		if err == nil {
			t.Error("marshalStringID() error = nil, want error for empty string")
		}
	})
}

func TestUnmarshalStringID(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		var result string
		err := unmarshalStringID([]byte(`"test-id"`), "TestType", "TestType cannot be empty", func(s string) {
			result = s
		})
		if err != nil {
			t.Errorf("unmarshalStringID() error = %v", err)
		}
		if result != "test-id" {
			t.Errorf("unmarshalStringID() result = %v, want 'test-id'", result)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		var result string
		err := unmarshalStringID([]byte(`""`), "TestType", "TestType cannot be empty", func(s string) {
			result = s
		})
		_ = result // Ensure variable is used
		if err == nil {
			t.Error("unmarshalStringID() error = nil, want error for empty string")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		var result string
		err := unmarshalStringID([]byte(`invalid`), "TestType", "TestType cannot be empty", func(s string) {
			result = s
		})
		_ = result // Ensure variable is used
		if err == nil {
			t.Error("unmarshalStringID() error = nil, want error for invalid JSON")
		}
	})
}

func TestMarshalUint(t *testing.T) {
	got, err := marshalUint(42)
	if err != nil {
		t.Errorf("marshalUint() error = %v", err)
	}
	if string(got) != "42" {
		t.Errorf("marshalUint() = %v, want '42'", string(got))
	}
}

func TestUnmarshalUint(t *testing.T) {
	t.Run("valid uint", func(t *testing.T) {
		var result uint
		err := unmarshalUint([]byte("42"), "TestType", func(n uint) {
			result = n
		})
		if err != nil {
			t.Errorf("unmarshalUint() error = %v", err)
		}
		if result != 42 {
			t.Errorf("unmarshalUint() result = %v, want 42", result)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		var result uint
		err := unmarshalUint([]byte("invalid"), "TestType", func(n uint) {
			result = n
		})
		_ = result // Ensure variable is used
		if err == nil {
			t.Error("unmarshalUint() error = nil, want error for invalid JSON")
		}
	})
}

func TestUnmarshalUintNonZero(t *testing.T) {
	t.Run("valid non-zero", func(t *testing.T) {
		var result uint
		err := unmarshalUintNonZero([]byte("42"), "TestType", "TestType cannot be zero", func(n uint) {
			result = n
		})
		if err != nil {
			t.Errorf("unmarshalUintNonZero() error = %v", err)
		}
		if result != 42 {
			t.Errorf("unmarshalUintNonZero() result = %v, want 42", result)
		}
	})

	t.Run("zero value", func(t *testing.T) {
		var result uint
		err := unmarshalUintNonZero([]byte("0"), "TestType", "TestType cannot be zero", func(n uint) {
			result = n
		})
		_ = result // Ensure variable is used
		if err == nil {
			t.Error("unmarshalUintNonZero() error = nil, want error for zero value")
		}
	})
}

func TestUnmarshalUintGeneric(t *testing.T) {
	t.Run("uint16", func(t *testing.T) {
		var result uint16
		err := unmarshalUintGeneric[uint16]([]byte("42"), "TestType", func(n uint16) {
			result = n
		})
		if err != nil {
			t.Errorf("unmarshalUintGeneric() error = %v", err)
		}
		if result != 42 {
			t.Errorf("unmarshalUintGeneric() result = %v, want 42", result)
		}
	})

	t.Run("uint32", func(t *testing.T) {
		var result uint32
		err := unmarshalUintGeneric[uint32]([]byte("42"), "TestType", func(n uint32) {
			result = n
		})
		if err != nil {
			t.Errorf("unmarshalUintGeneric() error = %v", err)
		}
		if result != 42 {
			t.Errorf("unmarshalUintGeneric() result = %v, want 42", result)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		var result uint16
		err := unmarshalUintGeneric[uint16]([]byte("invalid"), "TestType", func(n uint16) {
			result = n
		})
		_ = result // Ensure variable is used
		if err == nil {
			t.Error("unmarshalUintGeneric() error = nil, want error for invalid JSON")
		}
	})
}

// TestAnalysisJSONRoundTrip tests JSON marshaling/unmarshaling for Analysis.
func TestAnalysisJSONRoundTrip(t *testing.T) {
	original := Analysis{
		ID:        "test-1",
		State:     DetectionStateCompleted,
		Mode:      AnalysisModeFull,
		Threshold: 15,
		Stats: AnalysisStats{
			FilesAnalyzed:    10,
			TotalClones:      5,
			ProcessingTime:   1000,
			ComplexityScore:  2.5,
			DuplicationRatio: 0.3,
		},
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var result Analysis
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if result.ID != original.ID {
		t.Errorf("ID = %v, want %v", result.ID, original.ID)
	}
	if result.State != original.State {
		t.Errorf("State = %v, want %v", result.State, original.State)
	}
	if result.Mode != original.Mode {
		t.Errorf("Mode = %v, want %v", result.Mode, original.Mode)
	}
	if result.Threshold != original.Threshold {
		t.Errorf("Threshold = %v, want %v", result.Threshold, original.Threshold)
	}
}

// TestCloneGroupJSONRoundTrip tests JSON marshaling/unmarshaling for CloneGroup.
func TestCloneGroupJSONRoundTrip(t *testing.T) {
	original := CloneGroup{
		ID:       "group-1",
		Hash:     "abc123",
		Size:     10,
		Severity: CloneSeverityHigh,
		Status:   FileProcessingStateCompleted,
		Clones: []Clone{
			{
				StartLine: mustNewLineNumber(10),
				EndLine:   mustNewLineNumber(20),
				StartPos:  NewBytePosition(100),
				EndPos:    NewBytePosition(200),
				Status:    FileProcessingStateCompleted,
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var result CloneGroup
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if result.ID != original.ID {
		t.Errorf("ID = %v, want %v", result.ID, original.ID)
	}
	if result.Hash != original.Hash {
		t.Errorf("Hash = %v, want %v", result.Hash, original.Hash)
	}
	if result.Severity != original.Severity {
		t.Errorf("Severity = %v, want %v", result.Severity, original.Severity)
	}
}

// TestPoolStats tests PoolStats functionality.
func TestPoolStats(t *testing.T) {
	pool := NewStringInternPool(10)
	pool.Intern("file1.go")
	pool.Intern("file2.go")
	pool.Intern("file1.go") // Duplicate

	stats := pool.Stats()
	if stats.TotalStrings != 2 {
		t.Errorf("TotalStrings = %v, want 2", stats.TotalStrings)
	}
	if stats.TotalIDs != 2 {
		t.Errorf("TotalIDs = %v, want 2", stats.TotalIDs)
	}
}
