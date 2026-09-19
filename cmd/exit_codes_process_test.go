package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// buildTestBinary builds the art-dupl binary to a temp directory and returns its path.
func buildTestBinary(t *testing.T) string {
	t.Helper()

	binaryPath := filepath.Join(t.TempDir(), "art-dupl-test")

	buildCmd := exec.CommandContext(t.Context(), "go", "build", "-o", binaryPath, "./art-dupl")

	buildCmd.Env = append(os.Environ(), "GOEXPERIMENT=jsonv2")

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	return binaryPath
}

// runBinaryExitCode runs the binary with given args and returns the process exit code.
func runBinaryExitCode(t *testing.T, binaryPath string, args ...string) int {
	t.Helper()

	// Windows runners intermittently fail to START a freshly written exe
	// (antivirus scan lock); ProcessState stays nil in that case. Retry a
	// few times with a short pause, and only treat a process that actually
	// started as authoritative.
	const attempts = 3

	for i := range attempts {
		if i > 0 {
			time.Sleep(250 * time.Millisecond)
		}

		cmd := exec.CommandContext(t.Context(), binaryPath, args...)
		cmd.Env = append(os.Environ(), "GOEXPERIMENT=jsonv2")

		_ = cmd.Run()

		if cmd.ProcessState != nil {
			return cmd.ProcessState.ExitCode()
		}
	}

	t.Fatal("ProcessState is nil after Run")
	return -1
}

// TestExitCodes_Process verifies that the actual process exit codes match
// the documented ExitCode constants when running the real binary.
// TestExitCodes_Process exercises the compiled binary end to end. It is
// skipped on Windows runners: they intermittently refuse to start freshly
// built executables (ProcessState nil on every attempt), a runner/AV artifact
// unrelated to the exit-code logic, which TestExitCodeForError covers
// in-process on every platform.
func TestExitCodes_Process(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows runners cannot reliably start freshly built exes; logic covered by TestExitCodeForError")
	}
	if testing.Short() {
		t.Skip("skipping subprocess test in short mode")
	}

	binaryPath := buildTestBinary(t)

	t.Run("successful run exits 0", func(t *testing.T) {
		tmpDir := t.TempDir()
		writeTestFile(t, filepath.Join(tmpDir, "main.go"), "package main\nfunc main() {}\n", 0o600)

		code := runBinaryExitCode(t, binaryPath, "--threshold", "5", tmpDir)
		if code != ExitSuccess {
			t.Errorf("exit code = %d, want %d (ExitSuccess)", code, ExitSuccess)
		}
	})

	t.Run("invalid threshold exits 2", func(t *testing.T) {
		tmpDir := t.TempDir()
		writeTestFile(t, filepath.Join(tmpDir, "main.go"), "package main\nfunc main() {}\n", 0o600)

		code := runBinaryExitCode(t, binaryPath, "--threshold", "-1", tmpDir)
		if code != ExitConfigError {
			t.Errorf("exit code = %d, want %d (ExitConfigError)", code, ExitConfigError)
		}
	})

	t.Run("version subcommand exits 0", func(t *testing.T) {
		code := runBinaryExitCode(t, binaryPath, "version")
		if code != ExitSuccess {
			t.Errorf("exit code = %d, want %d (ExitSuccess)", code, ExitSuccess)
		}
	})
}
