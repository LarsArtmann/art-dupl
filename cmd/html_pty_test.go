package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// This is the end-to-end lock for the HTML auto-write path: the unit tests
// inject the TTY probe, so they verify the branch logic but not that a REAL
// terminal actually flows through it. Running the compiled binary under
// script(1) gives the child process a genuine PTY on stdout, exercising
// probe → auto-write → notice with no fakes.
//
// The test skips when script(1) is unavailable (non-Linux CI images, stripped
// containers) — the injected-probe unit tests still cover the branch logic.
func TestHTMLAutoWriteOnRealPTY(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("script(1) based PTY test is Linux-only")
	}

	scriptPath, err := exec.LookPath("script")
	if err != nil {
		t.Skip("script(1) not found; PTY end-to-end test skipped")
	}

	binary := buildTestBinary(t)

	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "a.go"), "package a\n\nfunc f() int {\n\tx := 1\n\ty := 1\n\treturn x + y\n}\n\nfunc g() int {\n\ta := 2\n\tb := 2\n\treturn a + b\n}\n", 0o600)
	writeTestFile(t, filepath.Join(dir, "b.go"), "package b\n\nfunc h() int {\n\tx := 3\n\ty := 3\n\treturn x + y\n}\n\nfunc i() int {\n\ta := 4\n\tb := 4\n\treturn a + b\n}\n", 0o600)

	typescript := filepath.Join(t.TempDir(), "typescript")

	// script(1) assigns the child a PTY on stdout, which makes the TTY probe
	// inside the binary report a terminal and trigger the auto-write branch.
	// The child's combined output lands in the typescript file.
	scriptCmd := exec.Command(scriptPath, "-q", "-e", "-c",
		binary+" --html --threshold 3 --min-lines 2 .", typescript)
	scriptCmd.Dir = dir
	scriptCmd.Env = os.Environ()

	output, err := scriptCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("script(1) run failed: %v\noutput:\n%s", err, output)
	}

	reportPath := filepath.Join(dir, defaultHTMLReportPath)

	content, readErr := os.ReadFile(reportPath)
	if readErr != nil {
		t.Fatalf("expected auto-written HTML report at %s on a real PTY, read error: %v\noutput:\n%s",
			reportPath, readErr, output)
	}

	if !strings.Contains(string(content), "<!DOCTYPE html>") && !strings.Contains(string(content), "<html") {
		t.Errorf("auto-written report does not look like HTML: %.120s", content)
	}

	combined := string(output) + "\n" + string(content)
	if !strings.Contains(combined, defaultHTMLReportPath) {
		t.Errorf("expected the auto-write notice to name %s, got:\n%s", defaultHTMLReportPath, combined)
	}
}
