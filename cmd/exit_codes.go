package cmd

import (
	"context"
	"errors"

	duplerrors "github.com/LarsArtmann/art-dupl/errors"
)

const (
	ExitSuccess       = 0
	ExitGeneralError  = 1
	ExitConfigError   = 2
	ExitInternalError = 3
	ExitInterrupted   = 130
)

// ExitCodeForError maps an error to the appropriate process exit code.
// nil error returns ExitSuccess. Context cancellation returns ExitInterrupted.
// Config/validation errors return ExitConfigError. Everything else returns
// ExitGeneralError.
func ExitCodeForError(err error) int {
	if err == nil {
		return ExitSuccess
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return ExitInterrupted
	}

	if duplerrors.Is(err, duplerrors.ValidationError) || duplerrors.Is(err, duplerrors.ConfigError) {
		return ExitConfigError
	}

	if duplerrors.Is(err, duplerrors.InternalError) {
		return ExitInternalError
	}

	return ExitGeneralError
}
