package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/utils"
)

// CommandResult holds the captured stdout and stderr from command execution.
type CommandResult struct {
	Stdout []byte
	Stderr []byte
}

// Combined returns stdout + stderr concatenated.
func (r *CommandResult) Combined() []byte {
	out := make([]byte, 0, len(r.Stdout)+len(r.Stderr))
	out = append(out, r.Stdout...)
	out = append(out, r.Stderr...)

	return out
}

// BDDTestSetup provides complete BDD test infrastructure with temporary directory.
//
// Execution is done in-process via the Executor field (set by the bdd package)
// instead of building and running a separate binary. This is faster and avoids
// race conditions like "text file busy".
type BDDTestSetup struct {
	T             *testing.T
	TmpDir        string
	FileProcessor *utils.FileProcessor
	// Executor runs art-dupl in-process and returns combined stdout+stderr output.
	// Set by the bdd package.
	Executor func(args ...string) ([]byte, error)
	// ExecutorResult runs art-dupl in-process and returns separated stdout/stderr.
	// Set by the bdd package. Used by RunSubcommandOutput for JSON parsing.
	ExecutorResult func(args ...string) (*CommandResult, error)
}

// NewBDDTestSetup creates a new BDD test setup with temporary directory.
// The caller is responsible for cleaning up using Cleanup() or manually.
//
// Note: The Executor field must be set before calling Run* methods.
func NewBDDTestSetup(t *testing.T) *BDDTestSetup {
	t.Helper()

	tmpDir := t.TempDir()

	return &BDDTestSetup{ //nolint:exhaustruct // Executor/ExecutorResult are set later by test setup
		T:             t,
		TmpDir:        tmpDir,
		FileProcessor: utils.NewFileProcessor(tmpDir),
	}
}

// NewBDDTestSetupForGinkgo creates a new BDD test setup without requiring *testing.T.
// Designed for use with Ginkgo's BeforeEach/AfterEach pattern.
// The caller must call Cleanup() in an AfterEach block.
//
// Note: The Executor field must be set before calling Run* methods.
func NewBDDTestSetupForGinkgo() (*BDDTestSetup, error) {
	tmpDir, err := os.MkdirTemp("", "art-dupl-bdd-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	return &BDDTestSetup{ //nolint:exhaustruct // T/Executor/ExecutorResult are set later by Ginkgo test setup
		TmpDir:        tmpDir,
		FileProcessor: utils.NewFileProcessor(tmpDir),
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

// appendFlagsToArgs converts a flag map to CLI arguments.
func appendFlagsToArgs(args []string, flags map[string]string) []string {
	for flag, value := range flags {
		if value != "" {
			args = append(args, "--"+flag, value)
		} else {
			args = append(args, "--"+flag)
		}
	}

	return args
}

// BuildArgsFromFlags converts a map of flags to command line arguments.
func BuildArgsFromFlags(baseArgs []string, flags map[string]string) []string {
	return appendFlagsToArgs(baseArgs, flags)
}

// GetFilePath returns full path for a file in test directory.
func (s *BDDTestSetup) GetFilePath(filename string) string {
	return filepath.Join(s.TmpDir, filename)
}
