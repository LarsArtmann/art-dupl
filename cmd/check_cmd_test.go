package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
)

// checkFixtureFamilyA/B are two functions with an identical body (one clone
// group), differing only in names so semantic mode reports a Type-2 clone.
const checkFixtureFamilyA = `package dup

func alpha(x int) int {
	a := x + 1
	b := a * 2
	c := b - 3
	d := c + 4
	e := d * 5
	f := e - 6
	return f
}
`

const checkFixtureFamilyB = `package dup

func beta(y int) int {
	a := y + 1
	b := a * 2
	c := b - 3
	d := c + 4
	e := d * 5
	f := e - 6
	return f
}
`

// checkFixtureFamilyC/D are a SECOND, structurally different clone family used
// to introduce a genuinely new group after a baseline was recorded.
const checkFixtureFamilyC = `package dup

func gamma(z int) int {
	p := z * 2
	q := p + 7
	r := q / 2
	s := r - 9
	t := s * 3
	u := t + 11
	return u
}
`

const checkFixtureFamilyD = `package dup

func delta(w int) int {
	p := w * 2
	q := p + 7
	r := q / 2
	s := r - 9
	t := s * 3
	u := t + 11
	return u
}
`

func writeCheckFixture(t *testing.T, dir, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600)
	if err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// recordCheckBaseline writes the baseline for the fixture directory and
// returns its path.
func recordCheckBaseline(t *testing.T, dir string, extraArgs ...string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "baseline.json")
	args := append([]string{binaryName, "baseline", "--baseline-path", path, "-t", "3", "--no-actionability"}, extraArgs...)
	args = append(args, dir)

	output, err := executeTestCommand(t, args)
	if err != nil {
		t.Fatalf("baseline failed: %v\nOutput: %s", err, output)
	}

	return path
}

