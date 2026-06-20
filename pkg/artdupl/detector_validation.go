package artdupl

import (
	"context"
	"fmt"
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
		return ctx.Err()
	default:
	}

	return nil
}

// validateFile checks if a file should be processed.
// Returns nil if the file passes all checks, or an error explaining why it was skipped.
// The caller logs the error and skips the file — it is not a pipeline failure.
func (d *detector) validateFile(filename string) error {
	// Check if file exists
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrFileNotFound, filename)
	}

	// Check file size
	if d.opts.MaxFileSize > 0 && info.Size() > d.opts.MaxFileSize {
		return fmt.Errorf(
			"%w: %s (size=%d, max=%d)",
			ErrFileTooLarge,
			filename,
			info.Size(),
			d.opts.MaxFileSize,
		)
	}

	// Check if file matches an ignore pattern
	for _, pattern := range d.opts.IgnoreFiles {
		if matched, _ := filepath.Match(pattern, filepath.Base(filename)); matched {
			return fmt.Errorf("%w: %s (matched pattern=%q)", ErrFileIgnored, filename, pattern)
		}
	}

	return nil
}
