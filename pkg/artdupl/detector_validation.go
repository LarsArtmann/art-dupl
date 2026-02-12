package artdupl

import (
	"context"
	"os"
	"path/filepath"
)

// validateInputs checks that inputs are valid for analysis.
func (d *detector) validateInputs(ctx context.Context, files []string) error {
	if len(files) == 0 {
		return ErrNoFilesProvided
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err() //nolint:wrapcheck // Context cancellation errors are already clear
	default:
	}

	return nil
}

// validateFile checks if a file should be processed.
func (d *detector) validateFile(filename string) error {
	// Check if file exists
	info, err := os.Stat(filename)
	if err != nil {
		return ErrFileNotFound
	}

	// Check file size
	if d.opts.MaxFileSize > 0 && info.Size() > d.opts.MaxFileSize {
		return ErrFileTooLarge
	}

	// Check if file should be ignored
	for _, pattern := range d.opts.IgnoreFiles {
		if matched, _ := filepath.Match(pattern, filepath.Base(filename)); matched {
			return ErrParsingFailed // Use generic error for ignored files
		}
	}

	return nil
}
