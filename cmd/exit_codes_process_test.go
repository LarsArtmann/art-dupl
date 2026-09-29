package cmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildTestBinary builds the art-dupl binary to a temp directory and returns its path.
// The -o name carries an explicit .exe suffix: cmd/go writes the EXACT -o name (no
// suffix appended on windows), and os/exec on windows only starts executables whose
// name ends with a PATHEXT extension (findExecutable returns ErrNotFound for an
// extension-less absolute path) — the extensionless name produced deterministic
// "ProcessState is nil" failures on windows runners that four rounds of AV
// hypotheses chased before the stdlib mechanism was proven.
func buildTestBinary(t *testing.T) string {
	t.Helper()

	binaryPath := filepath.Join(t.TempDir(), "art-dupl-test.exe")

	buildCmd := exec.CommandContext(t.Context(), "go", "build", "-o", binaryPath, "./art-dupl")

	buildCmd.Env = append(os.Environ(), "GOEXPERIMENT=jsonv2")

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	return binaryPath
}

// runBinaryExitCode runs the binary with given args and returns the process
// exit code. A failure to START the process is fatal (with the error text —
// the discarded-start-error blind spot cost weeks of AV hypotheses); a
// non-zero exit is a legitimate result the caller asserts on.
func runBinaryExitCode(t *testing.T, binaryPath string, args ...string) int {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), binaryPath, args...)
	cmd.Env = append(os.Environ(), "GOEXPERIMENT=jsonv2")

	err := cmd.Run()
	if err != nil {
		if _, started := errors.AsType[*exec.ExitError](err); !started {
			t.Fatalf("failed to start %s: %v", binaryPath, err)
		}
	}

	return cmd.ProcessState.ExitCode()
}

// TestExitCodes_Process verifies that the actual process exit codes match
// the documented ExitCode constants when running the real binary.
// TestExitCodes_Process exercises the compiled binary end to end and verifies
// the documented ExitCode constants on every platform. (It ran windows-skipped
// for weeks under an antivirus hypothesis; the real cause was the extension-
// less -o output name — see buildTestBinary.)
func TestExitCodes_Process(t *testing.T) {
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
