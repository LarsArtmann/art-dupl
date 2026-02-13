package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatsCommandIntegration(t *testing.T) {
	// Build the binary first
	binaryPath := filepath.Join(t.TempDir(), "art-dupl")
	if err := buildBinary(binaryPath); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name             string
		args             []string
		expectedInOutput []string
		wantErr          bool
	}{
		{
			name: "stats on current directory",
			args: []string{binaryPath, "stats", "."},
			expectedInOutput: []string{
				"Code Duplication Statistics",
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
			args: []string{binaryPath, "stats", "./printer"},
			expectedInOutput: []string{
				"Code Duplication Statistics",
				"Files Scanned:",
				"Top Files by Duplicate Lines:",
			},
			wantErr: false,
		},
		{
			name: "stats with threshold flag",
			args: []string{binaryPath, "stats", "-t", "20", "."},
			expectedInOutput: []string{
				"Threshold: 20 tokens",
				"Code Duplication Statistics",
			},
			wantErr: false,
		},
		{
			name: "stats with multiple paths",
			args: []string{binaryPath, "stats", "./cmd", "./printer"},
			expectedInOutput: []string{
				"Code Duplication Statistics",
				"Files Scanned:",
			},
			wantErr: false,
		},
		{
			name: "stats help",
			args: []string{binaryPath, "stats", "--help"},
			expectedInOutput: []string{
				"stats displays aggregated statistics about code duplication",
				"art-dupl stats",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find repo root for relative paths
			repoRoot, err := findRepoRoot()
			if err != nil {
				t.Fatalf("Failed to find repo root: %v", err)
			}

			cmd := &Command{
				Path: tt.args[0],
				Args: tt.args,
				Dir:  repoRoot,
			}

			output, err := cmd.CombinedOutput()

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
	binaryPath := filepath.Join(t.TempDir(), "art-dupl")
	if err := buildBinary(binaryPath); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "stats with non-existent path",
			args:    []string{binaryPath, "stats", "/non/existent/path"},
			wantErr: true,
		},
		{
			name:    "stats with invalid threshold",
			args:    []string{binaryPath, "stats", "-t", "-5", "."},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find repo root for relative paths
			repoRoot, err := findRepoRoot()
			if err != nil {
				t.Fatalf("Failed to find repo root: %v", err)
			}

			cmd := &Command{
				Path: tt.args[0],
				Args: tt.args,
				Dir:  repoRoot,
			}

			_, err = cmd.CombinedOutput()
			if (err != nil) != tt.wantErr {
				t.Errorf("Command error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStatsOutputFormat(t *testing.T) {
	binaryPath := filepath.Join(t.TempDir(), "art-dupl")
	if err := buildBinary(binaryPath); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Find repo root for relative paths
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("Failed to find repo root: %v", err)
	}

	// Test that stats output has structured format
	cmd := &Command{
		Path: binaryPath,
		Args: []string{binaryPath, "stats", "./printer"},
		Dir:  repoRoot,
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Stats command failed: %v", err)
	}

	outputStr := string(output)

	// Verify structure
	checks := []struct {
		name  string
		check func() bool
	}{
		{
			name: "has header section",
			check: func() bool {
				return strings.Contains(outputStr, "Code Duplication Statistics") &&
					strings.Contains(outputStr, "============================")
			},
		},
		{
			name: "has configuration section",
			check: func() bool {
				return strings.Contains(outputStr, "Configuration:") &&
					strings.Contains(outputStr, "Threshold:") &&
					strings.Contains(outputStr, "Detection Methods:")
			},
		},
		{
			name: "has overview section",
			check: func() bool {
				return strings.Contains(outputStr, "Overview:") &&
					strings.Contains(outputStr, "Files Scanned:") &&
					strings.Contains(outputStr, "Clone Groups:")
			},
		},
		{
			name: "has duplicate code section",
			check: func() bool {
				return strings.Contains(outputStr, "Duplicate Code:") &&
					strings.Contains(outputStr, "Total Duplicate Lines:")
			},
		},
		{
			name: "has size distribution section",
			check: func() bool {
				return strings.Contains(outputStr, "Clone Size Distribution:") ||
					strings.Contains(outputStr, "Top Files by Duplicate Lines:")
			},
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

// buildBinary builds the art-dupl binary for testing.
func buildBinary(outputPath string) error {
	// Change to repo root
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	// Save current directory and restore later
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }() // test cleanup

	if err := os.Chdir(repoRoot); err != nil {
		return err //nolint:wrapcheck // Test helper - pass through os.Chdir error
	}

	// Build the binary
	cmd := &Command{
		Path: "go",
		Args: []string{"go", "build", "-o", outputPath, "-ldflags", "-s -w", "-trimpath", "./cmd/art-dupl"},
	}

	return cmd.Run()
}

// findRepoRoot finds the repository root directory.
func findRepoRoot() (string, error) {
	// Start from current directory and look for git repo or go.mod
	current, err := os.Getwd()
	if err != nil {
		return "", err //nolint:wrapcheck // Test helper - pass through os.Getwd error
	}

	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", os.ErrNotExist
}

// Command is a simple command wrapper for testing.
type Command struct {
	Path string
	Args []string
	Env  []string
	Dir  string
}

// buildCmd creates and configures the exec.Command.
//
//nolint:funcorder // helper method
func (c *Command) buildCmd() *exec.Cmd {
	cmd := exec.Command(c.Path, c.Args[1:]...) // Args[0] is the binary path
	if c.Dir != "" {
		cmd.Dir = c.Dir
	}
	if len(c.Env) > 0 {
		cmd.Env = c.Env
	}
	return cmd
}

// CombinedOutput runs the command and returns its combined stdout and stderr.
func (c *Command) CombinedOutput() ([]byte, error) {
	return c.buildCmd().CombinedOutput() //nolint:wrapcheck // Test helper - pass through exec error
}

// Run runs the command.
func (c *Command) Run() error {
	return c.buildCmd().Run() //nolint:wrapcheck // Test helper - pass through exec error
}
