package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/printer/actionability"
	"github.com/spf13/pflag"
)

// Docs-health count gate: every hand-written count in the living docs is
// re-derived here from the artifact it describes, so drift fails CI instead
// of silently lying (the 29/30-patterns and 28-templ-nodes incidents).
// Canary rule: this gate was observed FAILING against an intentionally
// corrupted FEATURES number before it was trusted (SUPERB plan M08/F042).

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile("../" + rel)
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

func countOccurrences(t *testing.T, path, needle string) int {
	t.Helper()
	return strings.Count(readRepoFile(t, path), needle)
}

func requireDocContains(t *testing.T, doc, claim string) {
	t.Helper()
	if !strings.Contains(doc, claim) {
		t.Errorf("living doc drifted from code: %q not found in %s", claim, doc)
	}
}

func TestActionabilityPatternCountsMatchDocs(t *testing.T) {
	total := len(actionability.AllActionabilityPatterns())
	if total != 37 {
		t.Fatalf("AllActionabilityPatterns() = %d, want 37 (33 denylist + 4 property-engine); update the docs with the real count", total)
	}

	features := readRepoFile(t, "FEATURES.md")
	requireDocContains(t, features, fmt.Sprintf("%d actionability patterns (plus 4 property-engine labels)", total-4))
	requireDocContains(t, features, fmt.Sprintf("%d pattern labels (%d denylist + 4 property-engine)", total, total-4))
}

func TestCLIFlagCountMatchesDocs(t *testing.T) {
	root := NewRootCommand()
	AddFlags(root)
	root.InitDefaultHelpFlag()
	flags := 0
	root.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			flags++
		}
	})
	if flags == 0 {
		t.Fatal("NewRootCommand exposes zero flags; the count gate is measuring the wrong thing")
	}

	features := readRepoFile(t, "FEATURES.md")
	requireDocContains(t, features, fmt.Sprintf("%d flags for full control", flags))
}

func TestNodeTypesCountsMatchDocs(t *testing.T) {
	goNodeTypes := countOccurrences(t, "syntax/golang/transform.go", "case *ast.")
	if goNodeTypes != 49 {
		t.Fatalf("syntax/golang/transform.go dispatches %d node types, want 49; update the docs with the real count", goNodeTypes)
	}

	templNodeTypes := 0
	for _, f := range []string{
		"syntax/templ/transform.go",
		"syntax/templ/transform_node.go",
		"syntax/templ/transform_components.go",
	} {
		templNodeTypes += countOccurrences(t, f, "case *templparser.")
	}
	if templNodeTypes != 29 {
		t.Fatalf("templ transformer dispatches %d node types, want 29; update the docs with the real count", templNodeTypes)
	}

	features := readRepoFile(t, "FEATURES.md")
	readme := readRepoFile(t, "README.md")
	requireDocContains(t, features, fmt.Sprintf("%d node types", goNodeTypes))
	requireDocContains(t, readme, fmt.Sprintf("%d node types", goNodeTypes))
	requireDocContains(t, readme, fmt.Sprintf("%d node types", templNodeTypes))
}

// Guard against the claim text being reworded into a form the gate can no
// longer find: if these patterns stop matching, the gate is blind, not the
// docs correct. Adapt the patterns when the docs legitimately change shape.
func TestCountGateClaimPatternsArePresent(t *testing.T) {
	pattern := regexp.MustCompile(`\d+ (actionability patterns|pattern labels|node types|flags for full control)`)
	for _, doc := range []string{"FEATURES.md", "README.md"} {
		if !pattern.MatchString(readRepoFile(t, doc)) {
			t.Errorf("no countable claims found in %s; the docs-health gate lost its anchors", doc)
		}
	}
}
