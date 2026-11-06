package cmd

import (
	"encoding/json"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// TestSimpleJSONFormatEmitsImpactArray pins the documented `--simple-json`
// contract: a bare array of `{hash, score, instances}` (README: "lightweight
// JSON with impact scores"). The flag used to emit byte-identical FULL JSON,
// because OutputSimpleJSON had no caller — the "simple" format never reached
// the CLI.
func TestSimpleJSONFormatEmitsImpactArray(t *testing.T) {
	dir := t.TempDir()
	writeCheckFixture(t, dir, "a.go", checkFixtureFamilyA)
	writeCheckFixture(t, dir, "b.go", checkFixtureFamilyB)

	rootCmd := NewRootCommand()
	AddFlags(rootCmd)
	rootCmd.SetArgs([]string{"--simple-json", "-t", "3", "--no-actionability", dir})

	stdout, _, err := testutil.CaptureStdoutStderr(rootCmd.Execute)
	if err != nil {
		t.Fatalf("--simple-json failed: %v", err)
	}

	type instance struct {
		Filename   string `json:"filename"`
		LineStart  int    `json:"line_start"`
		TokenCount int    `json:"token_count"`
	}

	type group struct {
		Hash      string     `json:"hash"`
		Score     int        `json:"score"`
		Instances []instance `json:"instances"`
	}

	var parsed []group
	if decodeErr := json.Unmarshal(stdout, &parsed); decodeErr != nil {
		t.Fatalf("--simple-json must emit a JSON array, got %q: %v", stdout, decodeErr)
	}

	if len(parsed) != 1 {
		t.Fatalf("expected 1 group for the fixture pair, got %d", len(parsed))
	}

	if parsed[0].Hash == "" {
		t.Error("simple group must carry its hash")
	}

	if parsed[0].Score <= 0 {
		t.Errorf("simple group must carry a positive impact score, got %d", parsed[0].Score)
	}

	if len(parsed[0].Instances) != 2 || parsed[0].Instances[0].TokenCount <= 0 {
		t.Errorf("expected 2 scored instances, got %+v", parsed[0].Instances)
	}
}
