package testutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/utils"
)

// BDDTestSetup provides complete BDD test infrastructure with temporary directory and binary management.
type BDDTestSetup struct {
	T             *testing.T
	TmpDir        string
	FileProcessor *utils.FileProcessor
	BinaryPath    string
}

// NewBDDTestSetup creates a new BDD test setup with temporary directory and builds art-dupl binary.
// The caller is responsible for cleaning up the temporary directory using Cleanup() or manually.
func NewBDDTestSetup(t *testing.T) *BDDTestSetup {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "art-dupl-bdd-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	binaryPath := filepath.Join(tmpDir, "art-dupl-test")
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", binaryPath, "../cmd/art-dupl/main.go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to build art-dupl binary: %v\nOutput: %s", err, string(output))
	}

	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	return &BDDTestSetup{
		T:             t,
		TmpDir:        tmpDir,
		FileProcessor: utils.NewFileProcessor(tmpDir),
		BinaryPath:    binaryPath,
	}
}

// NewBDDTestSetupForGinkgo creates a new BDD test setup without requiring *testing.T.
// Designed for use with Ginkgo's BeforeEach/AfterEach pattern.
// The caller must call Cleanup() in an AfterEach block.
func NewBDDTestSetupForGinkgo() (*BDDTestSetup, error) {
	tmpDir, err := os.MkdirTemp("", "art-dupl-bdd-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	binaryPath := filepath.Join(tmpDir, "art-dupl-test")
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", binaryPath, "../cmd/art-dupl/main.go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to build art-dupl binary: %w\nOutput: %s", err, string(output))
	}

	return &BDDTestSetup{
		TmpDir:        tmpDir,
		FileProcessor: utils.NewFileProcessor(tmpDir),
		BinaryPath:    binaryPath,
	}, nil
}

// Cleanup removes the temporary directory and all its contents.
func (s *BDDTestSetup) Cleanup() error {
	return os.RemoveAll(s.TmpDir)
}

// CreateDuplicateFiles creates multiple files with identical content.
func (s *BDDTestSetup) CreateDuplicateFiles(filenames []string, content string) error {
	return s.FileProcessor.WriteDuplicateFiles(filenames, content)
}

// CreateTestFile creates a single test file with given content.
func (s *BDDTestSetup) CreateTestFile(filename, content string) error {
	return s.FileProcessor.WriteTextFile(filename, content)
}

// CreateTestFiles creates multiple test files from a map.
func (s *BDDTestSetup) CreateTestFiles(files map[string]string) error {
	return s.FileProcessor.WriteTestFiles(files)
}

// GetFilePath returns full path for a file in test directory.
func (s *BDDTestSetup) GetFilePath(filename string) string {
	return filepath.Join(s.TmpDir, filename)
}

// RunArtDupl executes art-dupl binary with given arguments and returns combined stdout/stderr.
func (s *BDDTestSetup) RunArtDupl(args ...string) ([]byte, error) {
	return s.RunArtDuplOnDir(s.TmpDir, args...)
}

// RunArtDuplOnDir executes art-dupl binary on a specific directory with given arguments.
func (s *BDDTestSetup) RunArtDuplOnDir(dir string, args ...string) ([]byte, error) {
	if s.T != nil {
		s.T.Helper()
	}
	cmd := exec.CommandContext(context.Background(), s.BinaryPath, append([]string{dir}, args...)...)
	return cmd.CombinedOutput()
}

// RunArtDuplWithFlags executes art-dupl binary with flag map and returns combined output.
func (s *BDDTestSetup) RunArtDuplWithFlags(flags map[string]string) ([]byte, error) {
	return s.RunArtDuplOnDirWithFlags(s.TmpDir, flags)
}

// RunArtDuplOnDirWithFlags executes art-dupl on directory with flag map.
func (s *BDDTestSetup) RunArtDuplOnDirWithFlags(dir string, flags map[string]string) ([]byte, error) {
	if s.T != nil {
		s.T.Helper()
	}

	args := []string{dir}
	for flag, value := range flags {
		if value != "" {
			args = append(args, fmt.Sprintf("--%s", flag), value)
		} else {
			args = append(args, fmt.Sprintf("--%s", flag))
		}
	}

	cmd := exec.CommandContext(context.Background(), s.BinaryPath, args...)
	return cmd.CombinedOutput()
}

// RunArtDuplWithStdin executes art-dupl with stdin input.
func (s *BDDTestSetup) RunArtDuplWithStdin(stdin string, flags map[string]string) ([]byte, error) {
	if s.T != nil {
		s.T.Helper()
	}

	args := []string{"--files"}
	for flag, value := range flags {
		if value != "" {
			args = append(args, fmt.Sprintf("--%s", flag), value)
		} else {
			args = append(args, fmt.Sprintf("--%s", flag))
		}
	}

	cmd := exec.CommandContext(context.Background(), s.BinaryPath, args...)
	cmd.Stdin = strings.NewReader(stdin)
	return cmd.CombinedOutput()
}

// RunArtDuplAndVerifyOutput executes art-dupl and verifies it completes successfully.
func (s *BDDTestSetup) RunArtDuplAndVerifyOutput(args ...string) string {
	if s.T != nil {
		s.T.Helper()
	}

	output, err := s.RunArtDupl(args...)
	if err != nil {
		if s.T != nil {
			s.T.Fatalf("art-dupl command failed: %v\nOutput: %s", err, string(output))
		} else {
			panic(fmt.Sprintf("art-dupl command failed: %v\nOutput: %s", err, string(output)))
		}
	}

	return string(output)
}

// RunArtDuplWithFlagsAndVerify executes art-dupl with flags and verifies success.
func (s *BDDTestSetup) RunArtDuplWithFlagsAndVerify(flags map[string]string) string {
	if s.T != nil {
		s.T.Helper()
	}

	output, err := s.RunArtDuplWithFlags(flags)
	if err != nil {
		if s.T != nil {
			s.T.Fatalf("art-dupl command failed: %v\nOutput: %s", err, string(output))
		} else {
			panic(fmt.Sprintf("art-dupl command failed: %v\nOutput: %s", err, string(output)))
		}
	}

	return string(output)
}
