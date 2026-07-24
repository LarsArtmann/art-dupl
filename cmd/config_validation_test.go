package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newValidationCmd(t *testing.T) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "test"}
	addSharedFlags(cmd)
	// incremental is a root-only flag; add it manually for tests
	cmd.Flags().Bool("incremental", false, "enable incremental analysis with AST caching")

	return cmd
}

func parseTestFlags(t *testing.T, cmd *cobra.Command, args ...string) {
	t.Helper()

	cmd.SetArgs(args)

	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags(%v) error: %v", args, err)
	}
}

func TestValidateMutualExclusionTypeAware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "type-aware plus structural errors",
			args:    []string{"--type-aware", "--structural"},
			wantErr: true,
			errMsg:  "type-aware",
		},
		{
			name:    "type-aware plus exact errors",
			args:    []string{"--type-aware", "--exact"},
			wantErr: true,
			errMsg:  "type-aware",
		},
		{
			name:    "type-aware alone is fine",
			args:    []string{"--type-aware"},
			wantErr: false,
		},
		{
			name:    "type-aware plus semantic is fine",
			args:    []string{"--type-aware", "--semantic"},
			wantErr: false,
		},
		{
			name:    "structural alone is fine",
			args:    []string{"--structural"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := newValidationCmd(t)
			parseTestFlags(t, cmd, tt.args...)

			err := validateMutualExclusion(cmd)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("validateMutualExclusion(%v) expected error, got nil", tt.args)
				}

				if !strings.Contains(strings.ToLower(err.Error()), tt.errMsg) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Fatalf("validateMutualExclusion(%v) unexpected error: %v", tt.args, err)
			}
		})
	}
}

func TestWarnTypeAwareIncremental(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantWarning bool
	}{
		{
			name:        "type-aware plus incremental warns",
			args:        []string{"--type-aware", "--incremental"},
			wantWarning: true,
		},
		{
			name:        "type-aware alone does not warn",
			args:        []string{"--type-aware"},
			wantWarning: false,
		},
		{
			name:        "incremental alone does not warn",
			args:        []string{"--incremental"},
			wantWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newValidationCmd(t)
			parseTestFlags(t, cmd, tt.args...)

			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			warnTypeAwareIncremental(cmd)

			_ = w.Close()

			os.Stderr = oldStderr

			var buf bytes.Buffer

			_, _ = io.Copy(&buf, r)
			output := buf.String()

			if tt.wantWarning {
				if !strings.Contains(strings.ToLower(output), "type-aware") {
					t.Errorf("expected warning containing 'type-aware', got %q", output)
				}
			} else if output != "" {
				t.Errorf("expected no warning, got %q", output)
			}
		})
	}
}
