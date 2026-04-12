package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strings"
)

// RunArtDupl executes art-dupl binary with given arguments and returns combined stdout/stderr.
func (s *BDDTestSetup) RunArtDupl(args ...string) ([]byte, error) {
	return s.RunArtDuplOnDir(s.TmpDir, args...)
}

// RunArtDuplOnDir executes art-dupl binary on a specific directory with given arguments.
func (s *BDDTestSetup) RunArtDuplOnDir(dir string, args ...string) ([]byte, error) {
	if s.T != nil {
		s.T.Helper()
	}

	cmd := exec.CommandContext(
		context.Background(),
		s.BinaryPath,
		append([]string{dir}, args...)...) // #nosec G204 -- Test helper running project binary

	return cmd.CombinedOutput()
}

// RunArtDuplWithFlags executes art-dupl binary with flag map and returns combined output.
func (s *BDDTestSetup) RunArtDuplWithFlags(flags map[string]string) ([]byte, error) {
	return s.RunArtDuplOnDirWithFlags(s.TmpDir, flags)
}

// RunArtDuplOnDirWithFlags executes art-dupl on directory with flag map.
func (s *BDDTestSetup) RunArtDuplOnDirWithFlags(
	dir string,
	flags map[string]string,
) ([]byte, error) {
	if s.T != nil {
		s.T.Helper()
	}

	args := BuildArgsFromFlags([]string{dir}, flags)

	cmd := exec.CommandContext(
		context.Background(),
		s.BinaryPath,
		args...) // #nosec G204 -- Test helper running project binary

	return cmd.CombinedOutput()
}

// RunArtDuplAllFormat runs art-dupl with --all flag to generate all output formats.
// The output directory will be created if it doesn't exist.
func (s *BDDTestSetup) RunArtDuplAllFormat(outputDir, threshold string) ([]byte, error) {
	return s.RunArtDuplOnDir(
		s.TmpDir,
		"--all",
		"--output-dir",
		outputDir,
		"--threshold",
		threshold,
	)
}

// RunArtDuplWithStdin executes art-dupl with stdin input.
func (s *BDDTestSetup) RunArtDuplWithStdin(stdin string, flags map[string]string) ([]byte, error) {
	if s.T != nil {
		s.T.Helper()
	}

	args := BuildArgsFromFlags([]string{"--files"}, flags)

	cmd := exec.CommandContext(
		context.Background(),
		s.BinaryPath,
		args...) // #nosec G204 -- Test helper running project binary
	cmd.Stdin = strings.NewReader(stdin)

	return cmd.CombinedOutput()
}

// prepareSubcommandArgs prepares arguments for a subcommand, adding the temp directory if needed.
// It returns the prepared arguments and a command ready to be executed.
func (s *BDDTestSetup) prepareSubcommandArgs(args ...string) ([]string, *exec.Cmd) {
	if s.T != nil {
		s.T.Helper()
	}

	// Append the temp directory at the end if not already specified
	hasDir := slices.Contains(args, s.TmpDir)

	if !hasDir {
		args = append(args, s.TmpDir)
	}

	cmd := exec.CommandContext(
		context.Background(),
		s.BinaryPath,
		args...) // #nosec G204 -- Test helper running project binary

	return args, cmd
}

// RunSubcommand executes an art-dupl subcommand (e.g., "stats") with given arguments.
// The subcommand name should be the first argument, followed by flags and the directory.
// Example: RunSubcommand("stats", "--format", "json", "--threshold", "10").
func (s *BDDTestSetup) RunSubcommand(args ...string) ([]byte, error) {
	_, cmd := s.prepareSubcommandArgs(args...)

	return cmd.CombinedOutput()
}

// RunSubcommandOutput executes an art-dupl subcommand and returns stdout only.
// Use this for JSON or other structured output where stderr contamination is undesirable.
func (s *BDDTestSetup) RunSubcommandOutput(args ...string) ([]byte, error) {
	_, cmd := s.prepareSubcommandArgs(args...)

	return cmd.Output()
}

// RunStatsSubcommandWithJSON runs the stats subcommand with JSON format and parses the result.
// The threshold parameter is required and specifies the minimum clone size to report.
func (s *BDDTestSetup) RunStatsSubcommandWithJSON(threshold string) (map[string]any, error) {
	_, cmd := s.prepareSubcommandArgs("stats", "--format", "json", "--threshold", threshold)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run stats command: %w\nOutput: %s", err, string(output))
	}

	var result map[string]any

	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal JSON output: %w\nOutput: %s",
			err,
			string(output),
		)
	}

	return result, nil
}

// runCommandAndVerify executes a command function and verifies it completes successfully.
func (s *BDDTestSetup) runCommandAndVerify(execute func() ([]byte, error)) string {
	if s.T != nil {
		s.T.Helper()
	}

	output, err := execute()
	if err != nil {
		if s.T != nil {
			s.T.Fatalf("art-dupl command failed: %v\nOutput: %s", err, string(output))
		} else {
			panic(fmt.Sprintf("art-dupl command failed: %v\nOutput: %s", err, string(output)))
		}
	}

	return string(output)
}

// RunArtDuplAndVerifyOutput executes art-dupl and verifies it completes successfully.
func (s *BDDTestSetup) RunArtDuplAndVerifyOutput(args ...string) string {
	return s.runCommandAndVerify(func() ([]byte, error) {
		return s.RunArtDupl(args...)
	})
}

// RunArtDuplWithFlagsAndVerify executes art-dupl with flags and verifies success.
func (s *BDDTestSetup) RunArtDuplWithFlagsAndVerify(flags map[string]string) string {
	return s.runCommandAndVerify(func() ([]byte, error) {
		return s.RunArtDuplWithFlags(flags)
	})
}

// RunArtDuplAndCapture executes art-dupl and captures stdout and stderr separately.
// Returns both outputs and any error that occurred.
//
//nolint:nonamedreturns // Multiple return values for stdout/stderr
func (s *BDDTestSetup) RunArtDuplAndCapture(args ...string) (stdout, stderr []byte, err error) {
	if s.T != nil {
		s.T.Helper()
	}

	// #nosec G204 -- Test helper running project binary
	cmd := exec.CommandContext(context.Background(), s.BinaryPath, args...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start command: %w", err)
	}

	stdout, err = io.ReadAll(stdoutPipe)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read stdout: %w", err)
	}

	stderr, err = io.ReadAll(stderrPipe)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read stderr: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return stdout, stderr, fmt.Errorf("command failed: %w", err)
	}

	return stdout, stderr, nil
}
