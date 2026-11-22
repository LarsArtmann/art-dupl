package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeExecutable(t *testing.T, dir, name string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}

	return path
}

// TestArtDuplPathCandidatesDedupesSymlinks pins the nix case: a system-path
// symlink and its store target must count as ONE copy, in first-seen order.
func TestArtDuplPathCandidatesDedupesSymlinks(t *testing.T) {
	t.Parallel()

	realDir := t.TempDir()
	target := writeExecutable(t, realDir, "art-dupl")

	linkDir := t.TempDir()
	if err := os.Symlink(target, filepath.Join(linkDir, "art-dupl")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	emptyDir := t.TempDir()

	got := artDuplPathCandidates(strings.Join([]string{linkDir, realDir, emptyDir}, string(os.PathListSeparator)))
	if len(got) != 1 || got[0] != target {
		t.Errorf("candidates = %v, want exactly [%s]", got, target)
	}
}

// TestArtDuplPathCandidatesSkipsNonExecutables pins that directory entries
// and non-executable files are not reported as copies.
func TestArtDuplPathCandidatesSkipsNonExecutables(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataDir, "art-dupl"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if got := artDuplPathCandidates(dataDir); len(got) != 0 {
		t.Errorf("candidates = %v, want none for a non-executable file", got)
	}
}

// TestWarnIfMultipleArtDuplInPath covers the warning text: both copies named
// with versions, the running one marked, the PATH winner marked. The version
// probe is stubbed so no child binaries execute.
func TestWarnIfMultipleArtDuplInPath(t *testing.T) {
	t.Parallel()

	dirA := t.TempDir()
	binA := writeExecutable(t, dirA, "art-dupl")
	dirB := t.TempDir()
	binB := writeExecutable(t, dirB, "art-dupl")

	original := probeArtDuplVersion
	probeArtDuplVersion = func(path string) string {
		if path == binB {
			return "v9.9.9-test"
		}

		return "v0.0.0-test"
	}
	t.Cleanup(func() { probeArtDuplVersion = original })

	pathEnv := strings.Join([]string{dirA, dirB}, string(os.PathListSeparator))

	var stderr strings.Builder
	warnIfMultipleArtDuplInPath(&stderr, pathEnv)

	out := stderr.String()
	for _, want := range []string{"2 times in PATH", binA, binB, "v9.9.9-test", "first in PATH, wins"} {
		if !strings.Contains(out, want) {
			t.Errorf("warning missing %q in:\n%s", want, out)
		}
	}
}

// TestWarnIfMultipleArtDuplInPathSilentForSingleCopy pins the no-noise
// contract: one copy (the normal case) prints nothing.
func TestWarnIfMultipleArtDuplInPathSilentForSingleCopy(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeExecutable(t, dir, "art-dupl")

	var stderr strings.Builder
	warnIfMultipleArtDuplInPath(&stderr, dir)

	if stderr.Len() != 0 {
		t.Errorf("warning printed for a single copy:\n%s", stderr.String())
	}
}
