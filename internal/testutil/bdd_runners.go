package testutil

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// errExecutorNil is returned when BDDTestSetup.Executor is nil.
var errExecutorNil = errors.New(
	"BDDTestSetup.Executor is nil — must be set before running commands",
)

// errNoStdinPaths is returned when stdin content has no file paths.
var errNoStdinPaths = errors.New("no file paths provided in stdin content")

// commandError wraps a command error with output context.
func commandError(msg string, err error, output []byte) error {
	return fmt.Errorf("%s: %w\nOutput: %s", msg, err, string(output))
}

// helper marks the calling function as a test helper when a *testing.T is bound.
// No-op when T is nil (Ginkgo panics-on-error mode).
func (s *BDDTestSetup) helper() {
	if s.T != nil {
		s.T.Helper()
	}
}

// runExecutor executes the in-process command and returns combined stdout+stderr.
func (s *BDDTestSetup) runExecutor(args ...string) ([]byte, error) {
	if s.Executor == nil {
		return nil, errExecutorNil
	}

	return s.Executor(args...)
}

// runExecutorStdout executes the in-process command and returns stdout only.
// Uses ExecutorResult for separated stdout/stderr, falls back to Executor if not set.
func (s *BDDTestSetup) runExecutorStdout(args ...string) ([]byte, error) {
	if s.ExecutorResult != nil {
		result, err := s.ExecutorResult(args...)
		if err != nil {
			return result.Stdout, err
		}

		return result.Stdout, nil
	}

	// Fallback: use combined executor (may include stderr noise)
	return s.runExecutor(args...)
}

// RunArtDupl executes art-dupl with given arguments and returns combined stdout/stderr.
func (s *BDDTestSetup) RunArtDupl(args ...string) ([]byte, error) {
	return s.RunArtDuplOnDir(s.TmpDir, args...)
}

// RunArtDuplOnDir executes art-dupl on a specific directory with given arguments.
func (s *BDDTestSetup) RunArtDuplOnDir(dir string, args ...string) ([]byte, error) {
	s.helper()

	allArgs := append([]string{dir}, args...)

	return s.runExecutor(allArgs...)
}

// RunArtDuplWithFlags executes art-dupl with flag map and returns combined output.
func (s *BDDTestSetup) RunArtDuplWithFlags(flags map[string]string) ([]byte, error) {
	return s.RunArtDuplOnDirWithFlags(s.TmpDir, flags)
}

// RunArtDuplOnDirWithFlags executes art-dupl on directory with flag map.
func (s *BDDTestSetup) RunArtDuplOnDirWithFlags(
	dir string,
	flags map[string]string,
) ([]byte, error) {
	s.helper()

	args := BuildArgsFromFlags([]string{dir}, flags)

	return s.runExecutor(args...)
}

// RunArtDuplAllFormat runs art-dupl with --all flag to generate all output formats.
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

// RunArtDuplWithStdin executes art-dupl with real stdin input via pipe injection.
// It replaces os.Stdin with a pipe containing the provided content and adds the
// --files flag so the CLI reads paths from stdin through feedFromStdin.
// This exercises the real stdin code path end-to-end.
func (s *BDDTestSetup) RunArtDuplWithStdin(
	stdinContent string,
	flags map[string]string,
) ([]byte, error) {
	s.helper()

	if strings.TrimSpace(stdinContent) == "" {
		return nil, errNoStdinPaths
	}

	if _, exists := flags["files"]; !exists {
		flags["files"] = ""
	}

	args := BuildArgsFromFlags(nil, flags)

	r, w, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe: %w", err)
	}

	go func() {
		defer func() { _ = w.Close() }()

		_, _ = io.WriteString(w, stdinContent)
	}()

	origStdin := os.Stdin

	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
		_ = r.Close()
	}()

	return s.runExecutor(args...)
}

// prepareSubcommandArgs prepares arguments for a subcommand, adding the temp directory if needed.
// Only appends TmpDir if no path-like argument is found (i.e., no non-flag arg besides subcommand names).
// It tracks flag-value pairs (e.g., --threshold 5) so numeric values aren't mistaken for paths.
func (s *BDDTestSetup) prepareSubcommandArgs(args ...string) []string {
	s.helper()

	subcommands := map[string]bool{
		"stats": true, "version": true, "completion": true, "man": true,
	}

	// Flags that take a value (non-boolean). Boolean flags like --vendor don't consume the next arg.
	valueFlags := map[string]bool{
		"--threshold": true, "-t": true,
		"--format":            true,
		"--sort":              true,
		"--detection-methods": true, "-m": true,
		"--output-dir": true,
		"--timeout":    true,
		"--only":       true,
		"--cache-dir":  true,
		"--diff":       true,
		"--config":     true, "-c": true,
		"--include-generated": true,
	}

	expectValue := false
	for _, arg := range args {
		if expectValue {
			expectValue = false

			continue
		}

		if strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "=") {
				continue
			}

			expectValue = valueFlags[arg]

			continue
		}

		if subcommands[arg] {
			continue
		}

		// Found a non-flag, non-subcommand, non-value arg — treat as path
		return args
	}

	// No path found — append TmpDir
	return append(args, s.TmpDir)
}

// RunSubcommand executes an art-dupl subcommand (e.g., "stats") with given arguments.
func (s *BDDTestSetup) RunSubcommand(args ...string) ([]byte, error) {
	allArgs := s.prepareSubcommandArgs(args...)

	return s.runExecutor(allArgs...)
}

// RunSubcommandOutput executes an art-dupl subcommand and returns stdout only.
// Use this for JSON or other structured output where stderr contamination is undesirable.
func (s *BDDTestSetup) RunSubcommandOutput(args ...string) ([]byte, error) {
	allArgs := s.prepareSubcommandArgs(args...)

	return s.runExecutorStdout(allArgs...)
}

// RunStatsSubcommandWithJSON runs the stats subcommand with JSON format and parses the result.
func (s *BDDTestSetup) RunStatsSubcommandWithJSON(threshold string) (map[string]any, error) {
	args := s.prepareSubcommandArgs("stats", "--format", "json", "--threshold", threshold)

	output, err := s.runExecutorStdout(args...)
	if err != nil {
		return nil, commandError(
			fmt.Sprintf("stats execution failed (threshold: %s)", threshold),
			err,
			output,
		)
	}

	var result map[string]any

	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, commandError("failed to unmarshal JSON output", err, output)
	}

	return result, nil
}

// runCommandAndVerify executes a command function and verifies it completes successfully.
func (s *BDDTestSetup) runCommandAndVerify(execute func() ([]byte, error)) string {
	s.helper()

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
//
//nolint:nonamedreturns // Multiple return values for stdout/stderr
func (s *BDDTestSetup) RunArtDuplAndCapture(args ...string) (stdout, stderr []byte, err error) {
	s.helper()

	if s.ExecutorResult != nil {
		result, execErr := s.ExecutorResult(args...)

		return result.Stdout, result.Stderr, execErr
	}

	// Fallback: can't separate
	output, execErr := s.runExecutor(args...)

	return output, nil, execErr
}
