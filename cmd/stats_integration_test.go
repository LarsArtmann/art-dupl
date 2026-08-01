package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// hasAllStrings returns a function that checks if all substrings exist in the given text.
func hasAllStrings(output string, substrings ...string) func() bool {
	return func() bool {
		return checkSubstrings(output, false, substrings)
	}
}

// containsAnySubstring returns a function that checks if any substring exists in given output.
func containsAnySubstring(output string, substrings ...string) func() bool {
	return func() bool {
		return checkSubstrings(output, true, substrings)
	}
}

// checkSubstrings checks if any (anyMode=true) or all (anyMode=false) substrings exist in output.
func checkSubstrings(output string, anyMode bool, substrings []string) bool {
	for _, sub := range substrings {
		found := strings.Contains(output, sub)
		if anyMode && found {
			return true
		}

		if !anyMode && !found {
			return false
		}
	}

	return !anyMode
}

const (
	statsSubCommand = "stats"
	statsHeaderText = "Code Duplication Statistics"
)

// executeTestCommand runs art-dupl in-process with the given arguments.
// It resolves relative paths against the repo root and captures output via fd duplication.
//
// Output capture mutates the process-global os.Stdout/os.Stderr, so tests that
// call this helper MUST be serial (no t.Parallel): see testutil.CaptureStdoutStderr.
func executeTestCommand(t *testing.T, args []string) ([]byte, error) {
	t.Helper()

	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("Failed to find repo root: %v", err)
	}

	// Resolve relative paths against repo root
	resolved := make([]string, 0, len(args))
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") && strings.Contains(arg, "/") && !filepath.IsAbs(arg) {
			resolved = append(resolved, filepath.Join(repoRoot, arg))
		} else {
			resolved = append(resolved, arg)
		}
	}

	rootCmd := NewRootCommand()
	AddFlags(rootCmd)
	rootCmd.SetArgs(resolved[1:])

	output, execErr := testutil.CaptureCombinedOutput(rootCmd.Execute)

	return output, execErr
}

func TestStatsCommandIntegration(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedInOutput []string
		wantErr          bool
	}{
		{
			name: "stats on current directory",
			args: []string{binaryName, statsSubCommand, "."},
			expectedInOutput: []string{
				statsHeaderText,
				"Files Scanned:",
				"Clone Groups:",
				"Total Clones:",
				"Total Duplicate Lines:",
				"Average Clone Size:",
				"Complexity Score:",
				"Impact Score:",
			},
			wantErr: false,
		},
		{
			name: "stats on printer directory",
			args: []string{binaryName, statsSubCommand, "-t", "1", "--no-accept-directives", "./printer"},
			expectedInOutput: []string{
				statsHeaderText,
				"Files Scanned:",
				"Top Files by Duplicate Lines:",
			},
			wantErr: false,
		},
		{
			name: "stats with threshold flag",
			args: []string{binaryName, statsSubCommand, "-t", "20", "."},
			expectedInOutput: []string{
				"Threshold: 20 tokens",
				statsHeaderText,
			},
			wantErr: false,
		},
		{
			name: "stats with multiple paths",
			args: []string{binaryName, statsSubCommand, "./cmd", "./printer"},
			expectedInOutput: []string{
				statsHeaderText,
				"Files Scanned:",
			},
			wantErr: false,
		},
		{
			name: "stats help",
			args: []string{binaryName, statsSubCommand, "--help"},
			expectedInOutput: []string{
				"Prints comprehensive duplication statistics",
				"art-dupl stats",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := executeTestCommand(t, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Command error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			outputStr := string(output)
			for _, expected := range tt.expectedInOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Output missing expected string: %q", expected)
					t.Logf("Got output: %s", outputStr)
				}
			}
		})
	}
}

func TestStatsCommandErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "stats with non-existent path",
			args:    []string{binaryName, statsSubCommand, "/non/existent/path"},
			wantErr: true,
		},
		{
			name:    "stats with invalid threshold",
			args:    []string{binaryName, statsSubCommand, "-t", "-5", "."},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeTestCommand(t, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Command error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStatsOutputFormat(t *testing.T) {
	output, err := executeTestCommand(
		t,
		[]string{binaryName, statsSubCommand, "-t", "1", "--no-accept-directives", "./printer"},
	)
	if err != nil {
		t.Fatalf("Stats command failed: %v", err)
	}

	outputStr := string(output)

	checks := []struct {
		name  string
		check func() bool
	}{
		{
			name: "has header section",
			check: hasAllStrings(
				outputStr,
				statsHeaderText,
				"============================",
			),
		},
		{
			name:  "has configuration section",
			check: hasAllStrings(outputStr, "Configuration:", "Threshold:", "Detection Methods:"),
		},
		{
			name:  "has overview section",
			check: hasAllStrings(outputStr, "Overview:", "Files Scanned:", "Clone Groups:"),
		},
		{
			name:  "has duplicate code section",
			check: hasAllStrings(outputStr, "Duplicate Code:", "Total Duplicate Lines:"),
		},
		{
			name: "has size distribution section",
			check: containsAnySubstring(
				outputStr,
				"Clone Size Distribution:",
				"Top Files by Duplicate Lines:",
			),
		},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if !check.check() {
				t.Errorf("Output doesn't contain expected section: %s", check.name)
			}
		})
	}
}

func TestCreateOutputWriter(t *testing.T) {
	t.Run("empty filename returns stdout", func(t *testing.T) {
		w, cleanup, err := createOutputWriter("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if w != os.Stdout {
			t.Error("expected os.Stdout")
		}

		cleanup()
	})

	t.Run("valid filename creates file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "stats-output.txt")

		w, cleanup, err := createOutputWriter(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, writeErr := w.Write([]byte("test data\n"))
		if writeErr != nil {
			t.Fatalf("write error: %v", writeErr)
		}

		cleanup()

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read error: %v", readErr)
		}

		if string(data) != "test data\n" {
			t.Errorf("got %q, want %q", string(data), "test data\n")
		}
	})

	t.Run("invalid path returns error", func(t *testing.T) {
		_, _, err := createOutputWriter("/nonexistent/dir/output.txt")
		if err == nil {
			t.Error("expected error for invalid path")
		}
	})
}

func TestStatsOutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "stats.txt")

	_, err := executeTestCommand(t, []string{
		binaryName, statsSubCommand,
		"--output-file", outputPath,
		"./printer",
	})
	if err != nil {
		t.Fatalf("Stats command with --output-file failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(data)
	if !strings.Contains(outputStr, statsHeaderText) {
		t.Errorf("Output file missing expected header %q", statsHeaderText)
		t.Logf("Got: %s", outputStr)
	}
}

// findRepoRoot finds the repository root directory.
func findRepoRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		for _, name := range []string{"go.mod", ".git"} {
			if _, err := os.Stat(filepath.Join(current, name)); err == nil {
				return current, nil
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}

		current = parent
	}

	return "", os.ErrNotExist
}
