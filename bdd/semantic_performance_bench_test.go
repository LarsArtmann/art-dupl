package bdd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func buildBenchBinary(b *testing.B) string {
	b.Helper()

	binaryPath := filepath.Join(b.TempDir(), "art-dupl-bench")

	buildCmd := exec.CommandContext(b.Context(), "go", "build", "-o", binaryPath, "../cmd/art-dupl")

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		b.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	return binaryPath
}

func runBenchBinary(b *testing.B, binaryPath, target, threshold string, extraArgs ...string) {
	b.Helper()

	args := append([]string{target, "--threshold", threshold}, extraArgs...)
	cmd := exec.CommandContext(b.Context(), binaryPath, args...)

	cmd.Env = append(os.Environ(), "ARTDUPL_NO_PROGRESS=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		b.Fatalf("art-dupl failed: %v\nOutput: %s", err, string(output))
	}
}

func getProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	return filepath.Join(wd, "..")
}

func writeBenchFiles(b *testing.B, dir string, count int, content string) {
	b.Helper()

	for i := range count {
		filename := filepath.Join(dir, fmt.Sprintf("file%d.go", i))

		err := os.WriteFile(filename, []byte(content), 0o644)
		if err != nil {
			b.Fatalf("Failed to write test file: %v", err)
		}
	}
}

func BenchmarkSemanticDetectionRealWorld(b *testing.B) {
	binaryPath := buildBenchBinary(b)
	projectRoot := getProjectRoot()

	b.Run("WithoutSemantic", func(b *testing.B) {
		for range b.N {
			runBenchBinary(b, binaryPath, projectRoot, "50")
		}
	})

	b.Run("WithSemantic", func(b *testing.B) {
		for range b.N {
			runBenchBinary(b, binaryPath, projectRoot, "50", "--semantic")
		}
	})
}

func BenchmarkSemanticDetectionSynthetic(b *testing.B) {
	binaryPath := buildBenchBinary(b)

	testDir := b.TempDir()
	code := `package test

type ServiceA struct{}
func (s *ServiceA) Process() error { return nil }
func (s *ServiceA) Validate() bool { return true }

type ServiceB struct{}
func (s *ServiceB) Process() error { return nil }
func (s *ServiceB) Validate() bool { return true }`

	writeBenchFiles(b, testDir, 20, code)

	b.Run("WithoutSemantic", func(b *testing.B) {
		for range b.N {
			runBenchBinary(b, binaryPath, testDir, "10")
		}
	})

	b.Run("WithSemantic", func(b *testing.B) {
		for range b.N {
			runBenchBinary(b, binaryPath, testDir, "10", "--semantic")
		}
	})
}
