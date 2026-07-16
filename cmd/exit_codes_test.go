package cmd

import (
	"context"
	"errors"
	"fmt"
	"testing"

	duplerrors "github.com/LarsArtmann/art-dupl/errors"
)

func TestExitCodeForError(t *testing.T) {
	t.Parallel()

	wrappedCancel := fmt.Errorf("wrapped: %w", context.Canceled)

	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil returns success", err: nil, want: ExitSuccess},
		{name: "context canceled returns interrupted", err: context.Canceled, want: ExitInterrupted},
		{name: "deadline exceeded returns interrupted", err: context.DeadlineExceeded, want: ExitInterrupted},
		{name: "wrapped cancel returns interrupted", err: wrappedCancel, want: ExitInterrupted},
		{
			name: "validation error returns config exit",
			err:  duplerrors.NewValidationError("bad threshold", nil),
			want: ExitConfigError,
		},
		{
			name: "config error returns config exit",
			err:  duplerrors.NewConfigError("bad config", nil),
			want: ExitConfigError,
		},
		{
			name: "wrapped validation returns config exit",
			err:  duplerrors.WrapValidation(errors.New("orig"), "ctx"),
			want: ExitConfigError,
		},
		{
			name: "internal error returns internal exit",
			err:  duplerrors.NewInternalError("boom", nil),
			want: ExitInternalError,
		},
		{name: "generic error returns general", err: errors.New("oops"), want: ExitGeneralError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ExitCodeForError(tt.err)
			if got != tt.want {
				t.Errorf("ExitCodeForError() = %d, want %d", got, tt.want)
			}
		})
	}
}
