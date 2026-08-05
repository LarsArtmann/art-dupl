package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/spf13/cobra"
)

func TestPrintBuildingStatus_QuietSuppresses(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{Quiet: true}

	printBuildingStatus(io.Discard, cfg, config.OutputFormatText, "verbose msg", "text msg")
}

func TestPrintBuildingStatus_NonQuietVerbose(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{Quiet: false, Verbose: true}

	printBuildingStatus(io.Discard, cfg, config.OutputFormatText, "verbose msg", "text msg")
}

func TestPrintBuildingStatus_NonQuietNonVerboseText(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{Quiet: false, Verbose: false}

	printBuildingStatus(io.Discard, cfg, config.OutputFormatText, "verbose msg", "text msg")
}

func TestPrintBuildingStatus_NonQuietNonVerboseNonText(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{Quiet: false, Verbose: false}

	printBuildingStatus(io.Discard, cfg, config.OutputFormatJSON, "verbose msg", "text msg")
}

func newVersionCmdWithFlags(t *testing.T, flags map[string]bool) *cobra.Command {
	t.Helper()

	cmd := NewVersionCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(&bytes.Buffer{})

	for name, val := range flags {
		if err := cmd.Flags().Set(name, strconv.FormatBool(val)); err != nil {
			t.Fatalf("failed to set flag %s: %v", name, err)
		}
	}

	return cmd
}

func TestVersionCommand_Text(t *testing.T) {
	t.Parallel()

	cmd := newVersionCmdWithFlags(t, nil)
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("version command returned error: %v", err)
	}
}

func TestVersionCommand_JSON(t *testing.T) {
	t.Parallel()

	cmd := newVersionCmdWithFlags(t, map[string]bool{"json": true})
	buf := cmd.OutOrStdout().(*bytes.Buffer)

	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("version --json returned error: %v", err)
	}

	for _, field := range []string{"version", "commit", "date", "goVersion", "compiler", "platform", "arch"} {
		if !bytes.Contains(buf.Bytes(), []byte(field)) {
			t.Errorf("version --json output missing field %q in output: %s", field, buf.String())
		}
	}
}

func TestVersionCommand_Short(t *testing.T) {
	t.Parallel()

	cmd := newVersionCmdWithFlags(t, map[string]bool{"short": true})
	buf := cmd.OutOrStdout().(*bytes.Buffer)

	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("version --short returned error: %v", err)
	}

	want := Version + "\n"
	if buf.String() != want {
		t.Errorf("version --short = %q, want %q", buf.String(), want)
	}
}

func newCmdWithOutputFlags(flags map[string]bool) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("html", false, "")
	cmd.Flags().Bool("plumbing", false, "")
	cmd.Flags().Bool("sarif", false, "")
	cmd.Flags().Bool("simple-json", false, "")
	cmd.Flags().Bool("json", false, "")

	for name, val := range flags {
		_ = cmd.Flags().Set(name, strconv.FormatBool(val))
	}

	return cmd
}

func TestParseOutputFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		flags map[string]bool
		want  config.OutputFormat
	}{
		{name: "no flags returns text", flags: nil, want: config.OutputFormatText},
		{name: "html flag returns html", flags: map[string]bool{"html": true}, want: config.OutputFormatHTML},
		{name: "json flag returns json", flags: map[string]bool{"json": true}, want: config.OutputFormatJSON},
		{
			name:  "plumbing returns plumbing",
			flags: map[string]bool{"plumbing": true},
			want:  config.OutputFormatPlumbing,
		},
		{name: "sarif returns sarif", flags: map[string]bool{"sarif": true}, want: config.OutputFormatSARIF},
		{
			name:  "simple-json returns simple-json",
			flags: map[string]bool{"simple-json": true},
			want:  config.OutputFormatSimpleJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := newCmdWithOutputFlags(tt.flags)

			got := parseOutputFormat(cmd)
			if got != tt.want {
				t.Errorf("parseOutputFormat() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestExitCodeForError_WrappedInternal(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("outer: %w", duplerrors.NewInternalError("inner boom", nil))

	got := ExitCodeForError(wrapped)
	if got != ExitInternalError {
		t.Errorf("wrapped internal: got %d, want %d", got, ExitInternalError)
	}
}

func TestExitCodeForError_WrappedValidation(t *testing.T) {
	t.Parallel()

	orig := duplerrors.NewValidationError("bad threshold", nil)
	wrapped := errors.Join(orig, errors.New("extra context"))

	got := ExitCodeForError(wrapped)
	if got != ExitConfigError {
		t.Errorf("errors.Join wrapped validation: got %d, want %d", got, ExitConfigError)
	}
}

func TestExitCodeForError_NestedWrappedCancel(t *testing.T) {
	t.Parallel()

	inner := fmt.Errorf("level2: %w", context.Canceled)
	outer := fmt.Errorf("level1: %w", inner)

	got := ExitCodeForError(outer)
	if got != ExitInterrupted {
		t.Errorf("nested wrapped cancel: got %d, want %d", got, ExitInterrupted)
	}
}
