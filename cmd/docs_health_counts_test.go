package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/pkg/artdupl"
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
	if goNodeTypes != 51 {
		t.Fatalf("syntax/golang/transform.go dispatches %d node types, want 51; update the docs with the real count", goNodeTypes)
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

// Duplicated-constant drift guard: pkg/artdupl cannot import config (arch
// lint), so DefaultThreshold exists on both sides by necessity. If they
// silently diverge, the SDK and CLI disagree about the default threshold.
func TestSDKDefaultThresholdMirrorsConfig(t *testing.T) {
	if artdupl.DefaultThreshold != config.DefaultThreshold {
		t.Errorf("SDK DefaultThreshold = %d but config.DefaultThreshold = %d; keep the deliberate duplication in sync",
			artdupl.DefaultThreshold, config.DefaultThreshold)
	}
}

// websiteFlagNames extracts the flag names documented in the website's
// cli-flags.mdx table rows (`| `--name` | ...`). The doc covers root AND
// stats-subcommand flags in one reference.
func websiteFlagNames(t *testing.T) map[string]bool {
	t.Helper()

	pattern := regexp.MustCompile(`(?m)^\|\s*` + "`" + `(--[a-z0-9-]+)` + "`")
	matches := pattern.FindAllStringSubmatch(readRepoFile(t, "website/src/content/docs/cli-flags.mdx"), -1)
	if len(matches) == 0 {
		t.Fatal("no flag rows found in website/src/content/docs/cli-flags.mdx; the gate lost its anchor")
	}

	names := make(map[string]bool, len(matches))
	for _, m := range matches {
		names[m[1]] = true
	}

	return names
}

// binaryFlagNames collects every non-hidden flag name across the root command
// and all subcommands. `--help` is excluded: cobra injects it everywhere and
// the docs deliberately document usage instead.
func binaryFlagNames(t *testing.T) map[string]bool {
	t.Helper()

	root := NewRootCommand()
	AddFlags(root)
	root.InitDefaultHelpFlag()

	names := make(map[string]bool)
	visit := func(c *pflag.FlagSet) {
		c.VisitAll(func(f *pflag.Flag) {
			if !f.Hidden && f.Name != "help" {
				names["--"+f.Name] = true
			}
		})
	}
	visit(root.LocalFlags())
	for _, sub := range root.Commands() {
		sub.InitDefaultHelpFlag()
		visit(sub.LocalFlags())
	}

	if len(names) == 0 {
		t.Fatal("binary exposes zero non-hidden flags; the gate is measuring the wrong thing")
	}

	return names
}

// TestWebsiteFlagsMatchBinary closes the manual website-vs-CLI diff from the
// 2026-09-28 pass: every documented flag must exist on the binary, and every
// non-hidden binary flag must be documented (both directions).
func TestWebsiteFlagsMatchBinary(t *testing.T) {
	doc := websiteFlagNames(t)
	bin := binaryFlagNames(t)

	docOnly := []string{}

	for name := range doc {
		if !bin[name] {
			docOnly = append(docOnly, name)
		}
	}
	binOnly := []string{}

	for name := range bin {
		if !doc[name] {
			binOnly = append(binOnly, name)
		}
	}

	if len(docOnly) > 0 || len(binOnly) > 0 {
		t.Fatalf("website flag table drifted from the binary:\n  doc-only (not on the binary): %v\n  binary-only (undocumented): %v\nfix the website/src/content/docs/cli-flags.mdx table — the binary is the truth source",
			docOnly, binOnly)
	}
}

// TestDetectionModeCountMatchesDocs derives the mode count from the enum and
// holds the docs to it (the "three matching modes" claim).
func TestDetectionModeCountMatchesDocs(t *testing.T) {
	modes := domain.AllDetectionModes()
	if len(modes) != 3 {
		t.Fatalf("domain.AllDetectionModes() = %d, want 3 (semantic/exact/structural); update the docs with the real count", len(modes))
	}

	readme := readRepoFile(t, "README.md")
	requireDocContains(t, readme, fmt.Sprintf("%d output formats", len(domain.AllOutputFormats())))
	requireDocContains(t, readme, "three matching modes")
}

// TestOutputFormatCountMatchesDocs derives the format count from the enum and
// holds the docs to it.
func TestOutputFormatCountMatchesDocs(t *testing.T) {
	formats := domain.AllOutputFormats()
	if len(formats) != 8 {
		t.Fatalf("domain.AllOutputFormats() = %d, want 8; update the docs with the real count", len(formats))
	}

	features := readRepoFile(t, "FEATURES.md")
	requireDocContains(t, features, fmt.Sprintf("%d output formats", len(formats)))
}
