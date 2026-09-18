package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/spf13/cobra"
)

func newHTMLTestCmd(htmlOut string) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("html-out", htmlOut, "write HTML report to a file instead of stdout (use with --html)")

	return cmd
}

func TestResolveHTMLOutputNonHTMLFormatAlwaysStdout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	explicit := filepath.Join(dir, "explicit.html")
	cmd := newHTMLTestCmd(explicit)

	var stderr bytes.Buffer
	w, cleanup, err := resolveHTMLOutput(cmd, config.OutputFormatText, "unused.html", func() bool { return true }, &stderr)
	if err != nil {
		t.Fatalf("resolveHTMLOutput() failed: %v", err)
	}

	if w != os.Stdout {
		t.Error("non-HTML format must always return stdout")
	}

	if cleanup != nil {
		t.Error("non-HTML format must not need cleanup")
	}

	if _, statErr := os.Stat(explicit); statErr == nil {
		t.Error("non-HTML format must not create a file even when --html-out is set")
	}

	if stderr.Len() > 0 {
		t.Errorf("non-HTML format must not print notices, got %q", stderr.String())
	}
}

func TestResolveHTMLOutputExplicitFlagWins(t *testing.T) {
	t.Parallel()

	explicit := filepath.Join(t.TempDir(), "explicit.html")
	cmd := newHTMLTestCmd(explicit)

	var stderr bytes.Buffer
	w, cleanup, err := resolveHTMLOutput(cmd, config.OutputFormatHTML, "auto.html", func() bool { return true }, &stderr)
	if err != nil {
		t.Fatalf("resolveHTMLOutput() failed: %v", err)
	}

	if w == os.Stdout {
		t.Error("explicit --html-out must produce a file writer, not stdout")
	}

	if cleanup == nil {
		t.Fatal("explicit --html-out must return a cleanup func")
	}

	if _, statErr := os.Stat(explicit); statErr != nil {
		t.Errorf("explicit --html-out file was not created: %v", statErr)
	}

	cleanup()

	if stderr.Len() > 0 {
		t.Errorf("explicit --html-out must not print an auto notice, got %q", stderr.String())
	}
}

func TestResolveHTMLOutputTTYAutoWritesDefaultFileWithNotice(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	auto := filepath.Join(dir, "auto.html")
	cmd := newHTMLTestCmd("")

	var stderr bytes.Buffer
	w, cleanup, err := resolveHTMLOutput(cmd, config.OutputFormatHTML, auto, func() bool { return true }, &stderr)
	if err != nil {
		t.Fatalf("resolveHTMLOutput() failed: %v", err)
	}

	if w == os.Stdout {
		t.Error("TTY + HTML must auto-write to the default file, not stdout")
	}

	if cleanup == nil {
		t.Fatal("TTY auto-write must return a cleanup func")
	}

	if _, statErr := os.Stat(auto); statErr != nil {
		t.Errorf("default HTML file was not created: %v", statErr)
	}

	cleanup()

	if !strings.Contains(stderr.String(), auto) {
		t.Errorf("notice should name the auto-written file, got %q", stderr.String())
	}
}

func TestResolveHTMLOutputPipedStdoutUnchanged(t *testing.T) {
	t.Parallel()

	cmd := newHTMLTestCmd("")

	var stderr bytes.Buffer
	w, cleanup, err := resolveHTMLOutput(cmd, config.OutputFormatHTML, "auto.html", func() bool { return false }, &stderr)
	if err != nil {
		t.Fatalf("resolveHTMLOutput() failed: %v", err)
	}

	if w != os.Stdout {
		t.Error("piped stdout + no --html-out must keep stdout output")
	}

	if cleanup != nil {
		t.Error("stdout output needs no cleanup")
	}

	if stderr.Len() > 0 {
		t.Errorf("piped stdout must not print notices, got %q", stderr.String())
	}
}

func TestResolveHTMLOutputUnwritablePathFails(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker.txt")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// A path beneath a regular file fails on MkdirAll regardless of user id
	// (ENOTDIR), unlike a merely missing directory which is now created.
	cmd := newHTMLTestCmd(filepath.Join(blocker, "sub", "out.html"))

	_, _, err := resolveHTMLOutput(cmd, config.OutputFormatHTML, "unused.html", func() bool { return true }, &bytes.Buffer{})
	if err == nil {
		t.Fatal("creating a file beneath a regular file must fail")
	}
}

func TestResolveHTMLOutputCreatesParentDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "reports", "nested", "report.html")
	cmd := newHTMLTestCmd(target)

	var stderr bytes.Buffer
	w, cleanup, err := resolveHTMLOutput(cmd, config.OutputFormatHTML, "auto.html", func() bool { return false }, &stderr)
	if err != nil {
		t.Fatalf("resolveHTMLOutput() failed: %v", err)
	}
	defer cleanup()

	if _, statErr := os.Stat(target); statErr != nil {
		t.Fatalf("expected HTML output file %q to be created (with parent dirs), stat error: %v", target, statErr)
	}

	if _, err := w.Write([]byte("<html></html>")); err != nil {
		t.Errorf("writing to HTML output failed: %v", err)
	}
}
