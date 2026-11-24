package gitignore

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestDiagWindowsGitignoreProbe is a TEMPORARY diagnostic for the windows CI
// failures of TestIsIgnoredDir / TestCrawlPrunesGitignoredDirs: it logs the
// exact intermediate values of the IsIgnoredDir probe chain (rules, Rel
// output, ToSlash, split components) so the failing step becomes visible in
// the CI log. Remove once the root cause is fixed and pinned.
func TestDiagWindowsGitignoreProbe(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("diagnostic targets the windows failure class")
	}

	tmpDir := t.TempDir()
	t.Logf("tmpDir=%q ToSlash=%q", tmpDir, filepath.ToSlash(tmpDir))

	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("logs/\nbuild\n*.tmp/\n!important.log\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, ".gitignore")); err != nil {
		t.Fatalf("Stat .gitignore: %v", err)
	}

	m := LoadGitignore([]string{tmpDir}, io.Discard)
	if m == nil {
		t.Fatalf("matcher nil: LoadGitignore(%q) found no rules", tmpDir)
	}

	t.Logf("rules=%d", len(m.rules))
	for i, r := range m.rules {
		t.Logf("rule[%d] baseDir=%q patterns=%d", i, r.baseDir, len(r.patterns))
		for _, p := range r.patterns {
			t.Logf("rule[%d] pattern raw=%q negated=%v dirOnly=%v anchored=%v", i, p.raw, p.negated, p.dirOnly, p.anchored)
		}
	}

	probe := filepath.Join(tmpDir, "logs")
	joined := filepath.Join(probe, "\x00")
	abs, err := filepath.Abs(joined)
	if err != nil {
		t.Fatalf("Abs: %v", err)
	}
	t.Logf("probe=%q joined=%q abs=%q", probe, joined, abs)

	for i, r := range m.rules {
		rel, err := filepath.Rel(r.baseDir, abs)
		if err != nil {
			t.Logf("rule[%d] Rel err=%v", i, err)
			continue
		}
		clean := filepath.ToSlash(rel)
		t.Logf("rule[%d] Rel=%q ToSlash=%q parts=%q", i, rel, clean, strings.Split(clean, "/"))
	}

	t.Logf("IsIgnoredDir=%v IsIgnored(joined)=%v IsIgnored(probe)=%v",
		m.IsIgnoredDir(probe), m.IsIgnored(joined), m.IsIgnored(probe))

	fileProbe := filepath.Join(tmpDir, "logs", "x.txt")
	relFile, relErr := filepath.Rel(m.rules[0].baseDir, fileProbe)
	t.Logf("fileProbe=%q rel=%q err=%v IsIgnored=%v", fileProbe, relFile, relErr, m.IsIgnored(fileProbe))

	t.Logf("Match sanity: Match(\"logs\", \"logs\")=%v Match(\"*.tmp\", \"cache.tmp\")=%v",
		matchGlob("logs", "logs"), matchGlob("*.tmp", "cache.tmp"))
}
