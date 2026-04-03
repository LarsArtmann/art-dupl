package bdd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// BenchmarkSemanticDetectionRealWorld benchmarks semantic detection
// using the actual art-dupl codebase as test data.
//
// This provides a realistic performance measurement for the --semantic flag.
func BenchmarkSemanticDetectionRealWorld(b *testing.B) {
	// Build binary once before benchmark
	binaryPath := filepath.Join(b.TempDir(), "art-dupl-bench")

	buildCmd := exec.CommandContext(b.Context(), "go", "build", "-o", binaryPath, "../cmd/art-dupl")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Get project root (two levels up from bdd/)
	//nolint:dogsled // Standard pattern for getting current file path
	_, currentFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..")

	b.Run("WithoutSemantic", func(b *testing.B) {
		for range b.N {
			cmd := exec.CommandContext(b.Context(), binaryPath, projectRoot, "--threshold", "50")

			cmd.Env = append(os.Environ(), "ARTDUPL_NO_PROGRESS=1")

			output, err := cmd.CombinedOutput()
			if err != nil {
				b.Fatalf("art-dupl failed: %v\nOutput: %s", err, string(output))
			}
		}
	})

	b.Run("WithSemantic", func(b *testing.B) {
		for range b.N {
			cmd := exec.CommandContext(
				b.Context(),
				binaryPath,
				projectRoot,
				"--threshold",
				"50",
				"--semantic",
			)

			cmd.Env = append(os.Environ(), "ARTDUPL_NO_PROGRESS=1")

			output, err := cmd.CombinedOutput()
			if err != nil {
				b.Fatalf("art-dupl failed: %v\nOutput: %s", err, string(output))
			}
		}
	})
}

// BenchmarkSemanticDetectionSynthetic benchmarks on synthetic test data
// to measure overhead without the noise of large file I/O.
func BenchmarkSemanticDetectionSynthetic(b *testing.B) {
	// Build binary once before benchmark
	binaryPath := filepath.Join(b.TempDir(), "art-dupl-bench")

	buildCmd := exec.CommandContext(b.Context(), "go", "build", "-o", binaryPath, "../cmd/art-dupl")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	// Create test directory with synthetic code
	testDir := b.TempDir()
	code := `package test

type ServiceA struct{}
func (s *ServiceA) Process() error { return nil }
func (s *ServiceA) Validate() bool { return true }

type ServiceB struct{}
func (s *ServiceB) Process() error { return nil }
func (s *ServiceB) Validate() bool { return true }
`
	// Create 20 test files
	for i := range 20 {
		filename := filepath.Join(testDir, fmt.Sprintf("file%d.go", i))
		err := os.WriteFile(filename, []byte(code), 0o644)
		if err != nil {
			b.Fatalf("Failed to write test file: %v", err)
		}
	}

	b.Run("WithoutSemantic", func(b *testing.B) {
		for range b.N {
			cmd := exec.CommandContext(b.Context(), binaryPath, testDir, "--threshold", "10")

			cmd.Env = append(os.Environ(), "ARTDUPL_NO_PROGRESS=1")
			if output, err := cmd.CombinedOutput(); err != nil {
				b.Fatalf("art-dupl failed: %v\nOutput: %s", err, string(output))
			}
		}
	})

	b.Run("WithSemantic", func(b *testing.B) {
		for range b.N {
			cmd := exec.CommandContext(
				b.Context(),
				binaryPath,
				testDir,
				"--threshold",
				"10",
				"--semantic",
			)

			cmd.Env = append(os.Environ(), "ARTDUPL_NO_PROGRESS=1")
			if output, err := cmd.CombinedOutput(); err != nil {
				b.Fatalf("art-dupl failed: %v\nOutput: %s", err, string(output))
			}
		}
	})
}
