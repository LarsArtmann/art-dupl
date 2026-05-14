package domain

import (
	"encoding/json"
	"testing"
)

func TestFilepath_Valid(t *testing.T) {
	fp, err := NewFilepath("src/main.go")
	if err != nil {
		t.Fatalf("NewFilepath() error: %v", err)
	}

	if fp.String() != "src/main.go" {
		t.Errorf("Filepath.String() = %q, want 'src/main.go'", fp.String())
	}
}

func TestFilepath_Empty(t *testing.T) {
	_, err := NewFilepath("")
	if err == nil {
		t.Error("NewFilepath(\"\") expected error, got nil")
	}
}

func TestFilepath_MarshalJSON(t *testing.T) {
	fp := Filepath("test.go")
	data, err := json.Marshal(fp)
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	if string(data) != `"test.go"` {
		t.Errorf("MarshalJSON() = %s, want \"test.go\"", data)
	}
}

func TestFilepath_UnmarshalJSON(t *testing.T) {
	var fp Filepath
	err := json.Unmarshal([]byte(`"test.go"`), &fp)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	if fp != "test.go" {
		t.Errorf("UnmarshalJSON() = %q, want 'test.go'", fp)
	}
}

func TestFilepath_UnmarshalJSON_Empty(t *testing.T) {
	var fp Filepath
	err := json.Unmarshal([]byte(`""`), &fp)

	if err == nil {
		t.Error("UnmarshalJSON(\"\") expected error, got nil")
	}
}

func TestLineNumber_Valid(t *testing.T) {
	ln, err := NewLineNumber(42)
	if err != nil {
		t.Fatalf("NewLineNumber() error: %v", err)
	}

	if ln.Uint16() != 42 {
		t.Errorf("LineNumber.Uint16() = %d, want 42", ln.Uint16())
	}
}

func TestLineNumber_Zero(t *testing.T) {
	_, err := NewLineNumber(0)
	if err == nil {
		t.Error("NewLineNumber(0) expected error, got nil")
	}
}

func TestLineNumber_MarshalJSON(t *testing.T) {
	ln := LineNumber(100)
	data, err := json.Marshal(ln)
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	if string(data) != "100" {
		t.Errorf("MarshalJSON() = %s, want 100", data)
	}
}

func TestLineNumber_MarshalJSON_Zero(t *testing.T) {
	ln := LineNumber(0)
	_, err := json.Marshal(ln)

	if err == nil {
		t.Error("MarshalJSON(0) expected error, got nil")
	}
}

func TestLineNumber_UnmarshalJSON(t *testing.T) {
	var ln LineNumber
	err := json.Unmarshal([]byte("200"), &ln)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	if ln != 200 {
		t.Errorf("UnmarshalJSON() = %d, want 200", ln)
	}
}

func TestLineNumber_UnmarshalJSON_Zero(t *testing.T) {
	var ln LineNumber
	err := json.Unmarshal([]byte("0"), &ln)
	if err == nil {
		t.Error("UnmarshalJSON(0) expected error, got nil")
	}
}

func TestCloneSeverity_Valid(t *testing.T) {
	for _, sev := range []CloneSeverity{
		CloneSeverityLow,
		CloneSeverityMedium,
		CloneSeverityHigh,
		CloneSeverityCritical,
	} {
		if !sev.IsValid() {
			t.Errorf("%q should be valid", sev)
		}
	}
}

func TestCloneSeverity_Invalid(t *testing.T) {
	sev := CloneSeverity("invalid")
	if sev.IsValid() {
		t.Error("'invalid' should not be valid")
	}
}

func TestCloneSeverity_MarshalJSON(t *testing.T) {
	sev := CloneSeverityHigh
	data, err := json.Marshal(sev)
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	if string(data) != `"high"` {
		t.Errorf("MarshalJSON() = %s, want \"high\"", data)
	}
}

func TestCloneSeverity_MarshalJSON_Invalid(t *testing.T) {
	sev := CloneSeverity("bad")
	_, err := json.Marshal(sev)

	if err == nil {
		t.Error("MarshalJSON('bad') expected error, got nil")
	}
}

func TestCloneSeverity_UnmarshalJSON(t *testing.T) {
	var sev CloneSeverity
	err := json.Unmarshal([]byte(`"medium"`), &sev)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	if sev != CloneSeverityMedium {
		t.Errorf("UnmarshalJSON() = %q, want medium", sev)
	}
}

func TestCloneSeverity_UnmarshalJSON_Invalid(t *testing.T) {
	var sev CloneSeverity
	err := json.Unmarshal([]byte(`"unknown"`), &sev)

	if err == nil {
		t.Error("UnmarshalJSON('unknown') expected error, got nil")
	}
}

func TestCloneClassification_Name(t *testing.T) {
	tests := []struct {
		c    CloneClassification
		want string
	}{
		{CloneUnknown, "Unknown"},
		{CloneExact, "Exact"},
		{CloneNearMiss, "Near-Miss"},
		{CloneStructural, "Structural"},
		{CloneFileDuplicate, "File Duplicate"},
		{CloneClassification(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.c.Name()
		if got != tt.want {
			t.Errorf("CloneClassification(%d).Name() = %q, want %q", tt.c, got, tt.want)
		}
	}
}

func TestProcessedClone_ZeroValue(t *testing.T) {
	var pc ProcessedClone
	if pc.Filename != "" {
		t.Errorf("zero value Filename = %q, want empty", pc.Filename)
	}
	if pc.LineStart != 0 {
		t.Errorf("zero value LineStart = %d, want 0", pc.LineStart)
	}
}

func TestProcessedCloneGroup_ZeroValue(t *testing.T) {
	var pg ProcessedCloneGroup
	if pg.Hash != "" {
		t.Errorf("zero value Hash = %q, want empty", pg.Hash)
	}
	if pg.Clones != nil {
		t.Errorf("zero value Clones = %v, want nil", pg.Clones)
	}
}

func TestProcessedCloneGroup_WithClones(t *testing.T) {
	pg := ProcessedCloneGroup{
		Hash: "abc123",
		Size: 50,
		Clones: []ProcessedClone{
			{Filename: "a.go", LineStart: 1, LineEnd: 10, Size: 25},
			{Filename: "b.go", LineStart: 5, LineEnd: 14, Size: 25},
		},
	}

	if len(pg.Clones) != 2 {
		t.Fatalf("len(Clones) = %d, want 2", len(pg.Clones))
	}

	if pg.Clones[0].Filename != "a.go" {
		t.Errorf("Clones[0].Filename = %q, want 'a.go'", pg.Clones[0].Filename)
	}
}
