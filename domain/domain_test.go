package domain

import (
	"encoding/json"
	"errors"
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

func TestClonePriority_Valid(t *testing.T) {
	for _, sev := range []ClonePriority{
		PriorityLow,
		PriorityMedium,
		PriorityHigh,
		PriorityCritical,
	} {
		if !sev.IsValid() {
			t.Errorf("%q should be valid", sev)
		}
	}
}

func TestClonePriority_Invalid(t *testing.T) {
	sev := ClonePriority("invalid")
	if sev.IsValid() {
		t.Error("'invalid' should not be valid")
	}
}

func TestClonePriority_MarshalJSON(t *testing.T) {
	sev := PriorityHigh

	data, err := json.Marshal(sev)
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	if string(data) != `"high"` {
		t.Errorf("MarshalJSON() = %s, want \"high\"", data)
	}
}

func TestClonePriority_MarshalJSON_Invalid(t *testing.T) {
	sev := ClonePriority("bad")

	_, err := json.Marshal(sev)
	if err == nil {
		t.Error("MarshalJSON('bad') expected error, got nil")
	}
}

func TestClonePriority_UnmarshalJSON(t *testing.T) {
	var sev ClonePriority

	err := json.Unmarshal([]byte(`"medium"`), &sev)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	if sev != PriorityMedium {
		t.Errorf("UnmarshalJSON() = %q, want medium", sev)
	}
}

func TestClonePriority_UnmarshalJSON_Invalid(t *testing.T) {
	var sev ClonePriority

	err := json.Unmarshal([]byte(`"unknown"`), &sev)
	if err == nil {
		t.Error("UnmarshalJSON('unknown') expected error, got nil")
	}
}

func TestProcessedCloneGroup_WithClones(t *testing.T) {
	pg := ProcessedCloneGroup{
		Hash:       "abc123",
		TokenCount: 50,
		Clones: []ProcessedClone{
			{Filename: "a.go", LineStart: 1, LineEnd: 10, TokenCount: 25},
			{Filename: "b.go", LineStart: 5, LineEnd: 14, TokenCount: 25},
		},
	}

	if pg.Hash != "abc123" {
		t.Errorf("Hash = %q, want 'abc123'", pg.Hash)
	}

	if pg.TokenCount != 50 {
		t.Errorf("TokenCount = %d, want 50", pg.TokenCount)
	}

	if len(pg.Clones) != 2 {
		t.Fatalf("len(Clones) = %d, want 2", len(pg.Clones))
	}

	if pg.Clones[0].Filename != "a.go" {
		t.Errorf("Clones[0].Filename = %q, want 'a.go'", pg.Clones[0].Filename)
	}
}

func TestProcessedClone_LineCount(t *testing.T) {
	tests := []struct {
		name      string
		lineStart int
		lineEnd   int
		want      int
	}{
		{"single line", 5, 5, 1},
		{"multi line", 1, 10, 10},
		{"same start/end", 3, 3, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := ProcessedClone{LineStart: tc.lineStart, LineEnd: tc.lineEnd}
			if got := c.LineCount(); got != tc.want {
				t.Errorf("LineCount() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestProcessedClone_Validate(t *testing.T) {
	tests := []struct {
		name    string
		clone   ProcessedClone
		wantErr error
	}{
		{
			name:    "valid clone",
			clone:   ProcessedClone{Filename: "main.go", LineStart: 1, LineEnd: 10, TokenCount: 15},
			wantErr: nil,
		},
		{
			name:    "empty filename",
			clone:   ProcessedClone{Filename: "", LineStart: 1, LineEnd: 10, TokenCount: 15},
			wantErr: ErrEmptyFilename,
		},
		{
			name:    "line end before start",
			clone:   ProcessedClone{Filename: "main.go", LineStart: 10, LineEnd: 5, TokenCount: 15},
			wantErr: ErrLineEndBeforeStart,
		},
		{
			name:    "negative token count",
			clone:   ProcessedClone{Filename: "main.go", LineStart: 1, LineEnd: 10, TokenCount: -1},
			wantErr: ErrNegativeTokenCount,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.clone.Validate()
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Validate() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestProcessedCloneGroup_TotalTokenCount(t *testing.T) {
	pg := ProcessedCloneGroup{
		Clones: []ProcessedClone{
			{TokenCount: 10},
			{TokenCount: 20},
			{TokenCount: 30},
		},
	}

	if got := pg.TotalTokenCount(); got != 60 {
		t.Errorf("TotalTokenCount() = %d, want 60", got)
	}

	empty := ProcessedCloneGroup{}
	if got := empty.TotalTokenCount(); got != 0 {
		t.Errorf("TotalTokenCount() on empty = %d, want 0", got)
	}
}

func TestProcessedCloneGroup_Validate(t *testing.T) {
	validClone := ProcessedClone{Filename: "test.go", LineStart: 1, LineEnd: 5, TokenCount: 10}

	t.Run("valid group", func(t *testing.T) {
		g := ProcessedCloneGroup{
			TokenCount: 20,
			Clones:     []ProcessedClone{validClone, validClone},
		}
		if err := g.Validate(); err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("empty group", func(t *testing.T) {
		g := ProcessedCloneGroup{}
		if err := g.Validate(); !errors.Is(err, ErrEmptyCloneGroup) {
			t.Errorf("Validate() error = %v, want ErrEmptyCloneGroup", err)
		}
	})

	t.Run("token count mismatch", func(t *testing.T) {
		g := ProcessedCloneGroup{
			TokenCount: 999, // Wrong — should be 20
			Clones:     []ProcessedClone{validClone, validClone},
		}
		if err := g.Validate(); !errors.Is(err, ErrTokenCountMismatch) {
			t.Errorf("Validate() error = %v, want ErrTokenCountMismatch", err)
		}
	})

	t.Run("invalid clone propagates", func(t *testing.T) {
		g := ProcessedCloneGroup{
			TokenCount: 10,
			Clones: []ProcessedClone{
				validClone,
				{Filename: "", LineStart: 1, LineEnd: 1, TokenCount: 0}, // Empty filename
			},
		}

		err := g.Validate()
		if err == nil {
			t.Fatal("Validate() expected error for invalid clone, got nil")
		}

		if !errors.Is(err, ErrEmptyFilename) {
			t.Errorf("Validate() error = %v, want ErrEmptyFilename", err)
		}
	})
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

func TestFindingType_IsValid(t *testing.T) {
	for _, ft := range []FindingType{FindingTypeTodo, FindingTypeLegacy} {
		if !ft.IsValid() {
			t.Errorf("%q should be valid", ft)
		}
	}

	if FindingType("invalid").IsValid() {
		t.Error("'invalid' should not be valid")
	}
}

func TestFindingType_String(t *testing.T) {
	if FindingTypeTodo.String() != "todo" {
		t.Errorf("FindingTypeTodo.String() = %q, want 'todo'", FindingTypeTodo.String())
	}

	if FindingTypeLegacy.String() != "legacy" {
		t.Errorf("FindingTypeLegacy.String() = %q, want 'legacy'", FindingTypeLegacy.String())
	}
}

func TestParseFindingType(t *testing.T) {
	ft, err := ParseFindingType("todo")
	if err != nil {
		t.Fatalf("ParseFindingType('todo') error: %v", err)
	}

	if ft != FindingTypeTodo {
		t.Errorf("ParseFindingType('todo') = %q, want 'todo'", ft)
	}

	_, err = ParseFindingType("bogus")
	if !errors.Is(err, ErrInvalidFindingType) {
		t.Errorf("ParseFindingType('bogus') error = %v, want ErrInvalidFindingType", err)
	}
}

func TestFindingType_MarshalJSON(t *testing.T) {
	data, err := json.Marshal(FindingTypeLegacy)
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	if string(data) != `"legacy"` {
		t.Errorf("MarshalJSON() = %s, want \"legacy\"", data)
	}
}

func TestFindingType_UnmarshalJSON(t *testing.T) {
	var ft FindingType

	err := json.Unmarshal([]byte(`"todo"`), &ft)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	if ft != FindingTypeTodo {
		t.Errorf("UnmarshalJSON() = %q, want 'todo'", ft)
	}
}

func TestFindingType_UnmarshalJSON_Invalid(t *testing.T) {
	var ft FindingType

	err := json.Unmarshal([]byte(`"bogus"`), &ft)
	if !errors.Is(err, ErrInvalidFindingType) {
		t.Errorf("UnmarshalJSON('bogus') error = %v, want ErrInvalidFindingType", err)
	}
}

func TestFinding_Validate(t *testing.T) {
	valid := Finding{
		Filename: "main.go",
		Line:     5,
		Type:     FindingTypeTodo,
		Priority: PriorityLow,
	}

	t.Run("valid finding", func(t *testing.T) {
		if err := valid.Validate(); err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("empty filename", func(t *testing.T) {
		f := valid
		f.Filename = ""
		if err := f.Validate(); !errors.Is(err, ErrEmptyFilename) {
			t.Errorf("Validate() error = %v, want ErrEmptyFilename", err)
		}
	})

	t.Run("zero line", func(t *testing.T) {
		f := valid
		f.Line = 0
		if err := f.Validate(); !errors.Is(err, ErrInvalidLineNumber) {
			t.Errorf("Validate() error = %v, want ErrInvalidLineNumber", err)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		f := valid
		f.Type = FindingType("bogus")
		if err := f.Validate(); !errors.Is(err, ErrInvalidFindingType) {
			t.Errorf("Validate() error = %v, want ErrInvalidFindingType", err)
		}
	})
}
