package bdd

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/cmd"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// executeInProcess runs art-dupl in-process via Cobra commands.
// It captures stdout and stderr by temporarily replacing os.Stdout/os.Stderr,
// which works cross-platform (the previous syscall.Dup/Dup2 approach was Unix-only).
func executeInProcess(args ...string) (*testutil.CommandResult, error) {
	rootCmd := cmd.NewRootCommand()
	cmd.AddFlags(rootCmd)

	rootCmd.Version = cmd.GetVersion()
	rootCmd.SetVersionTemplate("art-dupl version " + cmd.GetVersion() + "\n")

	rootCmd.SetArgs(args)

	stdout, stderr, err := testutil.CaptureStdoutStderr(rootCmd.Execute)

	result := &testutil.CommandResult{
		Stdout: stdout,
		Stderr: stderr,
	}

	if err != nil {
		return result, fmt.Errorf("art-dupl %v: %w", args, err)
	}

	return result, nil
}
