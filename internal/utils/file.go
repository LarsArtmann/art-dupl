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
	fp := &FileProcessor{} //nolint:exhaustruct // baseDir set below
	if len(baseDir) > 0 && baseDir[0] != "" {
		fp.baseDir = baseDir[0]
	}

	return fp
}

// resolvePath creates full path by joining baseDir with filename if baseDir is set.
func (fp *FileProcessor) resolvePath(filename string) string {
	if fp.baseDir != "" {
		return filepath.Join(fp.baseDir, filename)
	}

	return filename
}

// WriteFile writes content to a file with consistent error handling.
func (fp *FileProcessor) WriteFile(filename string, content []byte, perm os.FileMode) error {
	fullPath := fp.resolvePath(filename)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)

	err := os.MkdirAll(dir, 0o750) //nolint:mnd // Standard directory permission
	if err != nil {
		return errors.NewIOError(dir, "failed to create directory for "+filename, err)
	}

	// Write file
	err = os.WriteFile(fullPath, content, perm)
	if err != nil {
		return errors.NewIOError(fullPath, "failed to write file", err)
	}

	return nil
}

// WriteTextFile writes text content to a file with consistent permissions.
func (fp *FileProcessor) WriteTextFile(filename, content string) error {
	return fp.WriteFile(filename, []byte(content), 0o644) //nolint:mnd // Standard file permission
}

// ReadFile reads file content with consistent error handling.
func (fp *FileProcessor) ReadFile(filename string) ([]byte, error) {
	fullPath := fp.resolvePath(filename)

	data, err := os.ReadFile(
		fullPath,
	) // #nosec G304 -- Path is constructed from base directory and validated filename
	if err != nil {
		return nil, errors.NewIOError(fullPath, "failed to read file "+filename, err)
	}

	return data, nil
}

// WriteTestFiles creates multiple test files from content map.
func (fp *FileProcessor) WriteTestFiles(files map[string]string) error {
	return fp.writeFiles("test", files, func(filename, content string) error {
		return fp.WriteTextFile(filename, content)
	})
}

// WriteDuplicateFiles creates files with identical content for testing.
func (fp *FileProcessor) WriteDuplicateFiles(filenames []string, content string) error {
	files := make(map[string]string, len(filenames))
	for _, filename := range filenames {
		files[filename] = content
	}

	return fp.writeFiles("duplicate", files, fp.WriteTextFile)
}

// writeFiles is a helper that writes multiple files with a custom write function.
func (fp *FileProcessor) writeFiles(
	fileType string,
	files map[string]string,
	writeFunc func(string, string) error,
) error {
	for filename, content := range files {
		err := writeFunc(filename, content)
		if err != nil {
			return fmt.Errorf("failed to write %s file %s: %w", fileType, filename, err)
		}
	}

	return nil
}
