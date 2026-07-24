package domain

import (
	"encoding/json/v2"
	"testing"
)

func TestCloneRef_LineCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ref  CloneRef
		want int
	}{
		{
			name: "single line",
			ref:  CloneRef{Filename: "a.go", LineStart: 5, LineEnd: 5},
			want: 1,
		},
		{
			name: "multi line",
			ref:  CloneRef{Filename: "a.go", LineStart: 1, LineEnd: 10},
			want: 10,
		},
		{
			name: "zero range",
			ref:  CloneRef{Filename: "a.go", LineStart: 0, LineEnd: 0},
			want: 1,
		},
		{
			name: "large range",
			ref:  CloneRef{Filename: "a.go", LineStart: 100, LineEnd: 250},
			want: 151,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.ref.LineCount(); got != tc.want {
				t.Errorf("LineCount() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCloneRef_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := CloneRef{
		Filename:  "main.go",
		LineStart: 10,
		LineEnd:   25,
		Fragment:  "func main() { }",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded CloneRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded != original {
		t.Errorf("round-trip mismatch:\n  got  %+v\n  want %+v", decoded, original)
	}
}

func TestCloneRef_JSON_EmptyFragment(t *testing.T) {
	t.Parallel()

	ref := CloneRef{
		Filename:  "a.go",
		LineStart: 1,
		LineEnd:   5,
		Fragment:  "",
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Fragment has omitempty, so it should not appear in JSON
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map error: %v", err)
	}

	if _, exists := raw["fragment"]; exists {
		t.Error("expected fragment to be omitted when empty, but it was present")
	}
}

func TestCloneRef_JSON_NonEmptyFragment(t *testing.T) {
	t.Parallel()

	ref := CloneRef{
		Filename:  "a.go",
		LineStart: 1,
		LineEnd:   5,
		Fragment:  "code here",
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map error: %v", err)
	}

	fragment, exists := raw["fragment"]
	if !exists {
		t.Fatal("expected fragment to be present when non-empty")
	}

	if fragment != "code here" {
		t.Errorf("fragment = %v, want 'code here'", fragment)
	}
}
