package jsonutil_test

import (
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/jsonutil"
)

type sample struct {
	Name  string  `json:"name"`
	HTML  string  `json:"html"`
	Count int     `json:"count"`
	Ratio float64 `json:"ratio"`
}

func TestMarshalIndent_NoHTMLEscaping(t *testing.T) {
	tests := map[string]string{
		"less-than":    "a < b",
		"greater-than": "x > y",
		"ampersand":    "at&t",
		"combined":     "<div>&amp;</div>",
	}

	for name, in := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := jsonutil.MarshalIndent(map[string]string{"v": in}, "", "")
			if err != nil {
				t.Fatalf("MarshalIndent: %v", err)
			}

			if !strings.Contains(string(got), in) {
				t.Errorf("output lost raw characters: got %q, want it to contain %q", got, in)
			}

			for _, esc := range []string{"\\u003c", "\\u003e", "\\u0026"} {
				if strings.Contains(string(got), esc) {
					t.Errorf("output contains HTML escape %s: %q", esc, got)
				}
			}
		})
	}
}

func TestMarshalIndent_NoTrailingNewline(t *testing.T) {
	v := sample{Name: "x", HTML: "a&b", Count: 1, Ratio: 0.5}

	for name, marshal := range map[string]func() ([]byte, error){
		"MarshalIndent": func() ([]byte, error) { return jsonutil.MarshalIndent(v, "", "  ") },
		"Marshal":       func() ([]byte, error) { return jsonutil.Marshal(v) },
	} {
		t.Run(name, func(t *testing.T) {
			got, err := marshal()
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			if strings.HasSuffix(string(got), "\n") {
				t.Errorf("output has trailing newline: %q", got)
			}
		})
	}
}

func TestMarshalIndent_Indentation(t *testing.T) {
	v := sample{Name: "x", HTML: "a<b", Count: 2, Ratio: 1.5}

	got, err := jsonutil.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}

	for _, want := range []string{"\n  \"name\"", "\n  \"html\""} {
		if !strings.Contains(string(got), want) {
			t.Errorf("output missing indented field %q: %q", want, got)
		}
	}

	compact, err := jsonutil.MarshalIndent(v, "", "")
	if err != nil {
		t.Fatalf("MarshalIndent with empty indent: %v", err)
	}

	if strings.Contains(string(compact), "\n") {
		t.Errorf("empty indent should produce a single line, got %q", compact)
	}
}

func TestMarshal_MatchesMarshalIndentWithEmptyIndent(t *testing.T) {
	v := sample{Name: "x", HTML: "a&<b", Count: 3, Ratio: 2.25}

	viaMarshal, err := jsonutil.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	viaIndent, err := jsonutil.MarshalIndent(v, "", "")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}

	if string(viaMarshal) != string(viaIndent) {
		t.Errorf("Marshal = %q, want %q", viaMarshal, viaIndent)
	}
}

func TestMarshalIndent_Error(t *testing.T) {
	_, err := jsonutil.MarshalIndent(map[string]chan int{"c": make(chan int)}, "", "")
	if err == nil {
		t.Fatal("expected error for unmarshalable value, got nil")
	}

	if !strings.Contains(err.Error(), "encode json") {
		t.Errorf("error should be wrapped with context, got: %v", err)
	}
}
