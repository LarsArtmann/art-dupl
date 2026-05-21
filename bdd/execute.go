package bdd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/LarsArtmann/art-dupl/cmd"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// executeInProcess runs art-dupl in-process via Cobra commands.
// It captures stdout and stderr by duplicating file descriptors at the OS level,
// which preserves os.Stdout/os.Stderr for Go runtime use (Ginkgo output etc).
func executeInProcess(args ...string) (*testutil.CommandResult, error) {
	rootCmd := cmd.NewRootCommand()
	cmd.AddFlags(rootCmd)

	rootCmd.Version = cmd.GetVersion()
	rootCmd.SetVersionTemplate(cmd.GetVersion() + "\n")

	rootCmd.SetArgs(args)

	// Save original fd 1 (stdout) and fd 2 (stderr) by duplicating them
	savedStdout, _ := syscall.Dup(1)
	savedStderr, _ := syscall.Dup(2)

	// Create pipes
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	// Redirect fd 1 and fd 2 to the pipe write ends
	syscall.Dup2(int(stdoutW.Fd()), 1)
	syscall.Dup2(int(stderrW.Fd()), 2)

	var (
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)

	stdoutDone := make(chan struct{})

	go func() {
		_, _ = io.Copy(&stdoutBuf, stdoutR)

		close(stdoutDone)
	}()

	stderrDone := make(chan struct{})

	go func() {
		_, _ = io.Copy(&stderrBuf, stderrR)

		close(stderrDone)
	}()

	err := rootCmd.Execute()

	// Restore original fds
	syscall.Dup2(savedStdout, 1)
	syscall.Dup2(savedStderr, 2)

	syscall.Close(savedStdout)
	syscall.Close(savedStderr)

	stdoutW.Close()
	stderrW.Close()

	<-stdoutDone
	<-stderrDone

	result := &testutil.CommandResult{
		Stdout: stdoutBuf.Bytes(),
		Stderr: stderrBuf.Bytes(),
	}

	if err != nil {
		return result, fmt.Errorf("art-dupl %v: %w", args, err)
	}

	return result, nil
}
