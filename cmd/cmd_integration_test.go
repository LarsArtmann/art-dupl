package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// --- Integration Tests for Command Handlers ---

const flagKeyThreshold = "--threshold"

func writeTestFile(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}

func buildSuffixTreeWithFile(
	t *testing.T,
	tmpDir string,
	cfg *config.Config,
) (*suffixtree.STree, []*syntax.Node, job.ParseStats) {
	t.Helper()
	ctx := t.Context()

	result := buildSuffixTree(buildParams{
		ctx:          ctx,
		paths:        []string{tmpDir},
		cfg:          cfg,
		filterParam:  nil,
		outputFormat: config.OutputFormatText,
	})
	if result.err != nil {
		t.Fatalf("buildSuffixTree() error = %v", result.err)
	}

	if result.tree == nil {
		t.Error("buildSuffixTree() returned nil tree")
	}

	if len(result.data) == 0 {
		t.Error("buildSuffixTree() returned empty data")
	}

	if result.parseStats.FilesCount < 1 {
		t.Errorf("Expected at least 1 file parsed, got %d", result.parseStats.FilesCount)
	}

	return result.tree, result.data, result.parseStats
}

func TestRunCmd_Integration(t *testing.T) {
	t.Run("basic execution with duplicate code", func(t *testing.T) {
		tmpDir := t.TempDir()

		createDuplicateTestFiles(t, tmpDir)

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{flagKeyThreshold, "10", tmpDir})

		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("runCmd() error = %v", err)
		}
	})

	t.Run("invalid sort option returns error", func(t *testing.T) {
		tmpDir := t.TempDir()

		file1 := filepath.Join(tmpDir, "file1.go")
		writeTestFile(t, file1, "package main\n", 0o600)

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{"--sort", "invalid", tmpDir})

		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err == nil {
			t.Error("runCmd() expected error for invalid sort option")
		}
	})

	t.Run("invalid detection methods returns error", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{"--detection-methods", "invalid-method", tmpDir})

		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err == nil {
			t.Error("runCmd() expected error for invalid detection methods")
		}
	})

	t.Run("output format", func(t *testing.T) {
		for _, tc := range []struct {
			name       string
			formatFlag string
		}{
			{"json", "--json"},
			{"plumbing", "--plumbing"},
			{"html", "--html"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				runOutputFormatTest(t, tc.formatFlag)
			})
		}
	})

	t.Run("verbose output", func(t *testing.T) {
		tmpDir := t.TempDir()

		file1 := filepath.Join(tmpDir, "file1.go")
		writeTestFile(t, file1, "package main\nfunc main() {}\n", 0o600)

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{"--verbose", flagKeyThreshold, "10", tmpDir})

		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("runCmd() error = %v", err)
		}
	})

	t.Run("with vendor flag", func(t *testing.T) {
		tmpDir := t.TempDir()

		vendorDir := filepath.Join(tmpDir, "vendor")
		if err := os.MkdirAll(vendorDir, 0o750); err != nil {
			t.Fatalf("Failed to create vendor dir: %v", err)
		}

		file1 := filepath.Join(vendorDir, "file1.go")
		writeTestFile(t, file1, "package main\n", 0o600)

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{"--vendor", flagKeyThreshold, "10", tmpDir})

		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("runCmd() error = %v", err)
		}
	})
}

func TestWorkers_AutoDetection(t *testing.T) {
	t.Run("workers 0 uses parallel path without hanging", func(t *testing.T) {
		tmpDir := t.TempDir()
		createDuplicateTestFiles(t, tmpDir)

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{"--workers", "0", flagKeyThreshold, "10", tmpDir})

		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("workers=0: runCmd() error = %v", err)
		}
	})

	t.Run("workers 0 and workers 1 produce same result", func(t *testing.T) {
		tmpDir := t.TempDir()
		createDuplicateTestFiles(t, tmpDir)

		cfg0 := config.DefaultConfig()
		cfg0.Threshold = 10
		cfg0.Workers = 0
		cfg0.DetectionMethods = nil

		_, data0, ps0 := buildSuffixTreeWithFile(t, tmpDir, cfg0)

		cfg1 := config.DefaultConfig()
		cfg1.Threshold = 10
		cfg1.Workers = 1
		cfg1.DetectionMethods = nil

		_, data1, ps1 := buildSuffixTreeWithFile(t, tmpDir, cfg1)

		if ps0.FilesCount != ps1.FilesCount {
			t.Errorf("FilesCount: workers=0 got %d, workers=1 got %d", ps0.FilesCount, ps1.FilesCount)
		}

		if len(data0) != len(data1) {
			t.Errorf("data length: workers=0 got %d, workers=1 got %d", len(data0), len(data1))
		}
	})
}

