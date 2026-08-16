package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
)

// Tests for the three feedback-driven patterns added 2026-08-16:
// test-preamble, testmain-boilerplate, embed-directive.
//
// Positive cases run through the FULL pipeline (parse → serialize → suffix
// tree → FindSyntaxUnits → CloneNode) so cross-layer field loss (the
// serial() IsAlias bug class) cannot hide. Negative cases assert the
// patterns do NOT over-suppress real logic.

const embedBootstrapSrc = `package main

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed all:static
var staticFiles embed.FS

func main() {
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("static sub FS: %v", err)
	}
	_ = staticFS
}
`

const testMainSrc = `package foo_test

import (
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMain(m *testing.M) {
	code := m.Run()
	snaps.Clean(m)
	os.Exit(code)
}
`

const testPreambleSrc = `package foo_test

import (
	"context"
	"testing"
)

func TestA(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_ = ctx
}

func TestB(t *testing.T) {
	t.Parallel()
	reg := newRegistry()
	_ = reg
}
`

func TestPatternEmbedDirective(t *testing.T) {
	t.Parallel()

	result := runPipeline(t, map[string]string{
		"a_embed.go": embedBootstrapSrc,
		"b_embed.go": embedBootstrapSrc,
	}, 2)

	matched := false

	for _, group := range result.groups {
		label, verdict := EvaluateActionabilityWithLabel(group)
		if verdict == domain.NonActionable && label == PatternEmbedDirective {
			matched = true
		}
	}

	if !matched {
		t.Error("embed.FS bootstrap group should be non-actionable (embed-directive)")
	}
}

func TestPatternEmbedDirectiveNotOverSuppressed(t *testing.T) {
	t.Parallel()

	// A declaration sequence with real logic following must stay actionable.
	result := runPipeline(t, map[string]string{
		"a_logic.go": `package main

import "embed"

//go:embed all:static
var staticFiles embed.FS

func render(a int) int {
	total := 0
	for i := range a {
		total += i * 2
	}
	return total
}
`,
		"b_logic.go": `package main

import "embed"

//go:embed all:static
var staticFiles embed.FS

func render(a int) int {
	total := 0
	for i := range a {
		total += i * 2
	}
	return total
}
`,
	}, 2)

	for _, group := range result.groups {
		label, verdict := EvaluateActionabilityWithLabel(group)
		if label == PatternEmbedDirective && verdict == domain.NonActionable {
			// Only acceptable if the sequence is genuinely anchored on the
			// ValueSpec; logic bodies must not be caught.
			first := group[0][0]
			if !declaresEmbedFS(first) {
				t.Errorf(
					"embed-directive over-suppressed a group rooted at BaseType=%d",
					first.BaseType,
				)
			}
		}
	}
}

func TestPatternTestMainBoilerplate(t *testing.T) {
	t.Parallel()

	result := runPipeline(t, map[string]string{
		"c_main_test.go": testMainSrc,
		"d_main_test.go": testMainSrc,
	}, 2)

	matched := false

	for _, group := range result.groups {
		label, verdict := EvaluateActionabilityWithLabel(group)
		if verdict == domain.NonActionable && label == PatternTestMainBoilerplate {
			matched = true
		}
	}

	if !matched {
		t.Error("TestMain m.Run/snaps.Clean/os.Exit group should be non-actionable (testmain-boilerplate)")
	}
}

func TestPatternTestMainNotOverSuppressed(t *testing.T) {
	t.Parallel()

	// A TestMain with real setup logic in the middle must stay actionable.
	src := `package foo_test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	if err := setupInfrastructure(); err != nil {
		os.Exit(1)
	}
	os.Exit(code)
}
`

	result := runPipeline(t, map[string]string{
		"e_main_test.go": src,
		"f_main_test.go": src,
	}, 2)

	for _, group := range result.groups {
		label, _ := EvaluateActionabilityWithLabel(group)
		if label == PatternTestMainBoilerplate {
			t.Error("TestMain with an if-guard in the middle must NOT match testmain-boilerplate")
		}
	}
}

func TestPatternTestPreamble(t *testing.T) {
	t.Parallel()

	result := runPipeline(t, map[string]string{
		"g_pream_test.go": testPreambleSrc,
		"h_pream_test.go": testPreambleSrc,
	}, 2)

	matched := false

	for _, group := range result.groups {
		label, verdict := EvaluateActionabilityWithLabel(group)
		if verdict == domain.NonActionable && label == PatternTestPreamble {
			matched = true
		}
	}

	if !matched {
		t.Error("t.Parallel() + setup-line group should be non-actionable (test-preamble)")
	}
}

func TestPatternTestPreambleNotOverSuppressed(t *testing.T) {
	t.Parallel()

	// Preamble followed by assertion calls is real test logic.
	src := `package foo_test

import (
	"testing"
)

func TestA(t *testing.T) {
	t.Parallel()
	got := compute()
	if got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func TestB(t *testing.T) {
	t.Parallel()
	got := compute()
	if got != 7 {
		t.Errorf("got %d, want 7", got)
	}
}
`

	result := runPipeline(t, map[string]string{
		"i_pream_test.go": src,
		"j_pream_test.go": src,
	}, 3)

	for _, group := range result.groups {
		label, _ := EvaluateActionabilityWithLabel(group)
		if label == PatternTestPreamble {
			t.Error("preamble containing if-guards + t.Errorf must NOT match test-preamble")
		}
	}
}

// TestPatternTestPreambleProductionFileNotMatched: the test-file guard —
// identical preamble statements in production files stay actionable.
func TestPatternTestPreambleProductionFileNotMatched(t *testing.T) {
	t.Parallel()

	src := `package repo

func saveA() error {
	x := open()
	_ = x
	return nil
}

func saveB() error {
	x := open()
	_ = x
	return nil
}
`

	result := runPipeline(t, map[string]string{
		"k_prod.go": src,
		"l_prod.go": src,
	}, 2)

	for _, group := range result.groups {
		label, _ := EvaluateActionabilityWithLabel(group)
		if label == PatternTestPreamble {
			t.Error("production-file sequence must NOT match test-preamble (allFromTestFile guard)")
		}
	}
}
