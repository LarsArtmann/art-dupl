package testutil

import (
	"context"
	"fmt"
	"io"
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

// CreateSubdirectories creates multiple directories in test temporary directory.
// Each directory name is a relative path that will be created under the temp directory.
func (s *BDDTestSetup) CreateSubdirectories(paths ...string) error {
	if s.T != nil {
		s.T.Helper()
	}

	for _, path := range paths {
		fullPath := filepath.Join(s.TmpDir, path)
		if err := os.MkdirAll(fullPath, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

// CreateFileWithContent creates a file with specific content at a given subpath.
// The subpath is relative to the test temporary directory.
func (s *BDDTestSetup) CreateFileWithContent(subpath, content string) error {
	if s.T != nil {
		s.T.Helper()
	}

	fullPath := filepath.Join(s.TmpDir, subpath)
	dir := filepath.Dir(fullPath)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", subpath, err)
	}
	return nil
}

// RunArtDuplAndCapture executes art-dupl and captures stdout and stderr separately.
// Returns both outputs and any error that occurred.
func (s *BDDTestSetup) RunArtDuplAndCapture(args ...string) (stdout, stderr []byte, err error) {
	if s.T != nil {
		s.T.Helper()
	}

	cmd := exec.CommandContext(context.Background(), s.BinaryPath, args...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
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

	if err := cmd.Wait(); err != nil {
		return stdout, stderr, fmt.Errorf("command failed: %w", err)
	}

	return stdout, stderr, nil
}
