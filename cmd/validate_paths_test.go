package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePaths(t *testing.T) {
	t.Parallel()

	existingFile := filepath.Join(t.TempDir(), "present.go")

	if err := os.WriteFile(existingFile, nil, 0o600); err != nil {
		t.Fatalf("create fixture: %v", err)
	}

	tests := []struct {
		name           string
		paths          []string
		filesFromStdin bool
		wantErr        bool
	}{
		{
			name:    "no paths and no stdin flag is valid",
			paths:   nil,
			wantErr: false,
		},
		{
			name:    "existing path passes",
			paths:   []string{existingFile},
			wantErr: false,
		},
		{
			name:           "stdin flag with no positional paths is valid",
			paths:          nil,
			filesFromStdin: true,
			wantErr:        false,
		},
		{
			name: "stdin flag with positional paths is valid: paths are the gitignore discovery root",
			paths: []string{
				existingFile,
			},
			filesFromStdin: true,
			wantErr:        false,
		},
		{
			name:    "all nonexistent paths fail",
			paths:   []string{filepath.Join(t.TempDir(), "missing.go")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validatePaths(tt.paths, tt.filesFromStdin)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("validatePaths(%v, %v) = nil, want error", tt.paths, tt.filesFromStdin)
				}

				return
			}

			if err != nil {
				t.Fatalf("validatePaths(%v, %v) = %v, want nil", tt.paths, tt.filesFromStdin, err)
			}
		})
	}
}
