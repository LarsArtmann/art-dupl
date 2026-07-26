package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitignoreBasicExclusion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	gitignoreContent := "*.txt\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0o644); err != nil {
		t.Fatal(err)
	}

	trackedFile := filepath.Join(tmpDir, "main.go")
	ignoredFile := filepath.Join(tmpDir, "notes.txt")

	if err := os.WriteFile(trackedFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(ignoredFile, []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	matcher := LoadGitignore([]string{tmpDir})
	if matcher == nil {
		t.Fatal("LoadGitignore returned nil despite .gitignore existing")
	}

	if !matcher.IsIgnored(ignoredFile) {
		t.Error("expected notes.txt to be ignored")
	}

	if matcher.IsIgnored(trackedFile) {
		t.Error("expected main.go to NOT be ignored")
	}
}

func TestGitignoreDirectoryExclusion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	gitignoreContent := "vendor/\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0o644); err != nil {
		t.Fatal(err)
	}

	vendorFile := filepath.Join(tmpDir, "vendor", "dep.go")
	if err := os.MkdirAll(filepath.Dir(vendorFile), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(vendorFile, []byte("package vendor\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mainFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	matcher := LoadGitignore([]string{tmpDir})
	if matcher == nil {
		t.Fatal("LoadGitignore returned nil")
	}

	if !matcher.IsIgnored(vendorFile) {
		t.Error("expected vendor/dep.go to be ignored")
	}

	if matcher.IsIgnored(mainFile) {
		t.Error("expected main.go to NOT be ignored")
	}
}

func TestGitignoreNegation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	gitignoreContent := "*.go\n!important.go\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0o644); err != nil {
		t.Fatal(err)
	}

	regularGo := filepath.Join(tmpDir, "regular.go")
	importantGo := filepath.Join(tmpDir, "important.go")

	if err := os.WriteFile(regularGo, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(importantGo, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	matcher := LoadGitignore([]string{tmpDir})
	if matcher == nil {
		t.Fatal("LoadGitignore returned nil")
	}

	if !matcher.IsIgnored(regularGo) {
		t.Error("expected regular.go to be ignored")
	}

	if matcher.IsIgnored(importantGo) {
		t.Error("expected important.go to NOT be ignored (negated)")
	}
}

func TestGitignoreNilMatcher(t *testing.T) {
	var matcher *GitignoreMatcher

	if matcher.IsIgnored("anything.go") {
		t.Error("nil matcher should not ignore anything")
	}
}

func TestGitignoreNoGitignoreFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	matcher := LoadGitignore([]string{tmpDir})
	if matcher != nil {
		t.Error("expected nil matcher when no .gitignore exists")
	}
}

func TestGitignoreExcludesTemplGenerated(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	gitignoreContent := "*_templ.go\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0o644); err != nil {
		t.Fatal(err)
	}

	templFile := filepath.Join(tmpDir, "view_templ.go")
	mainFile := filepath.Join(tmpDir, "main.go")

	if err := os.WriteFile(templFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(mainFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	matcher := LoadGitignore([]string{tmpDir})
	if matcher == nil {
		t.Fatal("LoadGitignore returned nil")
	}

	if !matcher.IsIgnored(templFile) {
		t.Error("expected view_templ.go to be ignored by gitignore")
	}

	if matcher.IsIgnored(mainFile) {
		t.Error("expected main.go to NOT be ignored")
	}
}

func TestParseGitignoreFileScannerError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// A valid pattern, then a line exceeding the 64 KiB scanner buffer.
	validPattern := "*.tmp\n"
	longLine := strings.Repeat("a", maxScannerBufferSize+1)

	content := validPattern + longLine + "\n"

	path := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	pats, err := parseGitignoreFile(path)
	if err == nil {
		t.Fatal("expected error for oversized .gitignore line, got nil")
	}

	// Patterns parsed before the error should still be returned.
	if len(pats) != 1 || pats[0].raw != "*.tmp" {
		t.Fatalf("expected 1 valid pattern before error, got %d: %v", len(pats), pats)
	}
}

func TestParseGitignoreFileOpenError(t *testing.T) {
	t.Parallel()

	// Non-existent path should return an error from os.Open.
	pats, err := parseGitignoreFile(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}

	if pats != nil {
		t.Fatalf("expected nil patterns on error, got %v", pats)
	}
}
