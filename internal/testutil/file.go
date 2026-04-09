package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/utils"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// TestFileSetup provides a complete test file setup with temporary directory.
type TestFileSetup struct {
	TmpDir        string
	FileProcessor *utils.FileProcessor
}

// NewTestFileSetup creates a new test file setup with a temporary directory.
func NewTestFileSetup(t *testing.T) *TestFileSetup {
	t.Helper()
	tmpDir := t.TempDir()

	return &TestFileSetup{
		TmpDir:        tmpDir,
		FileProcessor: utils.NewFileProcessor(tmpDir),
	}
}

// CreateTestFile creates a single test Go file with given content.
func (s *TestFileSetup) CreateTestFile(filename, content string) error {
	return s.FileProcessor.WriteTextFile(
		filename,
		content,
	)
}

// CreateTestFiles creates multiple test Go files from a map.
func (s *TestFileSetup) CreateTestFiles(files map[string]string) error {
	return s.FileProcessor.WriteTestFiles(
		files,
	)
}

// CreateDuplicateFiles creates files with identical content.
func (s *TestFileSetup) CreateDuplicateFiles(filenames []string, content string) error {
	return s.FileProcessor.WriteDuplicateFiles(
		filenames,
		content,
	)
}

// GetFilePath returns the full path for a file in the test directory.
func (s *TestFileSetup) GetFilePath(filename string) string {
	return filepath.Join(s.TmpDir, filename)
}

// ParseFile parses a Go file and returns the AST node.
func ParseFile(t *testing.T, filePath string) *syntax.Node {
	t.Helper()

	node, err := golang.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse %s: %v", filePath, err)
	}

	return node
}

// ParseFiles parses multiple Go files and returns AST nodes.
func ParseFiles(t *testing.T, filePaths []string) []*syntax.Node {
	t.Helper()

	nodes := make([]*syntax.Node, 0, len(filePaths))
	for _, file := range filePaths {
		node, err := golang.Parse(file)
		if err != nil {
			t.Fatalf("Failed to parse %s: %v", file, err)
		}

		nodes = append(nodes, node)
	}

	return nodes
}

// WriteTestFile writes content to a file and fails the test if it fails.
func WriteTestFile(t *testing.T, filename, content string) {
	t.Helper()

	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write test file %s: %v", filename, err)
	}
}

// WriteAndParseFile writes a Go file and parses it into an AST node.
// This is a convenience function that combines os.WriteFile and golang.Parse
// with proper error handling, commonly used in test files.
func WriteAndParseFile(t *testing.T, filename, content string) *syntax.Node {
	t.Helper()

	WriteTestFile(t, filename, content)

	node, err := golang.Parse(filename)
	if err != nil {
		t.Fatalf("Failed to parse %s: %v", filename, err)
	}

	return node
}
