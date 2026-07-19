package cmd

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
)

func TestExitCodes_ConfigValidation(t *testing.T) {
	t.Parallel()

	t.Run("bad threshold maps to config exit code", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Threshold:         -1,
			MaxChildrenSerial: 5000,
			OutputFormat:      config.OutputFormatText,
			SortBy:            config.SortBySize,
			DiffMode:          domain.DiffModeDisabled,
			DetectionMethods:  []domain.DetectionMethod{domain.MethodArtDupl},
		}

		err := config.ValidateConfig(cfg)
		if err == nil {
			t.Fatal("expected validation error for threshold=-1")
		}

		wrapped := duplerrors.WrapValidation(err, "config validation failed")

		code := ExitCodeForError(wrapped)
		if code != ExitConfigError {
			t.Errorf("threshold=-1: exit code = %d, want %d", code, ExitConfigError)
		}
	})

	t.Run("bad sort maps to config exit code", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Threshold:         5,
			MaxChildrenSerial: 5000,
			OutputFormat:      config.OutputFormatText,
			SortBy:            config.SortCriteria("invalid"),
			DiffMode:          domain.DiffModeDisabled,
			DetectionMethods:  []domain.DetectionMethod{domain.MethodArtDupl},
		}

		err := config.ValidateConfig(cfg)
		if err == nil {
			t.Fatal("expected validation error for invalid sort")
		}

		code := ExitCodeForError(err)
		if code != ExitConfigError {
			t.Errorf("invalid sort: exit code = %d, want %d", code, ExitConfigError)
		}
	})

	t.Run("negative workers maps to config exit code", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Threshold:         5,
			MaxChildrenSerial: 5000,
			OutputFormat:      config.OutputFormatText,
			SortBy:            config.SortBySize,
			DiffMode:          domain.DiffModeDisabled,
			DetectionMethods:  []domain.DetectionMethod{domain.MethodArtDupl},
			Workers:           -3,
		}

		err := config.ValidateConfig(cfg)
		if err == nil {
			t.Fatal("expected validation error for workers=-3")
		}

		code := ExitCodeForError(err)
		if code != ExitConfigError {
			t.Errorf("workers=-3: exit code = %d, want %d", code, ExitConfigError)
		}
	})

	t.Run("negative min-lines maps to config exit code", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			Threshold:         5,
			MaxChildrenSerial: 5000,
			OutputFormat:      config.OutputFormatText,
			SortBy:            config.SortBySize,
			DiffMode:          domain.DiffModeDisabled,
			DetectionMethods:  []domain.DetectionMethod{domain.MethodArtDupl},
			MinLines:          -10,
		}

		err := config.ValidateConfig(cfg)
		if err == nil {
			t.Fatal("expected validation error for min-lines=-10")
		}

		code := ExitCodeForError(err)
		if code != ExitConfigError {
			t.Errorf("min-lines=-10: exit code = %d, want %d", code, ExitConfigError)
		}
	})

	t.Run("nil error maps to success exit code", func(t *testing.T) {
		t.Parallel()

		code := ExitCodeForError(nil)
		if code != ExitSuccess {
			t.Errorf("nil error: exit code = %d, want %d", code, ExitSuccess)
		}
	})
}