// TestCheckSummarySplitsBaselineFromNew (issue #12): the summary line must
// report how many detected groups the baseline already accepts and how many
// are new, instead of a single count that conflates the two populations.
func TestCheckSummarySplitsBaselineFromNew(t *testing.T) {
	dir := t.TempDir()
	writeCheckFixture(t, dir, "a.go", checkFixtureFamilyA)
	writeCheckFixture(t, dir, "b.go", checkFixtureFamilyB)

	baselinePath := recordCheckBaseline(t, dir)

	output, err := executeTestCommand(t, []string{
		binaryName, "check", "--baseline-path", baselinePath,
		"-t", "3", "--no-actionability", dir,
	})
	if err != nil {
		t.Fatalf("check with unchanged tree failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "clone groups (1 in baseline, 0 new)") {
		t.Errorf("expected baseline split '(1 in baseline, 0 new)', got:\n%s", output)
	}

	// Introduce a second clone family: the split must name it as new while
	// still counting the absorbed group.
	writeCheckFixture(t, dir, "c.go", checkFixtureFamilyC)
	writeCheckFixture(t, dir, "d.go", checkFixtureFamilyD)

	output, err = executeTestCommand(t, []string{
		binaryName, "check", "--baseline-path", baselinePath,
		"-t", "3", "--no-actionability", dir,
	})
	if err == nil {
		t.Fatalf("check with a new clone family must exit non-zero, got success:\n%s", output)
	}

	if !strings.Contains(string(output), "clone groups (1 in baseline, 1 new)") {
		t.Errorf("expected baseline split '(1 in baseline, 1 new)', got:\n%s", output)
	}
}

// TestCheckSummaryReportsSuppressedGroups: when filters suppress groups, the
// split must say so rather than presenting a green verdict that merely hides
// everything.
func TestCheckSummaryReportsSuppressedGroups(t *testing.T) {
	dir := t.TempDir()
	writeCheckFixture(t, dir, "a.go", checkFixtureFamilyA)
	writeCheckFixture(t, dir, "b.go", checkFixtureFamilyB)

	baselinePath := recordCheckBaseline(t, dir)

	output, err := executeTestCommand(t, []string{
		binaryName, "check", "--baseline-path", baselinePath,
		"-t", "3", "--no-actionability", "--min-lines", "99", dir,
	})
	if err != nil {
		t.Fatalf("check with suppression failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "clone groups (0 in baseline, 0 new, 1 suppressed)") {
		t.Errorf("expected baseline split naming the suppressed group, got:\n%s", output)
	}
}

// TestCheckBaselineVersionMustMatchAcceptsSamePolicy: a check run whose
// analyzer identity matches the baseline's must proceed normally.
func TestCheckBaselineVersionMustMatchAcceptsSamePolicy(t *testing.T) {
	dir := t.TempDir()
	writeCheckFixture(t, dir, "a.go", checkFixtureFamilyA)
	writeCheckFixture(t, dir, "b.go", checkFixtureFamilyB)

	baselinePath := recordCheckBaseline(t, dir)

	output, err := executeTestCommand(t, []string{
		binaryName, "check", "--baseline-path", baselinePath,
		"--" + baselineVersionMustMatchFlag,
		"-t", "3", "--no-actionability", dir,
	})
	if err != nil {
		t.Fatalf("matching identity must pass: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "0 new") {
		t.Errorf("expected a zero-new split, got:\n%s", output)
	}
}

// TestCheckBaselineVersionMustMatchRefusesDifferentPolicy (issue #13): a
// baseline recorded under different hash-affecting settings must be refused
// with both identities named, instead of reporting the whole tree as new.
func TestCheckBaselineVersionMustMatchRefusesDifferentPolicy(t *testing.T) {
	dir := t.TempDir()
	writeCheckFixture(t, dir, "a.go", checkFixtureFamilyA)
	writeCheckFixture(t, dir, "b.go", checkFixtureFamilyB)

	baselinePath := recordCheckBaseline(t, dir)

	output, err := executeTestCommand(t, []string{
		binaryName, "check", "--baseline-path", baselinePath,
		"--" + baselineVersionMustMatchFlag, "--exact",
		"-t", "3", "--no-actionability", dir,
	})
	if err == nil {
		t.Fatalf("mismatched identity must fail, got success:\n%s", output)
	}

	text := string(output)
	if !strings.Contains(text, "was recorded by art-dupl") {
		t.Errorf("error must name the recording analyzer, got:\n%s", text)
	}

	if !strings.Contains(text, "hashing policy") {
		t.Errorf("error must name the hashing policy, got:\n%s", text)
	}
}

// TestCheckBaselineVersionMustMatchRefusesLegacyBaseline: a baseline written
// before identity tracking cannot be verified, so the strict gate refuses it
// rather than silently checking unverified hashes.
func TestCheckBaselineVersionMustMatchRefusesLegacyBaseline(t *testing.T) {
	dir := t.TempDir()
	writeCheckFixture(t, dir, "a.go", checkFixtureFamilyA)
	writeCheckFixture(t, dir, "b.go", checkFixtureFamilyB)

	baselinePath := recordCheckBaseline(t, dir)

	legacyPath := filepath.Join(t.TempDir(), "legacy.json")
	stripAnalyzerIdentity(t, baselinePath, legacyPath)

	output, err := executeTestCommand(t, []string{
		binaryName, "check", "--baseline-path", legacyPath,
		"--" + baselineVersionMustMatchFlag,
		"-t", "3", "--no-actionability", dir,
	})
	if err == nil {
		t.Fatalf("legacy baseline must be refused by the strict gate, got success:\n%s", output)
	}

	if !strings.Contains(string(output), "carries no analyzer identity") {
		t.Errorf("error must explain the missing identity, got:\n%s", output)
	}
}

// TestCheckLegacyBaselineWithoutFlagStillWorks: the strict gate is opt-in, so
// existing CI setups keep comparing unverified hashes.
func TestCheckLegacyBaselineWithoutFlagStillWorks(t *testing.T) {
	dir := t.TempDir()
	writeCheckFixture(t, dir, "a.go", checkFixtureFamilyA)
	writeCheckFixture(t, dir, "b.go", checkFixtureFamilyB)

	baselinePath := recordCheckBaseline(t, dir)

	legacyPath := filepath.Join(t.TempDir(), "legacy.json")
	stripAnalyzerIdentity(t, baselinePath, legacyPath)

	output, err := executeTestCommand(t, []string{
		binaryName, "check", "--baseline-path", legacyPath,
		"-t", "3", "--no-actionability", dir,
	})
	if err != nil {
		t.Fatalf("legacy baseline without the strict flag must pass: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "1 in baseline") {
		t.Errorf("expected the baseline split line, got:\n%s", output)
	}
}

// stripAnalyzerIdentity rewrites a baseline without its analyzer identity,
// reproducing a file recorded by an older binary.
func stripAnalyzerIdentity(t *testing.T, src, dst string) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read baseline %s: %v", src, err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("decode baseline %s: %v", src, err)
	}

	delete(parsed, "analyzer")

	stripped, err := json.Marshal(parsed)
	if err != nil {
		t.Fatalf("encode stripped baseline: %v", err)
	}

	if err := os.WriteFile(dst, stripped, 0o600); err != nil {
		t.Fatalf("write stripped baseline %s: %v", dst, err)
	}
}

// analyzerIdentityForMode is a helper for the policy-sensitivity test below.
func analyzerIdentityForMode(t *testing.T, mutate func(*config.Config)) string {
	t.Helper()

	cfg := config.DefaultConfig()
	mutate(cfg)

	return analyzerIdentity(cfg).Policy
}

func TestAnalyzerIdentityTracksHashAffectingSettings(t *testing.T) {
	t.Parallel()

	baselinePolicy := analyzerIdentityForMode(t, func(*config.Config) {})

	mutations := map[string]func(*config.Config){
		"detection mode": func(c *config.Config) { c.DetectionMode = config.DetectionModeExact },
		"type aware":     func(c *config.Config) { c.TypeAware = true },
		"suggest generics": func(c *config.Config) {
			c.SuggestGenerics = true
		},
		"max children": func(c *config.Config) { c.MaxChildrenSerial = 7 },
		"detection methods": func(c *config.Config) {
			c.DetectionMethods = append(c.DetectionMethods, config.DetectionMethodHash)
		},
	}

	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := analyzerIdentityForMode(t, mutate); got == baselinePolicy {
				t.Errorf("%s must change the hashing policy, got identical policy %s", name, got)
			}
		})
	}
}

// TestDetectionMethodsPolicyKeyIsOrderIndependent: the policy keys on the
// detector SET, not its spelling order.
func TestDetectionMethodsPolicyKeyIsOrderIndependent(t *testing.T) {
	t.Parallel()

	a := detectionMethodsPolicyKey([]config.DetectionMethod{
		config.DetectionMethodArtDupl, config.DetectionMethodHash,
	})
	b := detectionMethodsPolicyKey([]config.DetectionMethod{
		config.DetectionMethodHash, config.DetectionMethodArtDupl,
	})

	if a != b {
		t.Errorf("method-set key depends on order: %q vs %q", a, b)
	}
}
