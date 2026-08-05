package cmd

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
)

func TestDumpTokensOutput(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "sample.go")
	writeTestFile(t, testFile, "package main\n\nfunc add(a, b int) int {\n\treturn a + b\n}\n", 0o600)

	cfg := config.DefaultConfig()
	cfg.Paths = []string{tmpDir}
	cfg.Threshold = 5

	buf := &bytes.Buffer{}

	err := dumpTokensOutput(t.Context(), cfg, buf, io.Discard)
	if err != nil {
		t.Fatalf("dumpTokensOutput() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("dumpTokensOutput() produced empty output")
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("dumpTokensOutput() produced no lines")
	}

	hasContent := false

	for _, line := range lines {
		if line == "---" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			t.Errorf("token line has fewer than 3 tab-separated fields: %q", line)

			continue
		}

		if !strings.Contains(fields[0], "sample.go") {
			t.Errorf("expected filename to contain sample.go, got %q", fields[0])
		}

		hasContent = true
	}

	if !hasContent {
		t.Error("expected at least one non-sentinel token line")
	}
}

func TestDumpTokensOutput_EmptyDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Paths = []string{tmpDir}

	buf := &bytes.Buffer{}

	err := dumpTokensOutput(t.Context(), cfg, buf, io.Discard)
	if err != nil {
		t.Fatalf("dumpTokensOutput() on empty dir error = %v", err)
	}

	if buf.Len() > 0 {
		t.Errorf("expected empty output for empty dir, got %q", buf.String())
	}
}
