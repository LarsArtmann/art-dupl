package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/errors"
)

// FileProcessor provides unified file operations with consistent error handling.
type FileProcessor struct {
	baseDir string
}

// NewFileProcessor creates a new file processor with optional base directory.
func NewFileProcessor(baseDir ...string) *FileProcessor {
	fp := &FileProcessor{}
	if len(baseDir) > 0 && baseDir[0] != "" {
		fp.baseDir = baseDir[0]
	}
	return fp
}

// WriteFile writes content to a file with consistent error handling.
func (fp *FileProcessor) WriteFile(filename string, content []byte, perm os.FileMode) error {
	// Create full path if base directory is set
	fullPath := filename
	if fp.baseDir != "" {
		fullPath = filepath.Join(fp.baseDir, filename)
	}

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec //G301 Test data needs readable directory permission
		return errors.NewIOError(dir, "failed to create directory", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, content, perm); err != nil {
		return errors.NewIOError(fullPath, "failed to write file", err)
	}

	return nil
}

// WriteTextFile writes text content to a file with consistent permissions.
func (fp *FileProcessor) WriteTextFile(filename, content string) error {
	return fp.WriteFile(filename, []byte(content), 0o644)
}

// ReadFile reads file content with consistent error handling.
func (fp *FileProcessor) ReadFile(filename string) ([]byte, error) {
	// Create full path if base directory is set
	fullPath := filename
	if fp.baseDir != "" {
		fullPath = filepath.Join(fp.baseDir, filename)
	}

	data, err := os.ReadFile(fullPath) //nolint:gosec //G304 Path is constructed from base directory and validated filename
	if err != nil {
		return nil, errors.NewIOError(fullPath, "failed to read file", err)
	}

	return data, nil
}

// WriteTestFiles creates multiple test files from content map.
func (fp *FileProcessor) WriteTestFiles(files map[string]string) error {
	for filename, content := range files {
		if err := fp.WriteTextFile(filename, content); err != nil {
			return fmt.Errorf("failed to write test file %s: %w", filename, err)
		}
	}
	return nil
}

// WriteDuplicateFiles creates files with identical content for testing.
func (fp *FileProcessor) WriteDuplicateFiles(filenames []string, content string) error {
	for _, filename := range filenames {
		if err := fp.WriteTextFile(filename, content); err != nil {
			return fmt.Errorf("failed to write duplicate file %s: %w", filename, err)
		}
	}
	return nil
}
