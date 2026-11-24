package cmd

import (
	"os"
	"path/filepath"
	"runtime"
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
// Expectations compare against the RESOLVED target: on macOS t.TempDir sits
// under /var (a symlink to /private/var), so candidates come back resolved
// (/private/var/...) while writeExecutable's return value is unresolved.
func TestArtDuplPathCandidatesDedupesSymlinks(t *testing.T) {
	t.Parallel()

	realDir := t.TempDir()
	target := writeExecutable(t, realDir, artDuplExecutableName())
	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		resolvedTarget = target
	}

	linkDir := t.TempDir()
	if err := os.Symlink(target, filepath.Join(linkDir, artDuplExecutableName())); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	emptyDir := t.TempDir()

	got := artDuplPathCandidates(strings.Join([]string{linkDir, realDir, emptyDir}, string(os.PathListSeparator)))
	if len(got) != 1 || got[0] != resolvedTarget {
		t.Errorf("candidates = %v, want exactly [%s]", got, resolvedTarget)
	}
}

// TestArtDuplPathCandidatesSkipsNonExecutables pins that directory entries
// and non-executable files are not reported as copies. On windows there is
// no exec bit: an existing art-dupl.exe runs regardless of its file mode, so
// production deliberately skips the perm check there and the candidate IS
// reported.
func TestArtDuplPathCandidatesSkipsNonExecutables(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataDir, artDuplExecutableName()), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got := artDuplPathCandidates(dataDir)
	if runtime.GOOS == "windows" {
		if len(got) != 1 {
			t.Errorf("candidates = %v, want exactly the existing .exe (windows has no exec bit)", got)
		}

		return
	}

	if len(got) != 0 {
		t.Errorf("candidates = %v, want none for a non-executable file", got)
	}
}

// TestWarnIfMultipleArtDuplInPath covers the warning text: both copies named
// with versions, the running one marked, the PATH winner marked. The version
// probe is stubbed so no child binaries execute. Assertions use the RESOLVED
// paths — candidates are realpath'd (macOS /var -> /private/var), and the
// stub matches the resolved form of binB for the same reason.
func TestWarnIfMultipleArtDuplInPath(t *testing.T) {
	t.Parallel()

	dirA := t.TempDir()
	binA := writeExecutable(t, dirA, artDuplExecutableName())
	dirB := t.TempDir()
	binB := writeExecutable(t, dirB, artDuplExecutableName())
	resolve := func(p string) string {
		r, err := filepath.EvalSymlinks(p)
		if err != nil {
			return p
		}
		return r
	}
	resolvedA, resolvedB := resolve(binA), resolve(binB)

	stub := func(path string) string {
		if path == resolvedB {
			return "v9.9.9-test"
		}

		return "v0.0.0-test"
	}
	// The probe rides the parameter (warnIfMultipleArtDuplInPathWithProbe):
	// writing the package-level seam var from a parallel test raced with
	// parallel readers of the version command (-race, -shuffle repro).

	pathEnv := strings.Join([]string{dirA, dirB}, string(os.PathListSeparator))

	var stderr strings.Builder
	warnIfMultipleArtDuplInPathWithProbe(&stderr, pathEnv, stub)

	out := stderr.String()
	for _, want := range []string{"2 times in PATH", resolvedA, resolvedB, "v9.9.9-test", "first in PATH, wins"} {
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