func TestRunStats_Integration(t *testing.T) {
	t.Run("basic stats execution", func(t *testing.T) {
		tmpDir := t.TempDir()

		createDuplicateTestFiles(t, tmpDir)

		cmd := NewStatsCommand()
		cmd.SetArgs([]string{flagKeyThreshold, "10", tmpDir})

		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("runStats() error = %v", err)
		}
	})

	t.Run("stats format", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			format string
		}{
			{"json", "json"},
			{"csv", "csv"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				runStatsFormatTest(t, tc.format)
			})
		}
	})

	t.Run("stats invalid format returns error", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := NewStatsCommand()
		cmd.SetArgs([]string{"--format", "invalid", tmpDir})

		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err == nil {
			t.Error("runStats() expected error for invalid format")
		}
	})

	t.Run("stats with verbose", func(t *testing.T) {
		tmpDir := t.TempDir()

		file1 := filepath.Join(tmpDir, "file1.go")
		writeTestFile(t, file1, "package main\nfunc main() {}\n", 0o600)

		cmd := NewStatsCommand()
		cmd.SetArgs([]string{"--verbose", flagKeyThreshold, "10", tmpDir})

		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("runStats() error = %v", err)
		}
	})
}

func TestRunAllModes_Integration(t *testing.T) {
	t.Run("generates all output formats", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputDir := filepath.Join(tmpDir, "reports")

		createDuplicateTestFiles(t, tmpDir)

		cmd := NewRootCommand()
		AddFlags(cmd)
		cmd.SetArgs([]string{"--all", "--output-dir", outputDir, flagKeyThreshold, "10", tmpDir})

		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.Execute()
		if err != nil {
			t.Errorf("runAllModes() error = %v", err)
		}

		formats := []string{"text", "html", "json", "plumbing", "simple-json"}
		for _, format := range formats {
			reportFile := filepath.Join(outputDir, "report."+format)
			if _, statErr := os.Stat(reportFile); os.IsNotExist(statErr) {
				t.Errorf("Expected report file %s to be created", reportFile)
			}
		}
	})
}

func TestExecuteAnalysis_Integration(t *testing.T) {
	t.Run("basic analysis", func(t *testing.T) {
		tmpDir := t.TempDir()

		createDuplicateTestFiles(t, tmpDir)

		cfg := &config.Config{
			Threshold:        10,
			DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
			IncludeTempl:     true,
			IncludeSQLC:      true,
		}

		ctx := t.Context()

		duplChan, parseStats, filterStats, err := executeAnalysis(
			ctx,
			cfg,
			[]string{tmpDir},
			config.OutputFormatText,
		)
		if err != nil {
			t.Fatalf("executeAnalysis() error = %v", err)
		}

		matchCount := 0
		for range duplChan {
			matchCount++
		}

		if parseStats.FilesCount < 2 {
			t.Errorf("Expected at least 2 files parsed, got %d", parseStats.FilesCount)
		}

		_ = filterStats
		_ = matchCount
	})

	t.Run("analysis with profile enabled", func(t *testing.T) {
		tmpDir := t.TempDir()

		file1 := filepath.Join(tmpDir, "file1.go")
		writeTestFile(t, file1, "package main\nfunc main() {}\n", 0o600)

		cfg := &config.Config{
			Threshold:        10,
			Profile:          true,
			DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
			IncludeTempl:     true,
			IncludeSQLC:      true,
		}

		ctx := t.Context()

		duplChan, _, _, err := executeAnalysis(ctx, cfg, []string{tmpDir}, config.OutputFormatText)
		if err != nil {
			t.Fatalf("executeAnalysis() error = %v", err)
		}

		for range duplChan {
			// Drain channel
		}
	})
}

func TestBuildSuffixTree_Integration(t *testing.T) {
	t.Run("builds tree from files", func(t *testing.T) {
		tmpDir := t.TempDir()

		code := `package test

func Example() int {
	return 42
}
`

		file1 := filepath.Join(tmpDir, "file1.go")
		if err := os.WriteFile(file1, []byte(code), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cfg := &config.Config{
			Threshold:        10,
			DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
		}

		_, _, _ = buildSuffixTreeWithFile(t, tmpDir, cfg)
	})

	t.Run("verbose output", func(t *testing.T) {
		tmpDir := t.TempDir()

		file1 := filepath.Join(tmpDir, "file1.go")
		writeTestFile(t, file1, "package main\n", 0o600)

		cfg := &config.Config{
			Verbose:          true,
			DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
		}

		tree, _, _ := buildSuffixTreeWithFile(t, tmpDir, cfg)
		_ = tree
	})
}
