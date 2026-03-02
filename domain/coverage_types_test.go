package domain

import (
	"encoding/json"
	"testing"
)

func newRepositoryTestCase(name, path, repoName, language string, wantErr bool) struct {
	name    string
	repo    Repository
	wantErr bool
} {
	return struct {
		name    string
		repo    Repository
		wantErr bool
	}{
		name: name,
		repo: Repository{
			Path:     path,
			Name:     repoName,
			Language: language,
		},
		wantErr: wantErr,
	}
}

// TestRepository_IsValid tests Repository.IsValid method.
func TestRepository_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		repo    Repository
		wantErr bool
	}{
		newRepositoryTestCase("valid repository", "/path/to/repo", "my-repo", "go", false),
		newRepositoryTestCase("empty path", "", "my-repo", "go", true),
		newRepositoryTestCase("empty name", "/path/to/repo", "", "go", true),
		newRepositoryTestCase("empty language", "/path/to/repo", "my-repo", "", true),
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

func newSourceFileTestCase(name, path, fileName string, size uint64, hash string, wantErr bool) struct {
	name    string
	file    SourceFile
	wantErr bool
} {
	return struct {
		name    string
		file    SourceFile
		wantErr bool
	}{
		name: name,
		file: SourceFile{
			Path: path,
			Name: fileName,
			Size: size,
			Hash: hash,
		},
		wantErr: wantErr,
	}
}

// TestSourceFile_IsValid tests SourceFile.IsValid method.
func TestSourceFile_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		file    SourceFile
		wantErr bool
	}{
		newSourceFileTestCase("valid source file", "/path/to/file.go", "file.go", 1024, "abc123", false),
		newSourceFileTestCase("empty path", "", "file.go", 1024, "abc123", true),
		newSourceFileTestCase("empty name", "/path/to/file.go", "", 1024, "abc123", true),
		newSourceFileTestCase("zero size", "/path/to/file.go", "file.go", 0, "abc123", true),
		newSourceFileTestCase("empty hash", "/path/to/file.go", "file.go", 1024, "", true),
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

	id1 := pool.Intern("hello")
	if id1 == 0 {
		t.Error("Intern() returned 0 for new string")
	}

	id2 := pool.Intern("hello")
	if id2 != id1 {
		t.Errorf("Intern() returned different IDs for same string: %v vs %v", id1, id2)
	}

	id3 := pool.Intern("")
	if id3 != 0 {
		t.Errorf("Intern('') = %v, want 0", id3)
	}

	id4 := pool.Intern("world")
	if id4 == 0 || id4 == id1 {
		t.Error("Intern() returned invalid ID for different string")
	}
}

func TestStringInternPool_Lookup(t *testing.T) {
	pool := NewStringInternPool(10)

	if got := pool.Lookup(StringID(999)); got != "" {
		t.Errorf("Lookup(non-existent) = %v, want empty string", got)
	}

	if got := pool.Lookup(StringID(0)); got != "" {
		t.Errorf("Lookup(0) = %v, want empty string", got)
	}

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

	pool.Intern("hello")

	if got := pool.Len(); got != 2 {
		t.Errorf("Len() = %v, want 2 (duplicate)", got)
	}
}

// TestPoolStats tests PoolStats functionality.
func TestPoolStats(t *testing.T) {
	pool := NewStringInternPool(10)
	pool.Intern("file1.go")
	pool.Intern("file2.go")
	pool.Intern("file1.go")

	stats := pool.Stats()
	if stats.TotalStrings != 2 {
		t.Errorf("TotalStrings = %v, want 2", stats.TotalStrings)
	}

	if stats.TotalIDs != 2 {
		t.Errorf("TotalIDs = %v, want 2", stats.TotalIDs)
	}
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
				StartLine: MustNewLineNumber(10),
				EndLine:   MustNewLineNumber(20),
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
