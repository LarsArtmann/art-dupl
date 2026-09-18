package cmd

import (
	"bytes"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/syntax"
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

func TestSourceLineTableLineCol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		offset  int32
		line    int32
		col     int32
	}{
		{name: "first byte", content: "package main\n\nfunc add() {}\n", offset: 0, line: 1, col: 1},
		{name: "start of line 3", content: "package main\n\nfunc add() {}\n", offset: 14, line: 3, col: 1},
		{name: "start of line 4", content: "package main\n\nfunc add() {}\n", offset: 28, line: 4, col: 1},
		{name: "mid line", content: "package main\n\nfunc add() {}\n", offset: 17, line: 3, col: 4},
		{name: "multibyte counts bytes", content: "héllo\nx\n", offset: 3, line: 1, col: 4},
		{name: "multibyte line 2 start", content: "héllo\nx\n", offset: 7, line: 2, col: 1},
		{name: "eof with trailing newline", content: "a\n", offset: 2, line: 2, col: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "sample.go")
			writeTestFile(t, path, tt.content, 0o600)

			tbl := newSourceLineTable(path)
			gotLine, gotCol := tbl.lineCol(tt.offset)
			if gotLine != tt.line || gotCol != tt.col {
				t.Errorf("lineCol(%d) = %d:%d, want %d:%d", tt.offset, gotLine, gotCol, tt.line, tt.col)
			}
		})
	}
}

func TestSourceLineTableMissingFile(t *testing.T) {
	t.Parallel()

	tbl := newSourceLineTable(filepath.Join(t.TempDir(), "does-not-exist.go"))
	if tbl != nil {
		t.Fatal("expected nil table for unreadable file")
	}

	line, col := tbl.lineCol(0)
	if line != 0 || col != 0 {
		t.Errorf("nil table lineCol(0) = %d:%d, want 0:0", line, col)
	}
}

func TestFormatTokenPosition(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	writeTestFile(t, path, "package main\n\nfunc add(a, b int) int {\n\treturn a + b\n}\n", 0o600)

	tables := make(map[string]*sourceLineTable)

	tests := []struct {
		name string
		pos  int32
		end  int32
		want string
	}{
		{name: "single line token", pos: 16, end: 20, want: "3:3"},
		{name: "multi line span", pos: 14, end: 55, want: "3:1-6:1"},
		{name: "statement body span", pos: 39, end: 53, want: "4:1-5:1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Each parallel subtest needs its own line-table cache: the
			// shared map would race on the lazy per-file insert.
			tables := make(map[string]*sourceLineTable)

			node := &syntax.Node{Filename: path, Pos: tt.pos, End: tt.end}
			if got := formatTokenPosition(tables, node); got != tt.want {
				t.Errorf("formatTokenPosition() = %q, want %q", got, tt.want)
			}
		})
	}

	node := &syntax.Node{Filename: filepath.Join(dir, "missing.go"), Pos: 0, End: 1}
	if got := formatTokenPosition(tables, node); got != "?" {
		t.Errorf("formatTokenPosition() for unreadable file = %q, want %q", got, "?")
	}
}

func TestDumpTokensOutput_PositionColumn(t *testing.T) {
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

	positionRe := regexp.MustCompile(`^\d+:\d+(-\d+:\d+)?$`)
	hasSpan := false

	for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		if line == "---" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 4 {
			t.Fatalf("token line has fewer than 4 tab-separated fields: %q", line)
		}

		if !positionRe.MatchString(fields[2]) {
			t.Errorf("position field %q does not match line:col format: %q", fields[2], line)
		}

		if strings.Contains(fields[2], "-") {
			hasSpan = true
		}
	}

	if !hasSpan {
		t.Error("expected at least one multi-line token with an endline:endcol span")
	}
}
