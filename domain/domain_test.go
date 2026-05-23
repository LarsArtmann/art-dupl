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

func TestProcessedCloneGroup_WithClones(t *testing.T) {
	pg := ProcessedCloneGroup{
		Hash: "abc123",
		Size: 50,
		Clones: []ProcessedClone{
			{Filename: "a.go", LineStart: 1, LineEnd: 10, Size: 25},
			{Filename: "b.go", LineStart: 5, LineEnd: 14, Size: 25},
		},
	}

	if pg.Hash != "abc123" {
		t.Errorf("Hash = %q, want 'abc123'", pg.Hash)
	}

	if pg.Size != 50 {
		t.Errorf("Size = %d, want 50", pg.Size)
	}

	if len(pg.Clones) != 2 {
		t.Fatalf("len(Clones) = %d, want 2", len(pg.Clones))
	}

	if pg.Clones[0].Filename != "a.go" {
		t.Errorf("Clones[0].Filename = %q, want 'a.go'", pg.Clones[0].Filename)
	}
}

func TestClonePriority_GetPriorityColor(t *testing.T) {
	tests := []struct {
		priority ClonePriority
		expected string
	}{
		{PriorityCritical, "var(--error)"},
		{PriorityHigh, "var(--warning)"},
		{PriorityMedium, "var(--accent)"},
		{PriorityLow, "var(--success)"},
		{ClonePriority("unknown"), "var(--text-secondary)"},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			result := tt.priority.GetPriorityColor()
			if result != tt.expected {
				t.Errorf("GetPriorityColor() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestClonePriority_GetPriorityEmoji(t *testing.T) {
	tests := []struct {
		priority ClonePriority
	}{
		{PriorityCritical},
		{PriorityHigh},
		{PriorityMedium},
		{PriorityLow},
		{ClonePriority("unknown")},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			result := tt.priority.GetPriorityEmoji()
			if result == "" {
				t.Error("GetPriorityEmoji() returned empty string")
			}
		})
	}
}

func TestCloneCategory_GetCategoryEmoji(t *testing.T) {
	categories := []CloneCategory{
		CategoryFunction,
		CategoryMethod,
		CategoryTest,
		CategoryStruct,
		CategoryInterface,
		CategoryHandler,
		CategoryLoop,
		CategoryConditional,
		CategoryAssignment,
		CategoryExpression,
		CategoryUnknown,
		CloneCategory("nonexistent"),
	}

	for _, cat := range categories {
		t.Run(string(cat), func(t *testing.T) {
			result := cat.GetCategoryEmoji()
			if result == "" {
				t.Errorf("GetCategoryEmoji() for %q returned empty string", cat)
			}
		})
	}
}
