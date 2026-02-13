package testutil

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
)

// BuildArtDuplBinary builds the art-dupl binary for testing.
func BuildArtDuplBinary(t *testing.T, outputPath string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", outputPath, "../../cmd/art-dupl/main.go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build art-dupl binary: %v\nOutput: %s", err, string(output))
	}
}

// BuildAndCleanArtDuplBinary builds the art-dupl binary and returns a cleanup function.
// The cleanup function should be called in defer to remove the binary after the test.
func BuildAndCleanArtDuplBinary(t *testing.T) string {
	t.Helper()
	binaryPath := filepath.Join(t.TempDir(), "art-dupl-test")
	BuildArtDuplBinary(t, binaryPath)
	return binaryPath
}

// RunArtDuplBinary executes the art-dupl binary with given arguments and returns output.
func RunArtDuplBinary(t *testing.T, binaryPath string, args ...string) ([]byte, error) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), binaryPath, args...)
	return cmd.CombinedOutput() //nolint:wrapcheck // Test helper - pass through exec error
}

// RunArtDuplBinaryOnDir executes art-dupl on a directory with given flags.
func RunArtDuplBinaryOnDir(t *testing.T, binaryPath, dir string, flags map[string]string) ([]byte, error) {
	t.Helper()
	args := []string{dir}
	for flag, value := range flags {
		if value != "" {
			args = append(args, "--"+flag, value)
		} else {
			args = append(args, "--"+flag)
		}
	}
	return RunArtDuplBinary(t, binaryPath, args...)
}
