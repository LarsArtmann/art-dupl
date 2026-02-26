package testutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/utils"
)

// sharedBinary holds the path to a pre-built binary shared across all test suites.
// This avoids concurrent go build commands which can cause hangs.
//
//nolint:gochecknoglobals // Shared binary path for test efficiency
var (
	sharedBinary     string
	sharedBinaryOnce sync.Once
	sharedBinaryErr  error
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
	cmd := exec.CommandContext(
		context.Background(),
		"go",
		"build",
		"-o",
		binaryPath,
		"../cmd/art-dupl/main.go",
	) // #nosec G204 -- Test helper building project binary

	output, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(tmpDir) // cleanup on error path

		t.Fatalf("Failed to build art-dupl binary: %v\nOutput: %s", err, string(output))
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir) // test cleanup
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
// Uses a shared binary to avoid concurrent build hangs.
func NewBDDTestSetupForGinkgo() (*BDDTestSetup, error) {
	tmpDir, err := os.MkdirTemp("", "art-dupl-bdd-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Build binary once using sync.Once to avoid concurrent builds
	sharedBinaryOnce.Do(func() {
		sharedBinary = filepath.Join(os.TempDir(), "art-dupl-bdd-shared")
		sharedBinaryErr = buildSharedBinary(sharedBinary)
	})

	if sharedBinaryErr != nil {
		_ = os.RemoveAll(tmpDir) // cleanup on error path

		return nil, sharedBinaryErr
	}

	// Check if binary still exists (may have been cleaned up by OS)
	if _, err := os.Stat(sharedBinary); err != nil {
		// Binary missing, rebuild it
		buildErr := buildSharedBinary(sharedBinary)
		if buildErr != nil {
			_ = os.RemoveAll(tmpDir)

			return nil, buildErr
		}
	}

	return &BDDTestSetup{
		TmpDir:        tmpDir,
		FileProcessor: utils.NewFileProcessor(tmpDir),
		BinaryPath:    sharedBinary,
	}, nil
}

// buildSharedBinary builds the art-dupl binary at the given path.
func buildSharedBinary(binaryPath string) error {
	cmd := exec.CommandContext(
		context.Background(),
		"go",
		"build",
		"-o",
		binaryPath,
		"../cmd/art-dupl/main.go",
	) // #nosec G204 -- Test helper building project binary

	output, buildErr := cmd.CombinedOutput()
	if buildErr != nil {
		return fmt.Errorf(
			"failed to build art-dupl binary: %w\nOutput: %s",
			buildErr,
			string(output),
		)
	}

	return nil
}

// Cleanup removes the temporary directory and all its contents.
func (s *BDDTestSetup) Cleanup() error {
	return os.RemoveAll(
		s.TmpDir,
	) //nolint:wrapcheck // Test cleanup - pass through os.RemoveAll error
}

// CreateDuplicateFiles creates multiple files with identical content.
func (s *BDDTestSetup) CreateDuplicateFiles(filenames []string, content string) error {
	return s.FileProcessor.WriteDuplicateFiles(
		filenames,
		content,
	) //nolint:wrapcheck // Test helper - pass through error
}

// CreateTestFile creates a single test file with given content.
func (s *BDDTestSetup) CreateTestFile(filename, content string) error {
	return s.FileProcessor.WriteTextFile(
		filename,
		content,
	) //nolint:wrapcheck // Test helper - pass through error
}

// CreateTestFiles creates multiple test files from a map.
func (s *BDDTestSetup) CreateTestFiles(files map[string]string) error {
	return s.FileProcessor.WriteTestFiles(
		files,
	) //nolint:wrapcheck // Test helper - pass through error
}

// GetFilePath returns full path for a file in test directory.
func (s *BDDTestSetup) GetFilePath(filename string) string {
	return filepath.Join(s.TmpDir, filename)
}
