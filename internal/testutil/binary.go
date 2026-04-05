// Package testutil provides testing utilities for art-dupl.
//
// This package contains helper functions for building and running the art-dupl
// binary during tests, as well as other testing utilities.
package testutil

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// BuildArtDuplBinary builds the art-dupl binary for testing.
func BuildArtDuplBinary(t *testing.T, outputPath string) {
	t.Helper()

	//nolint:gosec // G204: Go build command with controlled arguments is safe
	cmd := exec.CommandContext(
		t.Context(),
		"go",
		"build",
		"-o",
		outputPath,
		"../../cmd/art-dupl/main.go",
	)

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

	//nolint:gosec // G204: Safe - binaryPath from controlled build, args from test code
	cmd := exec.CommandContext(
		t.Context(),
		binaryPath,
		args...)

	return cmd.CombinedOutput()
}

// BuildArgsFromFlags converts a map of flags to command line arguments.
func BuildArgsFromFlags(baseArgs []string, flags map[string]string) []string {
	args := baseArgs

	for flag, value := range flags {
		if value != "" {
			args = append(args, "--"+flag, value)
		} else {
			args = append(args, "--"+flag)
		}
	}

	return args
}

// RunArtDuplBinaryOnDir executes art-dupl on a directory with given flags.
func RunArtDuplBinaryOnDir(
	t *testing.T,
	binaryPath, dir string,
	flags map[string]string,
) ([]byte, error) {
	t.Helper()

	args := BuildArgsFromFlags([]string{dir}, flags)

	return RunArtDuplBinary(t, binaryPath, args...)
}
